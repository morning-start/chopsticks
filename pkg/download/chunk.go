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

// Chunk 表示一个下载分片 - 按字段大小从大到小排列优化内存布局
// string: 16字节, int64: 8字节, int: 8字节 (64位系统), bool: 1字节
type Chunk struct {
	// 16字节字段
	URL      string
	TempPath string
	// 8字节字段 (指针)
	Downloader *AdaptiveDownloader
	// 8字节字段 (int64)
	Start      int64
	End        int64
	// 8字节字段 (int, 64位系统)
	Index int
	// 8字节字段 (需要8字节对齐)
	downloaded int64
	// 互斥锁
	mu sync.RWMutex
	// 1字节字段
	completed bool
}

// Download 下载分片
func (c *Chunk) Download(ctx context.Context) error {
	// 检查是否已存在临时文件（断点续传）
	if c.Downloader.config.EnableResume {
		_ = c.loadProgress()
	}

	req, err := http.NewRequestWithContext(ctx, MethodGet, c.URL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreateRequest, err)
	}

	// 设置 Range 头
	start := c.Start + c.downloaded
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, c.End))

	resp, err := c.Downloader.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDoRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != StatusPartialContent && resp.StatusCode != StatusOK {
		return fmt.Errorf("%w: %d", ErrInvalidStatusCode, resp.StatusCode)
	}

	// 打开临时文件
	flags := os.O_CREATE | os.O_WRONLY
	if c.downloaded > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(c.TempPath, flags, DefaultFilePerm)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenFile, err)
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
			_ = c.saveProgress()
		}
		return fmt.Errorf("%w: %v", ErrWriteFile, err)
	}

	c.completed = true
	atomic.AddInt32(&c.Downloader.stats.CompletedChunks, 1)

	// 清理进度文件
	if c.Downloader.config.EnableResume {
		_ = c.removeProgress()
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
		return ErrProgressMismatch
	}

	c.downloaded = progress
	return nil
}

// saveProgress 保存下载进度
func (c *Chunk) saveProgress() error {
	progressPath := c.TempPath + ".progress"
	data := fmt.Sprintf("%d", c.downloaded)
	return os.WriteFile(progressPath, []byte(data), DefaultFilePerm)
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
	if n <= 0 {
		return n, err
	}

	r.bytesRead += int64(n)
	r.chunk.mu.Lock()
	r.chunk.downloaded += int64(n)
	r.chunk.mu.Unlock()

	// 更新全局统计
	atomic.AddInt64(&r.downloader.stats.DownloadedBytes, int64(n))

	// 更新带宽监控
	r.downloader.bandwidthMonitor.RecordBytes(int64(n))

	return n, err
}
