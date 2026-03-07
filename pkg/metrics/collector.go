package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	// 任务指标
	taskSubmitted  int64
	taskCompleted  int64
	taskFailed     int64
	taskInProgress int64
	taskTotalTime  int64 // 纳秒

	// 下载指标
	downloadActive    int64
	downloadErrors    int64
	downloadTotal     int64 // 字节
	downloadTotalTime int64 // 纳秒

	// 搜索指标
	searchActive      int64
	searchCacheHits   int64
	searchCacheMisses int64

	// 安装指标
	installActive int64
	installQueue  int64

	// JS 池指标
	jsPoolSize    int64
	jsPoolActive  int64
	jsCacheHits   int64
	jsCacheMisses int64

	// 缓存指标
	appCacheHits    int64
	appCacheMisses  int64
	bucketCacheHits int64
	bucketCacheMisses int64
	indexCacheHits  int64
	indexCacheMisses int64
	cacheEvictions  int64
	batchReadHits   int64
	batchReadMisses int64

	// 历史记录
	history *MetricsHistory

	// 采样间隔
	sampleInterval time.Duration

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(sampleInterval time.Duration) *MetricsCollector {
	if sampleInterval <= 0 {
		sampleInterval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsCollector{
		history:        NewMetricsHistory(1000),
		sampleInterval: sampleInterval,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动指标收集
func (c *MetricsCollector) Start() {
	c.wg.Add(1)
	go c.collectLoop()
}

// Stop 停止指标收集
func (c *MetricsCollector) Stop() {
	c.cancel()
	c.wg.Wait()
}

// collectLoop 收集循环
func (c *MetricsCollector) collectLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.sampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			metrics := c.Collect()
			c.history.Add(MetricsSnapshot{
				Timestamp: time.Now(),
				Metrics:   metrics,
			})
		}
	}
}

