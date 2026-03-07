// Package cache 提供懒加载和批量读取功能。
//
// 该文件实现了懒加载索引、批量读取优化和预取机制，
// 减少启动时间和 I/O 操作，提升系统响应速度。

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

// ============================================================================
// 懒加载索引
// ============================================================================

// LazyIndex 懒加载索引
type LazyIndex struct {
	baseDir   string
	indexPath string
	mu        sync.RWMutex
	cache     *Cache
	loader    IndexLoader
	loaded    bool
	loadTime  time.Time
}

// IndexLoader 索引加载器接口
type IndexLoader interface {
	LoadIndex(ctx context.Context, path string) (interface{}, error)
}

// DepsIndexLoader 依赖索引加载器
type DepsIndexLoader struct{}

// LoadIndex 加载依赖索引
func (l *DepsIndexLoader) LoadIndex(ctx context.Context, path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &store.DepsIndex{
				GeneratedAt: time.Time{},
				Apps:        make(map[string]*store.AppDeps),
			}, nil
		}
		return nil, err
	}

	var index store.DepsIndex
	if err := parseJSON(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// BucketIndexLoader Bucket 索引加载器
type BucketIndexLoader struct{}

// LoadIndex 加载 Bucket 索引
func (l *BucketIndexLoader) LoadIndex(ctx context.Context, path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &store.BucketIndex{
				GeneratedAt: time.Time{},
				Buckets:     make(map[string]*store.BucketConfig),
			}, nil
		}
		return nil, err
	}

	var index store.BucketIndex
	if err := parseJSON(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// NewLazyIndex 创建懒加载索引
func NewLazyIndex(baseDir string, indexType string, loader IndexLoader) *LazyIndex {
	indexPath := filepath.Join(baseDir, indexType+"-index.json")

	cache := NewCache(CacheConfig{
		MaxSize:    64 * 1024 * 1024, // 64MB
		MaxEntries: 100,
		TTL:        5 * time.Minute,
	})

	return &LazyIndex{
		baseDir:   baseDir,
		indexPath: indexPath,
		cache:     cache,
		loader:    loader,
	}
}

// Get 获取索引（懒加载）
func (li *LazyIndex) Get(ctx context.Context) (interface{}, error) {
	li.mu.RLock()
	if li.loaded && !li.isStale() {
		// 已加载且未过期，从缓存获取
		cached, _ := li.cache.Get("index")
		li.mu.RUnlock()
		if cached != nil {
			return cached, nil
		}
	} else {
		li.mu.RUnlock()
	}

	// 需要加载
	li.mu.Lock()
	defer li.mu.Unlock()

	// 双重检查
	if li.loaded && !li.isStale() {
		cached, _ := li.cache.Get("index")
		if cached != nil {
			return cached, nil
		}
	}

	// 从磁盘加载
	index, err := li.loader.LoadIndex(ctx, li.indexPath)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	li.cache.Set("index", index)
	li.loaded = true
	li.loadTime = time.Now()

	return index, nil
}

// Reload 强制重新加载索引
func (li *LazyIndex) Reload(ctx context.Context) error {
	li.mu.Lock()
	defer li.mu.Unlock()

	index, err := li.loader.LoadIndex(ctx, li.indexPath)
	if err != nil {
		return err
	}

	li.cache.Set("index", index)
	li.loaded = true
	li.loadTime = time.Now()

	return nil
}

// isStale 检查索引是否过期
func (li *LazyIndex) isStale() bool {
	if !li.loaded {
		return true
	}

	// 检查文件是否被修改
	stat, err := os.Stat(li.indexPath)
	if err != nil {
		return true
	}

	return stat.ModTime().After(li.loadTime)
}

// Close 关闭懒加载索引
func (li *LazyIndex) Close() {
	li.cache.Close()
}

// ============================================================================
// 批量读取优化
// ============================================================================

// BatchReader 批量读取器
type BatchReader struct {
	storage    store.Storage
	cache      *Cache
	batchSize  int
	prefetcher *Prefetcher
}

// BatchReadResult 批量读取结果
type BatchReadResult struct {
	Items     []interface{}
	Errors    map[string]error
	CacheHits int
	CacheMisses int
	Duration  time.Duration
}

// NewBatchReader 创建批量读取器
func NewBatchReader(storage store.Storage, config CacheConfig) *BatchReader {
	cache := NewCache(config)

	return &BatchReader{
		storage:    storage,
		cache:      cache,
		batchSize:  DefaultBatchReadSize,
		prefetcher: NewPrefetcher(config),
	}
}

// SetBatchSize 设置批量大小
func (br *BatchReader) SetBatchSize(size int) {
	br.batchSize = size
}

// ReadApps 批量读取应用
func (br *BatchReader) ReadApps(ctx context.Context, names []string) *BatchReadResult {
	start := time.Now()
	result := &BatchReadResult{
		Items:  make([]interface{}, 0, len(names)),
		Errors: make(map[string]error),
	}

	// 分组：已缓存 vs 未缓存
	var uncachedNames []string
	for _, name := range names {
		key := fmt.Sprintf("app:%s", name)
		if cached, ok := br.cache.Get(key); ok {
			result.Items = append(result.Items, cached)
			result.CacheHits++
		} else {
			uncachedNames = append(uncachedNames, name)
			result.CacheMisses++
		}
	}

	// 批量读取未缓存的
	if len(uncachedNames) > 0 {
		br.batchReadApps(ctx, uncachedNames, result)
	}

	result.Duration = time.Since(start)
	return result
}

// batchReadApps 批量读取应用（内部方法）
func (br *BatchReader) batchReadApps(ctx context.Context, names []string, result *BatchReadResult) {
	// 并行读取
	type readResult struct {
		name string
		app  *store.AppManifest
		err  error
	}

	results := make(chan readResult, len(names))
	sem := make(chan struct{}, 8) // 最多 8 个并发

	for _, name := range names {
		go func(name string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			app, err := br.storage.GetApp(ctx, name)
			results <- readResult{name: name, app: app, err: err}
		}(name)
	}

	// 收集结果
	for i := 0; i < len(names); i++ {
		r := <-results
		if r.err != nil {
			result.Errors[r.name] = r.err
		} else {
			result.Items = append(result.Items, r.app)
			// 缓存结果
			key := fmt.Sprintf("app:%s", r.name)
			br.cache.Set(key, r.app)
		}
	}

	// 预取相关应用
	br.prefetcher.prefetchRelatedApps(ctx, names)
}

// ReadAllApps 读取所有应用（带批量优化）
func (br *BatchReader) ReadAllApps(ctx context.Context) (*BatchReadResult, error) {
	start := time.Now()

	// 先尝试从缓存获取所有
	allApps, err := br.storage.ListApps(ctx)
	if err != nil {
		return nil, err
	}

	result := &BatchReadResult{
		Items:    make([]interface{}, 0, len(allApps)),
		Errors:   make(map[string]error),
		Duration: time.Since(start),
	}

	// 批量缓存检查
	var names []string
	appMap := make(map[string]*store.AppManifest)
	for _, app := range allApps {
		names = append(names, app.Name)
		appMap[app.Name] = app
	}

	// 使用批量读取
	batchResult := br.ReadApps(ctx, names)
	result.Items = batchResult.Items
	result.Errors = batchResult.Errors
	result.CacheHits = batchResult.CacheHits
	result.CacheMisses = batchResult.CacheMisses
	result.Duration = time.Since(start)

	return result, nil
}

// Close 关闭批量读取器
func (br *BatchReader) Close() {
	br.cache.Close()
	if br.prefetcher != nil {
		br.prefetcher.Close()
	}
}

// ============================================================================
// 预取机制
// ============================================================================

// Prefetcher 预取器
type Prefetcher struct {
	cache     *Cache
	storage   store.Storage
	prefetchWg sync.WaitGroup
	stopChan  chan struct{}
}

// NewPrefetcher 创建预取器
func NewPrefetcher(config CacheConfig) *Prefetcher {
	cache := NewCache(config)

	return &Prefetcher{
		cache:    cache,
		stopChan: make(chan struct{}),
	}
}

// SetStorage 设置存储（用于预取）
func (p *Prefetcher) SetStorage(storage store.Storage) {
	p.storage = storage
}

// prefetchRelatedApps 预取相关应用
func (p *Prefetcher) prefetchRelatedApps(ctx context.Context, accessedNames []string) {
	if p.storage == nil {
		return
	}

	p.prefetchWg.Add(1)
	go func() {
		defer p.prefetchWg.Done()

		// 获取依赖关系
		depsIndex, err := p.storage.GetDepsIndex(ctx)
		if err != nil {
			return
		}

		// 收集需要预取的应用
		var toPrefetch []string
		seen := make(map[string]bool)

		for _, name := range accessedNames {
			if seen[name] {
				continue
			}
			seen[name] = true

			// 获取该应用的依赖
			if appDeps, ok := depsIndex.Apps[name]; ok {
				// 预取依赖的应用
				for _, depName := range appDeps.Dependencies {
					if !seen[depName] {
						toPrefetch = append(toPrefetch, depName)
						seen[depName] = true
					}
				}
			}
		}

		// 限制预取数量
		if len(toPrefetch) > DefaultPrefetchDistance {
			toPrefetch = toPrefetch[:DefaultPrefetchDistance]
		}

		// 异步预取
		if len(toPrefetch) > 0 {
			p.doPrefetch(ctx, toPrefetch)
		}
	}()
}

// doPrefetch 执行预取
func (p *Prefetcher) doPrefetch(ctx context.Context, names []string) {
	for _, name := range names {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		default:
		}

		// 检查是否已缓存
		key := fmt.Sprintf("app:%s", name)
		if _, ok := p.cache.Get(key); ok {
			continue
		}

		// 预取
		app, err := p.storage.GetApp(ctx, name)
		if err == nil {
			p.cache.Set(key, app)
		}
	}
}

