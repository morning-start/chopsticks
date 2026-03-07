package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"chopsticks/core/store"
)

func TestLRUCache_Basic(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024 * 1024, // 1MB
		MaxEntries: 100,
		TTL:        5 * time.Minute,
	}

	cache := NewCache(config)
	defer cache.Close()

	// 测试 Set 和 Get
	key := "test-key"
	value := "test-value"

	err := cache.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	got, exists := cache.Get(key)
	if !exists {
		t.Fatal("Cache miss for existing key")
	}

	if got != value {
		t.Fatalf("Expected %v, got %v", value, got)
	}

	// 测试删除
	deleted := cache.Delete(key)
	if !deleted {
		t.Fatal("Delete returned false for existing key")
	}

	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Cache hit for deleted key")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	config := CacheConfig{
		MaxSize:         1024 * 1024,
		MaxEntries:      100,
		TTL:             100 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
	}

	cache := NewCache(config)
	defer cache.Close()

	key := "ttl-key"
	value := "ttl-value"

	cache.Set(key, value)

	// 立即获取应该成功
	_, exists := cache.Get(key)
	if !exists {
		t.Fatal("Cache miss before TTL expiration")
	}

	// 等待 TTL 过期
	time.Sleep(200 * time.Millisecond)

	// 获取应该失败
	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Cache hit after TTL expiration")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024, // 1KB
		MaxEntries: 10,
		TTL:        5 * time.Minute,
	}

	cache := NewCache(config)
	defer cache.Close()

	// 填入超过最大条目数的数据
	for i := 0; i < 20; i++ {
		key := string(rune('a' + i))
		value := string(rune('0' + i))
		cache.Set(key, value)
	}

	stats := cache.Stats()
	if stats.Entries > config.MaxEntries {
		t.Fatalf("Cache entries %d exceeds max %d", stats.Entries, config.MaxEntries)
	}

	// 检查是否有淘汰
	if stats.Evictions == 0 {
		t.Fatal("Expected some evictions")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	config := CacheConfig{
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxEntries: 1000,
		TTL:        5 * time.Minute,
	}

	cache := NewCache(config)
	defer cache.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	opsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a'+id%26)) + string(rune('0'+j%10))
				value := id*1000 + j

				cache.Set(key, value)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	stats := cache.Stats()
	if stats.Hits+stats.Misses == 0 {
		t.Fatal("Expected some cache operations")
	}
}

func TestAppCacheManager(t *testing.T) {
	// 创建临时存储
	tempDir := t.TempDir()
	storage, err := store.NewFSStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	config := DefaultCacheConfig()
	config.MaxSize = 10 * 1024 * 1024
	config.MaxEntries = 100

	cacheMgr := NewAppCacheManager(storage, config)
	defer cacheMgr.Close()

	ctx := context.Background()

	// 创建测试应用
	app := &store.AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now(),
	}

	// 保存应用
	err = cacheMgr.SaveApp(ctx, app)
	if err != nil {
		t.Fatalf("Failed to save app: %v", err)
	}

	// 获取应用（应该从缓存）
	got, err := cacheMgr.GetApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("Failed to get app: %v", err)
	}

	if got.Name != app.Name {
		t.Fatalf("Expected app name %s, got %s", app.Name, got.Name)
	}

	// 删除应用
	err = cacheMgr.DeleteApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("Failed to delete app: %v", err)
	}

	// 验证删除
	_, err = cacheMgr.GetApp(ctx, "test-app")
	if err == nil {
		t.Fatal("Expected error for deleted app")
	}
}

func TestLazyIndex(t *testing.T) {
	tempDir := t.TempDir()
	loader := &DepsIndexLoader{}

	lazyIndex := NewLazyIndex(tempDir, "deps", loader)
	defer lazyIndex.Close()

	ctx := context.Background()

	// 第一次获取应该加载
	index, err := lazyIndex.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get lazy index: %v", err)
	}

	if index == nil {
		t.Fatal("Expected non-nil index")
	}

	// 第二次获取应该从缓存
	index2, err := lazyIndex.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get cached index: %v", err)
	}

	if index2 == nil {
		t.Fatal("Expected cached index")
	}
}

