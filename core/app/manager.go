package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

var (
	ErrAppNotFound          = errors.ErrAppNotFound
	ErrAppAlreadyInstalled  = errors.ErrAppAlreadyInstalled
	ErrVersionNotFound      = errors.ErrVersionNotFound
	ErrDependencyConflict   = errors.ErrDependencyConflict
)

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

type InstallSpec struct {
	Bucket  string
	Name    string
	Version string
}

type RemoveOptions struct {
	Purge bool
}

type UpdateOptions struct {
	Force bool
}

type SearchResult struct {
	Bucket string
	App    *manifest.AppRef
}

type manager struct {
	bucketMgr  bucket.Manager
	storage    store.Storage
	installer  Installer
	config     interface{}
	installDir string
}

var _ Manager = (*manager)(nil)

func NewManager(bucketMgr bucket.Manager, storage store.Storage, installer Installer, config interface{}, installDir string) Manager {
	return &manager{
		bucketMgr:  bucketMgr,
		storage:    storage,
		installer:  installer,
		config:     config,
		installDir: installDir,
	}
}

func (m *manager) Install(ctx context.Context, spec InstallSpec, opts InstallOptions) error {
	var bucketName string
	if spec.Bucket != "" {
		bucketName = spec.Bucket
	} else {
		bucketName = "main"
	}

	_, err := m.bucketMgr.GetBucket(ctx, bucketName)
	if err != nil {
		return errors.Wrap(err, "get bucket")
	}

	app, err := m.bucketMgr.GetApp(ctx, bucketName, spec.Name)
	if err != nil {
		return errors.Wrap(err, "get app info")
	}

	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(m.installDir, spec.Name)
	}

	installOpts := InstallOptions{
		Arch:       opts.Arch,
		Force:      opts.Force,
		InstallDir: installDir,
	}

	return m.installer.Install(ctx, app, installOpts)
}

func (m *manager) Remove(ctx context.Context, name string, opts RemoveOptions) error {
	_, err := m.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	uninstallOpts := UninstallOptions{
		Purge: opts.Purge,
	}

	return m.installer.Uninstall(ctx, name, uninstallOpts)
}

func (m *manager) Update(ctx context.Context, name string, opts UpdateOptions) error {
	installed, err := m.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	bucketName := installed.Bucket
	if bucketName == "" {
		bucketName = "main"
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
		fmt.Printf("[%d/%d] 正在更新 %s...\n", i+1, total, app.Name)
		if err := m.Update(ctx, app.Name, opts); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ 更新失败: %v\n", err)
			results = append(results, updateResult{name: app.Name, success: false, err: err})
		} else {
			fmt.Printf("  ✓ %s 更新成功\n", app.Name)
			results = append(results, updateResult{name: app.Name, success: true})
		}
	}

	// 汇总结果
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
	fmt.Printf("更新完成: 成功 %d, 失败 %d\n", successCount, failCount)
	if failCount > 0 {
		fmt.Println("失败的软件包:")
		for _, name := range failedApps {
			fmt.Printf("  - %s\n", name)
		}
		return fmt.Errorf("部分软件包更新失败")
	}

	return nil
}

func (m *manager) Switch(ctx context.Context, name, version string) error {
	return m.installer.Switch(ctx, name, version)
}

func (m *manager) ListInstalled() ([]*manifest.InstalledApp, error) {
	return m.storage.ListInstalledApps(context.Background())
}

func (m *manager) Info(ctx context.Context, bucketName, name string) (*manifest.AppInfo, error) {
	if bucketName == "" {
		bucketName = "main"
	}

	app, err := m.bucketMgr.GetApp(ctx, bucketName, name)
	if err != nil {
		return nil, errors.Wrap(err, "get app info")
	}

	installed, err := m.storage.GetInstalledApp(ctx, name)
	isInstalled := err == nil && installed != nil

	info := &manifest.AppInfo{
		Name:              app.Script.Name,
		Description:       app.Script.Description,
		Homepage:          app.Script.Homepage,
		License:           app.Script.License,
		Category:          app.Script.Category,
		Tags:              app.Script.Tags,
		Version:           app.Meta.Version,
		Bucket:            app.Script.Bucket,
		Installed:         isInstalled,
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

func matchesQuery(app *manifest.AppRef, query string) bool {
	lowerQuery := toLower(query)
	return contains(toLower(app.Name), lowerQuery) ||
		contains(toLower(app.Description), lowerQuery)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
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
