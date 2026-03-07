// Package cache 提供性能监控和分析功能。
//
// 该文件实现了缓存性能监控、IO 统计和性能分析工具，
// 帮助诊断性能瓶颈和优化系统表现。

package cache

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"chopsticks/pkg/metrics"
)

// ============================================================================
// 性能分析器
// ============================================================================

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	mu         sync.RWMutex
	cacheStats []CacheStats
	ioStats    []IOStats
	samples    []PerformanceSample
	maxSamples int
	sampleRate time.Duration
	stopChan   chan struct{}
	doneChan   chan struct{}
	running    bool
}

// PerformanceSample 性能样本
type PerformanceSample struct {
	Timestamp time.Time
	Cache     CacheStats
	IO        IOStats
	Memory    MemoryStats
}

// IOStats IO 统计
type IOStats struct {
	ReadOps       int64         // 读取操作数
	WriteOps      int64         // 写入操作数
	ReadBytes     int64         // 读取字节数
	WriteBytes    int64         // 写入字节数
	ReadDuration  time.Duration // 读取耗时
	WriteDuration time.Duration // 写入耗时
	BatchReads    int64         // 批量读取次数
	Prefetches    int64         // 预取次数
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc   uint64 // 已分配内存
	Sys     uint64 // 系统内存
	GCCount uint32 // GC 次数
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(sampleRate time.Duration, maxSamples int) *PerformanceAnalyzer {
	if sampleRate <= 0 {
		sampleRate = 5 * time.Second
	}
	if maxSamples <= 0 {
		maxSamples = 1000
	}

	return &PerformanceAnalyzer{
		sampleRate: sampleRate,
		maxSamples: maxSamples,
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
	}
}

// Start 启动分析器
func (pa *PerformanceAnalyzer) Start() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	if pa.running {
		return
	}

	pa.running = true
	go pa.analyzeLoop()
}

// Stop 停止分析器
func (pa *PerformanceAnalyzer) Stop() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	if !pa.running {
		return
	}

	close(pa.stopChan)
	<-pa.doneChan
	pa.running = false
}

// analyzeLoop 分析循环
func (pa *PerformanceAnalyzer) analyzeLoop() {
	defer close(pa.doneChan)

	ticker := time.NewTicker(pa.sampleRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pa.collectSample()
		case <-pa.stopChan:
			return
		}
	}
}

// collectSample 收集样本
func (pa *PerformanceAnalyzer) collectSample() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	sample := PerformanceSample{
		Timestamp: time.Now(),
		Cache:     pa.getCacheStats(),
		IO:        pa.getIOStats(),
		Memory:    pa.getMemoryStats(),
	}

	pa.samples = append(pa.samples, sample)

	// 限制样本数量
	if len(pa.samples) > pa.maxSamples {
		pa.samples = pa.samples[1:]
	}
}

// getCacheStats 获取缓存统计
func (pa *PerformanceAnalyzer) getCacheStats() CacheStats {
	m := metrics.GetMetrics()
	return CacheStats{
		Hits:      int64(m.AppCacheSize),
		Misses:    0,
		Size:      0,
		MaxSize:   0,
		HitRate:   m.AppCacheHitRate,
		Entries:   m.AppCacheSize,
		Evictions: m.CacheEvictions,
	}
}

// getIOStats 获取 IO 统计（需要外部实现）
func (pa *PerformanceAnalyzer) getIOStats() IOStats {
	// 这里需要集成到实际的 IO 操作中
	// 暂时返回空统计
	return IOStats{}
}

// getMemoryStats 获取内存统计
func (pa *PerformanceAnalyzer) getMemoryStats() MemoryStats {
	_, _, _, memStats := metrics.GetSystemMetrics()
	return MemoryStats{
		Alloc:   memStats.Alloc,
		Sys:     memStats.Sys,
		GCCount: memStats.NumGC,
	}
}

// GetSamples 获取样本
func (pa *PerformanceAnalyzer) GetSamples(n int) []PerformanceSample {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if n > len(pa.samples) {
		n = len(pa.samples)
	}
	if n <= 0 {
		return nil
	}

	result := make([]PerformanceSample, n)
	copy(result, pa.samples[len(pa.samples)-n:])
	return result
}