// Close 关闭预取器
func (p *Prefetcher) Close() {
	close(p.stopChan)
	p.prefetchWg.Wait()
	p.cache.Close()
}

// ============================================================================
// 辅助函数
// ============================================================================

// parseJSON 解析 JSON 数据
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ============================================================================
// 应用列表懒加载
// ============================================================================

// LazyAppList 懒加载应用列表
type LazyAppList struct {
	appsDir   string
	mu        sync.RWMutex
	apps      map[string]*AppEntry
	loaded    bool
	loadTime  time.Time
	cache     *Cache
}

// AppEntry 应用条目
type AppEntry struct {
	Name       string
	Manifest   *store.AppManifest
	Loaded     bool
	LoadError  error
	AccessTime time.Time
}

// NewLazyAppList 创建懒加载应用列表
func NewLazyAppList(appsDir string, config CacheConfig) *LazyAppList {
	cache := NewCache(config)

	return &LazyAppList{
		appsDir: appsDir,
		apps:    make(map[string]*AppEntry),
		cache:   cache,
	}
}

// List 列出所有应用（懒加载）
func (lal *LazyAppList) List(ctx context.Context) ([]string, error) {
	lal.mu.Lock()
	defer lal.mu.Unlock()

	// 检查是否需要重新扫描
	if !lal.loaded || lal.isStale() {
		if err := lal.scanAppsLocked(); err != nil {
			return nil, err
		}
	}

	// 返回应用名称列表
	names := make([]string, 0, len(lal.apps))
	for name := range lal.apps {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// Get 获取应用信息（懒加载）
func (lal *LazyAppList) Get(ctx context.Context, name string) (*store.AppManifest, error) {
	lal.mu.RLock()
	entry, exists := lal.apps[name]
	lal.mu.RUnlock()

	if !exists {
		// 应用不存在，重新扫描
		lal.mu.Lock()
		if err := lal.scanAppsLocked(); err != nil {
			lal.mu.Unlock()
			return nil, errors.NewAppNotFound(name)
		}
		entry, exists = lal.apps[name]
		lal.mu.Unlock()

		if !exists {
			return nil, errors.NewAppNotFound(name)
		}
	}

	// 懒加载 manifest
	if !entry.Loaded {
		return lal.loadAppManifest(ctx, name, entry)
	}

	if entry.LoadError != nil {
		return nil, entry.LoadError
	}

	return entry.Manifest, nil
}

// loadAppManifest 加载应用 manifest
func (lal *LazyAppList) loadAppManifest(ctx context.Context, name string, entry *AppEntry) (*store.AppManifest, error) {
	lal.mu.Lock()
	defer lal.mu.Unlock()

	// 双重检查
	if entry.Loaded {
		return entry.Manifest, entry.LoadError
	}

	// 从磁盘加载
	manifestPath := filepath.Join(lal.appsDir, name, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		entry.Loaded = true
		entry.LoadError = err
		return nil, err
	}

	var app store.AppManifest
	if err := parseJSON(data, &app); err != nil {
		entry.Loaded = true
		entry.LoadError = err
		return nil, err
	}

	entry.Manifest = &app
	entry.Loaded = true
	entry.LoadError = nil

	// 缓存
	key := fmt.Sprintf("app:%s", name)
	lal.cache.Set(key, &app)

	return &app, nil
}

// scanAppsLocked 扫描应用目录（需要持有锁）
func (lal *LazyAppList) scanAppsLocked() error {
	entries, err := os.ReadDir(lal.appsDir)
	if err != nil {
		if os.IsNotExist(err) {
			lal.apps = make(map[string]*AppEntry)
			lal.loaded = true
			lal.loadTime = time.Now()
			return nil
		}
		return err
	}

	newApps := make(map[string]*AppEntry)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// 保留已有条目
		if oldEntry, ok := lal.apps[name]; ok {
			newApps[name] = oldEntry
		} else {
			newApps[name] = &AppEntry{
				Name: name,
			}
		}
	}

	lal.apps = newApps
	lal.loaded = true
	lal.loadTime = time.Now()

	return nil
}

// isStale 检查是否过期
func (lal *LazyAppList) isStale() bool {
	stat, err := os.Stat(lal.appsDir)
	if err != nil {
		return true
	}

	return stat.ModTime().After(lal.loadTime)
}

// Close 关闭懒加载应用列表
func (lal *LazyAppList) Close() {
	lal.cache.Close()
}

// GetStats 获取统计信息
func (lal *LazyAppList) GetStats() CacheStats {
	return lal.cache.Stats()
}
