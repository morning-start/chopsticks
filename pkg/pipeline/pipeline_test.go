package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestNewPipeline 测试创建流水线
func TestNewPipeline(t *testing.T) {
	p := NewPipeline(nil)

	if p == nil {
		t.Fatal("NewPipeline returned nil")
	}

	if p.bufferSize != 10 {
		t.Errorf("bufferSize = %d, want 10", p.bufferSize)
	}

	if p.errorPolicy != StopOnError {
		t.Errorf("errorPolicy = %v, want StopOnError", p.errorPolicy)
	}

	if len(p.stages) != 0 {
		t.Errorf("stages length = %d, want 0", len(p.stages))
	}
}

// TestNewPipelineWithConfig 测试使用配置创建流水线
func TestNewPipelineWithConfig(t *testing.T) {
	config := &Config{
		BufferSize:  20,
		ErrorPolicy: ContinueOnError,
	}

	p := NewPipeline(config)

	if p.bufferSize != 20 {
		t.Errorf("bufferSize = %d, want 20", p.bufferSize)
	}

	if p.errorPolicy != ContinueOnError {
		t.Errorf("errorPolicy = %v, want ContinueOnError", p.errorPolicy)
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BufferSize != 10 {
		t.Errorf("BufferSize = %d, want 10", config.BufferSize)
	}

	if config.ErrorPolicy != StopOnError {
		t.Errorf("ErrorPolicy = %v, want StopOnError", config.ErrorPolicy)
	}
}

// TestPipeline_AddStage 测试添加阶段
func TestPipeline_AddStage(t *testing.T) {
	p := NewPipeline(nil)

	stage := &FuncStage{
		Name_:       "test-stage",
		Handler:     func(ctx context.Context, item Item) (Item, error) { return item, nil },
		Parallelism: 1,
	}

	p.AddStage(stage)

	if len(p.stages) != 1 {
		t.Errorf("stages length = %d, want 1", len(p.stages))
	}
}

// TestPipeline_AddStageFunc 测试添加函数阶段
func TestPipeline_AddStageFunc(t *testing.T) {
	p := NewPipeline(nil)

	fn := func(ctx context.Context, item Item) (Item, error) {
		return item, nil
	}

	p.AddStageFunc("func-stage", fn, 2)

	if len(p.stages) != 1 {
		t.Errorf("stages length = %d, want 1", len(p.stages))
	}
}

// TestPipeline_Run 测试执行流水线
func TestPipeline_Run(t *testing.T) {
	p := NewPipeline(nil)

	// 添加一个简单的阶段
	p.AddStageFunc("double", func(ctx context.Context, item Item) (Item, error) {
		if num, ok := item.Data.(int); ok {
			item.Data = num * 2
		}
		return item, nil
	}, 1)

	// 执行流水线
	input := []interface{}{1, 2, 3, 4, 5}
	results, err := p.Run(context.Background(), input)

	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("results length = %d, want 5", len(results))
	}

	// 验证结果
	expected := []int{2, 4, 6, 8, 10}
	for i, item := range results {
		if num, ok := item.Data.(int); !ok || num != expected[i] {
			t.Errorf("result[%d] = %v, want %d", i, item.Data, expected[i])
		}
	}
}

// TestPipeline_Run_EmptyStages 测试空阶段流水线
func TestPipeline_Run_EmptyStages(t *testing.T) {
	p := NewPipeline(nil)

	input := []interface{}{1, 2, 3}
	_, err := p.Run(context.Background(), input)

	if err == nil {
		t.Error("Expected error for empty stages")
	}
}