// Collect 收集当前指标
func (c *MetricsCollector) Collect() PerformanceMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 获取系统指标
	_, memoryMB, goroutines, memStats := GetSystemMetrics()

	// 计算任务速率
	now := time.Now()
	recentSnapshots := c.history.GetRecent(2)

	var taskSubmitRate, taskCompleteRate float64
	if len(recentSnapshots) >= 2 {
		duration := recentSnapshots[1].Timestamp.Sub(recentSnapshots[0].Timestamp).Seconds()
		if duration > 0 {
			taskSubmitRate = float64(atomic.LoadInt64(&c.taskSubmitted)-recentSnapshots[0].Metrics.TotalTasks) / duration
			taskCompleteRate = float64(atomic.LoadInt64(&c.taskCompleted)-recentSnapshots[0].Metrics.CompletedTasks) / duration
		}
	}

	// 计算平均任务时间
	avgTaskDuration := time.Duration(0)
	completed := atomic.LoadInt64(&c.taskCompleted)
	if completed > 0 {
		avgTaskDuration = time.Duration(atomic.LoadInt64(&c.taskTotalTime) / completed)
	}

	// 计算下载速度
	avgDownloadSpeed := float64(0)
	totalDownloadTime := atomic.LoadInt64(&c.downloadTotalTime)
	if totalDownloadTime > 0 {
		avgDownloadSpeed = float64(atomic.LoadInt64(&c.downloadTotal)) / float64(totalDownloadTime) * 1e9 / 1024 / 1024 // MB/s
	}

	// 计算缓存命中率
	jsCacheHitRate := float64(0)
	jsHits := atomic.LoadInt64(&c.jsCacheHits)
	jsMisses := atomic.LoadInt64(&c.jsCacheMisses)
	if jsHits+jsMisses > 0 {
		jsCacheHitRate = float64(jsHits) / float64(jsHits+jsMisses) * 100
	}

	searchCacheHitRate := float64(0)
	searchHits := atomic.LoadInt64(&c.searchCacheHits)
	searchMisses := atomic.LoadInt64(&c.searchCacheMisses)
	if searchHits+searchMisses > 0 {
		searchCacheHitRate = float64(searchHits) / float64(searchHits+searchMisses) * 100
	}

	// 计算 JS 池利用率
	jsPoolUtilization := float64(0)
	poolSize := atomic.LoadInt64(&c.jsPoolSize)
	poolActive := atomic.LoadInt64(&c.jsPoolActive)
	if poolSize > 0 {
		jsPoolUtilization = float64(poolActive) / float64(poolSize) * 100
	}

	// 计算缓存命中率
	appCacheHitRate := float64(0)
	appHits := atomic.LoadInt64(&c.appCacheHits)
	appMisses := atomic.LoadInt64(&c.appCacheMisses)
	if appHits+appMisses > 0 {
		appCacheHitRate = float64(appHits) / float64(appHits+appMisses) * 100
	}

	bucketCacheHitRate := float64(0)
	bucketHits := atomic.LoadInt64(&c.bucketCacheHits)
	bucketMisses := atomic.LoadInt64(&c.bucketCacheMisses)
	if bucketHits+bucketMisses > 0 {
		bucketCacheHitRate = float64(bucketHits) / float64(bucketHits+bucketMisses) * 100
	}

	indexCacheHitRate := float64(0)
	indexHits := atomic.LoadInt64(&c.indexCacheHits)
	indexMisses := atomic.LoadInt64(&c.indexCacheMisses)
	if indexHits+indexMisses > 0 {
		indexCacheHitRate = float64(indexHits) / float64(indexHits+indexMisses) * 100
	}

	batchReadEfficiency := float64(0)
	batchHits := atomic.LoadInt64(&c.batchReadHits)
	batchMisses := atomic.LoadInt64(&c.batchReadMisses)
	if batchHits+batchMisses > 0 {
		batchReadEfficiency = float64(batchHits) / float64(batchHits+batchMisses) * 100
	}

	return PerformanceMetrics{
		Timestamp:          now,
		TaskSubmitRate:     taskSubmitRate,
		TaskCompleteRate:   taskCompleteRate,
		TaskInProgress:     int(atomic.LoadInt64(&c.taskInProgress)),
		AvgTaskDuration:    avgTaskDuration,
		TotalTasks:         atomic.LoadInt64(&c.taskSubmitted),
		CompletedTasks:     completed,
		FailedTasks:        atomic.LoadInt64(&c.taskFailed),
		CPUUsage:           0, // 需要通过外部工具获取
		MemoryUsage:        memoryMB,
		MemoryAlloc:        memStats.Alloc,
		MemorySys:          memStats.Sys,
		GoroutineCount:     goroutines,
		NumGC:              memStats.NumGC,
		JSPoolSize:         int(poolSize),
		JSPoolActive:       int(poolActive),
		JSPoolUtilization:  jsPoolUtilization,
		JSCacheHitRate:      jsCacheHitRate,
		JSCacheSize:         int(jsHits + jsMisses),
		DownloadSpeed:       avgDownloadSpeed,
		ActiveDownloads:     int(atomic.LoadInt64(&c.downloadActive)),
		DownloadErrors:      atomic.LoadInt64(&c.downloadErrors),
		TotalDownloaded:     atomic.LoadInt64(&c.downloadTotal),
		AvgDownloadSpeed:    avgDownloadSpeed,
		SearchCacheHitRate:  searchCacheHitRate,
		SearchCacheSize:     int(searchHits + searchMisses),
		ActiveSearches:      int(atomic.LoadInt64(&c.searchActive)),
		ActiveInstalls:      int(atomic.LoadInt64(&c.installActive)),
		InstallQueueSize:    int(atomic.LoadInt64(&c.installQueue)),
		AppCacheHitRate:     appCacheHitRate,
		AppCacheSize:        int(appHits + appMisses),
		BucketCacheHitRate:  bucketCacheHitRate,
		BucketCacheSize:     int(bucketHits + bucketMisses),
		IndexCacheHitRate:   indexCacheHitRate,
		IndexCacheSize:      int(indexHits + indexMisses),
		CacheEvictions:      atomic.LoadInt64(&c.cacheEvictions),
		BatchReadEfficiency: batchReadEfficiency,
	}
}

// GetCurrentMetrics 获取当前指标
func (c *MetricsCollector) GetCurrentMetrics() PerformanceMetrics {
	return c.Collect()
}

// GetHistory 获取指标历史
func (c *MetricsCollector) GetHistory() *MetricsHistory {
	return c.history
}

// Task 相关方法

// RecordTaskSubmitted 记录任务提交
func (c *MetricsCollector) RecordTaskSubmitted() {
	atomic.AddInt64(&c.taskSubmitted, 1)
	atomic.AddInt64(&c.taskInProgress, 1)
}

// RecordTaskCompleted 记录任务完成
func (c *MetricsCollector) RecordTaskCompleted(duration time.Duration) {
	atomic.AddInt64(&c.taskCompleted, 1)
	atomic.AddInt64(&c.taskInProgress, -1)
	atomic.AddInt64(&c.taskTotalTime, int64(duration))
}

// RecordTaskFailed 记录任务失败
func (c *MetricsCollector) RecordTaskFailed() {
	atomic.AddInt64(&c.taskFailed, 1)
	atomic.AddInt64(&c.taskInProgress, -1)
}

// Download 相关方法

// RecordDownloadStart 记录下载开始
func (c *MetricsCollector) RecordDownloadStart() {
	atomic.AddInt64(&c.downloadActive, 1)
}

