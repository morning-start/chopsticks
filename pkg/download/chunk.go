package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Chunk 表示一个下载分片
type Chunk struct {
	Index      int
	Start      int64
	End        int64
	URL        string
	TempPath   string
	Downloader *AdaptiveDownloader

	// 下载状态
	downloaded int64
	completed  bool
	mu         sync.RWMutex
}

// Download 下载分片
func (c *Chunk) Download(ctx context.Context) error {
	// 检查是否已存在临时文件（断点续传）
	if c.Downloader.config.EnableResume {
		if err := c.loadProgress(); err == nil {
			// 已有进度，从上次位置继续
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Range 头
	start := c.Start + c.downloaded
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, c.End))

	resp, err := c.Downloader.client.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 打开临时文件
	flags := os.O_CREATE | os.O_WRONLY
	if c.downloaded > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(c.TempPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("打开临时文件失败: %w", err)
	}
	defer file.Close()

	// 创建带统计的 Reader
	statsReader := &statsReader{
		Reader:     resp.Body,
		chunk:      c,
		downloader: c.Downloader,
		startTime:  time.Now(),
	}

	// 复制数据
	_, err = io.Copy(file, statsReader)
	if err != nil {
		// 保存进度以便断点续传
		if c.Downloader.config.EnableResume {
			c.saveProgress()
		}
		return fmt.Errorf("写入数据失败: %w", err)
	}

	c.completed = true
	atomic.AddInt32(&c.Downloader.stats.CompletedChunks, 1)

	// 清理进度文件
	if c.Downloader.config.EnableResume {
		c.removeProgress()
	}

	return nil
}

// loadProgress 加载下载进度
func (c *Chunk) loadProgress() error {
	progressPath := c.TempPath + ".progress"
	data, err := os.ReadFile(progressPath)
	if err != nil {
		return err
	}

	var progress int64
	_, err = fmt.Sscanf(string(data), "%d", &progress)
	if err != nil {
		return err
	}

	// 验证文件大小
	info, err := os.Stat(c.TempPath)
	if err != nil {
		return err
	}

	if info.Size() != progress {
		// 进度不匹配，重新开始
		c.downloaded = 0
		return fmt.Errorf("进度不匹配")
	}

	c.downloaded = progress
	return nil
}

// saveProgress 保存下载进度
func (c *Chunk) saveProgress() error {
	progressPath := c.TempPath + ".progress"
	data := fmt.Sprintf("%d", c.downloaded)
	return os.WriteFile(progressPath, []byte(data), 0644)
}

// removeProgress 删除进度文件
func (c *Chunk) removeProgress() error {
	progressPath := c.TempPath + ".progress"
	return os.Remove(progressPath)
}

// GetDownloaded 获取已下载字节数
func (c *Chunk) GetDownloaded() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.downloaded
}

// IsCompleted 检查是否完成
func (c *Chunk) IsCompleted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.completed
}

// statsReader 带统计功能的 Reader
type statsReader struct {
	Reader     io.Reader
	chunk      *Chunk
	downloader *AdaptiveDownloader
	startTime  time.Time
	bytesRead  int64
}

func (r *statsReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 {
		r.bytesRead += int64(n)
		r.chunk.mu.Lock()
		r.chunk.downloaded += int64(n)
		r.chunk.mu.Unlock()

		// 更新全局统计
		atomic.AddInt64(&r.downloader.stats.DownloadedBytes, int64(n))

		// 更新带宽监控
		r.downloader.bandwidthMonitor.RecordBytes(int64(n))
	}
	return n, err
}
