package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Downloader 下载器接口
type Downloader interface {
	Download(ctx context.Context, url, destPath string, progress ProgressCallback) error
	GetStats() DownloadStats
}

// SimpleDownloader 简单下载器（单线程）
type SimpleDownloader struct {
	client *http.Client
	stats  DownloadStats
}

// NewSimpleDownloader 创建简单下载器
func NewSimpleDownloader() *SimpleDownloader {
	return &SimpleDownloader{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		stats: DownloadStats{},
	}
}

// Download 单线程下载
func (d *SimpleDownloader) Download(ctx context.Context, url, destPath string, progress ProgressCallback) error {
	d.stats.StartTime = time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	d.stats.TotalBytes = resp.ContentLength

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件
	tempPath := destPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 带进度监控的复制
	reader := &progressReader{
		Reader:     resp.Body,
		total:      resp.ContentLength,
		callback:   progress,
		startTime:  time.Now(),
		onProgress: d.updateProgress,
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 关闭文件
	file.Close()

	// 重命名为目标文件
	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	d.stats.DownloadedBytes = d.stats.TotalBytes
	return nil
}

// updateProgress 更新进度
func (d *SimpleDownloader) updateProgress(downloaded int64) {
	atomic.StoreInt64(&d.stats.DownloadedBytes, downloaded)
}

// GetStats 获取统计信息
func (d *SimpleDownloader) GetStats() DownloadStats {
	return d.stats
}

// progressReader 带进度报告的 Reader
type progressReader struct {
	Reader     io.Reader
	total      int64
	callback   ProgressCallback
	startTime  time.Time
	onProgress func(int64)

	downloaded int64
	lastReport time.Time
	mu         sync.Mutex
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		pr.onProgress(pr.downloaded)

		// 限制回调频率，每 100ms 报告一次
		pr.mu.Lock()
		if time.Since(pr.lastReport) > 100*time.Millisecond {
			if pr.callback != nil {
				elapsed := time.Since(pr.startTime).Seconds()
				var speed float64
				if elapsed > 0 {
					speed = float64(pr.downloaded) / elapsed
				}
				pr.callback(pr.downloaded, pr.total, speed)
			}
			pr.lastReport = time.Now()
		}
		pr.mu.Unlock()
	}
	return n, err
}

// mergeChunks 合并分片文件
func (d *AdaptiveDownloader) mergeChunks(chunks []*Chunk, destPath string) error {
	// 按索引排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	// 创建目标文件
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 依次合并分片
	for _, chunk := range chunks {
		if !chunk.IsCompleted() {
			return fmt.Errorf("分片 %d 未完成下载", chunk.Index)
		}

		chunkFile, err := os.Open(chunk.TempPath)
		if err != nil {
			return fmt.Errorf("打开分片文件失败: %w", err)
		}

		_, err = io.Copy(destFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return fmt.Errorf("合并分片 %d 失败: %w", chunk.Index, err)
		}

		// 删除临时文件
		os.Remove(chunk.TempPath)
	}

	return nil
}

// writeToFile 写入文件（带进度）
func (d *AdaptiveDownloader) writeToFile(resp *http.Response, destPath string, start, total int64, progress ProgressCallback) error {
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件
	tempPath := destPath + d.config.TempFileSuffix
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 带进度监控的复制
	reader := &bandwidthReader{
		Reader:     resp.Body,
		downloader: d,
	}

	stopProgress := make(chan struct{})
	if progress != nil {
		go d.reportProgress(stopProgress, progress)
	}

	_, err = io.Copy(file, reader)
	
	if progress != nil {
		close(stopProgress)
	}

	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	file.Close()

	// 重命名为目标文件
	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// bandwidthReader 带带宽监控的 Reader
type bandwidthReader struct {
	Reader     io.Reader
	downloader *AdaptiveDownloader
}

func (br *bandwidthReader) Read(p []byte) (n int, err error) {
	n, err = br.Reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&br.downloader.stats.DownloadedBytes, int64(n))
		br.downloader.bandwidthMonitor.RecordBytes(int64(n))
	}
	return n, err
}

// shouldResume 检查是否应该断点续传
func shouldResume(tempPath string, info *FileInfo) (int64, bool) {
	// 检查临时文件是否存在
	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		return 0, false
	}

	// 检查进度文件
	progressPath := tempPath + ".progress"
	progressData, err := os.ReadFile(progressPath)
	if err != nil {
		return 0, false
	}

	var progress int64
	_, err = fmt.Sscanf(string(progressData), "%d", &progress)
	if err != nil {
		return 0, false
	}

	// 验证进度与文件大小是否匹配
	if fileInfo.Size() != progress {
		return 0, false
	}

	// 检查文件是否已下载完成
	if info.Size > 0 && progress >= info.Size {
		return 0, false
	}

	return progress, true
}
