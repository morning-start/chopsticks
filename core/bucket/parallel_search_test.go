package bucket

import (
	"context"
	"testing"
	"time"

	"chopsticks/core/manifest"
)

// mockManager 是一个模拟的 Manager 实现
type mockManager struct {
	buckets map[string]map[string]*manifest.AppRef
}

func newMockManager() *mockManager {
	return &mockManager{
		buckets: map[string]map[string]*manifest.AppRef{
			"main": {
				"git": {
					Name:        "git",
					Description: "Distributed version control system",
					Category:    "devel",
					Tags:        []string{"vcs", "scm"},
				},
				"go": {
					Name:        "go",
					Description: "Go programming language",
					Category:    "devel",
					Tags:        []string{"compiler", "language"},
				},
				"python": {
					Name:        "python",
					Description: "Python programming language",
					Category:    "devel",
					Tags:        []string{"interpreter", "language"},
				},
			},
			"extras": {
				"nodejs": {
					Name:        "nodejs",
					Description: "JavaScript runtime",
					Category:    "devel",
					Tags:        []string{"javascript", "runtime"},
				},
				"rust": {
					Name:        "rust",
					Description: "Rust programming language",
					Category:    "devel",
					Tags:        []string{"compiler", "language"},
				},
			},
			"games": {
				"steam": {
					Name:        "steam",
					Description: "Game distribution platform",
					Category:    "games",
					Tags:        []string{"launcher", "store"},
				},
			},
		},
	}
}

func (m *mockManager) Add(ctx context.Context, name, url string, opts AddOptions) error {
	return nil
}

func (m *mockManager) Remove(ctx context.Context, name string, purge bool) error {
	return nil
}

func (m *mockManager) Update(ctx context.Context, name string) error {
	return nil
}

func (m *mockManager) UpdateAll(ctx context.Context) error {
	return nil
}

func (m *mockManager) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	return nil, nil
}

func (m *mockManager) GetApp(ctx context.Context, bucket, name string) (*manifest.App, error) {
	return nil, nil
}

func (m *mockManager) ListApps(ctx context.Context, bucket string) (map[string]*manifest.AppRef, error) {
	apps, ok := m.buckets[bucket]
	if !ok {
		return nil, ErrBucketNotFound
	}
	return apps, nil
}

func (m *mockManager) ListBuckets(ctx context.Context) ([]string, error) {
	buckets := make([]string, 0, len(m.buckets))
	for name := range m.buckets {
		buckets = append(buckets, name)
	}
	return buckets, nil
}

func (m *mockManager) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult

	bucketsToSearch := []string{opts.Bucket}
	if opts.Bucket == "" {
		bucketsToSearch, _ = m.ListBuckets(ctx)
	}

	for _, bucketName := range bucketsToSearch {
		apps, err := m.ListApps(ctx, bucketName)
		if err != nil {
			continue
		}

		for _, app := range apps {
			if matchesSearchQuery(app, query, opts) {
				results = append(results, SearchResult{
					Bucket: bucketName,
					App:    app,
				})
			}
		}
	}

	return results, nil
}

func TestNewParallelSearcher(t *testing.T) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 5)

	if searcher == nil {
		t.Fatal("NewParallelSearcher returned nil")
	}

	if searcher.maxWorkers != 5 {
		t.Errorf("maxWorkers = %d, want 5", searcher.maxWorkers)
	}

	if searcher.cache == nil {
		t.Error("cache should not be nil")
	}
}

func TestNewParallelSearcher_DefaultWorkers(t *testing.T) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 0)

	if searcher.maxWorkers != 10 {
		t.Errorf("maxWorkers = %d, want 10", searcher.maxWorkers)
	}
}

func TestParallelSearcher_SearchAllBuckets(t *testing.T) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 3)
	ctx := context.Background()

	results, err := searcher.SearchAllBuckets(ctx, "go", SearchOptions{})
	if err != nil {
		t.Fatalf("SearchAllBuckets failed: %v", err)
	}

	// 应该找到 "go" 和 "cargo" (如果存在)
	found := false
	for _, r := range results {
		if r.App.Name == "go" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'go' in search results")
	}

	t.Logf("Found %d results for 'go'", len(results))
}

func TestParallelSearcher_SearchAllBuckets_WithCategory(t *testing.T) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 3)
	ctx := context.Background()

	results, err := searcher.SearchAllBuckets(ctx, "", SearchOptions{
		Category: "games",
	})
	if err != nil {
		t.Fatalf("SearchAllBuckets failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].App.Name != "steam" {
		t.Errorf("Expected 'steam', got '%s'", results[0].App.Name)
	}
}

