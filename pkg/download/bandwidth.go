package download

import (
	"sync"
	"sync/atomic"
	"time"
)

// 带宽监控相关常量
const (
	// 默认速度采样窗口大小
	DefaultSpeedWindowSize = 5 * time.Second
	// 默认采样容量
	DefaultSampleCapacity = 100
	// 默认最小并发数
	DefaultMinConcurrency = 1
	// 默认最大并发数
	DefaultMaxConcurrency = 16
	// 默认慢启动阈值
	DefaultSlowStartThreshold = 8.0
	// 默认拥塞窗口初始值
	DefaultInitialCongestionWindow = 1.0
	// 最小慢启动阈值
	MinSlowStartThreshold = 2.0
)

// BandwidthMonitor 带宽监控器 - 按字段大小从大到小排列优化内存布局
// time.Time: 24字节, int64: 8字节, 指针: 8字节
type BandwidthMonitor struct {
	// 24字节字段
	startTime time.Time
	// 8字节字段
	currentSpeed int64
	avgSpeed     int64
	peakSpeed    int64
	totalBytes   int64
	// 指针字段
	window *SpeedWindow
	// 互斥锁
	mu sync.RWMutex
}

// SpeedWindow 速度采样窗口
type SpeedWindow struct {
	samples    []SpeedSample
	windowSize time.Duration
	mu         sync.RWMutex
}

// SpeedSample 速度样本
type SpeedSample struct {
	Timestamp time.Time
	Bytes     int64
}

// NewBandwidthMonitor 创建带宽监控器
func NewBandwidthMonitor() *BandwidthMonitor {
	return &BandwidthMonitor{
		startTime: time.Now(),
		window: &SpeedWindow{
			samples:    make([]SpeedSample, 0, DefaultSampleCapacity),
			windowSize: DefaultSpeedWindowSize,
		},
	}
}

// RecordBytes 记录传输的字节数
func (bm *BandwidthMonitor) RecordBytes(bytes int64) {
	atomic.AddInt64(&bm.totalBytes, bytes)

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// 添加样本
	bm.window.mu.Lock()
	bm.window.samples = append(bm.window.samples, SpeedSample{
		Timestamp: time.Now(),
		Bytes:     bytes,
	})

	// 清理过期样本
	cutoff := time.Now().Add(-bm.window.windowSize)
	newSamples := make([]SpeedSample, 0, len(bm.window.samples))
	for _, sample := range bm.window.samples {
		if sample.Timestamp.After(cutoff) {
			newSamples = append(newSamples, sample)
		}
	}
	bm.window.samples = newSamples
	bm.window.mu.Unlock()

	// 计算当前速度
	bm.calculateSpeed()
}

// calculateSpeed 计算当前速度
func (bm *BandwidthMonitor) calculateSpeed() {
	bm.window.mu.RLock()
	samples := make([]SpeedSample, len(bm.window.samples))
	copy(samples, bm.window.samples)
	bm.window.mu.RUnlock()

	if len(samples) < 2 {
		return
	}

	var totalBytes int64
	var earliest time.Time
	var latest time.Time

	for i, sample := range samples {
		totalBytes += sample.Bytes
		if i == 0 || sample.Timestamp.Before(earliest) {
			earliest = sample.Timestamp
		}
		if i == 0 || sample.Timestamp.After(latest) {
			latest = sample.Timestamp
		}
	}

	duration := latest.Sub(earliest).Seconds()
	if duration <= 0 {
		return
	}

	speed := int64(float64(totalBytes) / duration)
	atomic.StoreInt64(&bm.currentSpeed, speed)

	// 更新峰值速度
	for {
		peak := atomic.LoadInt64(&bm.peakSpeed)
		if speed <= peak || atomic.CompareAndSwapInt64(&bm.peakSpeed, peak, speed) {
			break
		}
	}

	// 更新平均速度
	total := atomic.LoadInt64(&bm.totalBytes)
	elapsed := time.Since(bm.startTime).Seconds()
	if elapsed > 0 {
		atomic.StoreInt64(&bm.avgSpeed, int64(float64(total)/elapsed))
	}
}

// GetCurrentSpeed 获取当前速度
func (bm *BandwidthMonitor) GetCurrentSpeed() int64 {
	return atomic.LoadInt64(&bm.currentSpeed)
}

