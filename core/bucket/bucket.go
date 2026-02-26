// Package bucket 提供软件源（Bucket）管理功能。
package bucket

import (
	"context"
	"errors"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

// 常用 sentinel 错误。
var (
	// ErrBucketNotFound 表示指定的软件源不存在。
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrBucketAlreadyExists 表示软件源已存在。
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	// ErrInvalidURL 表示无效的软件源地址。
	ErrInvalidURL = errors.New("invalid bucket URL")
)

// Manager 定义软件源管理接口。
type Manager interface {
	Add(ctx context.Context, name, url string, opts AddOptions) error
	Remove(ctx context.Context, name string, purge bool) error
	Update(ctx context.Context, name string) error
	UpdateAll(ctx context.Context) error
	List() []*manifest.Bucket
	Get(name string) (*manifest.Bucket, error)
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
}

// AddOptions 包含添加软件源的选项。
type AddOptions struct {
	Branch string // 分支
	Depth  int    // 克隆深度
}

// SearchOptions 包含搜索选项。
type SearchOptions struct {
	Bucket   string   // 指定软件源
	Category string   // 分类
	Tags     []string // 标签
}

// SearchResult 表示搜索结果。
type SearchResult struct {
	Bucket string           // 软件源名称
	App    *manifest.AppRef // 应用引用
}

// manager 是 Manager 的实现。
type manager struct {
	buckets map[string]*manifest.Bucket
	db      store.Storage
	config  interface{} // 使用 interface{} 避免循环依赖，实际为 *app.Config
}

// 编译时接口检查。
var _ Manager = (*manager)(nil)

// NewManager 创建新的 Manager。
func NewManager(db store.Storage, config interface{}) Manager {
	return &manager{
		buckets: make(map[string]*manifest.Bucket),
		db:      db,
		config:  config,
	}
}

func (m *manager) Add(ctx context.Context, name, url string, opts AddOptions) error {
	// TODO: 实现
	return nil
}

func (m *manager) Remove(ctx context.Context, name string, purge bool) error {
	// TODO: 实现
	return nil
}

func (m *manager) Update(ctx context.Context, name string) error {
	// TODO: 实现
	return nil
}

func (m *manager) UpdateAll(ctx context.Context) error {
	// TODO: 实现
	return nil
}

func (m *manager) List() []*manifest.Bucket {
	// TODO: 实现
	return nil
}

func (m *manager) Get(name string) (*manifest.Bucket, error) {
	// TODO: 实现
	return nil, nil
}

func (m *manager) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// TODO: 实现
	return nil, nil
}
