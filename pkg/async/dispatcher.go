package async

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// DispatcherConfig 调度器配置
type DispatcherConfig struct {
	// I/O 任务并发数 (高并发，因为主要是等待)
	MaxIOWorkers int

	// CPU 任务并发数 (限制并发，避免 CPU 过载)
	MaxCPUWorkers int

	// JS 任务并发数 (VM 池大小)
	MaxJSWorkers int

	// 是否启用动态调整
	EnableAdaptive bool

	// 动态调整间隔
	AdjustmentInterval time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() DispatcherConfig {
	cpuCount := runtime.NumCPU()
	return DispatcherConfig{
		MaxIOWorkers:       16,
		MaxCPUWorkers:      cpuCount,
		MaxJSWorkers:       4,
		EnableAdaptive:     true,
		AdjustmentInterval: 10 * time.Second,
	}
}

// SmartDispatcher 智能任务调度器
type SmartDispatcher struct {
	config DispatcherConfig

	// 分类处理池 - 使用信号量控制并发
	ioSemaphore  chan struct{}
	cpuSemaphore chan struct{}
	jsSemaphore  chan struct{}

	// 任务队列 (按优先级排序)
	taskQueues map[TaskCategory][]Task
	queueMu    sync.RWMutex

	// 运行状态
	running int32
	wg      sync.WaitGroup

	// 统计信息
	stats DispatcherStats

	// 动态调整
	adaptiveStop chan struct{}
}

// DispatcherStats 调度器统计信息
type DispatcherStats struct {
	SubmittedTasks   int64
	CompletedTasks   int64
	FailedTasks      int64
	ActiveIOTasks    int32
	ActiveCPUTasks   int32
	ActiveJSTasks    int32
	ActiveMixedTasks int32
}

// NewSmartDispatcher 创建智能调度器
func NewSmartDispatcher(config DispatcherConfig) *SmartDispatcher {
	d := &SmartDispatcher{
		config:       config,
		ioSemaphore:  make(chan struct{}, config.MaxIOWorkers),
		cpuSemaphore: make(chan struct{}, config.MaxCPUWorkers),
		jsSemaphore:  make(chan struct{}, config.MaxJSWorkers),
		taskQueues:   make(map[TaskCategory][]Task),
		adaptiveStop: make(chan struct{}),
	}

	// 初始化任务队列
	d.taskQueues[CategoryIO] = make([]Task, 0)
	d.taskQueues[CategoryCPU] = make([]Task, 0)
	d.taskQueues[CategoryJS] = make([]Task, 0)
	d.taskQueues[CategoryMixed] = make([]Task, 0)

	return d
}

// Start 启动调度器
func (d *SmartDispatcher) Start() {
	if !atomic.CompareAndSwapInt32(&d.running, 0, 1) {
		return // 已经在运行
	}

	if d.config.EnableAdaptive {
		d.wg.Add(1)
		go d.adaptiveLoop()
	}
}

// Stop 停止调度器
func (d *SmartDispatcher) Stop() {
	if !atomic.CompareAndSwapInt32(&d.running, 1, 0) {
		return // 已经停止
	}

	close(d.adaptiveStop)
	d.wg.Wait()
}

// Dispatch 分发任务
func (d *SmartDispatcher) Dispatch(ctx context.Context, task Task) (*TaskFuture, error) {
	if atomic.LoadInt32(&d.running) == 0 {
		return nil, fmt.Errorf("dispatcher not started")
	}

	profile := task.Profile()
	resultChan := make(chan TaskResult, 1)

	atomic.AddInt64(&d.stats.SubmittedTasks, 1)

	// 根据任务类型选择调度策略
	switch profile.Category {
	case CategoryIO:
		go d.executeIOTask(ctx, task, resultChan)
	case CategoryCPU:
		go d.executeCPUTask(ctx, task, resultChan)
	case CategoryJS:
		go d.executeJSTask(ctx, task, resultChan)
	case CategoryMixed:
		go d.executeMixedTask(ctx, task, resultChan)
	default:
		return nil, fmt.Errorf("unknown task category: %v", profile.Category)
	}

	return &TaskFuture{
		ResultChan: resultChan,
		TaskID:     task.ID(),
	}, nil
}

// DispatchBatch 批量分发任务
func (d *SmartDispatcher) DispatchBatch(ctx context.Context, tasks []Task) ([]*TaskFuture, error) {
	futures := make([]*TaskFuture, len(tasks))
	for i, task := range tasks {
		future, err := d.Dispatch(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to dispatch task %s: %w", task.ID(), err)
		}
		futures[i] = future
	}
	return futures, nil
}

// executeIOTask 执行 I/O 密集型任务
func (d *SmartDispatcher) executeIOTask(ctx context.Context, task Task, resultChan chan<- TaskResult) {
	start := time.Now()
	atomic.AddInt32(&d.stats.ActiveIOTasks, 1)
	defer atomic.AddInt32(&d.stats.ActiveIOTasks, -1)

	// 获取 I/O 信号量
	select {
	case d.ioSemaphore <- struct{}{}:
		defer func() { <-d.ioSemaphore }()
	case <-ctx.Done():
		resultChan <- TaskResult{TaskID: task.ID(), Error: ctx.Err(), Duration: time.Since(start)}
		return
	}

	// 执行任务
	err := task.Execute(ctx)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&d.stats.FailedTasks, 1)
	} else {
		atomic.AddInt64(&d.stats.CompletedTasks, 1)
	}

	resultChan <- TaskResult{TaskID: task.ID(), Error: err, Duration: duration}
}

