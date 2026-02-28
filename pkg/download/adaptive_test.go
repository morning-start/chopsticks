package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadStrategy_String(t *testing.T) {
	tests := []struct {
		strategy DownloadStrategy
		want     string
	}{
		{StrategySingleThread, "SingleThread"},
		{StrategyMultiConnection, "MultiConnection"},
		{StrategyChunked, "Chunked"},
		{DownloadStrategy(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ChunkSize != 10*1024*1024 {
		t.Errorf("ChunkSize = %d, want %d", config.ChunkSize, 10*1024*1024)
	}

	if config.MaxChunks != 16 {
		t.Errorf("MaxChunks = %d, want 16", config.MaxChunks)
	}

	if config.MinChunkSize != 1*1024*1024 {
		t.Errorf("MinChunkSize = %d, want %d", config.MinChunkSize, 1*1024*1024)
	}

	if config.SingleThreadThreshold != 5*1024*1024 {
		t.Errorf("SingleThreadThreshold = %d, want %d", config.SingleThreadThreshold, 5*1024*1024)
	}

	if config.MaxConcurrentConns != 8 {
		t.Errorf("MaxConcurrentConns = %d, want 8", config.MaxConcurrentConns)
	}

	if !config.EnableResume {
		t.Error("EnableResume should be true")
	}
}

func TestAdaptiveDownloader_probeFile(t *testing.T) {
	// 创建测试服务器
	content := []byte("Hello, World!")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("ETag", "\"abc123\"")
			w.Header().Set("Last-Modified", "Wed, 21 Oct 2025 07:28:00 GMT")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	downloader := NewAdaptiveDownloader(nil)
	ctx := context.Background()

	info, err := downloader.probeFile(ctx, server.URL)
	if err != nil {
		t.Fatalf("probeFile failed: %v", err)
	}

	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}

	if info.ContentType != "text/plain" {
		t.Errorf("ContentType = %s, want text/plain", info.ContentType)
	}

	if !info.AcceptRanges {
		t.Error("AcceptRanges should be true")
	}

	if info.ETag != `"abc123"` {
		t.Errorf("ETag = %s, want \"abc123\"", info.ETag)
	}

	if !info.SupportsResume {
		t.Error("SupportsResume should be true")
	}
}

func TestAdaptiveDownloader_selectStrategy(t *testing.T) {
	config := DefaultConfig()
	downloader := NewAdaptiveDownloader(config)

	tests := []struct {
		name     string
		info     *FileInfo
		expected DownloadStrategy
	}{
		{
			name:     "未知大小文件",
			info:     &FileInfo{Size: -1},
			expected: StrategySingleThread,
		},
		{
			name:     "小文件",
			info:     &FileInfo{Size: 1 * 1024 * 1024}, // 1MB
			expected: StrategySingleThread,
		},
		{
			name:     "不支持Range的中等文件",
			info:     &FileInfo{Size: 10 * 1024 * 1024, AcceptRanges: false}, // 10MB
			expected: StrategyMultiConnection,
		},
		{
			name:     "支持Range的中等文件",
			info:     &FileInfo{Size: 10 * 1024 * 1024, AcceptRanges: true}, // 10MB
			expected: StrategyMultiConnection,
		},
		{
			name:     "大文件",
			info:     &FileInfo{Size: 100 * 1024 * 1024, AcceptRanges: true}, // 100MB
			expected: StrategyChunked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := downloader.selectStrategy(tt.info)
			if strategy != tt.expected {
				t.Errorf("selectStrategy() = %v, want %v", strategy, tt.expected)
			}
		})
	}
}

func TestAdaptiveDownloader_calculateOptimalChunks(t *testing.T) {
	config := DefaultConfig()
	downloader := NewAdaptiveDownloader(config)

	tests := []struct {
		fileSize         int64
		minExpectedChunks int
		maxExpectedChunks int
	}{
		{1 * 1024 * 1024, 1, 1},       // 1MB
		{10 * 1024 * 1024, 1, 2},      // 10MB
		{100 * 1024 * 1024, 8, 16},    // 100MB
		{1024 * 1024 * 1024, 16, 16},  // 1GB
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%dMB", tt.fileSize/1024/1024), func(t *testing.T) {
			numChunks, chunkSize := downloader.calculateOptimalChunks(tt.fileSize)

			if numChunks < tt.minExpectedChunks || numChunks > tt.maxExpectedChunks {
				t.Errorf("numChunks = %d, expected between %d and %d",
					numChunks, tt.minExpectedChunks, tt.maxExpectedChunks)
			}

			// 验证分片大小
			totalSize := int64(numChunks) * chunkSize
			if totalSize < tt.fileSize {
				t.Errorf("total chunk size %d < file size %d", totalSize, tt.fileSize)
			}

			t.Logf("File size: %d MB, Chunks: %d, Chunk size: %d MB",
				tt.fileSize/1024/1024, numChunks, chunkSize/1024/1024)
		})
	}
}

