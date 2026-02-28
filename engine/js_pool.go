package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// PoolConfig 引擎池配置
type PoolConfig struct {
	MaxEngines  int           // 最大引擎数
	MaxIdle     int           // 最大空闲引擎数
	Prewarm     bool          // 是否预热
	PrewarmSize int           // 预热数量
	CacheSize   int64         // 缓存大小
	MaxCacheEntries int       // 最大缓存条目数
}

// DefaultPoolConfig 返回默认池配置
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxEngines:      4,
		MaxIdle:         2,
		Prewarm:         true,
		PrewarmSize:     2,
		CacheSize:       100 * 1024 * 1024, // 100MB
		MaxCacheEntries: 100,
	}
}

// JSEnginePool JS 引擎池
type JSEnginePool struct {
	config      PoolConfig
	available   chan *PooledEngine
	inUse       sync.Map
	scriptCache *ScriptCache
	stats       PoolStats
	mu          sync.RWMutex
	closed      int32
}

// PooledEngine 池中的引擎
type PooledEngine struct {
	*JSEngine
	id         string
	acquiredAt time.Time
	useCount   int
}

// PoolStats 池统计信息
type PoolStats struct {
	TotalEngines   int32
	ActiveEngines  int32
	IdleEngines    int32
	TotalRequests  int64
	CacheHits      int64
	CacheMisses    int64
	WaitTime       int64 // 等待时间 (纳秒)
}

// NewJSEnginePool 创建 JS 引擎池
func NewJSEnginePool(config PoolConfig) *JSEnginePool {
	pool := &JSEnginePool{
		config:      config,
		available:   make(chan *PooledEngine, config.MaxEngines),
		scriptCache: NewScriptCache(config.CacheSize, config.MaxCacheEntries),
	}

	// 预热引擎
	if config.Prewarm {
		pool.prewarm()
	}

	return pool
}

// prewarm 预热引擎池
func (p *JSEnginePool) prewarm() {
	count := p.config.PrewarmSize
	if count > p.config.MaxEngines {
		count = p.config.MaxEngines
	}

	for i := 0; i < count; i++ {
		engine := p.createEngine()
		if engine != nil {
			p.available <- engine
			atomic.AddInt32(&p.stats.IdleEngines, 1)
			atomic.AddInt32(&p.stats.TotalEngines, 1)
		}
	}
}

// createEngine 创建新引擎
func (p *JSEnginePool) createEngine() *PooledEngine {
	jsEngine := NewJSEngine()
	if jsEngine == nil {
		return nil
	}

	return &PooledEngine{
		JSEngine: jsEngine,
		id:       uuid.New().String(),
		useCount: 0,
	}
}

