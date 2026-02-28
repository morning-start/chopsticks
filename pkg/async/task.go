package async

import (
	"context"
	"time"
)

// TaskCategory 任务分类
type TaskCategory int

const (
	// CategoryIO I/O 密集型 - 可以大量并发，主要等待网络/磁盘
	CategoryIO TaskCategory = iota

	// CategoryCPU CPU 密集型 - 需要限制并发，避免 CPU 过载
	CategoryCPU

	// CategoryJS JS 执行型 - 使用 VM 池，单 VM 单线程但多 VM 并行
	CategoryJS

	// CategoryMixed 混合型 - 包含 I/O 和计算，动态调整
	CategoryMixed
)

// String 返回任务分类的字符串表示
func (c TaskCategory) String() string {
	switch c {
	case CategoryIO:
		return "IO"
	case CategoryCPU:
		return "CPU"
	case CategoryJS:
		return "JS"
	case CategoryMixed:
		return "Mixed"
	default:
		return "Unknown"
	}
}

// ResourceReq 资源需求
type ResourceReq struct {
	CPU     float64 // CPU 核心数需求
	Memory  int64   // 内存需求 (MB)
	Disk    int64   // 磁盘 I/O 需求
	Network bool    // 是否需要网络
}

// TaskProfile 任务画像
type TaskProfile struct {
	Category    TaskCategory  // 任务分类
	Priority    int           // 优先级 (1-10, 10 最高)
	EstDuration time.Duration // 预估执行时间
	Resources   ResourceReq   // 资源需求
}

// Task 异步任务接口
type Task interface {
	// Execute 执行任务
	Execute(ctx context.Context) error

	// Profile 返回任务画像
	Profile() TaskProfile

	// ID 返回任务唯一标识
	ID() string
}

// TaskFunc 函数类型的任务
type TaskFunc struct {
	id      string
	fn      func(ctx context.Context) error
	profile TaskProfile
}

// NewTaskFunc 创建函数类型的任务
func NewTaskFunc(id string, profile TaskProfile, fn func(ctx context.Context) error) *TaskFunc {
	return &TaskFunc{
		id:      id,
		fn:      fn,
		profile: profile,
	}
}

// Execute 执行任务
func (t *TaskFunc) Execute(ctx context.Context) error {
	return t.fn(ctx)
}

// Profile 返回任务画像
func (t *TaskFunc) Profile() TaskProfile {
	return t.profile
}

// ID 返回任务唯一标识
func (t *TaskFunc) ID() string {
	return t.id
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID   string
	Error    error
	Duration time.Duration
}

// TaskFuture 任务未来结果
type TaskFuture struct {
	ResultChan <-chan TaskResult
	TaskID     string
}

// Wait 等待任务完成并返回结果
func (f *TaskFuture) Wait() TaskResult {
	return <-f.ResultChan
}

// WaitTimeout 带超时的等待
func (f *TaskFuture) WaitTimeout(timeout time.Duration) (TaskResult, bool) {
	select {
	case result := <-f.ResultChan:
		return result, true
	case <-time.After(timeout):
		return TaskResult{TaskID: f.TaskID, Error: context.DeadlineExceeded}, false
	}
}

// SimpleTaskProfile 创建简单的任务画像
func SimpleTaskProfile(category TaskCategory, priority int) TaskProfile {
	return TaskProfile{
		Category: category,
		Priority: priority,
		Resources: ResourceReq{
			CPU:     1.0,
			Memory:  64,
			Network: category == CategoryIO || category == CategoryMixed,
		},
	}
}

// IOTaskProfile 创建 I/O 密集型任务画像
func IOTaskProfile(priority int) TaskProfile {
	return TaskProfile{
		Category: CategoryIO,
		Priority: priority,
		Resources: ResourceReq{
			CPU:     0.5,
			Memory:  32,
			Network: true,
		},
	}
}

// CPUTaskProfile 创建 CPU 密集型任务画像
func CPUTaskProfile(priority int) TaskProfile {
	return TaskProfile{
		Category: CategoryCPU,
		Priority: priority,
		Resources: ResourceReq{
			CPU:     2.0,
			Memory:  128,
			Network: false,
		},
	}
}

// JSTaskProfile 创建 JS 执行任务画像
func JSTaskProfile(priority int) TaskProfile {
	return TaskProfile{
		Category: CategoryJS,
		Priority: priority,
		Resources: ResourceReq{
			CPU:     1.0,
			Memory:  256,
			Network: false,
		},
	}
}

// MixedTaskProfile 创建混合型任务画像
func MixedTaskProfile(priority int) TaskProfile {
	return TaskProfile{
		Category: CategoryMixed,
		Priority: priority,
		Resources: ResourceReq{
			CPU:     1.5,
			Memory:  128,
			Network: true,
		},
	}
}