func TestSimpleDownloader_Download(t *testing.T) {
	// 创建测试服务器
	content := []byte("Hello, World! This is a test file content.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	// 创建临时目录
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test_download.txt")

	downloader := NewSimpleDownloader()
	ctx := context.Background()

	var progressCalled bool
	progress := func(downloaded, total int64, speed float64) {
		progressCalled = true
		t.Logf("Progress: %d/%d bytes, speed: %.2f bytes/s", downloaded, total, speed)
	}

	err := downloader.Download(ctx, server.URL, destPath, progress)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file does not exist")
	}

	// 验证文件内容
	downloadedContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(downloadedContent) != string(content) {
		t.Errorf("Downloaded content mismatch: got %s, want %s", downloadedContent, content)
	}

	if !progressCalled {
		t.Error("Progress callback was not called")
	}

	// 验证统计信息
	stats := downloader.GetStats()
	if stats.TotalBytes != int64(len(content)) {
		t.Errorf("TotalBytes = %d, want %d", stats.TotalBytes, len(content))
	}
}

func TestBandwidthMonitor(t *testing.T) {
	bm := NewBandwidthMonitor()

	// 模拟数据传输
	bm.RecordBytes(1024 * 1024) // 1MB
	time.Sleep(100 * time.Millisecond)
	bm.RecordBytes(1024 * 1024) // 1MB
	time.Sleep(100 * time.Millisecond)
	bm.RecordBytes(1024 * 1024) // 1MB

	// 获取统计
	currentSpeed := bm.GetCurrentSpeed()
	avgSpeed := bm.GetAverageSpeed()
	totalBytes := bm.GetTotalBytes()
	peakSpeed := bm.GetPeakSpeed()

	if totalBytes != 3*1024*1024 {
		t.Errorf("TotalBytes = %d, want %d", totalBytes, 3*1024*1024)
	}

	if currentSpeed <= 0 {
		t.Error("CurrentSpeed should be > 0")
	}

	if avgSpeed <= 0 {
		t.Error("AverageSpeed should be > 0")
	}

	if peakSpeed <= 0 {
		t.Error("PeakSpeed should be > 0")
	}

	t.Logf("Current: %d B/s, Avg: %d B/s, Peak: %d B/s, Total: %d MB",
		currentSpeed, avgSpeed, peakSpeed, totalBytes/1024/1024)
}

func TestCongestionController(t *testing.T) {
	cc := NewCongestionController()

	// 初始状态
	if cc.GetCurrentConcurrency() != 1 {
		t.Errorf("Initial concurrency = %d, want 1", cc.GetCurrentConcurrency())
	}

	// 模拟成功下载（慢启动阶段）
	for i := 0; i < 3; i++ {
		cc.OnSuccess()
	}

	concurrency := cc.GetCurrentConcurrency()
	if concurrency <= 1 {
		t.Errorf("Concurrency after successes = %d, want > 1", concurrency)
	}

	window := cc.GetCongestionWindow()
	if window <= 1 {
		t.Errorf("Congestion window after successes = %f, want > 1", window)
	}

	// 模拟失败（拥塞发生）
	cc.OnFailure()

	newConcurrency := cc.GetCurrentConcurrency()
	if newConcurrency >= concurrency {
		t.Errorf("Concurrency after failure = %d, should be less than %d", newConcurrency, concurrency)
	}

	t.Logf("Final concurrency: %d, window: %f", cc.GetCurrentConcurrency(), cc.GetCongestionWindow())
}

func TestChunk_Progress(t *testing.T) {
	chunk := &Chunk{
		Index:    0,
		Start:    0,
		End:      1023,
		TempPath: filepath.Join(t.TempDir(), "chunk0.tmp"),
	}

	// 初始状态
	if chunk.GetDownloaded() != 0 {
		t.Errorf("Initial downloaded = %d, want 0", chunk.GetDownloaded())
	}

	if chunk.IsCompleted() {
		t.Error("Initial completed should be false")
	}

	// 模拟下载进度
	chunk.mu.Lock()
	chunk.downloaded = 512
	chunk.mu.Unlock()

	if chunk.GetDownloaded() != 512 {
		t.Errorf("Downloaded = %d, want 512", chunk.GetDownloaded())
	}
}