// Acquire 获取引擎
func (p *JSEnginePool) Acquire(ctx context.Context) (*PooledEngine, error) {
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, fmt.Errorf("pool is closed")
	}

	atomic.AddInt64(&p.stats.TotalRequests, 1)
	start := time.Now()

	select {
	case engine := <-p.available:
		// 从池中获取
		engine.acquiredAt = time.Now()
		engine.useCount++
		p.inUse.Store(engine.id, engine)
		atomic.AddInt32(&p.stats.IdleEngines, -1)
		atomic.AddInt32(&p.stats.ActiveEngines, 1)
		atomic.AddInt64(&p.stats.WaitTime, time.Since(start).Nanoseconds())
		return engine, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// 池已满，创建临时引擎
		if atomic.LoadInt32(&p.stats.TotalEngines) < int32(p.config.MaxEngines) {
			engine := p.createEngine()
			if engine != nil {
				engine.acquiredAt = time.Now()
				engine.useCount++
				p.inUse.Store(engine.id, engine)
				atomic.AddInt32(&p.stats.TotalEngines, 1)
				atomic.AddInt32(&p.stats.ActiveEngines, 1)
				atomic.AddInt64(&p.stats.WaitTime, time.Since(start).Nanoseconds())
				return engine, nil
			}
		}

		// 等待可用引擎
		select {
		case engine := <-p.available:
			engine.acquiredAt = time.Now()
			engine.useCount++
			p.inUse.Store(engine.id, engine)
			atomic.AddInt32(&p.stats.IdleEngines, -1)
			atomic.AddInt32(&p.stats.ActiveEngines, 1)
			atomic.AddInt64(&p.stats.WaitTime, time.Since(start).Nanoseconds())
			return engine, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Release 归还引擎
func (p *JSEnginePool) Release(engine *PooledEngine) {
	if engine == nil {
		return
	}

	// 从使用中移除
	p.inUse.Delete(engine.id)

	// 重置引擎状态
	p.resetEngine(engine)

	// 检查池是否已关闭
	if atomic.LoadInt32(&p.closed) == 1 {
		engine.Close()
		atomic.AddInt32(&p.stats.TotalEngines, -1)
		atomic.AddInt32(&p.stats.ActiveEngines, -1)
		return
	}

	// 尝试归还到池中
	select {
	case p.available <- engine:
		// 成功归还
		atomic.AddInt32(&p.stats.IdleEngines, 1)
		atomic.AddInt32(&p.stats.ActiveEngines, -1)
	default:
		// 池已满，关闭引擎
		engine.Close()
		atomic.AddInt32(&p.stats.TotalEngines, -1)
		atomic.AddInt32(&p.stats.ActiveEngines, -1)
	}
}

// resetEngine 重置引擎状态
func (p *JSEnginePool) resetEngine(engine *PooledEngine) {
	// 重置安装上下文
	engine.installCtx = make(map[string]interface{})
	
	// 重置 VM 中的全局变量（如果需要）
	// 这里可以根据需要添加更多重置逻辑
}

// Execute 使用池中的引擎执行任务
func (p *JSEnginePool) Execute(ctx context.Context, fn func(*JSEngine) error) error {
	engine, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire engine: %w", err)
	}
	defer p.Release(engine)

	return fn(engine.JSEngine)
}

// ExecuteWithCache 使用缓存执行脚本
func (p *JSEnginePool) ExecuteWithCache(ctx context.Context, scriptPath string, fn func(*JSEngine, *goja.Program) error) error {
	engine, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire engine: %w", err)
	}
	defer p.Release(engine)

	// 使用缓存加载和编译脚本
	program, err := p.scriptCache.LoadAndCompile(engine.vm, scriptPath)
	if err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	return fn(engine.JSEngine, program)
}

// GetCacheStats 获取缓存统计
func (p *JSEnginePool) GetCacheStats() CacheStats {
	return p.scriptCache.Stats()
}

// GetStats 获取池统计
func (p *JSEnginePool) GetStats() PoolStats {
	return PoolStats{
		TotalEngines:  atomic.LoadInt32(&p.stats.TotalEngines),
		ActiveEngines: atomic.LoadInt32(&p.stats.ActiveEngines),
		IdleEngines:   atomic.LoadInt32(&p.stats.IdleEngines),
		TotalRequests: atomic.LoadInt64(&p.stats.TotalRequests),
		CacheHits:     atomic.LoadInt64(&p.stats.CacheHits),
		CacheMisses:   atomic.LoadInt64(&p.stats.CacheMisses),
		WaitTime:      atomic.LoadInt64(&p.stats.WaitTime),
	}
}

// Close 关闭引擎池
func (p *JSEnginePool) Close() {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return // 已经关闭
	}

	// 关闭所有可用引擎
	close(p.available)
	for engine := range p.available {
		engine.Close()
	}

	// 关闭所有使用中的引擎
	p.inUse.Range(func(key, value interface{}) bool {
		if engine, ok := value.(*PooledEngine); ok {
			engine.Close()
		}
		return true
	})

	// 清空缓存
	p.scriptCache.Clear()
}

// IsClosed 检查池是否已关闭
func (p *JSEnginePool) IsClosed() bool {
	return atomic.LoadInt32(&p.closed) == 1
}