// executeCPUTask 执行 CPU 密集型任务
func (d *SmartDispatcher) executeCPUTask(ctx context.Context, task Task, resultChan chan<- TaskResult) {
	start := time.Now()
	atomic.AddInt32(&d.stats.ActiveCPUTasks, 1)
	defer atomic.AddInt32(&d.stats.ActiveCPUTasks, -1)

	// 获取 CPU 信号量
	select {
	case d.cpuSemaphore <- struct{}{}:
		defer func() { <-d.cpuSemaphore }()
	case <-ctx.Done():
		resultChan <- TaskResult{TaskID: task.ID(), Error: ctx.Err(), Duration: time.Since(start)}
		return
	}

	// 执行任务
	err := task.Execute(ctx)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&d.stats.FailedTasks, 1)
	} else {
		atomic.AddInt64(&d.stats.CompletedTasks, 1)
	}

	resultChan <- TaskResult{TaskID: task.ID(), Error: err, Duration: duration}
}

// executeJSTask 执行 JS 任务
func (d *SmartDispatcher) executeJSTask(ctx context.Context, task Task, resultChan chan<- TaskResult) {
	start := time.Now()
	atomic.AddInt32(&d.stats.ActiveJSTasks, 1)
	defer atomic.AddInt32(&d.stats.ActiveJSTasks, -1)

	// 获取 JS 信号量
	select {
	case d.jsSemaphore <- struct{}{}:
		defer func() { <-d.jsSemaphore }()
	case <-ctx.Done():
		resultChan <- TaskResult{TaskID: task.ID(), Error: ctx.Err(), Duration: time.Since(start)}
		return
	}

	// 执行任务
	err := task.Execute(ctx)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&d.stats.FailedTasks, 1)
	} else {
		atomic.AddInt64(&d.stats.CompletedTasks, 1)
	}

	resultChan <- TaskResult{TaskID: task.ID(), Error: err, Duration: duration}
}

// executeMixedTask 执行混合型任务
func (d *SmartDispatcher) executeMixedTask(ctx context.Context, task Task, resultChan chan<- TaskResult) {
	start := time.Now()
	atomic.AddInt32(&d.stats.ActiveMixedTasks, 1)
	defer atomic.AddInt32(&d.stats.ActiveMixedTasks, -1)

	// 混合型任务需要同时获取 I/O 和 CPU 信号量
	// 先获取 CPU (限制更严格)
	select {
	case d.cpuSemaphore <- struct{}{}:
		defer func() { <-d.cpuSemaphore }()
	case <-ctx.Done():
		resultChan <- TaskResult{TaskID: task.ID(), Error: ctx.Err(), Duration: time.Since(start)}
		return
	}

	// 再获取 I/O
	select {
	case d.ioSemaphore <- struct{}{}:
		defer func() { <-d.ioSemaphore }()
	case <-ctx.Done():
		resultChan <- TaskResult{TaskID: task.ID(), Error: ctx.Err(), Duration: time.Since(start)}
		return
	}

	// 执行任务
	err := task.Execute(ctx)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&d.stats.FailedTasks, 1)
	} else {
		atomic.AddInt64(&d.stats.CompletedTasks, 1)
	}

	resultChan <- TaskResult{TaskID: task.ID(), Error: err, Duration: duration}
}

// adaptiveLoop 动态调整循环
func (d *SmartDispatcher) adaptiveLoop() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.config.AdjustmentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.adjustConcurrency()
		case <-d.adaptiveStop:
			return
		}
	}
}

// adjustConcurrency 动态调整并发数
func (d *SmartDispatcher) adjustConcurrency() {
	// 获取当前系统负载
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 简单的自适应逻辑：根据内存使用情况调整
	// 实际生产环境可以接入更复杂的指标
	memUsage := float64(m.Alloc) / float64(m.Sys)

	if memUsage > 0.8 {
		// 内存使用率过高，降低并发
		d.reduceConcurrency()
	} else if memUsage < 0.5 {
		// 内存使用率较低，可以适当提高并发
		d.increaseConcurrency()
	}
}

// reduceConcurrency 降低并发数
func (d *SmartDispatcher) reduceConcurrency() {
	// 通过减小信号量容量来实现
	// 注意：这里只是示例，实际实现需要更复杂的逻辑
}

// increaseConcurrency 增加并发数
func (d *SmartDispatcher) increaseConcurrency() {
	// 通过增加信号量容量来实现
	// 注意：这里只是示例，实际实现需要更复杂的逻辑
}

// GetStats 获取统计信息
func (d *SmartDispatcher) GetStats() DispatcherStats {
	return DispatcherStats{
		SubmittedTasks:   atomic.LoadInt64(&d.stats.SubmittedTasks),
		CompletedTasks:   atomic.LoadInt64(&d.stats.CompletedTasks),
		FailedTasks:      atomic.LoadInt64(&d.stats.FailedTasks),
		ActiveIOTasks:    atomic.LoadInt32(&d.stats.ActiveIOTasks),
		ActiveCPUTasks:   atomic.LoadInt32(&d.stats.ActiveCPUTasks),
		ActiveJSTasks:    atomic.LoadInt32(&d.stats.ActiveJSTasks),
		ActiveMixedTasks: atomic.LoadInt32(&d.stats.ActiveMixedTasks),
	}
}

// Wait 等待所有任务完成
func (d *SmartDispatcher) Wait() {
	// 等待所有活跃任务完成
	for {
		stats := d.GetStats()
		active := stats.ActiveIOTasks + stats.ActiveCPUTasks + stats.ActiveJSTasks + stats.ActiveMixedTasks
		if active == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// WaitWithTimeout 带超时的等待
func (d *SmartDispatcher) WaitWithTimeout(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		d.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
