// Package cache 提供高性能内存缓存层。
//
// 该包实现了基于 LRU 策略的内存缓存，用于缓存应用元数据、依赖索引、
// Bucket 信息等，减少文件系统 I/O 操作，提升系统性能。
package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"chopsticks/core/store"
)

// ============================================================================
// 常量定义
// ============================================================================

const (
	// 默认缓存配置
	DefaultMaxCacheSize     = 256 * 1024 * 1024 // 256MB
	DefaultMaxCacheEntries  = 10000             // 10000 个条目
	DefaultCacheTTL         = 10 * time.Minute  // 10 分钟 TTL
	DefaultCleanupInterval  = 2 * time.Minute   // 2 分钟清理间隔
	DefaultBatchReadSize    = 100               // 批量读取大小
	DefaultPrefetchDistance = 3                 // 预取距离
)

// ============================================================================
// 数据结构定义
// ============================================================================

// CacheEntry 缓存条目
type CacheEntry struct {
	Key       string
	Value     interface{}
	Size      int64         // 缓存条目大小 (字节)
	CreatedAt time.Time     // 创建时间
	ExpiresAt time.Time     // 过期时间
	AccessAt  time.Time     // 最后访问时间
	HitCount  int64         // 命中次数
	Element   *list.Element // LRU 列表元素
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits        int64     `json:"hits"`        // 命中次数
	Misses      int64     `json:"misses"`      // 未命中次数
	Evictions   int64     `json:"evictions"`   // 淘汰次数
	Size        int64     `json:"size"`        // 当前缓存大小
	Entries     int       `json:"entries"`     // 当前条目数
	MaxSize     int64     `json:"max_size"`    // 最大缓存大小
	MaxEntries  int       `json:"max_entries"` // 最大条目数
	HitRate     float64   `json:"hit_rate"`    // 命中率
	AvgAccessAt time.Time `json:"avg_access"`  // 平均访问时间
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxSize         int64         `json:"max_size"`         // 最大缓存大小 (字节)
	MaxEntries      int           `json:"max_entries"`      // 最大条目数
	TTL             time.Duration `json:"ttl"`              // 默认 TTL
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
}

// DefaultCacheConfig 返回默认缓存配置
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:         DefaultMaxCacheSize,
		MaxEntries:      DefaultMaxCacheEntries,
		TTL:             DefaultCacheTTL,
		CleanupInterval: DefaultCleanupInterval,
	}
}

// Cache LRU 缓存实现
type Cache struct {
	mu          sync.RWMutex
	entries     map[string]*CacheEntry
	lruList     *list.List
	config      CacheConfig
	currentSize int64
	hits        int64
	misses      int64
	evictions   int64
	cleanupStop chan struct{}
	cleanupDone chan struct{}
	closed      bool
	valueSizer  ValueSizer // 可选的值大小计算器
}

// ValueSizer 计算值大小的接口
type ValueSizer interface {
	SizeOf(value interface{}) int64
}

// ============================================================================
// 缓存核心功能
// ============================================================================

// NewCache 创建新的 LRU 缓存
func NewCache(config CacheConfig) *Cache {
	if config.MaxSize <= 0 {
		config.MaxSize = DefaultMaxCacheSize
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = DefaultMaxCacheEntries
	}
	if config.TTL <= 0 {
		config.TTL = DefaultCacheTTL
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = DefaultCleanupInterval
	}

	cache := &Cache{
		entries:     make(map[string]*CacheEntry),
		lruList:     list.New(),
		config:      config,
		cleanupStop: make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}

	// 启动后台清理协程
	go cache.cleanupLoop()

	return cache
}

// SetSizer 设置值大小计算器
func (c *Cache) SetSizer(sizer ValueSizer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.valueSizer = sizer
}

// Get 获取缓存值
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// 检查是否过期
	if c.isExpired(entry) {
		c.Delete(key)
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// 更新访问时间和 LRU 位置
	c.mu.Lock()
	c.lruList.MoveToFront(entry.Element)
	entry.AccessAt = time.Now()
	entry.HitCount++
	c.hits++
	c.mu.Unlock()

	return entry.Value, true
}

