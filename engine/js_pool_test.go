package engine

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"chopsticks/pkg/async"

	"github.com/dop251/goja"
)

func TestNewJSEnginePool(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	if pool == nil {
		t.Fatal("NewJSEnginePool() returned nil")
	}

	if pool.config.MaxEngines != 4 {
		t.Errorf("MaxEngines = %d, want 4", pool.config.MaxEngines)
	}

	if pool.scriptCache == nil {
		t.Error("scriptCache is nil")
	}

	// 检查预热
	stats := pool.GetStats()
	if stats.TotalEngines != int32(config.PrewarmSize) {
		t.Errorf("TotalEngines = %d, want %d", stats.TotalEngines, config.PrewarmSize)
	}
}

func TestJSEnginePool_AcquireRelease(t *testing.T) {
	config := DefaultPoolConfig()
	config.Prewarm = false // 不预热，手动测试
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	// 获取引擎
	engine, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if engine == nil {
		t.Fatal("Acquire() returned nil engine")
	}

	if engine.id == "" {
		t.Error("Engine ID is empty")
	}

	// 检查统计
	stats := pool.GetStats()
	if stats.ActiveEngines != 1 {
		t.Errorf("ActiveEngines = %d, want 1", stats.ActiveEngines)
	}

	// 归还引擎
	pool.Release(engine)

	// 检查统计
	stats = pool.GetStats()
	if stats.ActiveEngines != 0 {
		t.Errorf("ActiveEngines after release = %d, want 0", stats.ActiveEngines)
	}
}

