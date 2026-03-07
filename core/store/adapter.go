// Package store 提供存储适配器，用于向后兼容旧的 Storage 接口。
package store

import (
	"context"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"
)

// LegacyStorage 提供向后兼容的 Storage 接口。
//
// Deprecated: 请直接使用 Storage 接口。
type LegacyStorage interface {
	// 已安装的应用
	SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error
	GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error)
	DeleteInstalledApp(ctx context.Context, name string) error
	ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error)
	IsInstalled(ctx context.Context, name string) (bool, error)

	// 软件源
	SaveBucket(ctx context.Context, b *manifest.BucketConfig) error
	GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error)

	Close() error
}

// StorageAdapter 将新的 Storage 接口适配到旧的 LegacyStorage 接口。
type StorageAdapter struct {
	storage     Storage
	installBase string // 安装基础目录
}

// 编译时接口检查。
var _ LegacyStorage = (*StorageAdapter)(nil)

// NewStorageAdapter 创建新的存储适配器。
func NewStorageAdapter(storage Storage, installBase string) *StorageAdapter {
	return &StorageAdapter{
		storage:     storage,
		installBase: installBase,
	}
}

// SaveInstalledApp 保存已安装应用（向后兼容）。
func (a *StorageAdapter) SaveInstalledApp(ctx context.Context, app *manifest.InstalledApp) error {
	// 转换为 AppManifest 格式
	manifest := &AppManifest{
		Name:              app.Name,
		Bucket:            app.Bucket,
		CurrentVersion:    app.Version,
		InstalledVersions: []string{app.Version},
		InstalledAt:       app.InstalledAt,
	}
	return a.storage.SaveApp(ctx, manifest)
}

// GetInstalledApp 获取已安装应用（向后兼容）。
func (a *StorageAdapter) GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error) {
	appManifest, err := a.storage.GetApp(ctx, name)
	if err != nil {
		return nil, err
	}

	// 转换为 manifest.InstalledApp 格式
	return &manifest.InstalledApp{
		Name:        appManifest.Name,
		Version:     appManifest.CurrentVersion,
		Bucket:      appManifest.Bucket,
		InstallDir:  filepath.Join(a.installBase, appManifest.Name),
		InstalledAt: appManifest.InstalledAt,
		UpdatedAt:   appManifest.InstalledAt, // 简化处理
	}, nil
}

// DeleteInstalledApp 删除已安装应用（向后兼容）。
func (a *StorageAdapter) DeleteInstalledApp(ctx context.Context, name string) error {
	return a.storage.DeleteApp(ctx, name)
}

// ListInstalledApps 列出所有已安装应用（向后兼容）。
func (a *StorageAdapter) ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error) {
	appManifests, err := a.storage.ListApps(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 manifest.InstalledApp 格式
	apps := make([]*manifest.InstalledApp, 0, len(appManifests))
	for _, appManifest := range appManifests {
		apps = append(apps, &manifest.InstalledApp{
			Name:        appManifest.Name,
			Version:     appManifest.CurrentVersion,
			Bucket:      appManifest.Bucket,
			InstallDir:  filepath.Join(a.installBase, appManifest.Name),
			InstalledAt: appManifest.InstalledAt,
			UpdatedAt:   appManifest.InstalledAt,
		})
	}
	return apps, nil
}

// IsInstalled 检查应用是否已安装（向后兼容）。
func (a *StorageAdapter) IsInstalled(ctx context.Context, name string) (bool, error) {
	return a.storage.IsInstalled(ctx, name)
}

// SaveBucket 保存软件源配置（向后兼容）。
func (a *StorageAdapter) SaveBucket(ctx context.Context, bucket *manifest.BucketConfig) error {
	// 转换为 store.BucketConfig 格式
	storeBucket := &BucketConfig{
		ID:          bucket.ID,
		Name:        bucket.Name,
		Author:      bucket.Author,
		Description: bucket.Description,
		Homepage:    bucket.Homepage,
		License:     bucket.License,
		Repository:  bucket.Repository,
		AddedAt:     time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	return a.storage.SaveBucket(ctx, storeBucket)
}

// GetBucket 获取软件源配置（向后兼容）。
func (a *StorageAdapter) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	storeBucket, err := a.storage.GetBucket(ctx, name)
	if err != nil {
		return nil, err
	}

	// 转换为 manifest.BucketConfig 格式
	return &manifest.BucketConfig{
		ID:          storeBucket.ID,
		Name:        storeBucket.Name,
		Author:      storeBucket.Author,
		Description: storeBucket.Description,
		Homepage:    storeBucket.Homepage,
		License:     storeBucket.License,
		Repository:  storeBucket.Repository,
	}, nil
}

// DeleteBucket 删除软件源配置（向后兼容）。
func (a *StorageAdapter) DeleteBucket(ctx context.Context, name string) error {
	return a.storage.DeleteBucket(ctx, name)
}

// ListBuckets 列出所有软件源（向后兼容）。
func (a *StorageAdapter) ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error) {
	storeBuckets, err := a.storage.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 manifest.BucketConfig 格式
	buckets := make([]*manifest.BucketConfig, 0, len(storeBuckets))
	for _, storeBucket := range storeBuckets {
		buckets = append(buckets, &manifest.BucketConfig{
			ID:          storeBucket.ID,
			Name:        storeBucket.Name,
			Author:      storeBucket.Author,
			Description: storeBucket.Description,
			Homepage:    storeBucket.Homepage,
			License:     storeBucket.License,
			Repository:  storeBucket.Repository,
		})
	}
	return buckets, nil
}

// Close 关闭存储（向后兼容）。
func (a *StorageAdapter) Close() error {
	return a.storage.Close()
}