// GetAverageStats 获取平均统计
func (pa *PerformanceAnalyzer) GetAverageStats() (CacheStats, IOStats, MemoryStats) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if len(pa.samples) == 0 {
		return CacheStats{}, IOStats{}, MemoryStats{}
	}

	var avgCache CacheStats
	var avgIO IOStats
	var avgMemory MemoryStats

	count := len(pa.samples)
	for _, sample := range pa.samples {
		avgCache.Hits += sample.Cache.Hits
		avgCache.HitRate += sample.Cache.HitRate
		avgCache.Evictions += sample.Cache.Evictions

		avgIO.ReadOps += sample.IO.ReadOps
		avgIO.WriteOps += sample.IO.WriteOps
		avgIO.ReadBytes += sample.IO.ReadBytes
		avgIO.WriteBytes += sample.IO.WriteBytes
		avgIO.BatchReads += sample.IO.BatchReads
		avgIO.Prefetches += sample.IO.Prefetches

		avgMemory.Alloc += sample.Memory.Alloc
		avgMemory.Sys += sample.Memory.Sys
		avgMemory.GCCount += sample.Memory.GCCount
	}

	avgCache.HitRate /= float64(count)
	avgMemory.Alloc /= uint64(count)
	avgMemory.Sys /= uint64(count)
	avgMemory.GCCount /= uint32(count)

	return avgCache, avgIO, avgMemory
}

// Report 生成性能报告
func (pa *PerformanceAnalyzer) Report(w io.Writer) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w, "Performance Analysis Report")
	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w)

	if len(pa.samples) == 0 {
		fmt.Fprintln(w, "No performance data collected yet.")
		return
	}

	// 基本信息
	fmt.Fprintf(w, "Samples: %d\n", len(pa.samples))
	fmt.Fprintf(w, "Sample Rate: %v\n", pa.sampleRate)
	fmt.Fprintln(w)

	// 平均统计
	avgCache, avgIO, avgMemory := pa.GetAverageStats()

	fmt.Fprintln(w, "Average Cache Statistics:")
	fmt.Fprintf(w, "  Hit Rate: %.2f%%\n", avgCache.HitRate)
	fmt.Fprintf(w, "  Evictions: %d\n", avgCache.Evictions)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Average IO Statistics:")
	fmt.Fprintf(w, "  Read Ops: %d\n", avgIO.ReadOps)
	fmt.Fprintf(w, "  Write Ops: %d\n", avgIO.WriteOps)
	fmt.Fprintf(w, "  Batch Reads: %d\n", avgIO.BatchReads)
	fmt.Fprintf(w, "  Prefetches: %d\n", avgIO.Prefetches)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Average Memory Statistics:")
	fmt.Fprintf(w, "  Allocated: %d MB\n", avgMemory.Alloc/1024/1024)
	fmt.Fprintf(w, "  System: %d MB\n", avgMemory.Sys/1024/1024)
	fmt.Fprintf(w, "  GC Count: %d\n", avgMemory.GCCount)
	fmt.Fprintln(w)

	// 趋势分析
	pa.analyzeTrends(w)
}

// analyzeTrends 分析趋势
func (pa *PerformanceAnalyzer) analyzeTrends(w io.Writer) {
	if len(pa.samples) < 2 {
		return
	}

	fmt.Fprintln(w, "Performance Trends:")

	// 缓存命中率趋势
	first := pa.samples[0]
	last := pa.samples[len(pa.samples)-1]

	hitRateChange := last.Cache.HitRate - first.Cache.HitRate
	if hitRateChange > 5 {
		fmt.Fprintf(w, "  ✓ Cache hit rate improving (+%.2f%%)\n", hitRateChange)
	} else if hitRateChange < -5 {
		fmt.Fprintf(w, "  ⚠ Cache hit rate decreasing (%.2f%%)\n", hitRateChange)
	}

	// 内存趋势
	memoryChange := int64(last.Memory.Alloc) - int64(first.Memory.Alloc)
	if memoryChange > 10*1024*1024 {
		fmt.Fprintf(w, "  ⚠ Memory usage increasing (+%d MB)\n", memoryChange/1024/1024)
	} else if memoryChange < -10*1024*1024 {
		fmt.Fprintf(w, "  ✓ Memory usage decreasing (%d MB)\n", memoryChange/1024/1024)
	}

	fmt.Fprintln(w)
}

