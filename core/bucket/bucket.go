package bucket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/infra/git"
)

var (
	ErrBucketNotFound     = fmt.Errorf("bucket not found")
	ErrBucketAlreadyExists = fmt.Errorf("bucket already exists")
	ErrInvalidURL        = fmt.Errorf("invalid bucket URL")
)

type Manager interface {
	Add(ctx context.Context, name, url string, opts AddOptions) error
	Remove(ctx context.Context, name string, purge bool) error
	Update(ctx context.Context, name string) error
	UpdateAll(ctx context.Context) error
	GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error)
	GetApp(ctx context.Context, bucket, name string) (*manifest.App, error)
	ListApps(ctx context.Context, bucket string) (map[string]*manifest.AppRef, error)
	ListBuckets(ctx context.Context) ([]string, error)
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
}

type AddOptions struct {
	Branch string
	Depth  int
}

type SearchOptions struct {
	Bucket   string
	Category string
	Tags     []string
}

type SearchResult struct {
	Bucket string
	App    *manifest.AppRef
}

type manager struct {
	buckets    map[string]*manifest.Bucket
	db         store.Storage
	config     interface{}
	bucketsDir string
	git        git.Git
}

var _ Manager = (*manager)(nil)

func NewManager(db store.Storage, config interface{}, bucketsDir string) Manager {
	return &manager{
		buckets:    make(map[string]*manifest.Bucket),
		db:         db,
		config:     config,
		bucketsDir: bucketsDir,
		git:        git.New(),
	}
}

func (m *manager) Add(ctx context.Context, name, url string, opts AddOptions) error {
	if name == "" {
		return fmt.Errorf("软件源名称不能为空")
	}
	if url == "" {
		return fmt.Errorf("软件源 URL 不能为空")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err == nil {
		return fmt.Errorf("软件源 %s 已存在", name)
	}

	fmt.Printf("克隆软件源: %s -> %s\n", url, bucketPath)
	if err := m.git.Clone(ctx, url, bucketPath); err != nil {
		return fmt.Errorf("克隆软件源失败: %w", err)
	}

	bucketConfig := &manifest.BucketConfig{
		ID:          name,
		Name:        name,
		Repository: manifest.RepositoryInfo{
			URL:    url,
			Branch: opts.Branch,
		},
		Description: "",
	}

	if err := m.db.SaveBucket(ctx, bucketConfig); err != nil {
		return fmt.Errorf("保存软件源配置失败: %w", err)
	}

	fmt.Printf("✓ 软件源 %s 添加成功\n", name)
	return nil
}

func (m *manager) Remove(ctx context.Context, name string, purge bool) error {
	if name == "" {
		return fmt.Errorf("软件源名称不能为空")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err != nil {
		return fmt.Errorf("软件源 %s 不存在", name)
	}

	if purge {
		if err := os.RemoveAll(bucketPath); err != nil {
			return fmt.Errorf("删除软件源目录失败: %w", err)
		}
	}

	if err := m.db.DeleteBucket(ctx, name); err != nil {
		return fmt.Errorf("删除软件源配置失败: %w", err)
	}

	fmt.Printf("✓ 软件源 %s 已删除\n", name)
	return nil
}

func (m *manager) Update(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("软件源名称不能为空")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err != nil {
		return fmt.Errorf("软件源 %s 不存在", name)
	}

	fmt.Printf("更新软件源: %s\n", name)
	if err := m.git.Pull(ctx, bucketPath); err != nil {
		return fmt.Errorf("更新软件源失败: %w", err)
	}

	fmt.Printf("✓ 软件源 %s 更新成功\n", name)
	return nil
}

func (m *manager) UpdateAll(ctx context.Context) error {
	buckets, err := m.db.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("获取软件源列表失败: %w", err)
	}

	for _, bucket := range buckets {
		if err := m.Update(ctx, bucket.ID); err != nil {
			fmt.Fprintf(os.Stderr, "更新软件源 %s 失败: %v\n", bucket.ID, err)
		}
	}

	return nil
}

func (m *manager) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	if name == "" {
		return nil, fmt.Errorf("软件源名称不能为空")
	}

	bucket, err := m.db.GetBucket(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("软件源 %s 不存在: %w", name, err)
	}

	return bucket, nil
}

func (m *manager) GetApp(ctx context.Context, bucket, name string) (*manifest.App, error) {
	bucketPath := filepath.Join(m.bucketsDir, bucket)
	if _, err := os.Stat(bucketPath); err != nil {
		return nil, fmt.Errorf("软件源 %s 不存在", bucket)
	}

	loader := NewLoader()
	b, err := loader.Load(bucketPath)
	if err != nil {
		return nil, fmt.Errorf("加载软件源失败: %w", err)
	}

	ref, ok := b.Apps[name]
	if !ok {
		return nil, fmt.Errorf("应用 %s 不存在", name)
	}

	return &manifest.App{
		Script: &manifest.AppScript{
			Name:        ref.Name,
			Description: ref.Description,
			Bucket:      bucket,
		},
		Meta: &manifest.AppMeta{
			Version: ref.Version,
		},
		Ref: ref,
	}, nil
}

func (m *manager) ListApps(ctx context.Context, bucket string) (map[string]*manifest.AppRef, error) {
	bucketPath := filepath.Join(m.bucketsDir, bucket)
	if _, err := os.Stat(bucketPath); err != nil {
		return nil, fmt.Errorf("软件源 %s 不存在", bucket)
	}

	loader := NewLoader()
	b, err := loader.Load(bucketPath)
	if err != nil {
		return nil, fmt.Errorf("加载软件源失败: %w", err)
	}

	return b.Apps, nil
}

func (m *manager) ListBuckets(ctx context.Context) ([]string, error) {
	buckets, err := m.db.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(buckets))
	for i, b := range buckets {
		names[i] = b.ID
	}

	if len(names) == 0 {
		names = []string{"main"}
	}

	return names, nil
}

func (m *manager) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult

	bucketsToSearch := []string{opts.Bucket}
	if opts.Bucket == "" {
		buckets, err := m.ListBuckets(ctx)
		if err != nil {
			return nil, err
		}
		bucketsToSearch = buckets
	}

	for _, bucket := range bucketsToSearch {
		apps, err := m.ListApps(ctx, bucket)
		if err != nil {
			continue
		}

		for _, app := range apps {
			if matchesSearchQuery(app, query, opts) {
				results = append(results, SearchResult{
					Bucket: bucket,
					App:    app,
				})
			}
		}
	}

	return results, nil
}

func matchesSearchQuery(app *manifest.AppRef, query string, opts SearchOptions) bool {
	lowerQuery := lower(query)
	if !contains(lower(app.Name), lowerQuery) && !contains(lower(app.Description), lowerQuery) {
		return false
	}

	if opts.Category != "" && app.Category != opts.Category {
		return false
	}

	if len(opts.Tags) > 0 {
		hasTag := false
		for _, tag := range opts.Tags {
			for _, appTag := range app.Tags {
				if tag == appTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func lower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
