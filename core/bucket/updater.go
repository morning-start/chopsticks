package bucket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"
)

// Updater 定义软件源更新器接口。
type Updater interface {
	Update(ctx context.Context, bucket *manifest.Bucket) error
	UpdateApp(ctx context.Context, bucket *manifest.Bucket, appName string) error
	CheckUpdates(ctx context.Context, bucket *manifest.Bucket) ([]string, error)
}

// updater 是 Updater 的实现。
type updater struct {
	loader Loader
}

// 编译时接口检查。
var _ Updater = (*updater)(nil)

// NewUpdater 创建新的 Updater。
func NewUpdater(loader Loader) Updater {
	return &updater{
		loader: loader,
	}
}

// Update 更新整个软件源。
func (u *updater) Update(ctx context.Context, bucket *manifest.Bucket) error {
	// 对于 Git 仓库，执行 git pull
	gitDir := filepath.Join(bucket.Path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// TODO: 执行 git pull
		return fmt.Errorf("Git 更新暂未实现")
	}

	// 重新扫描应用
	apps, err := u.loader.ScanApps(bucket.Path)
	if err != nil {
		return fmt.Errorf("重新扫描应用: %w", err)
	}

	bucket.Apps = apps
	bucket.LastUpdated = time.Now()

	return nil
}

// UpdateApp 更新单个应用。
func (u *updater) UpdateApp(ctx context.Context, bucket *manifest.Bucket, appName string) error {
	ref, ok := bucket.Apps[appName]
	if !ok {
		return fmt.Errorf("应用不存在: %s", appName)
	}

	// TODO: 调用应用的 check_version 函数获取最新版本
	_ = ref

	return fmt.Errorf("应用更新暂未实现")
}

// CheckUpdates 检查哪些应用有更新。
func (u *updater) CheckUpdates(ctx context.Context, bucket *manifest.Bucket) ([]string, error) {
	var updates []string

	// TODO: 实现更新检查逻辑
	// 1. 遍历所有应用
	// 2. 调用 check_version 函数
	// 3. 与当前版本比较
	// 4. 返回有更新的应用列表

	return updates, nil
}
