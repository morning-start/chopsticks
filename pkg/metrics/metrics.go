package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics 性能指标结构体
type PerformanceMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// 任务统计指标
	TaskSubmitRate   float64       `json:"task_submit_rate"`   // 每秒提交任务数
	TaskCompleteRate float64       `json:"task_complete_rate"` // 每秒完成任务数
	TaskInProgress   int           `json:"task_in_progress"`   // 当前进行中的任务数
	AvgTaskDuration  time.Duration `json:"avg_task_duration"`  // 平均任务执行时间
	TotalTasks       int64         `json:"total_tasks"`        // 总任务数
	CompletedTasks   int64         `json:"completed_tasks"`    // 已完成任务数
	FailedTasks      int64         `json:"failed_tasks"`       // 失败任务数

	// 资源使用指标
	CPUUsage       float64 `json:"cpu_usage"`       // CPU 使用率 (百分比)
	MemoryUsage    float64 `json:"memory_usage"`    // 内存使用 (MB)
	MemoryAlloc    uint64  `json:"memory_alloc"`    // 已分配内存
	MemorySys      uint64  `json:"memory_sys"`      // 系统内存
	GoroutineCount int     `json:"goroutine_count"` // Goroutine 数量
	NumGC          uint32  `json:"num_gc"`          // GC 次数

	// JS 引擎池指标
	JSPoolSize        int     `json:"js_pool_size"`        // JS 池大小
	JSPoolActive      int     `json:"js_pool_active"`      // 活跃 JS 引擎数
	JSPoolUtilization float64 `json:"js_pool_utilization"` // JS 池利用率
	JSCacheHitRate    float64 `json:"js_cache_hit_rate"`   // JS 缓存命中率
	JSCacheSize       int     `json:"js_cache_size"`       // JS 缓存大小

	// 下载指标
	DownloadSpeed    float64 `json:"download_speed"`     // 下载速度 (MB/s)
	ActiveDownloads  int     `json:"active_downloads"`   // 活跃下载数
	DownloadErrors   int64   `json:"download_errors"`    // 下载错误数
	TotalDownloaded  int64   `json:"total_downloaded"`   // 总下载字节数
	AvgDownloadSpeed float64 `json:"avg_download_speed"` // 平均下载速度

	// 搜索指标
	SearchCacheHitRate float64 `json:"search_cache_hit_rate"` // 搜索缓存命中率
	SearchCacheSize    int     `json:"search_cache_size"`     // 搜索缓存大小
	ActiveSearches     int     `json:"active_searches"`       // 活跃搜索数

	// 安装指标
	ActiveInstalls   int `json:"active_installs"`    // 活跃安装数
	InstallQueueSize int `json:"install_queue_size"` // 安装队列大小

	// 缓存指标 (新增)
	AppCacheHitRate    float64 `json:"app_cache_hit_rate"`    // 应用缓存命中率
	AppCacheSize       int     `json:"app_cache_size"`        // 应用缓存大小
	BucketCacheHitRate float64 `json:"bucket_cache_hit_rate"` // Bucket 缓存命中率
	BucketCacheSize    int     `json:"bucket_cache_size"`     // Bucket 缓存大小
	IndexCacheHitRate  float64 `json:"index_cache_hit_rate"`  // 索引缓存命中率
	IndexCacheSize     int     `json:"index_cache_size"`      // 索引缓存大小
	CacheEvictions     int64   `json:"cache_evictions"`       // 缓存淘汰次数
	BatchReadEfficiency float64 `json:"batch_read_efficiency"` // 批量读取效率
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	Timestamp time.Time          `json:"timestamp"`
	Metrics   PerformanceMetrics `json:"metrics"`
}

// MetricsHistory 指标历史
type MetricsHistory struct {
	snapshots []MetricsSnapshot
	maxSize   int
	mu        sync.RWMutex
}

// NewMetricsHistory 创建指标历史
func NewMetricsHistory(maxSize int) *MetricsHistory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &MetricsHistory{
		snapshots: make([]MetricsSnapshot, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Add 添加指标快照
func (h *MetricsHistory) Add(snapshot MetricsSnapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.snapshots = append(h.snapshots, snapshot)
	if len(h.snapshots) > h.maxSize {
		h.snapshots = h.snapshots[1:]
	}
}

// GetRecent 获取最近的指标快照
func (h *MetricsHistory) GetRecent(n int) []MetricsSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n > len(h.snapshots) {
		n = len(h.snapshots)
	}
	if n <= 0 {
		return nil
	}

	result := make([]MetricsSnapshot, n)
	copy(result, h.snapshots[len(h.snapshots)-n:])
	return result
}

// GetAll 获取所有指标快照
func (h *MetricsHistory) GetAll() []MetricsSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]MetricsSnapshot, len(h.snapshots))
	copy(result, h.snapshots)
	return result
}

// Clear 清空历史
func (h *MetricsHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.snapshots = h.snapshots[:0]
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	Submitted   int64         `json:"submitted"`
	Completed   int64         `json:"completed"`
	Failed      int64         `json:"failed"`
	InProgress  int64         `json:"in_progress"`
	TotalTime   time.Duration `json:"total_time"`
	AvgDuration time.Duration `json:"avg_duration"`
}

// JSPoolMetrics JS 引擎池指标
type JSPoolMetrics struct {
	PoolSize      int     `json:"pool_size"`
	Active        int     `json:"active"`
	Idle          int     `json:"idle"`
	Utilization   float64 `json:"utilization"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	CacheSize     int     `json:"cache_size"`
	WaitQueueSize int     `json:"wait_queue_size"`
}

// DownloadMetrics 下载指标
type DownloadMetrics struct {
	ActiveCount    int           `json:"active_count"`
	QueueSize      int           `json:"queue_size"`
	CurrentSpeed   float64       `json:"current_speed"` // MB/s
	AvgSpeed       float64       `json:"avg_speed"`     // MB/s
	TotalBytes     int64         `json:"total_bytes"`
	ErrorCount     int64         `json:"error_count"`
	TotalTime      time.Duration `json:"total_time"`
	BytesPerSecond float64       `json:"bytes_per_second"`
}

// GetSystemMetrics 获取系统指标
func GetSystemMetrics() (cpuUsage float64, memoryMB float64, goroutines int, memStats runtime.MemStats) {
	runtime.ReadMemStats(&memStats)

	// 获取内存使用 (MB)
	memoryMB = float64(memStats.Alloc) / 1024 / 1024

	// 获取 Goroutine 数量
	goroutines = runtime.NumGoroutine()

	// CPU 使用率需要通过外部工具获取，这里返回 0
	// 实际实现可以使用 github.com/shirou/gopsutil
	cpuUsage = 0

	return cpuUsage, memoryMB, goroutines, memStats
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
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

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d μs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
	return fmt.Sprintf("%.2f min", d.Minutes())
}

// CalculateRate 计算速率
func CalculateRate(count int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(count) / duration.Seconds()
}

// CalculatePercentage 计算百分比
func CalculatePercentage(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