// ============================================================================
// IO 追踪器
// ============================================================================

// IOTracker IO 追踪器
type IOTracker struct {
	mu            sync.RWMutex
	operations    []IOOperation
	maxOperations int
}

// IOOperation IO 操作
type IOOperation struct {
	Type      string        // read, write, batch_read, prefetch
	Path      string        // 文件路径
	Size      int64         // 操作大小
	Duration  time.Duration // 耗时
	Timestamp time.Time     // 时间戳
	Success   bool          // 是否成功
}

// NewIOTracker 创建 IO 追踪器
func NewIOTracker(maxOperations int) *IOTracker {
	if maxOperations <= 0 {
		maxOperations = 10000
	}

	return &IOTracker{
		maxOperations: maxOperations,
	}
}

// Record 记录 IO 操作
func (iot *IOTracker) Record(op IOOperation) {
	iot.mu.Lock()
	defer iot.mu.Unlock()

	iot.operations = append(iot.operations, op)

	// 限制操作记录数量
	if len(iot.operations) > iot.maxOperations {
		iot.operations = iot.operations[1:]
	}
}

// GetStats 获取统计信息
func (iot *IOTracker) GetStats() IOStats {
	iot.mu.RLock()
	defer iot.mu.RUnlock()

	var stats IOStats
	for _, op := range iot.operations {
		switch op.Type {
		case "read":
			stats.ReadOps++
			stats.ReadBytes += op.Size
			stats.ReadDuration += op.Duration
		case "write":
			stats.WriteOps++
			stats.WriteBytes += op.Size
			stats.WriteDuration += op.Duration
		case "batch_read":
			stats.BatchReads++
			stats.ReadOps++
			stats.ReadBytes += op.Size
			stats.ReadDuration += op.Duration
		case "prefetch":
			stats.Prefetches++
			stats.ReadOps++
			stats.ReadBytes += op.Size
			stats.ReadDuration += op.Duration
		}
	}

	return stats
}

// GetSlowOperations 获取慢操作
func (iot *IOTracker) GetSlowOperations(threshold time.Duration, n int) []IOOperation {
	iot.mu.RLock()
	defer iot.mu.RUnlock()

	var slowOps []IOOperation
	for _, op := range iot.operations {
		if op.Duration >= threshold {
			slowOps = append(slowOps, op)
		}
	}

	// 按耗时排序
	sort.Slice(slowOps, func(i, j int) bool {
		return slowOps[i].Duration > slowOps[j].Duration
	})

	if n > 0 && len(slowOps) > n {
		return slowOps[:n]
	}

	return slowOps
}

// Report 生成 IO 报告
func (iot *IOTracker) Report(w io.Writer) {
	stats := iot.GetStats()

	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w, "IO Statistics Report")
	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Read Operations:  %d\n", stats.ReadOps)
	fmt.Fprintf(w, "Write Operations: %d\n", stats.WriteOps)
	fmt.Fprintf(w, "Batch Reads:      %d\n", stats.BatchReads)
	fmt.Fprintf(w, "Prefetches:       %d\n", stats.Prefetches)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Total Read Bytes:    %s\n", formatBytes(stats.ReadBytes))
	fmt.Fprintf(w, "Total Write Bytes:   %s\n", formatBytes(stats.WriteBytes))
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Avg Read Duration:  %v\n", stats.ReadDuration/time.Duration(stats.ReadOps))
	fmt.Fprintf(w, "Avg Write Duration: %v\n", stats.WriteDuration/time.Duration(stats.WriteOps))
	fmt.Fprintln(w)

	// 慢操作
	slowOps := iot.GetSlowOperations(100*time.Millisecond, 10)
	if len(slowOps) > 0 {
		fmt.Fprintln(w, "Slow Operations (>100ms):")
		for i, op := range slowOps {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "  %d. %s %s - %v\n", i+1, op.Type, op.Path, op.Duration)
		}
		fmt.Fprintln(w)
	}
}

// ============================================================================
// 缓存健康检查
// ============================================================================

