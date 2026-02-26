// Package app 提供应用（软件包）管理功能。
package app

import (
	"context"
	"errors"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

// 常用 sentinel 错误。
var (
	// ErrAppNotFound 表示指定的应用不存在。
	ErrAppNotFound = errors.New("app not found")
	// ErrAppAlreadyInstalled 表示应用已安装。
	ErrAppAlreadyInstalled = errors.New("app already installed")
	// ErrVersionNotFound 表示指定的版本不存在。
	ErrVersionNotFound = errors.New("version not found")
	// ErrDependencyConflict 表示依赖冲突。
	ErrDependencyConflict = errors.New("dependency conflict")
)

// Manager 定义应用管理接口。
type Manager interface {
	Install(ctx context.Context, spec InstallSpec, opts InstallOptions) error
	Remove(ctx context.Context, name string, opts RemoveOptions) error
	Update(ctx context.Context, name string, opts UpdateOptions) error
	UpdateAll(ctx context.Context, opts UpdateOptions) error
	Switch(ctx context.Context, name, version string) error
	ListInstalled() ([]*manifest.InstalledApp, error)
	Info(ctx context.Context, bucket, name string) (*manifest.AppInfo, error)
	Search(ctx context.Context, query string, bucket string) ([]SearchResult, error)
}

// InstallSpec 定义安装规格。
type InstallSpec struct {
	Bucket  string // 软件源名称
	Name    string // 应用名称
	Version string // 版本号
}

// RemoveOptions 包含卸载选项。
type RemoveOptions struct {
	Purge bool // 彻底清除
}

// UpdateOptions 包含更新选项。
type UpdateOptions struct {
	Force bool // 强制更新
}

// SearchResult 表示搜索结果。
type SearchResult struct {
	Bucket string           // 软件源名称
	App    *manifest.AppRef // 应用引用
}

// manager 是 Manager 的实现。
type manager struct {
	bucketMgr bucket.Manager
	storage   store.Storage
	installer Installer
	config    interface{}
}

var _ Manager = (*manager)(nil)

func NewManager(bucketMgr bucket.Manager, storage store.Storage, installer Installer, config interface{}) Manager {
	return &manager{
		bucketMgr: bucketMgr,
		storage:   storage,
		installer: installer,
		config:    config,
	}
}

func (m *manager) Install(ctx context.Context, spec InstallSpec, opts InstallOptions) error {
	// TODO: 实现
	return nil
}

func (m *manager) Remove(ctx context.Context, name string, opts RemoveOptions) error {
	// TODO: 实现
	return nil
}

func (m *manager) Update(ctx context.Context, name string, opts UpdateOptions) error {
	// TODO: 实现
	return nil
}

func (m *manager) UpdateAll(ctx context.Context, opts UpdateOptions) error {
	// TODO: 实现
	return nil
}

func (m *manager) Switch(ctx context.Context, name, version string) error {
	// TODO: 实现
	return nil
}

func (m *manager) ListInstalled() ([]*manifest.InstalledApp, error) {
	// TODO: 实现
	return nil, nil
}

func (m *manager) Info(ctx context.Context, bucket, name string) (*manifest.AppInfo, error) {
	// TODO: 实现
	return nil, nil
}

func (m *manager) Search(ctx context.Context, query string, bucket string) ([]SearchResult, error) {
	// TODO: 实现
	return nil, nil
}
