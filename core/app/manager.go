// Package app 提供应用管理功能。
//
// 该包实现了应用的安装、卸载、更新和查询等核心功能，
// 支持依赖解析、版本切换和批量操作。
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/bucket"
	"chopsticks/core/dep"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

var (
	ErrAppNotFound         = errors.ErrAppNotFound
	ErrAppAlreadyInstalled = errors.ErrAppAlreadyInstalled
	ErrVersionNotFound     = errors.ErrVersionNotFound
	ErrDependencyConflict  = errors.ErrDependencyConflict
)

// AppManager 定义应用管理器接口。
type AppManager interface {
	Install(ctx context.Context, spec InstallSpec, opts InstallOptions) error
	Remove(ctx context.Context, name string, opts RemoveOptions) error
	Update(ctx context.Context, name string, opts UpdateOptions) error
	UpdateAll(ctx context.Context, opts UpdateOptions) error
	Switch(ctx context.Context, name, version string) error
	ListInstalled() ([]*manifest.InstalledApp, error)
	Info(ctx context.Context, bucket, name string) (*manifest.AppInfo, error)
	Search(ctx context.Context, query string, bucket string) ([]SearchResult, error)
}

// InstallSpec 安装规格
type InstallSpec struct {
	Bucket  string
	Name    string
	Version string
}

// RemoveOptions 移除选项
type RemoveOptions struct {
	Purge bool
}

// UpdateOptions 更新选项 - 字段已优化
type UpdateOptions struct {
	Force bool
}

// SearchResult 搜索结果
type SearchResult struct {
	Bucket string
	App    *manifest.AppRef
}

// manager 管理器 - 字段按大小从大到小排列
type manager struct {
	bucketMgr  bucket.BucketManager // 16 bytes (interface)
	storage    store.LegacyStorage  // 16 bytes (interface)
	installer  Installer            // 16 bytes (interface)
	depMgr     dep.Manager          // 16 bytes (interface)
	config     interface{}          // 16 bytes (interface)
	installDir string               // 16 bytes (string header)
}

var _ AppManager = (*manager)(nil)

// NewManager 创建应用管理器
func NewManager(bucketMgr bucket.BucketManager, storage store.LegacyStorage, installer Installer, config interface{}, installDir string) AppManager {
	// 创建依赖管理器
	depMgr, err := dep.NewDependencyManager(bucketMgr, storage, installDir)
	if err != nil {
		fmt.Printf("warning: failed to create dependency manager: %v\n", err)
	}

	return &manager{
		bucketMgr:  bucketMgr,
		storage:    storage,
		installer:  installer,
		depMgr:     depMgr,
		config:     config,
		installDir: installDir,
	}
}

func (m *manager) Install(ctx context.Context, spec InstallSpec, opts InstallOptions) error {
	bucketName := spec.Bucket
	if bucketName == "" {
		bucketName = DefaultBucket
	}

	_, err := m.bucketMgr.GetBucket(ctx, bucketName)
	if err != nil {
		return errors.Wrap(err, "get bucket")
	}

	app, err := m.bucketMgr.GetApp(ctx, bucketName, spec.Name)
	if err != nil {
		return errors.Wrap(err, "get app info")
	}

	// 使用依赖管理器解析依赖
	var depGraph *dep.DependencyGraph
	if m.depMgr != nil && !opts.NoDeps {
		depGraph, err = m.depMgr.Resolve(ctx, app)
		if err != nil {
			return errors.Wrap(err, "resolve dependencies")
		}

		// 安装依赖（按拓扑排序顺序）
		if len(depGraph.Order) > 1 {
			fmt.Printf("Installing dependencies for %s (%d packages)...\n", spec.Name, len(depGraph.Order)-1)
			for _, depName := range depGraph.Order {
				if depName == spec.Name {
					continue // 跳过主应用
				}

				// 检查是否已安装
				if _, err := m.storage.GetInstalledApp(ctx, depName); err == nil {
					fmt.Printf("  ✓ %s already installed\n", depName)
					continue
				}

				// 安装依赖
				depNode := depGraph.Nodes[depName]
				if depNode == nil || depNode.App == nil {
					continue
				}

				fmt.Printf("  → Installing %s...\n", depName)
				depOpts := InstallOptions{
					Arch:       opts.Arch,
					Force:      false,
					InstallDir: filepath.Join(m.installDir, depName),
					NoDeps:     true, // 依赖不再递归安装依赖
				}
				if err := m.installer.Install(ctx, depNode.App, depOpts); err != nil {
					return errors.Wrapf(err, "install dependency %s", depName)
				}

				// 更新依赖索引
				if m.depMgr != nil {
					deps := extractDependencyNames(depNode.App)
					if err := m.depMgr.UpdateDepsIndex(ctx, depName, deps); err != nil {
						fmt.Printf("warning: failed to update deps index for %s: %v\n", depName, err)
					}
				}
			}
			fmt.Println()
		}
	}

	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(m.installDir, spec.Name)
	}

	installOpts := InstallOptions{
		Arch:       opts.Arch,
		Force:      opts.Force,
		InstallDir: installDir,
		NoDeps:     true, // 主应用不再递归安装依赖
	}

	if err := m.installer.Install(ctx, app, installOpts); err != nil {
		return err
	}

	// 更新依赖索引
	if m.depMgr != nil {
		deps := extractDependencyNames(app)
		if err := m.depMgr.UpdateDepsIndex(ctx, spec.Name, deps); err != nil {
			fmt.Printf("warning: failed to update deps index for %s: %v\n", spec.Name, err)
		}
	}

	return nil
}

// extractDependencyNames 提取依赖名称列表
func extractDependencyNames(app *manifest.App) []string {
	if app == nil || app.Script == nil {
		return nil
	}

	var deps []string
	for _, dep := range app.Script.Dependencies {
		deps = append(deps, dep.Name)
	}
	return deps
}