// RecordDownloadComplete 记录下载完成
func (c *MetricsCollector) RecordDownloadComplete(bytes int64, duration time.Duration) {
	atomic.AddInt64(&c.downloadActive, -1)
	atomic.AddInt64(&c.downloadTotal, bytes)
	atomic.AddInt64(&c.downloadTotalTime, int64(duration))
}

// RecordDownloadError 记录下载错误
func (c *MetricsCollector) RecordDownloadError() {
	atomic.AddInt64(&c.downloadErrors, 1)
	atomic.AddInt64(&c.downloadActive, -1)
}

// Search 相关方法

// RecordSearchStart 记录搜索开始
func (c *MetricsCollector) RecordSearchStart() {
	atomic.AddInt64(&c.searchActive, 1)
}

// RecordSearchComplete 记录搜索完成
func (c *MetricsCollector) RecordSearchComplete(cacheHit bool) {
	atomic.AddInt64(&c.searchActive, -1)
	if cacheHit {
		atomic.AddInt64(&c.searchCacheHits, 1)
	} else {
		atomic.AddInt64(&c.searchCacheMisses, 1)
	}
}

// Install 相关方法

// RecordInstallStart 记录安装开始
func (c *MetricsCollector) RecordInstallStart() {
	atomic.AddInt64(&c.installActive, 1)
}

// RecordInstallComplete 记录安装完成
func (c *MetricsCollector) RecordInstallComplete() {
	atomic.AddInt64(&c.installActive, -1)
}

// RecordInstallQueueSize 记录安装队列大小
func (c *MetricsCollector) RecordInstallQueueSize(size int) {
	atomic.StoreInt64(&c.installQueue, int64(size))
}

// JS Pool 相关方法

// SetJSPoolSize 设置 JS 池大小
func (c *MetricsCollector) SetJSPoolSize(size int) {
	atomic.StoreInt64(&c.jsPoolSize, int64(size))
}

// SetJSPoolActive 设置 JS 池活跃数
func (c *MetricsCollector) SetJSPoolActive(active int) {
	atomic.StoreInt64(&c.jsPoolActive, int64(active))
}

// RecordJSCacheHit 记录 JS 缓存命中
func (c *MetricsCollector) RecordJSCacheHit() {
	atomic.AddInt64(&c.jsCacheHits, 1)
}

// RecordJSCacheMiss 记录 JS 缓存未命中
func (c *MetricsCollector) RecordJSCacheMiss() {
	atomic.AddInt64(&c.jsCacheMisses, 1)
}

// Cache 相关方法

// RecordAppCacheHit 记录应用缓存命中
func (c *MetricsCollector) RecordAppCacheHit() {
	atomic.AddInt64(&c.appCacheHits, 1)
}

// RecordAppCacheMiss 记录应用缓存未命中
func (c *MetricsCollector) RecordAppCacheMiss() {
	atomic.AddInt64(&c.appCacheMisses, 1)
}

// RecordBucketCacheHit 记录 Bucket 缓存命中
func (c *MetricsCollector) RecordBucketCacheHit() {
	atomic.AddInt64(&c.bucketCacheHits, 1)
}

// RecordBucketCacheMiss 记录 Bucket 缓存未命中
func (c *MetricsCollector) RecordBucketCacheMiss() {
	atomic.AddInt64(&c.bucketCacheMisses, 1)
}

// RecordIndexCacheHit 记录索引缓存命中
func (c *MetricsCollector) RecordIndexCacheHit() {
	atomic.AddInt64(&c.indexCacheHits, 1)
}

// RecordIndexCacheMiss 记录索引缓存未命中
func (c *MetricsCollector) RecordIndexCacheMiss() {
	atomic.AddInt64(&c.indexCacheMisses, 1)
}

// RecordCacheEviction 记录缓存淘汰
func (c *MetricsCollector) RecordCacheEviction() {
	atomic.AddInt64(&c.cacheEvictions, 1)
}

// RecordBatchReadHit 记录批量读取命中
func (c *MetricsCollector) RecordBatchReadHit() {
	atomic.AddInt64(&c.batchReadHits, 1)
}

// RecordBatchReadMiss 记录批量读取未命中
func (c *MetricsCollector) RecordBatchReadMiss() {
	atomic.AddInt64(&c.batchReadMisses, 1)
}

// GlobalCollector 全局指标收集器
var GlobalCollector = NewMetricsCollector(5 * time.Second)

// Init 初始化全局收集器
func Init() {
	GlobalCollector.Start()
}

// Shutdown 关闭全局收集器
func Shutdown() {
	GlobalCollector.Stop()
}

// GetMetrics 获取当前指标（便捷函数）
func GetMetrics() PerformanceMetrics {
	return GlobalCollector.GetCurrentMetrics()
}

// GetMetricsHistory 获取指标历史（便捷函数）
func GetMetricsHistory() *MetricsHistory {
	return GlobalCollector.GetHistory()
}