// GetAverageSpeed 获取平均速度
func (bm *BandwidthMonitor) GetAverageSpeed() int64 {
	return atomic.LoadInt64(&bm.avgSpeed)
}

// GetPeakSpeed 获取峰值速度
func (bm *BandwidthMonitor) GetPeakSpeed() int64 {
	return atomic.LoadInt64(&bm.peakSpeed)
}

// GetTotalBytes 获取总字节数
func (bm *BandwidthMonitor) GetTotalBytes() int64 {
	return atomic.LoadInt64(&bm.totalBytes)
}

// Reset 重置监控器
func (bm *BandwidthMonitor) Reset() {
	atomic.StoreInt64(&bm.currentSpeed, 0)
	atomic.StoreInt64(&bm.avgSpeed, 0)
	atomic.StoreInt64(&bm.peakSpeed, 0)
	atomic.StoreInt64(&bm.totalBytes, 0)

	bm.mu.Lock()
	bm.startTime = time.Now()
	bm.window.mu.Lock()
	bm.window.samples = bm.window.samples[:0]
	bm.window.mu.Unlock()
	bm.mu.Unlock()
}

// CongestionController 拥塞控制器 - 按字段大小从大到小排列优化内存布局
// float64: 8字节, int32: 4字节
type CongestionController struct {
	// 8字节字段 (float64)
	congestionWindow   float64
	slowStartThreshold float64
	// 4字节字段 (int32)
	currentConcurrency int32
	maxConcurrency     int32
	minConcurrency     int32
	// 1字节字段
	state CongestionState
	// 互斥锁
	mu sync.RWMutex
}

// CongestionState 拥塞状态
type CongestionState int

const (
	StateSlowStart CongestionState = iota
	StateCongestionAvoidance
	StateFastRecovery
)

// NewCongestionController 创建拥塞控制器
func NewCongestionController() *CongestionController {
	return &CongestionController{
		currentConcurrency: DefaultMinConcurrency,
		maxConcurrency:     DefaultMaxConcurrency,
		minConcurrency:     DefaultMinConcurrency,
		congestionWindow:   DefaultInitialCongestionWindow,
		slowStartThreshold: DefaultSlowStartThreshold,
		state:              StateSlowStart,
	}
}

// OnSuccess 下载成功回调
func (cc *CongestionController) OnSuccess() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	switch cc.state {
	case StateSlowStart:
		// 慢启动阶段：指数增长
		cc.congestionWindow *= 2
		if cc.congestionWindow >= cc.slowStartThreshold {
			cc.state = StateCongestionAvoidance
		}
	case StateCongestionAvoidance:
		// 拥塞避免阶段：线性增长
		cc.congestionWindow += 1.0 / cc.congestionWindow
	case StateFastRecovery:
		// 快速恢复阶段
		cc.congestionWindow = cc.slowStartThreshold
		cc.state = StateCongestionAvoidance
	}

	cc.updateConcurrency()
}

// OnFailure 下载失败回调
func (cc *CongestionController) OnFailure() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// 发生丢包，降低窗口
	cc.slowStartThreshold = cc.congestionWindow / 2
	if cc.slowStartThreshold < MinSlowStartThreshold {
		cc.slowStartThreshold = MinSlowStartThreshold
	}
	cc.congestionWindow = DefaultInitialCongestionWindow
	cc.state = StateSlowStart

	cc.updateConcurrency()
}

// updateConcurrency 更新并发数
func (cc *CongestionController) updateConcurrency() {
	newConcurrency := int32(cc.congestionWindow)
	if newConcurrency < cc.minConcurrency {
		newConcurrency = cc.minConcurrency
	}
	if newConcurrency > cc.maxConcurrency {
		newConcurrency = cc.maxConcurrency
	}
	atomic.StoreInt32(&cc.currentConcurrency, newConcurrency)
}

// GetCurrentConcurrency 获取当前建议并发数
func (cc *CongestionController) GetCurrentConcurrency() int32 {
	return atomic.LoadInt32(&cc.currentConcurrency)
}

// GetCongestionWindow 获取拥塞窗口大小
func (cc *CongestionController) GetCongestionWindow() float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.congestionWindow
}

// SetMaxConcurrency 设置最大并发数
func (cc *CongestionController) SetMaxConcurrency(max int32) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.maxConcurrency = max
	cc.updateConcurrency()
}