func TestParallelSearcher_SearchWithCache(t *testing.T) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 3)
	ctx := context.Background()

	// 第一次搜索
	results1, err := searcher.SearchWithCache(ctx, "python", SearchOptions{})
	if err != nil {
		t.Fatalf("First search failed: %v", err)
	}

	// 第二次搜索（应该从缓存获取）
	results2, err := searcher.SearchWithCache(ctx, "python", SearchOptions{})
	if err != nil {
		t.Fatalf("Second search failed: %v", err)
	}

	// 验证结果相同
	if len(results1) != len(results2) {
		t.Errorf("Cached results differ: first=%d, second=%d", len(results1), len(results2))
	}

	// 检查缓存统计
	stats := searcher.GetCacheStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}

	t.Logf("Cache stats: hits=%d, misses=%d, hitRate=%.2f", stats.Hits, stats.Misses, stats.HitRate)
}

func TestSearchCache(t *testing.T) {
	cache := NewSearchCache(100 * time.Millisecond)

	results := []SearchResult{
		{Bucket: "main", App: &manifest.AppRef{Name: "test"}},
	}

	// 设置缓存
	cache.Set("test", results)

	// 立即获取
	cached, found := cache.Get("test")
	if !found {
		t.Error("Expected to find cached result")
	}
	if len(cached) != len(results) {
		t.Error("Cached result mismatch")
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)

	// 再次获取（应该过期）
	_, found = cache.Get("test")
	if found {
		t.Error("Expected cache entry to expire")
	}
}

func TestSearchCache_Clear(t *testing.T) {
	cache := NewSearchCache(5 * time.Minute)

	cache.Set("key1", []SearchResult{})
	cache.Set("key2", []SearchResult{})

	stats := cache.GetStats()
	if stats.EntryCount != 2 {
		t.Errorf("Expected 2 entries, got %d", stats.EntryCount)
	}

	cache.Clear()

	stats = cache.GetStats()
	if stats.EntryCount != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.EntryCount)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		query string
		opts  SearchOptions
		want  string
	}{
		{
			query: "git",
			opts:  SearchOptions{},
			want:  "git",
		},
		{
			query: "git",
			opts:  SearchOptions{Bucket: "main"},
			want:  "git:bucket=main",
		},
		{
			query: "git",
			opts:  SearchOptions{Category: "devel"},
			want:  "git:category=devel",
		},
		{
			query: "git",
			opts:  SearchOptions{Tags: []string{"vcs", "scm"}},
			want:  "git:tags=vcs,scm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := generateCacheKey(tt.query, tt.opts)
			if got != tt.want {
				t.Errorf("generateCacheKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterResults(t *testing.T) {
	results := []SearchResult{
		{Bucket: "main", App: &manifest.AppRef{Name: "git", Category: "devel"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "steam", Category: "games"}},
		{Bucket: "extras", App: &manifest.AppRef{Name: "nodejs", Category: "devel"}},
	}

	filtered := FilterResults(results, func(app *manifest.AppRef) bool {
		return app.Category == "devel"
	})

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(filtered))
	}
}

func TestSortResultsByName(t *testing.T) {
	results := []SearchResult{
		{Bucket: "main", App: &manifest.AppRef{Name: "python"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "git"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "go"}},
	}

	SortResultsByName(results)

	expected := []string{"git", "go", "python"}
	for i, r := range results {
		if r.App.Name != expected[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expected[i], r.App.Name)
		}
	}
}

func TestDeduplicateResults(t *testing.T) {
	results := []SearchResult{
		{Bucket: "main", App: &manifest.AppRef{Name: "git"}},
		{Bucket: "extras", App: &manifest.AppRef{Name: "git"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "go"}},
	}

	unique := DeduplicateResults(results)

	// 注意：DeduplicateResults 是基于 bucket+name 去重的
	// 所以 main/git 和 extras/git 是不同的条目
	// 这个测试验证去重逻辑是否正确工作
	if len(unique) != 3 {
		t.Errorf("Expected 3 unique results (different buckets), got %d", len(unique))
	}

	// 测试真正的重复（相同 bucket 和 name）
	resultsWithDup := []SearchResult{
		{Bucket: "main", App: &manifest.AppRef{Name: "git"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "git"}},
		{Bucket: "main", App: &manifest.AppRef{Name: "go"}},
	}

	unique2 := DeduplicateResults(resultsWithDup)
	if len(unique2) != 2 {
		t.Errorf("Expected 2 unique results after dedup, got %d", len(unique2))
	}
}

func TestDefaultParallelSearchOptions(t *testing.T) {
	opts := DefaultParallelSearchOptions()

	if opts.MaxWorkers != 10 {
		t.Errorf("MaxWorkers = %d, want 10", opts.MaxWorkers)
	}

	if !opts.UseCache {
		t.Error("UseCache should be true")
	}

	if opts.CacheTTL != 5*time.Minute {
		t.Errorf("CacheTTL = %v, want 5m", opts.CacheTTL)
	}
}

func BenchmarkParallelSearch(b *testing.B) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searcher.SearchAllBuckets(ctx, "go", SearchOptions{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedSearch(b *testing.B) {
	mgr := newMockManager()
	searcher := NewParallelSearcher(mgr, 10)
	ctx := context.Background()

	// 预热缓存
	searcher.SearchWithCache(ctx, "go", SearchOptions{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searcher.SearchWithCache(ctx, "go", SearchOptions{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
