package download

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// DownloadStrategy 下载策略类型
type DownloadStrategy int

const (
	StrategySingleThread DownloadStrategy = iota
	StrategyMultiConnection
	StrategyChunked
)

func (s DownloadStrategy) String() string {
	switch s {
	case StrategySingleThread:
		return "SingleThread"
	case StrategyMultiConnection:
		return "MultiConnection"
	case StrategyChunked:
		return "Chunked"
	default:
		return "Unknown"
	}
}

// FileInfo 文件信息
type FileInfo struct {
	LastModified   time.Time
	URL            string
	ContentType    string
	ETag           string
	Size           int64
	AcceptRanges   bool
	SupportsResume bool
}

// AdaptiveDownloader 自适应下载器
type AdaptiveDownloader struct {
	config           *Config
	client           *http.Client
	transport        *http.Transport
	bandwidthMonitor *BandwidthMonitor
	congestionCtrl   *CongestionController
	stats            *DownloadStats
}

// Config 下载器配置
type Config struct {
	// 分片大小 (默认 10MB)
	ChunkSize int64
	// 最大分片数 (默认 16)
	MaxChunks int
	// 最小分片大小 (默认 1MB)
	MinChunkSize int64
	// 单线程下载阈值 (默认 5MB)
	SingleThreadThreshold int64
	// 多连接下载阈值 (默认 50MB)
	MultiConnThreshold int64
	// 最大并发连接数 (默认 8)
	MaxConcurrentConns int
	// 连接超时
	ConnectTimeout time.Duration
	// 读取超时
	ReadTimeout time.Duration
	// 是否启用断点续传
	EnableResume bool
	// 临时文件后缀
	TempFileSuffix string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ChunkSize:             10 * 1024 * 1024, // 10MB
		MaxChunks:             16,
		MinChunkSize:          1 * 1024 * 1024,  // 1MB
		SingleThreadThreshold: 5 * 1024 * 1024,  // 5MB
		MultiConnThreshold:    50 * 1024 * 1024, // 50MB
		MaxConcurrentConns:    8,
		ConnectTimeout:        30 * time.Second,
		ReadTimeout:           5 * time.Minute,
		EnableResume:          true,
		TempFileSuffix:        ".tmp",
	}
}

// DownloadStats 下载统计
type DownloadStats struct {
	TotalBytes      int64
	DownloadedBytes int64
	StartTime       time.Time
	CurrentSpeed    int64 // bytes per second
	AvgSpeed        int64 // bytes per second
	ActiveChunks    int32
	CompletedChunks int32
	FailedChunks    int32
	mu              sync.RWMutex
}

// NewAdaptiveDownloader 创建自适应下载器
func NewAdaptiveDownloader(config *Config) *AdaptiveDownloader {
	if config == nil {
		config = DefaultConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:        config.MaxConcurrentConns * 2,
		MaxIdleConnsPerHost: config.MaxConcurrentConns,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ReadTimeout,
	}

	return &AdaptiveDownloader{
		config:           config,
		client:           client,
		transport:        transport,
		bandwidthMonitor: NewBandwidthMonitor(),
		congestionCtrl:   NewCongestionController(),
		stats:            &DownloadStats{},
	}
}

// probeFile 探测文件信息
func (d *AdaptiveDownloader) probeFile(ctx context.Context, url string) (*FileInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 HEAD 请求失败: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行 HEAD 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	info := &FileInfo{
		URL:          url,
		Size:         resp.ContentLength,
		ContentType:  resp.Header.Get("Content-Type"),
		AcceptRanges: resp.Header.Get("Accept-Ranges") == "bytes",
		ETag:         resp.Header.Get("ETag"),
	}

	// 解析 Last-Modified
	if lastMod := resp.Header.Get("Last-Modified"); lastMod != "" {
		info.LastModified, _ = http.ParseTime(lastMod)
	}

	// 判断是否支持断点续传
	info.SupportsResume = info.AcceptRanges && info.Size > 0

	return info, nil
}

// selectStrategy 选择下载策略
func (d *AdaptiveDownloader) selectStrategy(info *FileInfo) DownloadStrategy {
	if info.Size <= 0 {
		// 未知大小，使用单线程
		return StrategySingleThread
	}

	if info.Size <= d.config.SingleThreadThreshold {
		// 小文件使用单线程
		return StrategySingleThread
	}

	if !info.AcceptRanges {
		// 不支持 Range 请求，使用多连接
		return StrategyMultiConnection
	}

	if info.Size <= d.config.MultiConnThreshold {
		// 中等文件使用多连接
		return StrategyMultiConnection
	}

	// 大文件使用分片下载
	return StrategyChunked
}

