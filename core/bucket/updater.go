package bucket

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"

	"github.com/go-git/go-git/v5"
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
	gitDir := filepath.Join(bucket.Path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// 是 Git 仓库，执行 git pull
		if err := u.updateGitRepo(ctx, bucket); err != nil {
			return fmt.Errorf("Git 更新失败: %w", err)
		}
	}

	// 重新扫描应用
	apps, err := u.loader.ScanApps(ctx, bucket.Path)
	if err != nil {
		return fmt.Errorf("重新扫描应用: %w", err)
	}

	bucket.Apps = apps
	bucket.LastUpdated = time.Now()

	return nil
}

// updateGitRepo 执行 Git 仓库的更新
func (u *updater) updateGitRepo(ctx context.Context, bucket *manifest.Bucket) error {
	repo, err := git.PlainOpen(bucket.Path)
	if err != nil {
		return fmt.Errorf("打开 Git 仓库: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区: %w", err)
	}

	// 执行 pull
	pullOpts := &git.PullOptions{
		RemoteName: "origin",
	}

	if err := worktree.PullContext(ctx, pullOpts); err != nil {
		// 如果已经是最新，返回 nil
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return fmt.Errorf("拉取更新: %w", err)
	}

	return nil
}

// UpdateApp 更新单个应用。
func (u *updater) UpdateApp(ctx context.Context, bucket *manifest.Bucket, appName string) error {
	ref, ok := bucket.Apps[appName]
	if !ok {
		return fmt.Errorf("应用不存在: %s", appName)
	}

	// 重新加载应用引用以获取最新信息
	newRef, err := u.loadAppRef(appName, ref.ScriptPath)
	if err != nil {
		return fmt.Errorf("重新加载应用信息: %w", err)
	}

	// 更新应用引用
	bucket.Apps[appName] = newRef
	bucket.LastUpdated = time.Now()

	return nil
}

// loadAppRef 加载单个应用的引用信息（复用 loader 的逻辑）
func (u *updater) loadAppRef(name, scriptPath string) (*manifest.AppRef, error) {
	// 这里简化处理，实际应该调用 loader 的方法
	// 由于 loader 是接口，我们需要重新实现或暴露该方法
	ref := &manifest.AppRef{
		Name:       name,
		ScriptPath: scriptPath,
	}

	// 尝试读取对应的 .meta.json 文件获取更多信息
	dir := filepath.Dir(scriptPath)
	metaPath := filepath.Join(dir, name+".meta.json")

	if data, err := os.ReadFile(metaPath); err == nil {
		var meta struct {
			Description string   `json:"description"`
			Version     string   `json:"version"`
			Category    string   `json:"category"`
			Tags        []string `json:"tags"`
		}
		if err := json.Unmarshal(data, &meta); err == nil {
			ref.Description = meta.Description
			ref.Version = meta.Version
			ref.Category = meta.Category
			ref.Tags = meta.Tags
			ref.MetaPath = metaPath
		}
	}

	return ref, nil
}

// CheckUpdates 检查哪些应用有更新。
func (u *updater) CheckUpdates(ctx context.Context, bucket *manifest.Bucket) ([]string, error) {
	var updates []string

	// 1. 遍历所有应用
	for appName, ref := range bucket.Apps {
		// 2. 重新加载应用信息
		newRef, err := u.loadAppRef(appName, ref.ScriptPath)
		if err != nil {
			continue
		}

		// 3. 与当前版本比较
		if newRef.Version != "" && newRef.Version != ref.Version {
			updates = append(updates, appName)
		}
	}

	return updates, nil
}
