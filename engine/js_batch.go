package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"chopsticks/pkg/async"

	"github.com/dop251/goja"
	"golang.org/x/sync/errgroup"
)

// BatchConfig 批量执行配置
type BatchConfig struct {
	MaxConcurrency  int           // 最大并发数
	Timeout         time.Duration // 单个任务超时
	ContinueOnError bool          // 出错时是否继续
}

// DefaultBatchConfig 返回默认批量配置
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxConcurrency:  4,
		Timeout:         30 * time.Second,
		ContinueOnError: false,
	}
}

// BatchResult 批量执行结果
type BatchResult struct {
	Total    int
	Success  int
	Failed   int
	Errors   map[string]error
	Duration time.Duration
}

// ExecuteBatch 批量执行 JS 任务
func (p *JSEnginePool) ExecuteBatch(
	ctx context.Context,
	tasks []async.Task,
	config BatchConfig,
) (*BatchResult, error) {
	if p.IsClosed() {
		return nil, fmt.Errorf("pool is closed")
	}

	result := &BatchResult{
		Total:  len(tasks),
		Errors: make(map[string]error),
	}

	start := time.Now()

	// 使用 errgroup 控制并发
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.MaxConcurrency)

	var successCount int32
	var failedCount int32
	var mu sync.Mutex

	for _, task := range tasks {
		task := task // 捕获循环变量
		g.Go(func() error {
			// 创建带超时的上下文
			taskCtx, cancel := context.WithTimeout(ctx, config.Timeout)
			defer cancel()

			// 获取引擎并执行任务
			err := p.Execute(taskCtx, func(engine *JSEngine) error {
				return task.Execute(taskCtx)
			})

			if err != nil {
				atomic.AddInt32(&failedCount, 1)
				mu.Lock()
				result.Errors[task.ID()] = err
				mu.Unlock()

				if !config.ContinueOnError {
					return err // 终止其他任务
				}
			} else {
				atomic.AddInt32(&successCount, 1)
			}

			return nil
		})
	}

	// 等待所有任务完成
	err := g.Wait()

	result.Duration = time.Since(start)
	result.Success = int(atomic.LoadInt32(&successCount))
	result.Failed = int(atomic.LoadInt32(&failedCount))

	if err != nil && !config.ContinueOnError {
		return result, err
	}

	return result, nil
}

// ExecuteScriptsBatch 批量执行脚本文件
func (p *JSEnginePool) ExecuteScriptsBatch(
	ctx context.Context,
	scriptPaths []string,
	config BatchConfig,
) (*BatchResult, error) {
	if p.IsClosed() {
		return nil, fmt.Errorf("pool is closed")
	}

	result := &BatchResult{
		Total:  len(scriptPaths),
		Errors: make(map[string]error),
	}

	start := time.Now()

	// 使用 errgroup 控制并发
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.MaxConcurrency)

	var successCount int32
	var failedCount int32
	var mu sync.Mutex

	for _, path := range scriptPaths {
		path := path // 捕获循环变量
		g.Go(func() error {
			// 创建带超时的上下文
			taskCtx, cancel := context.WithTimeout(ctx, config.Timeout)
			defer cancel()

			// 使用缓存执行脚本
			err := p.ExecuteWithCache(taskCtx, path, func(engine *JSEngine, program *goja.Program) error {
				// 执行编译后的脚本
				_, runErr := engine.vm.RunProgram(program)
				return runErr
			})

			if err != nil {
				atomic.AddInt32(&failedCount, 1)
				mu.Lock()
				result.Errors[path] = err
				mu.Unlock()

				if !config.ContinueOnError {
					return err
				}
			} else {
				atomic.AddInt32(&successCount, 1)
			}

			return nil
		})
	}

	// 等待所有任务完成
	err := g.Wait()

	result.Duration = time.Since(start)
	result.Success = int(atomic.LoadInt32(&successCount))
	result.Failed = int(atomic.LoadInt32(&failedCount))

	if err != nil && !config.ContinueOnError {
		return result, err
	}

	return result, nil
}

// ExecuteWithResult 执行并返回结果
func (p *JSEnginePool) ExecuteWithResult(
	ctx context.Context,
	fn func(*JSEngine) (interface{}, error),
) (interface{}, error) {
	var result interface{}
	var resultErr error

	err := p.Execute(ctx, func(engine *JSEngine) error {
		result, resultErr = fn(engine)
		return resultErr
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Warmup 预热引擎池
func (p *JSEnginePool) Warmup(count int) error {
	if p.IsClosed() {
		return fmt.Errorf("pool is closed")
	}

	if count <= 0 {
		count = p.config.PrewarmSize
	}

	if count > p.config.MaxEngines {
		count = p.config.MaxEngines
	}

	currentIdle := int(atomic.LoadInt32(&p.stats.IdleEngines))
	needed := count - currentIdle

	if needed <= 0 {
		return nil // 已经有足够的引擎
	}

	for i := 0; i < needed; i++ {
		engine := p.createEngine()
		if engine == nil {
			return fmt.Errorf("failed to create engine %d", i)
		}

		select {
		case p.available <- engine:
			atomic.AddInt32(&p.stats.IdleEngines, 1)
			atomic.AddInt32(&p.stats.TotalEngines, 1)
		default:
			engine.Close()
			return fmt.Errorf("failed to add engine to pool")
		}
	}

	return nil
}

// Resize 调整池大小
func (p *JSEnginePool) Resize(maxEngines int) error {
	if p.IsClosed() {
		return fmt.Errorf("pool is closed")
	}

	if maxEngines <= 0 {
		return fmt.Errorf("maxEngines must be positive")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	oldMax := p.config.MaxEngines
	p.config.MaxEngines = maxEngines

	// 如果新大小小于旧大小，需要关闭多余的引擎
	if maxEngines < oldMax {
		currentTotal := int(atomic.LoadInt32(&p.stats.TotalEngines))
		toClose := currentTotal - maxEngines

		if toClose > 0 {
			// 关闭空闲引擎
			closed := 0
			for closed < toClose {
				select {
				case engine := <-p.available:
					engine.Close()
					atomic.AddInt32(&p.stats.IdleEngines, -1)
					atomic.AddInt32(&p.stats.TotalEngines, -1)
					closed++
				default:
					// 没有更多空闲引擎了
					return nil
				}
			}
		}
	}

	return nil
}

// Health 健康检查
func (p *JSEnginePool) Health() PoolHealth {
	stats := p.GetStats()

	return PoolHealth{
		Healthy:       !p.IsClosed(),
		TotalEngines:  int(stats.TotalEngines),
		ActiveEngines: int(stats.ActiveEngines),
		IdleEngines:   int(stats.IdleEngines),
		CacheHitRate:  p.GetCacheStats().HitRate,
	}
}

// PoolHealth 池健康状态
type PoolHealth struct {
	Healthy       bool
	TotalEngines  int
	ActiveEngines int
	IdleEngines   int
	CacheHitRate  float64
}