// Set 设置缓存值（使用默认 TTL）
func (c *Cache) Set(key string, value interface{}) error {
	return c.SetWithTTL(key, value, c.config.TTL)
}

// SetWithTTL 设置缓存值和自定义 TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("cache is closed")
	}

	// 计算值大小
	size := c.calculateSize(value)

	// 如果值太大，拒绝缓存
	if size > c.config.MaxSize {
		return fmt.Errorf("value size %d exceeds max cache size %d", size, c.config.MaxSize)
	}

	// 检查是否已存在
	if entry, exists := c.entries[key]; exists {
		// 更新现有条目
		c.updateEntry(entry, value, size, ttl)
		return nil
	}

	// 检查是否需要淘汰
	for c.currentSize+size > c.config.MaxSize || len(c.entries) >= c.config.MaxEntries {
		if !c.evictLRU() {
			break
		}
	}

	// 创建新条目
	now := time.Now()
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		Size:      size,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		AccessAt:  now,
		HitCount:  0,
	}

	entry.Element = c.lruList.PushFront(entry)
	c.entries[key] = entry
	c.currentSize += size

	return nil
}

// Delete 删除缓存条目
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return false
	}

	c.removeEntry(entry)
	return true
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.lruList.Init()
	c.currentSize = 0
}

// Close 关闭缓存
func (c *Cache) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	// 停止后台清理协程
	close(c.cleanupStop)
	<-c.cleanupDone
}

// Stats 获取缓存统计信息
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	var avgAccessAt time.Time
	if len(c.entries) > 0 {
		var totalAccessTime int64
		for _, entry := range c.entries {
			totalAccessTime += entry.AccessAt.UnixNano()
		}
		avgAccessAt = time.Unix(0, totalAccessTime/int64(len(c.entries)))
	}

	return CacheStats{
		Hits:        c.hits,
		Misses:      c.misses,
		Evictions:   c.evictions,
		Size:        c.currentSize,
		Entries:     len(c.entries),
		MaxSize:     c.config.MaxSize,
		MaxEntries:  c.config.MaxEntries,
		HitRate:     hitRate,
		AvgAccessAt: avgAccessAt,
	}
}

// ============================================================================
// 内部辅助方法
// ============================================================================

// cleanupLoop 后台清理过期条目的循环
func (c *Cache) cleanupLoop() {
	defer close(c.cleanupDone)

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.cleanupStop:
			return
		}
	}
}

// cleanupExpired 清理过期条目
func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiredKeys []string
	for key, entry := range c.entries {
		if c.isExpired(entry) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if entry, exists := c.entries[key]; exists {
			c.removeEntry(entry)
		}
	}
}

// isExpired 检查条目是否过期
func (c *Cache) isExpired(entry *CacheEntry) bool {
	if entry.ExpiresAt.IsZero() {
		return false // TTL 为 0 表示永不过期
	}
	return time.Now().After(entry.ExpiresAt)
}

// calculateSize 计算值的大小
func (c *Cache) calculateSize(value interface{}) int64 {
	if c.valueSizer != nil {
		return c.valueSizer.SizeOf(value)
	}

	// 默认大小计算：基于值的类型和大小
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case *store.AppManifest:
		return int64(len(v.Name) + 256) // 估算
	default:
		return 256 // 默认估算大小
	}
}

// updateEntry 更新现有条目
func (c *Cache) updateEntry(entry *CacheEntry, value interface{}, size int64, ttl time.Duration) {
	// 更新大小
	c.currentSize -= entry.Size
	c.currentSize += size

	// 更新值
	entry.Value = value
	entry.Size = size
	entry.CreatedAt = time.Now()
	entry.ExpiresAt = entry.CreatedAt.Add(ttl)
	entry.HitCount = 0

	c.lruList.MoveToFront(entry.Element)
}

// evictLRU 淘汰最久未使用的条目
func (c *Cache) evictLRU() bool {
	elem := c.lruList.Back()
	if elem == nil {
		return false
	}

	entry := elem.Value.(*CacheEntry)
	c.removeEntry(entry)
	c.evictions++
	return true
}

// removeEntry 删除条目
func (c *Cache) removeEntry(entry *CacheEntry) {
	c.lruList.Remove(entry.Element)
	delete(c.entries, entry.Key)
	c.currentSize -= entry.Size
}

