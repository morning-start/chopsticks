package bucket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/infra/git"
	"chopsticks/pkg/errors"
)

var (
	ErrBucketNotFound      = errors.ErrBucketNotFound
	ErrBucketAlreadyExists = errors.ErrBucketAlreadyExists
	ErrInvalidURL          = errors.ErrInvalidBucketURL
)

// BucketManager 定义软件源管理器接口。
type BucketManager interface {
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

var _ BucketManager = (*manager)(nil)

// NewManager 创建软件源管理器
func NewManager(db store.Storage, config interface{}, bucketsDir string) BucketManager {return &manager{
		buckets:    make(map[string]*manifest.Bucket),
		db:         db,
		config:     config,
		bucketsDir: bucketsDir,
		git:        git.New(),
	}
}

func (m *manager) Add(ctx context.Context, name, url string, opts AddOptions) error {
	if name == "" {
		return errors.Newf(errors.KindInvalidInput, "bucket name is required")
	}
	if url == "" {
		return errors.Newf(errors.KindInvalidInput, "bucket URL is required")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err == nil {
		return errors.NewBucketAlreadyExists(name)
	}

	fmt.Printf("cloning bucket: %s -> %s\n", url, bucketPath)
	if err := m.git.Clone(ctx, url, bucketPath); err != nil {
		return errors.NewBucketLoadFailed(name, err)
	}

	bucketConfig := &manifest.BucketConfig{
		ID:   name,
		Name: name,
		Repository: manifest.RepositoryInfo{
			URL:    url,
			Branch: opts.Branch,
		},
		Description: "",
	}

	if err := m.db.SaveBucket(ctx, bucketConfig); err != nil {
		return errors.Wrap(err, "save bucket config")
	}

	fmt.Printf("✓ bucket %s added successfully\n", name)
	return nil
}

func (m *manager) Remove(ctx context.Context, name string, purge bool) error {
	if name == "" {
		return errors.Newf(errors.KindInvalidInput, "bucket name is required")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err != nil {
		return errors.NewBucketNotFound(name)
	}

	if purge {
		if err := os.RemoveAll(bucketPath); err != nil {
			return errors.Wrapf(err, "remove bucket directory %s", bucketPath)
		}
	}

	if err := m.db.DeleteBucket(ctx, name); err != nil {
		return errors.Wrap(err, "delete bucket config")
	}

	fmt.Printf("✓ bucket %s removed\n", name)
	return nil
}

func (m *manager) Update(ctx context.Context, name string) error {
	if name == "" {
		return errors.Newf(errors.KindInvalidInput, "bucket name is required")
	}

	bucketPath := filepath.Join(m.bucketsDir, name)
	if _, err := os.Stat(bucketPath); err != nil {
		return errors.NewBucketNotFound(name)
	}

	fmt.Printf("updating bucket: %s\n", name)
	if err := m.git.Pull(ctx, bucketPath); err != nil {
		return errors.NewBucketUpdateFailed(name, err)
	}

	fmt.Printf("✓ bucket %s updated\n", name)
	return nil
}

func (m *manager) UpdateAll(ctx context.Context) error {
	buckets, err := m.db.ListBuckets(ctx)
	if err != nil {
		return errors.Wrap(err, "list buckets")
	}

	for _, bucket := range buckets {
		if err := m.Update(ctx, bucket.ID); err != nil {
			fmt.Fprintf(os.Stderr, "update bucket %s failed: %v\n", bucket.ID, err)
		}
	}

	return nil
}

func (m *manager) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	if name == "" {
		return nil, errors.Newf(errors.KindInvalidInput, "bucket name is required")
	}

	bucket, err := m.db.GetBucket(ctx, name)
	if err != nil {
		return nil, errors.NewBucketNotFound(name)
	}

	return bucket, nil
}

func (m *manager) GetApp(ctx context.Context, bucket, name string) (*manifest.App, error) {
	bucketPath := filepath.Join(m.bucketsDir, bucket)
	if _, err := os.Stat(bucketPath); err != nil {
		return nil, errors.NewBucketNotFound(bucket)
	}

	loader := NewLoader()
	b, err := loader.Load(ctx, bucketPath)
	if err != nil {
		return nil, errors.NewBucketLoadFailed(bucket, err)
	}

	ref, ok := b.Apps[name]
	if !ok {
		return nil, errors.NewAppManifestNotFound(bucket, name)
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
		return nil, errors.NewBucketNotFound(bucket)
	}

	loader := NewLoader()
	b, err := loader.Load(ctx, bucketPath)
	if err != nil {
		return nil, errors.NewBucketLoadFailed(bucket, err)
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
