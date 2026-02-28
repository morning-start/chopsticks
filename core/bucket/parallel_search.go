package bucket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"chopsticks/core/manifest"
	"golang.org/x/sync/errgroup"
)

// ParallelSearcher 并行搜索器
type ParallelSearcher struct {
	mgr        Manager
	maxWorkers int
	cache      *SearchCache
}

// NewParallelSearcher 创建并行搜索器
func NewParallelSearcher(mgr Manager, maxWorkers int) *ParallelSearcher {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	return &ParallelSearcher{
		mgr:        mgr,
		maxWorkers: maxWorkers,
		cache:      NewSearchCache(5 * time.Minute),
	}
}

// SearchAllBuckets 并行搜索所有 bucket
func (s *ParallelSearcher) SearchAllBuckets(
	ctx context.Context,
	query string,
	opts SearchOptions,
) ([]SearchResult, error) {
	// 获取所有 bucket
	buckets, err := s.mgr.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}

	if len(buckets) == 0 {
		return []SearchResult{}, nil
	}

	// 使用 errgroup 控制并发
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.maxWorkers)

	// 结果通道
	resultChan := make(chan []SearchResult, len(buckets))
	var mu sync.Mutex
	var searchErrors []error

	// 并行搜索每个 bucket
	for _, bucketName := range buckets {
		bucketName := bucketName // 捕获循环变量
		g.Go(func() error {
			bucketOpts := SearchOptions{
				Bucket:   bucketName,
				Category: opts.Category,
				Tags:     opts.Tags,
			}

			results, err := s.mgr.Search(ctx, query, bucketOpts)
			if err != nil {
				mu.Lock()
				searchErrors = append(searchErrors, 
					fmt.Errorf("search bucket %s failed: %w", bucketName, err))
				mu.Unlock()
				return nil // 继续处理其他 bucket
			}

			resultChan <- results
			return nil
		})
	}

	// 等待所有搜索完成
	go func() {
		if err := g.Wait(); err != nil {
			// errgroup 已经处理了错误
		}
		close(resultChan)
	}()

	// 聚合结果
	var allResults []SearchResult
	for results := range resultChan {
		allResults = append(allResults, results...)
	}

	// 检查是否有错误
	if len(allResults) == 0 && len(searchErrors) > 0 {
		return nil, searchErrors[0]
	}

	return allResults, nil
}

// SearchWithCache 带缓存的搜索
func (s *ParallelSearcher) SearchWithCache(
	ctx context.Context,
	query string,
	opts SearchOptions,
) ([]SearchResult, error) {
	// 生成缓存键
	cacheKey := generateCacheKey(query, opts)

	// 尝试从缓存获取
	if cached, found := s.cache.Get(cacheKey); found {
		return cached, nil
	}

	// 执行搜索
	results, err := s.SearchAllBuckets(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	s.cache.Set(cacheKey, results)

	return results, nil
}

// ClearCache 清除搜索缓存
func (s *ParallelSearcher) ClearCache() {
	s.cache.Clear()
}

// GetCacheStats 获取缓存统计
func (s *ParallelSearcher) GetCacheStats() CacheStats {
	return s.cache.GetStats()
}

// generateCacheKey 生成缓存键
func generateCacheKey(query string, opts SearchOptions) string {
	key := query
	if opts.Bucket != "" {
		key += ":bucket=" + opts.Bucket
	}
	if opts.Category != "" {
		key += ":category=" + opts.Category
	}
	if len(opts.Tags) > 0 {
		key += ":tags="
		for i, tag := range opts.Tags {
			if i > 0 {
				key += ","
			}
			key += tag
		}
	}
	return key
}

// SearchCache 搜索缓存
type SearchCache struct {
	data    map[string]cacheEntry
	ttl     time.Duration
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

type cacheEntry struct {
	results   []SearchResult
	timestamp time.Time
}

// NewSearchCache 创建搜索缓存
func NewSearchCache(ttl time.Duration) *SearchCache {
	cache := &SearchCache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}

	// 启动清理协程
	go cache.cleanupLoop()

	return cache
}

// Get 从缓存获取结果
func (c *SearchCache) Get(key string) ([]SearchResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.timestamp) > c.ttl {
		c.misses++
		return nil, false
	}

	c.hits++
	return entry.results, true
}

// Set 将结果存入缓存
func (c *SearchCache) Set(key string, results []SearchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		results:   results,
		timestamp: time.Now(),
	}
}

// Clear 清除所有缓存
func (c *SearchCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheEntry)
	c.hits = 0
	c.misses = 0
}

// cleanupLoop 定期清理过期缓存
func (c *SearchCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期条目
func (c *SearchCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.data, key)
		}
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits       int64
	Misses     int64
	HitRate    float64
	EntryCount int
}

// GetStats 获取缓存统计
func (c *SearchCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:       c.hits,
		Misses:     c.misses,
		HitRate:    hitRate,
		EntryCount: len(c.data),
	}
}

// ParallelSearchOptions 并行搜索选项
type ParallelSearchOptions struct {
	MaxWorkers int
	UseCache   bool
	CacheTTL   time.Duration
}

// DefaultParallelSearchOptions 返回默认并行搜索选项
func DefaultParallelSearchOptions() ParallelSearchOptions {
	return ParallelSearchOptions{
		MaxWorkers: 10,
		UseCache:   true,
		CacheTTL:   5 * time.Minute,
	}
}

// SearchAppsInBucket 在单个 bucket 中搜索应用
func (s *ParallelSearcher) SearchAppsInBucket(
	ctx context.Context,
	bucketName string,
	query string,
	opts SearchOptions,
) ([]SearchResult, error) {
	bucketOpts := SearchOptions{
		Bucket:   bucketName,
		Category: opts.Category,
		Tags:     opts.Tags,
	}

	return s.mgr.Search(ctx, query, bucketOpts)
}

// FilterResults 过滤搜索结果
func FilterResults(results []SearchResult, filter func(*manifest.AppRef) bool) []SearchResult {
	var filtered []SearchResult
	for _, result := range results {
		if filter(result.App) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// SortResultsByName 按名称排序搜索结果
func SortResultsByName(results []SearchResult) {
	// 使用冒泡排序（简单实现）
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].App.Name > results[j+1].App.Name {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// DeduplicateResults 去重搜索结果
func DeduplicateResults(results []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	var unique []SearchResult

	for _, result := range results {
		key := result.Bucket + "/" + result.App.Name
		if !seen[key] {
			seen[key] = true
			unique = append(unique, result)
		}
	}

	return unique
}