func TestBatchReader(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := store.NewFSStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	config := DefaultCacheConfig()
	config.MaxSize = 10 * 1024 * 1024

	reader := NewBatchReader(storage, config)
	defer reader.Close()

	ctx := context.Background()

	// 创建测试数据
	for i := 0; i < 5; i++ {
		app := &store.AppManifest{
			Name:           string(rune('a' + i)),
			Bucket:         "main",
			CurrentVersion: "1.0.0",
			InstalledAt:    time.Now(),
		}
		storage.SaveApp(ctx, app)
	}

	// 批量读取
	names := []string{"a", "b", "c", "d", "e"}
	result := reader.ReadApps(ctx, names)

	if len(result.Items) != 5 {
		t.Fatalf("Expected 5 items, got %d", len(result.Items))
	}

	if result.CacheMisses != 5 {
		t.Fatalf("Expected 5 cache misses, got %d", result.CacheMisses)
	}

	// 再次读取应该命中缓存
	result2 := reader.ReadApps(ctx, names)
	if result2.CacheHits != 5 {
		t.Fatalf("Expected 5 cache hits, got %d", result2.CacheHits)
	}
}

func TestPerformanceAnalyzer(t *testing.T) {
	analyzer := NewPerformanceAnalyzer(100*time.Millisecond, 10)
	analyzer.Start()
	defer analyzer.Stop()

	// 等待收集样本
	time.Sleep(300 * time.Millisecond)

	samples := analyzer.GetSamples(5)
	if len(samples) == 0 {
		t.Fatal("Expected some performance samples")
	}
}

func TestIOTracker(t *testing.T) {
	tracker := NewIOTracker(1000)

	// 记录一些操作
	tracker.Record(IOOperation{
		Type:      "read",
		Path:      "/test/file.txt",
		Size:      1024,
		Duration:  10 * time.Millisecond,
		Timestamp: time.Now(),
		Success:   true,
	})

	tracker.Record(IOOperation{
		Type:      "batch_read",
		Path:      "/test/batch",
		Size:      5 * 1024,
		Duration:  50 * time.Millisecond,
		Timestamp: time.Now(),
		Success:   true,
	})

	stats := tracker.GetStats()
	if stats.ReadOps != 2 {
		t.Fatalf("Expected 2 read ops, got %d", stats.ReadOps)
	}

	if stats.BatchReads != 1 {
		t.Fatalf("Expected 1 batch read, got %d", stats.BatchReads)
	}
}

func TestHealthChecker(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024 * 1024,
		MaxEntries: 100,
		TTL:        5 * time.Minute,
	}

	cache := NewCache(config)
	defer cache.Close()

	// 填入一些数据
	for i := 0; i < 50; i++ {
		key := string(rune('a' + i%26))
		value := i
		cache.Set(key, value)
	}

	checker := NewHealthChecker(cache)
	status := checker.Check()

	// 应该基本健康
	if !status.Healthy {
		t.Logf("Health issues: %v", status.Issues)
		t.Logf("Health warnings: %v", status.Warnings)
	}
}

func TestCacheStats(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024 * 1024,
		MaxEntries: 100,
		TTL:        5 * time.Minute,
	}

	cache := NewCache(config)
	defer cache.Close()

	// 一些操作
	for i := 0; i < 20; i++ {
		key := string(rune('a' + i))
		value := i
		cache.Set(key, value)
		cache.Get(key)
	}

	stats := cache.Stats()

	if stats.Hits == 0 {
		t.Fatal("Expected some cache hits")
	}

	if stats.Entries != 20 {
		t.Fatalf("Expected 20 entries, got %d", stats.Entries)
	}

	t.Logf("Cache Stats: Hits=%d, Misses=%d, HitRate=%.2f%%, Entries=%d",
		stats.Hits, stats.Misses, stats.HitRate, stats.Entries)
}