// ============================================================================
// 应用缓存管理器
// ============================================================================

// AppCacheManager 应用缓存管理器
type AppCacheManager struct {
	cache       *Cache
	storage     store.Storage
	bucketCache *Cache // Bucket 信息缓存
	indexCache  *Cache // 索引缓存
}

// NewAppCacheManager 创建应用缓存管理器
func NewAppCacheManager(storage store.Storage, config CacheConfig) *AppCacheManager {
	cache := NewCache(config)

	// 创建专门的 Bucket 缓存
	bucketConfig := config
	bucketConfig.MaxSize = config.MaxSize / 4 // 25% 用于 Bucket 缓存
	bucketCache := NewCache(bucketConfig)

	// 创建专门的索引缓存
	indexConfig := config
	indexConfig.MaxSize = config.MaxSize / 4 // 25% 用于索引缓存
	indexConfig.TTL = 5 * time.Minute        // 索引缓存 TTL 更短
	indexCache := NewCache(indexConfig)

	return &AppCacheManager{
		cache:       cache,
		storage:     storage,
		bucketCache: bucketCache,
		indexCache:  indexCache,
	}
}

// GetApp 获取应用信息（带缓存）
func (m *AppCacheManager) GetApp(ctx context.Context, name string) (*store.AppManifest, error) {
	key := fmt.Sprintf("app:%s", name)

	// 尝试从缓存获取
	if cached, ok := m.cache.Get(key); ok {
		if app, ok := cached.(*store.AppManifest); ok {
			return app, nil
		}
	}

	// 从存储加载
	app, err := m.storage.GetApp(ctx, name)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.cache.Set(key, app)
	return app, nil
}

// SaveApp 保存应用信息（更新缓存）
func (m *AppCacheManager) SaveApp(ctx context.Context, app *store.AppManifest) error {
	key := fmt.Sprintf("app:%s", app.Name)

	// 保存到存储
	if err := m.storage.SaveApp(ctx, app); err != nil {
		return fmt.Errorf("保存应用 [%s] 到存储失败：%w", app.Name, err)
	}

	// 更新缓存
	m.cache.Set(key, app)
	return nil
}

// DeleteApp 删除应用信息（清除缓存）
func (m *AppCacheManager) DeleteApp(ctx context.Context, name string) error {
	key := fmt.Sprintf("app:%s", name)

	// 从存储删除
	if err := m.storage.DeleteApp(ctx, name); err != nil {
		return fmt.Errorf("从存储删除应用 [%s] 失败：%w", name, err)
	}

	// 清除缓存
	m.cache.Delete(key)
	return nil
}

// GetBucket 获取 Bucket 信息（带缓存）
func (m *AppCacheManager) GetBucket(ctx context.Context, name string) (*store.BucketConfig, error) {
	key := fmt.Sprintf("bucket:%s", name)

	// 尝试从缓存获取
	if cached, ok := m.bucketCache.Get(key); ok {
		if bucket, ok := cached.(*store.BucketConfig); ok {
			return bucket, nil
		}
	}

	// 从存储加载
	bucket, err := m.storage.GetBucket(ctx, name)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.bucketCache.Set(key, bucket)
	return bucket, nil
}

// GetDepsIndex 获取依赖索引（带缓存）
func (m *AppCacheManager) GetDepsIndex(ctx context.Context) (*store.DepsIndex, error) {
	key := "index:deps"

	// 尝试从缓存获取
	if cached, ok := m.indexCache.Get(key); ok {
		if index, ok := cached.(*store.DepsIndex); ok {
			return index, nil
		}
	}

	// 从存储加载
	index, err := m.storage.GetDepsIndex(ctx)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.indexCache.Set(key, index)
	return index, nil
}

// Close 关闭缓存管理器
func (m *AppCacheManager) Close() {
	m.cache.Close()
	m.bucketCache.Close()
	m.indexCache.Close()
}

// GetStats 获取所有缓存统计
func (m *AppCacheManager) GetStats() map[string]CacheStats {
	return map[string]CacheStats{
		"app":    m.cache.Stats(),
		"bucket": m.bucketCache.Stats(),
		"index":  m.indexCache.Stats(),
	}
}