// TestPipeline_Run_WithError 测试错误处理
func TestPipeline_Run_WithError(t *testing.T) {
	p := NewPipeline(nil)

	// 添加一个会出错的阶段
	p.AddStageFunc("error-stage", func(ctx context.Context, item Item) (Item, error) {
		if num, ok := item.Data.(int); ok && num == 3 {
			return item, errors.New("test error")
		}
		return item, nil
	}, 1)

	input := []interface{}{1, 2, 3, 4, 5}
	results, err := p.Run(context.Background(), input)

	// Run 方法本身不返回错误，错误在结果中
	if err != nil {
		t.Errorf("Unexpected error from Run: %v", err)
	}

	// 验证结果中有一个错误
	hasError := false
	for _, item := range results {
		if item.Error != nil {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("Expected at least one error in results")
	}
}

// TestPipeline_RunAsync 测试异步执行
func TestPipeline_RunAsync(t *testing.T) {
	p := NewPipeline(nil)

	p.AddStageFunc("delay", func(ctx context.Context, item Item) (Item, error) {
		time.Sleep(10 * time.Millisecond)
		return item, nil
	}, 1)

	input := []interface{}{1, 2, 3}
	outputChan := p.RunAsync(context.Background(), input)

	var results []Item
	for item := range outputChan {
		results = append(results, item)
	}

	if len(results) != 3 {
		t.Errorf("results length = %d, want 3", len(results))
	}
}

// TestPipeline_Cancel 测试取消流水线
func TestPipeline_Cancel(t *testing.T) {
	p := NewPipeline(nil)

	p.AddStageFunc("slow", func(ctx context.Context, item Item) (Item, error) {
		time.Sleep(100 * time.Millisecond)
		return item, nil
	}, 1)

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 异步执行
	input := []interface{}{1, 2, 3, 4, 5}
	go p.Run(ctx, input)

	// 立即取消
	time.Sleep(10 * time.Millisecond)
	cancel()

	// 验证上下文已取消
	select {
	case <-ctx.Done():
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled")
	}
}

// TestPipeline_GetStats 测试获取统计信息
func TestPipeline_GetStats(t *testing.T) {
	p := NewPipeline(nil)
	p.AddStageFunc("stage1", func(ctx context.Context, item Item) (Item, error) {
		return item, nil
	}, 1)
	p.AddStageFunc("stage2", func(ctx context.Context, item Item) (Item, error) {
		return item, nil
	}, 1)

	stats := p.GetStats()

	if stats.StageCount != 2 {
		t.Errorf("StageCount = %d, want 2", stats.StageCount)
	}

	if stats.BufferSize != 10 {
		t.Errorf("BufferSize = %d, want 10", stats.BufferSize)
	}

	if stats.ErrorPolicy != "stop_on_error" {
		t.Errorf("ErrorPolicy = %s, want stop_on_error", stats.ErrorPolicy)
	}
}

// TestPipelineBuilder 测试流水线构建器
func TestPipelineBuilder(t *testing.T) {
	builder := NewPipelineBuilder()

	pipeline := builder.
		WithBufferSize(20).
		WithErrorPolicy(ContinueOnError).
		WithStageFunc("stage1", func(ctx context.Context, item Item) (Item, error) {
			return item, nil
		}, 2).
		WithStageFunc("stage2", func(ctx context.Context, item Item) (Item, error) {
			return item, nil
		}, 2).
		Build()

	if pipeline.bufferSize != 20 {
		t.Errorf("bufferSize = %d, want 20", pipeline.bufferSize)
	}

	if pipeline.errorPolicy != ContinueOnError {
		t.Errorf("errorPolicy = %v, want ContinueOnError", pipeline.errorPolicy)
	}

	if len(pipeline.stages) != 2 {
		t.Errorf("stages length = %d, want 2", len(pipeline.stages))
	}
}

// TestFuncStage_ProcessSequential 测试顺序处理
func TestFuncStage_ProcessSequential(t *testing.T) {
	stage := &FuncStage{
		Name_:       "test",
		Parallelism: 1,
		Handler: func(ctx context.Context, item Item) (Item, error) {
			if num, ok := item.Data.(int); ok {
				item.Data = num + 1
			}
			return item, nil
		},
	}

	input := make(chan Item, 3)
	input <- Item{Data: 1}
	input <- Item{Data: 2}
	input <- Item{Data: 3}
	close(input)

	output := make(chan Item, 3)
	go func() {
		stage.Process(context.Background(), input, output)
		close(output)
	}()

	var results []int
	for item := range output {
		if num, ok := item.Data.(int); ok {
			results = append(results, num)
		}
	}

	expected := []int{2, 3, 4}
	for i, v := range expected {
		if results[i] != v {
			t.Errorf("result[%d] = %d, want %d", i, results[i], v)
		}
	}
}

// TestFuncStage_ProcessParallel 测试并行处理
func TestFuncStage_ProcessParallel(t *testing.T) {
	stage := &FuncStage{
		Name_:       "test",
		Parallelism: 3,
		Handler: func(ctx context.Context, item Item) (Item, error) {
			time.Sleep(10 * time.Millisecond)
			if num, ok := item.Data.(int); ok {
				item.Data = num * 2
			}
			return item, nil
		},
	}

	input := make(chan Item, 5)
	for i := 1; i <= 5; i++ {
		input <- Item{Data: i}
	}
	close(input)

	output := make(chan Item, 5)
	go func() {
		stage.Process(context.Background(), input, output)
		close(output)
	}()

	var results []int
	for item := range output {
		if num, ok := item.Data.(int); ok {
			results = append(results, num)
		}
	}

	if len(results) != 5 {
		t.Errorf("results length = %d, want 5", len(results))
	}
}

// TestFilterStage 测试过滤阶段
func TestFilterStage(t *testing.T) {
	stage := &FilterStage{
		Name_: "filter-even",
		Filter: func(item Item) bool {
			if num, ok := item.Data.(int); ok {
				return num%2 == 0
			}
			return false
		},
	}

	input := make(chan Item, 5)
	for i := 1; i <= 5; i++ {
		input <- Item{Data: i}
	}
	close(input)

	output := make(chan Item, 5)
	go func() {
		stage.Process(context.Background(), input, output)
		close(output)
	}()

	var results []int
	for item := range output {
		if num, ok := item.Data.(int); ok {
			results = append(results, num)
		}
	}

	expected := []int{2, 4}
	if len(results) != len(expected) {
		t.Errorf("results length = %d, want %d", len(results), len(expected))
	}
}

// TestMapStage 测试映射阶段
func TestMapStage(t *testing.T) {
	stage := &MapStage{
		Name_: "square",
		Mapper: func(item Item) Item {
			if num, ok := item.Data.(int); ok {
				item.Data = num * num
			}
			return item
		},
	}

	input := make(chan Item, 3)
	input <- Item{Data: 2}
	input <- Item{Data: 3}
	input <- Item{Data: 4}
	close(input)

	output := make(chan Item, 3)
	go func() {
		stage.Process(context.Background(), input, output)
		close(output)
	}()

	var results []int
	for item := range output {
		if num, ok := item.Data.(int); ok {
			results = append(results, num)
		}
	}

	expected := []int{4, 9, 16}
	for i, v := range expected {
		if results[i] != v {
			t.Errorf("result[%d] = %d, want %d", i, results[i], v)
		}
	}
}

// TestChainStages 测试阶段链
func TestChainStages(t *testing.T) {
	stage1 := &MapStage{
		Name_: "add-one",
		Mapper: func(item Item) Item {
			if num, ok := item.Data.(int); ok {
				item.Data = num + 1
			}
			return item
		},
	}

	stage2 := &MapStage{
		Name_: "double",
		Mapper: func(item Item) Item {
			if num, ok := item.Data.(int); ok {
				item.Data = num * 2
			}
			return item
		},
	}

	chained := ChainStages(stage1, stage2)

	input := make(chan Item, 3)
	input <- Item{Data: 1}
	input <- Item{Data: 2}
	input <- Item{Data: 3}
	close(input)

	output := make(chan Item, 3)
	go func() {
		chained.Process(context.Background(), input, output)
		close(output)
	}()

	var results []int
	for item := range output {
		if num, ok := item.Data.(int); ok {
			results = append(results, num)
		}
	}

	// (1+1)*2=4, (2+1)*2=6, (3+1)*2=8
	expected := []int{4, 6, 8}
	for i, v := range expected {
		if results[i] != v {
			t.Errorf("result[%d] = %d, want %d", i, results[i], v)
		}
	}
}

// BenchmarkPipeline 基准测试流水线
func BenchmarkPipeline(b *testing.B) {
	p := NewPipeline(nil)

	p.AddStageFunc("increment", func(ctx context.Context, item Item) (Item, error) {
		if num, ok := item.Data.(int); ok {
			item.Data = num + 1
		}
		return item, nil
	}, 1)

	input := make([]interface{}, 100)
	for i := range input {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Run(context.Background(), input)
	}
}

// BenchmarkPipelineParallel 基准测试并行流水线
func BenchmarkPipelineParallel(b *testing.B) {
	p := NewPipeline(nil)

	p.AddStageFunc("increment", func(ctx context.Context, item Item) (Item, error) {
		if num, ok := item.Data.(int); ok {
			item.Data = num + 1
		}
		return item, nil
	}, 4)

	input := make([]interface{}, 100)
	for i := range input {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Run(context.Background(), input)
	}
}