func TestJSEnginePool_Execute(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	executed := false
	err := pool.Execute(ctx, func(engine *JSEngine) error {
		executed = true
		if engine == nil {
			t.Error("Engine is nil in Execute")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !executed {
		t.Error("Execute function was not called")
	}
}

func TestJSEnginePool_ExecuteWithCache(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	// 创建临时脚本文件
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.js")
	scriptContent := `var result = 1 + 1;`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	ctx := context.Background()

	// 第一次执行（缓存未命中）
	err := pool.ExecuteWithCache(ctx, scriptPath, func(engine *JSEngine, program *goja.Program) error {
		if program == nil {
			t.Error("Compiled program is nil")
		}
		return nil
	})

	if err != nil {
		t.Errorf("ExecuteWithCache() first call error = %v", err)
	}

	// 检查缓存统计
	cacheStats := pool.GetCacheStats()
	if cacheStats.Misses != 1 {
		t.Errorf("Cache misses = %d, want 1", cacheStats.Misses)
	}

	// 第二次执行（缓存命中）
	err = pool.ExecuteWithCache(ctx, scriptPath, func(engine *JSEngine, program *goja.Program) error {
		return nil
	})

	if err != nil {
		t.Errorf("ExecuteWithCache() second call error = %v", err)
	}

	// 检查缓存统计
	cacheStats = pool.GetCacheStats()
	if cacheStats.Hits != 1 {
		t.Errorf("Cache hits = %d, want 1", cacheStats.Hits)
	}
	if cacheStats.HitRate != 0.5 {
		t.Errorf("Cache hit rate = %f, want 0.5", cacheStats.HitRate)
	}
}

func TestJSEnginePool_ExecuteBatch(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	// 创建任务
	tasks := make([]async.Task, 5)
	var counter int32
	for i := 0; i < 5; i++ {
		taskID := string(rune('a' + i))
		tasks[i] = async.NewTaskFunc(
			taskID,
			async.JSTaskProfile(5),
			func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				return nil
			},
		)
	}

	batchConfig := DefaultBatchConfig()
	result, err := pool.ExecuteBatch(ctx, tasks, batchConfig)

	if err != nil {
		t.Errorf("ExecuteBatch() error = %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteBatch() returned nil result")
	}

	if result.Total != 5 {
		t.Errorf("Total = %d, want 5", result.Total)
	}

	if result.Success != 5 {
		t.Errorf("Success = %d, want 5", result.Success)
	}

	if counter != 5 {
		t.Errorf("Counter = %d, want 5", counter)
	}
}

func TestJSEnginePool_ExecuteBatch_WithError(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	// 创建任务，其中一个会失败
	tasks := []async.Task{
		async.NewTaskFunc("task-1", async.JSTaskProfile(5), func(ctx context.Context) error {
			return nil
		}),
		async.NewTaskFunc("task-2", async.JSTaskProfile(5), func(ctx context.Context) error {
			return context.Canceled
		}),
		async.NewTaskFunc("task-3", async.JSTaskProfile(5), func(ctx context.Context) error {
			return nil
		}),
	}

	batchConfig := DefaultBatchConfig()
	batchConfig.ContinueOnError = true

	result, err := pool.ExecuteBatch(ctx, tasks, batchConfig)

	// ContinueOnError = true，所以不应该返回错误
	if err != nil {
		t.Errorf("ExecuteBatch() with ContinueOnError should not return error, got %v", err)
	}

	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}

	if result.Success != 2 {
		t.Errorf("Success = %d, want 2", result.Success)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Errors count = %d, want 1", len(result.Errors))
	}
}

func TestJSEnginePool_Warmup(t *testing.T) {
	config := DefaultPoolConfig()
	config.Prewarm = false // 不自动预热
	pool := NewJSEnginePool(config)
	defer pool.Close()

	// 初始状态
	stats := pool.GetStats()
	if stats.TotalEngines != 0 {
		t.Errorf("Initial TotalEngines = %d, want 0", stats.TotalEngines)
	}

	// 预热
	err := pool.Warmup(3)
	if err != nil {
		t.Errorf("Warmup() error = %v", err)
	}

	// 检查预热结果
	stats = pool.GetStats()
	if stats.TotalEngines != 3 {
		t.Errorf("TotalEngines after warmup = %d, want 3", stats.TotalEngines)
	}
	if stats.IdleEngines != 3 {
		t.Errorf("IdleEngines after warmup = %d, want 3", stats.IdleEngines)
	}
}

func TestJSEnginePool_Resize(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	// 初始大小
	if pool.config.MaxEngines != 4 {
		t.Errorf("Initial MaxEngines = %d, want 4", pool.config.MaxEngines)
	}

	// 缩小
	err := pool.Resize(2)
	if err != nil {
		t.Errorf("Resize() error = %v", err)
	}

	if pool.config.MaxEngines != 2 {
		t.Errorf("MaxEngines after resize = %d, want 2", pool.config.MaxEngines)
	}

	// 扩大
	err = pool.Resize(6)
	if err != nil {
		t.Errorf("Resize() error = %v", err)
	}

	if pool.config.MaxEngines != 6 {
		t.Errorf("MaxEngines after resize = %d, want 6", pool.config.MaxEngines)
	}
}

func TestJSEnginePool_Health(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	health := pool.Health()

	if !health.Healthy {
		t.Error("Health should be healthy")
	}

	if health.TotalEngines < 0 {
		t.Error("TotalEngines should be non-negative")
	}
}

func TestJSEnginePool_Close(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)

	// 关闭前应该是健康的
	if pool.IsClosed() {
		t.Error("Pool should not be closed initially")
	}

	// 关闭
	pool.Close()

	// 关闭后应该标记为已关闭
	if !pool.IsClosed() {
		t.Error("Pool should be closed after Close()")
	}

	// 关闭后获取引擎应该失败
	ctx := context.Background()
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Acquire() should return error when pool is closed")
	}
}

func TestJSEnginePool_Acquire_ContextCancel(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxEngines = 1 // 限制为1个引擎
	config.Prewarm = false
	pool := NewJSEnginePool(config)
	defer pool.Close()

	// 获取唯一的引擎
	ctx := context.Background()
	engine, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	// 使用取消的上下文尝试获取
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	_, err = pool.Acquire(cancelCtx)
	if err != context.Canceled {
		t.Errorf("Acquire() with canceled context should return context.Canceled, got %v", err)
	}

	// 归还引擎
	pool.Release(engine)
}

func TestJSEnginePool_Stats(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	// 记录初始统计
	initialStats := pool.GetStats()

	// 执行一些操作
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		pool.Execute(ctx, func(engine *JSEngine) error {
			return nil
		})
	}

	// 检查统计
	stats := pool.GetStats()
	if stats.TotalRequests <= initialStats.TotalRequests {
		t.Error("TotalRequests should increase")
	}
}

// Benchmarks

func BenchmarkJSEnginePool_Execute(b *testing.B) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Execute(ctx, func(engine *JSEngine) error {
			return nil
		})
	}
}

func BenchmarkJSEnginePool_Execute_Parallel(b *testing.B) {
	config := DefaultPoolConfig()
	pool := NewJSEnginePool(config)
	defer pool.Close()

	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Execute(ctx, func(engine *JSEngine) error {
				return nil
			})
		}
	})
}
