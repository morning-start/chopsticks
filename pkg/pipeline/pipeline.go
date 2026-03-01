package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// 常量定义
const (
	// DefaultBufferSize 默认通道缓冲区大小
	DefaultBufferSize = 10
	// DefaultTimeout 默认超时时间（秒）
	DefaultTimeout = 30
	// DefaultParallelism 默认工作线程数
	DefaultParallelism = 1
)

// 预定义错误变量
var (
	// ErrPipelineClosed 流水线已关闭
	ErrPipelineClosed = errors.New("pipeline is closed")
	// ErrStageFailed 阶段执行失败
	ErrStageFailed = errors.New("stage execution failed")
	// ErrNoStages 流水线没有阶段
	ErrNoStages = errors.New("pipeline has no stages")
	// ErrMismatchedInputCount 输入数量与流水线数量不匹配
	ErrMismatchedInputCount = errors.New("pipeline count does not match input count")
)

// Item 流水线数据项
type Item struct {
	Data  interface{}
	Error error
	Index int
}

// Stage 流水线阶段接口
type Stage interface {
	Name() string
	Process(ctx context.Context, input <-chan Item, output chan<- Item) error
}

// StageFunc 阶段处理函数类型
type StageFunc func(ctx context.Context, item Item) (Item, error)

// Pipeline 流水线
type Pipeline struct {
	mu          sync.Mutex
	stages      []Stage
	bufferSize  int
	errorPolicy ErrorPolicy
}

// ErrorPolicy 错误处理策略
type ErrorPolicy int

const (
	// StopOnError 遇到错误立即停止
	StopOnError ErrorPolicy = iota
	// ContinueOnError 遇到错误继续处理其他数据
	ContinueOnError
	// SkipOnError 遇到错误跳过当前数据
	SkipOnError
)

// Config 流水线配置
type Config struct {
	BufferSize  int
	ErrorPolicy ErrorPolicy
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BufferSize:  DefaultBufferSize,
		ErrorPolicy: StopOnError,
	}
}

// NewPipeline 创建流水线
func NewPipeline(config *Config) *Pipeline {
	if config == nil {
		config = DefaultConfig()
	}

	return &Pipeline{
		stages:      make([]Stage, 0),
		bufferSize:  config.BufferSize,
		errorPolicy: config.ErrorPolicy,
	}
}

// AddStage 添加阶段
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
	p.stages = append(p.stages, stage)
	return p
}

// AddStageFunc 添加函数阶段
func (p *Pipeline) AddStageFunc(name string, fn StageFunc, parallelism int) *Pipeline {
	stage := &FuncStage{
		Name_:       name,
		Handler:     fn,
		Parallelism: parallelism,
		BufferSize:  p.bufferSize,
	}
	p.stages = append(p.stages, stage)
	return p
}

// Run 执行流水线
func (p *Pipeline) Run(ctx context.Context, input []interface{}) ([]Item, error) {
	if len(p.stages) == 0 {
		return nil, fmt.Errorf("%w", ErrNoStages)
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建输入通道
	inputChan := make(chan Item, p.bufferSize)
	go func() {
		defer close(inputChan)
		for i, data := range input {
			select {
			case <-ctx.Done():
				return
			case inputChan <- Item{Data: data, Index: i}:
			}
		}
	}()

	// 连接所有阶段
	var currentChan <-chan Item = inputChan
	var wg sync.WaitGroup

	for _, stage := range p.stages {
		outputChan := make(chan Item, p.bufferSize)
		wg.Add(1)

		go func(s Stage, in <-chan Item, out chan<- Item) {
			defer wg.Done()
			defer close(out)

			if err := s.Process(ctx, in, out); err != nil {
				// 根据错误策略处理
				if p.errorPolicy == StopOnError {
					cancel()
				}
			}
		}(stage, currentChan, outputChan)

		currentChan = outputChan
	}

	// 收集结果
	var results []Item
	for item := range currentChan {
		results = append(results, item)
		if item.Error != nil && p.errorPolicy == StopOnError {
			cancel()
			break
		}
	}

	// 等待所有阶段完成
	wg.Wait()

	return results, nil
}

// RunAsync 异步执行流水线
func (p *Pipeline) RunAsync(ctx context.Context, input []interface{}) <-chan Item {
	outputChan := make(chan Item, p.bufferSize)

	go func() {
		defer close(outputChan)

		results, err := p.Run(ctx, input)
		if err != nil {
			outputChan <- Item{Error: err}
			return
		}

		for _, item := range results {
			select {
			case <-ctx.Done():
				return
			case outputChan <- item:
			}
		}
	}()

	return outputChan
}

// SetErrorPolicy 设置错误处理策略
func (p *Pipeline) SetErrorPolicy(policy ErrorPolicy) *Pipeline {
	p.errorPolicy = policy
	return p
}

// Stats 流水线统计
type Stats struct {
	StageCount  int
	BufferSize  int
	ErrorPolicy string
}

// GetStats 获取流水线统计
func (p *Pipeline) GetStats() Stats {
	policyStr := "unknown"
	switch p.errorPolicy {
	case StopOnError:
		policyStr = "stop_on_error"
	case ContinueOnError:
		policyStr = "continue_on_error"
	case SkipOnError:
		policyStr = "skip_on_error"
	}

	return Stats{
		StageCount:  len(p.stages),
		BufferSize:  p.bufferSize,
		ErrorPolicy: policyStr,
	}
}

// PipelineBuilder 流水线构建器
type PipelineBuilder struct {
	config *Config
	stages []Stage
}

// NewPipelineBuilder 创建流水线构建器
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		config: DefaultConfig(),
		stages: make([]Stage, 0),
	}
}