// HealthChecker 健康检查器
type HealthChecker struct {
	cache *Cache
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(cache *Cache) *HealthChecker {
	return &HealthChecker{
		cache: cache,
	}
}

// Check 执行健康检查
func (hc *HealthChecker) Check() HealthStatus {
	stats := hc.cache.Stats()

	status := HealthStatus{
		Timestamp: time.Now(),
		Healthy:   true,
		Issues:    []string{},
	}

	// 检查命中率
	if stats.HitRate < 50 {
		status.Issues = append(status.Issues, fmt.Sprintf("Low cache hit rate: %.2f%%", stats.HitRate))
		status.Warnings = append(status.Warnings, "Consider increasing cache size or TTL")
	}

	// 检查淘汰率
	totalOps := stats.Hits + stats.Misses
	if totalOps > 0 {
		evictionRate := float64(stats.Evictions) / float64(totalOps) * 100
		if evictionRate > 20 {
			status.Issues = append(status.Issues, fmt.Sprintf("High eviction rate: %.2f%%", evictionRate))
			status.Warnings = append(status.Warnings, "Cache size may be too small")
		}
	}

	// 检查缓存使用率
	if stats.MaxSize > 0 {
		usageRate := float64(stats.Size) / float64(stats.MaxSize) * 100
		if usageRate > 95 {
			status.Issues = append(status.Issues, fmt.Sprintf("Cache nearly full: %.2f%%", usageRate))
			status.Warnings = append(status.Warnings, "Consider increasing max cache size")
		}
	}

	if len(status.Issues) > 0 {
		status.Healthy = false
	}

	return status
}

// HealthStatus 健康状态
type HealthStatus struct {
	Timestamp time.Time
	Healthy   bool
	Issues    []string
	Warnings  []string
}

// Report 生成健康报告
func (hs HealthStatus) Report(w io.Writer) {
	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w, "Cache Health Status")
	fmt.Fprintln(w, "========================================")
	fmt.Fprintln(w)

	if hs.Healthy {
		fmt.Fprintln(w, "Status: ✓ Healthy")
	} else {
		fmt.Fprintln(w, "Status: ⚠ Issues Detected")
	}
	fmt.Fprintln(w)

	if len(hs.Issues) > 0 {
		fmt.Fprintln(w, "Issues:")
		for _, issue := range hs.Issues {
			fmt.Fprintf(w, "  - %s\n", issue)
		}
		fmt.Fprintln(w)
	}

	if len(hs.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings:")
		for _, warning := range hs.Warnings {
			fmt.Fprintf(w, "  - %s\n", warning)
		}
		fmt.Fprintln(w)
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetCacheReport 获取缓存报告
func GetCacheReport(cache *Cache) string {
	var sb strings.Builder

	stats := cache.Stats()

	sb.WriteString("Cache Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Hits: %d\n", stats.Hits))
	sb.WriteString(fmt.Sprintf("  Misses: %d\n", stats.Misses))
	sb.WriteString(fmt.Sprintf("  Hit Rate: %.2f%%\n", stats.HitRate))
	sb.WriteString(fmt.Sprintf("  Size: %s / %s\n", formatBytes(stats.Size), formatBytes(stats.MaxSize)))
	sb.WriteString(fmt.Sprintf("  Entries: %d / %d\n", stats.Entries, stats.MaxEntries))
	sb.WriteString(fmt.Sprintf("  Evictions: %d\n", stats.Evictions))

	return sb.String()
}

// PrintCacheReport 打印缓存报告
func PrintCacheReport(cache *Cache, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	stats := cache.Stats()

	fmt.Fprintln(w, "Cache Statistics:")
	fmt.Fprintf(w, "  Hits: %d\n", stats.Hits)
	fmt.Fprintf(w, "  Misses: %d\n", stats.Misses)
	fmt.Fprintf(w, "  Hit Rate: %.2f%%\n", stats.HitRate)
	fmt.Fprintf(w, "  Size: %s / %s\n", formatBytes(stats.Size), formatBytes(stats.MaxSize))
	fmt.Fprintf(w, "  Entries: %d / %d\n", stats.Entries, stats.MaxEntries)
	fmt.Fprintf(w, "  Evictions: %d\n", stats.Evictions)
}