func (m *manager) Remove(ctx context.Context, name string, opts RemoveOptions) error {
	_, err := m.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	// 使用依赖管理器检查反向依赖
	if m.depMgr != nil {
		dependents, err := m.depMgr.GetDependents(ctx, name)
		if err != nil {
			return errors.Wrap(err, "check dependents")
		}

		if len(dependents) > 0 {
			return errors.NewDependencyConflict(
				name,
				fmt.Sprintf("the following apps depend on %s: %s", name, strings.Join(dependents, ", ")),
			)
		}
	}

	uninstallOpts := UninstallOptions{
		Purge: opts.Purge,
	}

	if err := m.installer.Uninstall(ctx, name, uninstallOpts); err != nil {
		return err
	}

	// 更新依赖索引
	if m.depMgr != nil {
		if err := m.depMgr.UpdateDepsIndex(ctx, name, []string{}); err != nil {
			fmt.Printf("warning: failed to update deps index for %s: %v\n", name, err)
		}
	}

	return nil
}

func (m *manager) Update(ctx context.Context, name string, opts UpdateOptions) error {
	installed, err := m.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	bucketName := installed.Bucket
	if bucketName == "" {
		bucketName = DefaultBucket
	}

	app, err := m.bucketMgr.GetApp(ctx, bucketName, name)
	if err != nil {
		return errors.Wrap(err, "get app info")
	}

	refreshOpts := RefreshOptions{
		Force: opts.Force,
	}

	return m.installer.Refresh(ctx, app, installed, refreshOpts)
}

func (m *manager) UpdateAll(ctx context.Context, opts UpdateOptions) error {
	installedApps, err := m.storage.ListInstalledApps(ctx)
	if err != nil {
		return errors.Wrap(err, "list installed apps")
	}

	total := len(installedApps)
	if total == 0 {
		return nil
	}

	type updateResult struct {
		name    string
		success bool
		err     error
	}

	results := make([]updateResult, 0, total)

	for i, app := range installedApps {
		fmt.Printf("[%d/%d] Updating %s...\n", i+1, total, app.Name)
		if err := m.Update(ctx, app.Name, opts); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Update failed: %v\n", err)
			results = append(results, updateResult{name: app.Name, success: false, err: err})
		} else {
			fmt.Printf("  ✓ %s updated successfully\n", app.Name)
			results = append(results, updateResult{name: app.Name, success: true})
		}
	}

	// Summary
	var successCount, failCount int
	var failedApps []string

	for _, r := range results {
		if r.success {
			successCount++
		} else {
			failCount++
			failedApps = append(failedApps, r.name)
		}
	}

	fmt.Println()
	fmt.Printf("Update complete: %d succeeded, %d failed\n", successCount, failCount)
	if failCount > 0 {
		fmt.Println("Failed packages:")
		for _, name := range failedApps {
			fmt.Printf("  - %s\n", name)
		}
		return fmt.Errorf("some packages failed to update")
	}

	return nil
}

func (m *manager) Switch(ctx context.Context, name, version string) error {
	return m.installer.Switch(ctx, name, version)
}

func (m *manager) ListInstalled() ([]*manifest.InstalledApp, error) {
	apps, err := m.storage.ListInstalledApps(context.Background())
	if err != nil {
		return nil, err
	}
	// Return empty slice instead of nil for consistency
	if apps == nil {
		return []*manifest.InstalledApp{}, nil
	}
	return apps, nil
}

func (m *manager) Info(ctx context.Context, bucketName, name string) (*manifest.AppInfo, error) {
	if bucketName == "" {
		bucketName = DefaultBucket
	}

	app, err := m.bucketMgr.GetApp(ctx, bucketName, name)
	if err != nil {
		return nil, errors.Wrap(err, "get app info")
	}

	installed, err := m.storage.GetInstalledApp(ctx, name)
	isInstalled := err == nil && installed != nil

	info := &manifest.AppInfo{
		Name:             app.Script.Name,
		Description:      app.Script.Description,
		Homepage:         app.Script.Homepage,
		License:          app.Script.License,
		Category:         app.Script.Category,
		Tags:             app.Script.Tags,
		Version:          app.Meta.Version,
		Bucket:           app.Script.Bucket,
		Installed:        isInstalled,
		InstalledVersion: "",
	}

	if isInstalled {
		info.InstalledVersion = installed.Version
	}

	return info, nil
}

func (m *manager) Search(ctx context.Context, query string, bucketName string) ([]SearchResult, error) {
	var results []SearchResult

	if bucketName != "" {
		buckets := []string{bucketName}
		for _, b := range buckets {
			apps, err := m.bucketMgr.ListApps(ctx, b)
			if err != nil {
				continue
			}
			for _, app := range apps {
				if matchesQuery(app, query) {
					results = append(results, SearchResult{
						Bucket: b,
						App:    app,
					})
				}
			}
		}
	} else {
		buckets, err := m.bucketMgr.ListBuckets(ctx)
		if err != nil {
			return nil, err
		}
		for _, b := range buckets {
			apps, err := m.bucketMgr.ListApps(ctx, b)
			if err != nil {
				continue
			}
			for _, app := range apps {
				if matchesQuery(app, query) {
					results = append(results, SearchResult{
						Bucket: b,
						App:    app,
					})
				}
			}
		}
	}

	return results, nil
}

// matchesQuery checks if app matches the search query
func matchesQuery(app *manifest.AppRef, query string) bool {
	lowerQuery := strings.ToLower(query)
	return strings.Contains(strings.ToLower(app.Name), lowerQuery) ||
		strings.Contains(strings.ToLower(app.Description), lowerQuery)
}