// WithBufferSize 设置缓冲区大小
func (b *PipelineBuilder) WithBufferSize(size int) *PipelineBuilder {
	b.config.BufferSize = size
	return b
}

// WithErrorPolicy 设置错误策略
func (b *PipelineBuilder) WithErrorPolicy(policy ErrorPolicy) *PipelineBuilder {
	b.config.ErrorPolicy = policy
	return b
}

// WithStage 添加阶段
func (b *PipelineBuilder) WithStage(stage Stage) *PipelineBuilder {
	b.stages = append(b.stages, stage)
	return b
}

// WithStageFunc 添加函数阶段
func (b *PipelineBuilder) WithStageFunc(name string, fn StageFunc, parallelism int) *PipelineBuilder {
	stage := &FuncStage{
		Name_:       name,
		Handler:     fn,
		Parallelism: parallelism,
		BufferSize:  b.config.BufferSize,
	}
	b.stages = append(b.stages, stage)
	return b
}

// Build 构建流水线
func (b *PipelineBuilder) Build() *Pipeline {
	pipeline := NewPipeline(b.config)
	pipeline.stages = b.stages
	return pipeline
}

// ParallelPipeline 并行流水线
type ParallelPipeline struct {
	pipelines []*Pipeline
	merger    func([]Item) Item
}

// NewParallelPipeline 创建并行流水线
func NewParallelPipeline(merger func([]Item) Item) *ParallelPipeline {
	return &ParallelPipeline{
		pipelines: make([]*Pipeline, 0),
		merger:    merger,
	}
}

// AddPipeline 添加子流水线
func (pp *ParallelPipeline) AddPipeline(pipeline *Pipeline) *ParallelPipeline {
	pp.pipelines = append(pp.pipelines, pipeline)
	return pp
}

// Run 执行并行流水线
func (pp *ParallelPipeline) Run(ctx context.Context, inputs [][]interface{}) ([]Item, error) {
	if len(pp.pipelines) != len(inputs) {
		return nil, fmt.Errorf("%w: pipeline count (%d) does not match input count (%d)",
			ErrMismatchedInputCount, len(pp.pipelines), len(inputs))
	}

	var wg sync.WaitGroup
	results := make([][]Item, len(pp.pipelines))
	errChan := make(chan error, len(pp.pipelines))

	for i, pipeline := range pp.pipelines {
		wg.Add(1)
		go func(idx int, p *Pipeline, input []interface{}) {
			defer wg.Done()
			var err error
			results[idx], err = p.Run(ctx, input)
			if err != nil {
				errChan <- err
			}
		}(i, pipeline, inputs[i])
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// 合并结果
	var merged []Item
	maxLen := 0
	for _, r := range results {
		if len(r) > maxLen {
			maxLen = len(r)
		}
	}

	for i := 0; i < maxLen; i++ {
		items := make([]Item, 0, len(results))
		for _, r := range results {
			if i < len(r) {
				items = append(items, r[i])
			}
		}
		if pp.merger != nil {
			merged = append(merged, pp.merger(items))
			continue
		}
		// 默认合并：取第一个
		if len(items) > 0 {
			merged = append(merged, items[0])
		}
	}

	return merged, nil
}