// calculateOptimalChunks 计算最优分片数
func (d *AdaptiveDownloader) calculateOptimalChunks(fileSize int64) (int, int64) {
	if fileSize <= d.config.MinChunkSize {
		return 1, fileSize
	}

	// 计算理论分片数
	numChunks := int(fileSize / d.config.ChunkSize)
	if numChunks < 1 {
		numChunks = 1
	}
	if numChunks > d.config.MaxChunks {
		numChunks = d.config.MaxChunks
	}

	// 根据网络状况调整
	currentSpeed := d.bandwidthMonitor.GetCurrentSpeed()
	if currentSpeed > 0 {
		// 如果网速很快，可以增加分片数
		targetChunks := int(currentSpeed / d.config.ChunkSize)
		if targetChunks > numChunks && targetChunks <= d.config.MaxChunks {
			numChunks = targetChunks
		}
	}

	// 计算实际分片大小
	chunkSize := fileSize / int64(numChunks)
	if chunkSize < d.config.MinChunkSize {
		chunkSize = d.config.MinChunkSize
		numChunks = int(fileSize / chunkSize)
		if numChunks < 1 {
			numChunks = 1
		}
	}

	return numChunks, chunkSize
}

// Download 下载文件
func (d *AdaptiveDownloader) Download(ctx context.Context, url, destPath string, progress ProgressCallback) error {
	// 探测文件信息
	info, err := d.probeFile(ctx, url)
	if err != nil {
		return fmt.Errorf("探测文件信息失败: %w", err)
	}

	// 选择下载策略
	strategy := d.selectStrategy(info)

	switch strategy {
	case StrategySingleThread:
		return d.singleThreadDownload(ctx, info, destPath, progress)
	case StrategyMultiConnection:
		return d.multiConnectionDownload(ctx, info, destPath, progress)
	case StrategyChunked:
		return d.chunkedParallelDownload(ctx, info, destPath, progress)
	default:
		return fmt.Errorf("未知的下载策略: %v", strategy)
	}
}

// ProgressCallback 进度回调函数
type ProgressCallback func(downloaded, total int64, speed float64)

// singleThreadDownload 单线程下载
func (d *AdaptiveDownloader) singleThreadDownload(ctx context.Context, info *FileInfo, destPath string, progress ProgressCallback) error {
	d.stats.StartTime = time.Now()
	d.stats.TotalBytes = info.Size

	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
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

	return d.writeToFile(resp, destPath, progress)
}

// multiConnectionDownload 多连接下载 (简单实现，实际可以打开多个连接)
func (d *AdaptiveDownloader) multiConnectionDownload(ctx context.Context, info *FileInfo, destPath string, progress ProgressCallback) error {
	// 对于不支持 Range 的文件，退化为单线程下载
	return d.singleThreadDownload(ctx, info, destPath, progress)
}

// chunkedParallelDownload 分片并行下载
func (d *AdaptiveDownloader) chunkedParallelDownload(ctx context.Context, info *FileInfo, destPath string, progress ProgressCallback) error {
	numChunks, chunkSize := d.calculateOptimalChunks(info.Size)

	d.stats.StartTime = time.Now()
	d.stats.TotalBytes = info.Size

	// 创建分片下载任务
	chunks := make([]*Chunk, numChunks)
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = info.Size - 1
		}

		chunks[i] = &Chunk{
			Index:      i,
			Start:      start,
			End:        end,
			URL:        info.URL,
			TempPath:   fmt.Sprintf("%s.%d%s", destPath, i, d.config.TempFileSuffix),
			Downloader: d,
		}
	}

	// 使用信号量控制并发
	sem := make(chan struct{}, d.config.MaxConcurrentConns)
	var wg sync.WaitGroup
	errChan := make(chan error, numChunks)

	// 启动进度报告协程
	stopProgress := make(chan struct{})
	go d.reportProgress(stopProgress, progress)

	for _, chunk := range chunks {
		wg.Add(1)
		go func(c *Chunk) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if err := c.Download(ctx); err != nil {
				errChan <- fmt.Errorf("分片 %d 下载失败: %w", c.Index, err)
			}
		}(chunk)
	}

	wg.Wait()
	close(stopProgress)
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// 合并分片
	return d.mergeChunks(chunks, destPath)
}

// reportProgress 报告下载进度
func (d *AdaptiveDownloader) reportProgress(stop chan struct{}, callback ProgressCallback) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if callback != nil {
				downloaded := atomic.LoadInt64(&d.stats.DownloadedBytes)
				total := d.stats.TotalBytes
				speed := d.bandwidthMonitor.GetCurrentSpeed()
				callback(downloaded, total, float64(speed))
			}
		}
	}
}

// GetStats 获取下载统计
func (d *AdaptiveDownloader) GetStats() DownloadStats {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()
	return DownloadStats{
		TotalBytes:      d.stats.TotalBytes,
		DownloadedBytes: d.stats.DownloadedBytes,
		StartTime:       d.stats.StartTime,
		CurrentSpeed:    d.stats.CurrentSpeed,
		AvgSpeed:        d.stats.AvgSpeed,
		ActiveChunks:    d.stats.ActiveChunks,
		CompletedChunks: d.stats.CompletedChunks,
		FailedChunks:    d.stats.FailedChunks,
	}
}
