package app

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/manifest"
	"chopsticks/pkg/async"
)

// PipelineStage 流水线阶段
type PipelineStage int

const (
	PipelineDownload PipelineStage = iota
	PipelineVerify
	PipelineExtract
	PipelineExecute
	PipelineRegister
	PipelineComplete
)

func (s PipelineStage) String() string {
	switch s {
	case PipelineDownload:
		return "Download"
	case PipelineVerify:
		return "Verify"
	case PipelineExtract:
		return "Extract"
	case PipelineExecute:
		return "Execute Script"
	case PipelineRegister:
		return "Register"
	case PipelineComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// StageTask 阶段任务
type StageTask struct {
	App    *manifest.App
	Stage  PipelineStage
	Input  interface{}
	Output interface{}
	Error  error
	Index  int
}

// PipelineInstaller 流水线安装器
type PipelineInstaller struct {
	dispatcher *async.SmartDispatcher
	stages     map[PipelineStage]StageProcessor
	bufferSize int
}

// StageProcessor 阶段处理器
type StageProcessor interface {
	Process(ctx context.Context, task *StageTask) (*StageTask, error)
	GetCategory() async.TaskCategory
}

// StageFunc 阶段处理函数
type StageFunc func(ctx context.Context, task *StageTask) (*StageTask, error)

// funcStage 函数阶段处理器
type funcStage struct {
	fn       StageFunc
	category async.TaskCategory
}

func (s *funcStage) Process(ctx context.Context, task *StageTask) (*StageTask, error) {
	return s.fn(ctx, task)
}

func (s *funcStage) GetCategory() async.TaskCategory {
	return s.category
}

// NewPipelineInstaller 创建流水线安装器
func NewPipelineInstaller(dispatcher *async.SmartDispatcher) *PipelineInstaller {
	return &PipelineInstaller{
		dispatcher: dispatcher,
		stages:     make(map[PipelineStage]StageProcessor),
		bufferSize: 10,
	}
}

// RegisterStage 注册阶段处理器
func (p *PipelineInstaller) RegisterStage(stage PipelineStage, processor StageProcessor) {
	p.stages[stage] = processor
}

// RegisterStageFunc 注册阶段处理函数
func (p *PipelineInstaller) RegisterStageFunc(stage PipelineStage, fn StageFunc, category async.TaskCategory) {
	p.stages[stage] = &funcStage{fn: fn, category: category}
}

// InstallWithPipeline 使用流水线安装应用
func (p *PipelineInstaller) InstallWithPipeline(
	ctx context.Context,
	apps []*manifest.App,
	opts PipelineOptions,
) error {
	if len(apps) == 0 {
		return nil
	}

	// Create stage channels
	stages := []PipelineStage{
		PipelineDownload,
		PipelineVerify,
		PipelineExtract,
		PipelineExecute,
		PipelineRegister,
	}

	channels := make([]chan *StageTask, len(stages)+1)
	for i := range channels {
		channels[i] = make(chan *StageTask, p.bufferSize)
	}

	// Start stage processors
	var wg sync.WaitGroup
	for i, stage := range stages {
		wg.Add(1)
		go func(stageIdx int, stage PipelineStage) {
			defer wg.Done()
			p.runStage(ctx, stage, channels[stageIdx], channels[stageIdx+1], opts)
		}(i, stage)
	}

	// Send initial tasks
	go func() {
		for i, app := range apps {
			channels[0] <- &StageTask{
				App:   app,
				Stage: PipelineDownload,
				Index: i,
			}
		}
		close(channels[0])
	}()

	// Collect results
	results := make([]*StageTask, len(apps))
	completed := 0
	failed := 0

	for task := range channels[len(stages)] {
		results[task.Index] = task
		if task.Error != nil {
			failed++
			if opts.OnError != nil {
				opts.OnError(task.App, task.Error)
			}
		} else {
			completed++
			if opts.OnComplete != nil {
				opts.OnComplete(task.App)
			}
		}

		if completed+failed >= len(apps) {
			break
		}
	}

	// Wait for all stages to complete
	wg.Wait()

	if failed > 0 {
		return fmt.Errorf("installation complete: %d succeeded, %d failed", completed, failed)
	}

	return nil
}

// runStage 运行单个阶段
func (p *PipelineInstaller) runStage(
	ctx context.Context,
	stage PipelineStage,
	inputChan <-chan *StageTask,
	outputChan chan<- *StageTask,
	opts PipelineOptions,
) {
	processor, ok := p.stages[stage]
	if !ok {
		// If no processor, pass through
		for task := range inputChan {
			task.Stage = stage + 1
			outputChan <- task
		}
		return
	}

	// Use dispatcher for concurrency control
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.MaxConcurrency)

	for task := range inputChan {
		if task.Error != nil {
			// If previous stage failed, pass through
			outputChan <- task
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(t *StageTask) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if opts.OnStageStart != nil {
				opts.OnStageStart(t.App, stage)
			}

			result, err := processor.Process(ctx, t)
			if err != nil {
				t.Error = err
			} else {
				t.Output = result.Output
			}

			t.Stage = stage + 1

			if opts.OnStageComplete != nil {
				opts.OnStageComplete(t.App, stage, err)
			}

			outputChan <- t
		}(task)
	}

	wg.Wait()
	close(outputChan)
}

// PipelineOptions 流水线选项
type PipelineOptions struct {
	MaxConcurrency  int
	OnStageStart    func(*manifest.App, PipelineStage)
	OnStageComplete func(*manifest.App, PipelineStage, error)
	OnComplete      func(*manifest.App)
	OnError         func(*manifest.App, error)
}

// DefaultPipelineOptions 默认流水线选项
func DefaultPipelineOptions() PipelineOptions {
	return PipelineOptions{
		MaxConcurrency: 4,
	}
}

// StageMetrics 阶段指标
type StageMetrics struct {
	Stage          PipelineStage
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	AvgDuration    float64 // milliseconds
}

// PipelineMetrics 流水线指标
type PipelineMetrics struct {
	Stages     []StageMetrics
	TotalTime  float64 // milliseconds
	Throughput float64 // tasks/second
}

// PipelineBuilder 流水线构建器
type PipelineBuilder struct {
	installer *PipelineInstaller
}

// NewPipelineBuilder 创建流水线构建器
func NewPipelineBuilder(dispatcher *async.SmartDispatcher) *PipelineBuilder {
	return &PipelineBuilder{
		installer: NewPipelineInstaller(dispatcher),
	}
}

// WithDownloadStage 添加下载阶段
func (b *PipelineBuilder) WithDownloadStage(processor StageProcessor) *PipelineBuilder {
	b.installer.RegisterStage(PipelineDownload, processor)
	return b
}

// WithVerifyStage 添加校验阶段
func (b *PipelineBuilder) WithVerifyStage(processor StageProcessor) *PipelineBuilder {
	b.installer.RegisterStage(PipelineVerify, processor)
	return b
}

// WithExtractStage 添加解压阶段
func (b *PipelineBuilder) WithExtractStage(processor StageProcessor) *PipelineBuilder {
	b.installer.RegisterStage(PipelineExtract, processor)
	return b
}

// WithExecuteStage 添加执行阶段
func (b *PipelineBuilder) WithExecuteStage(processor StageProcessor) *PipelineBuilder {
	b.installer.RegisterStage(PipelineExecute, processor)
	return b
}

// WithRegisterStage 添加注册阶段
func (b *PipelineBuilder) WithRegisterStage(processor StageProcessor) *PipelineBuilder {
	b.installer.RegisterStage(PipelineRegister, processor)
	return b
}

// Build 构建流水线安装器
func (b *PipelineBuilder) Build() *PipelineInstaller {
	return b.installer
}

// SimplePipelineInstaller 简单流水线安装器（用于快速创建）
func SimplePipelineInstaller(
	dispatcher *async.SmartDispatcher,
	downloader StageProcessor,
	verifier StageProcessor,
	extractor StageProcessor,
) *PipelineInstaller {
	installer := NewPipelineInstaller(dispatcher)
	installer.RegisterStage(PipelineDownload, downloader)
	installer.RegisterStage(PipelineVerify, verifier)
	installer.RegisterStage(PipelineExtract, extractor)
	return installer
}

// ParallelStage 并行阶段处理器
type ParallelStage struct {
	processor StageProcessor
	parallel  int
}

// NewParallelStage 创建并行阶段
func NewParallelStage(processor StageProcessor, parallel int) *ParallelStage {
	return &ParallelStage{
		processor: processor,
		parallel:  parallel,
	}
}

// Process 处理任务
func (p *ParallelStage) Process(ctx context.Context, tasks []*StageTask) ([]*StageTask, error) {
	results := make([]*StageTask, len(tasks))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, p.parallel)

	for i, task := range tasks {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, t *StageTask) {
			defer wg.Done()
			defer func() { <-semaphore }()

			result, err := p.processor.Process(ctx, t)
			if err != nil {
				t.Error = err
			} else {
				t.Output = result.Output
			}
			results[idx] = t
		}(i, task)
	}

	wg.Wait()
	return results, nil
}
