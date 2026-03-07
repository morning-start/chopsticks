// Package testutil 提供集成测试的辅助工具。
package testutil

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"chopsticks/core/app"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

// MockInstaller 模拟安装器。
type MockInstaller struct {
	mu              sync.RWMutex
	InstalledApps   map[string]*InstallRecord
	UninstalledApps map[string]*UninstallRecord
	RefreshedApps   map[string]*RefreshRecord
	SwitchedApps    map[string]*SwitchRecord
	storage         store.LegacyStorage

	// InstallError 模拟安装错误
	InstallError error
	// UninstallError 模拟卸载错误
	UninstallError error
	// RefreshError 模拟刷新错误
	RefreshError error
	// SwitchError 模拟切换错误
	SwitchError error
}

// InstallRecord 安装记录。
type InstallRecord struct {
	App  *manifest.App
	Opts app.InstallOptions
}

// UninstallRecord 卸载记录。
type UninstallRecord struct {
	Name string
	Opts app.UninstallOptions
}

// RefreshRecord 刷新记录。
type RefreshRecord struct {
	App       *manifest.App
	Installed *manifest.InstalledApp
	Opts      app.RefreshOptions
}

// SwitchRecord 切换记录。
type SwitchRecord struct {
	Name    string
	Version string
}

// NewMockInstaller 创建新的 MockInstaller。
func NewMockInstaller(storage store.LegacyStorage) *MockInstaller {
	return &MockInstaller{
		InstalledApps:   make(map[string]*InstallRecord),
		UninstalledApps: make(map[string]*UninstallRecord),
		RefreshedApps:   make(map[string]*RefreshRecord),
		SwitchedApps:    make(map[string]*SwitchRecord),
		storage:         storage,
	}
}

// Install 模拟安装操作。
func (m *MockInstaller) Install(ctx context.Context, app *manifest.App, opts app.InstallOptions) error {
	if m.InstallError != nil {
		return m.InstallError
	}

	if app == nil || app.Script == nil {
		return fmt.Errorf("invalid app")
	}

	// 创建安装目录
	if opts.InstallDir != "" {
		if err := os.MkdirAll(opts.InstallDir, 0755); err != nil {
			return fmt.Errorf("创建安装目录失败: %w", err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.InstalledApps[app.Script.Name] = &InstallRecord{
		App:  app,
		Opts: opts,
	}

	// 如果 storage 不为 nil，保存安装记录到 storage
	if m.storage != nil {
		// 获取版本信息
		version := ""
		if app.Meta != nil {
			version = app.Meta.Version
		}
		// 获取 bucket 信息
		bucket := ""
		if app.Script != nil {
			bucket = app.Script.Bucket
		}
		installedApp := &manifest.InstalledApp{
			Name:        app.Script.Name,
			Version:     version,
			Bucket:      bucket,
			InstallDir:  opts.InstallDir,
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := m.storage.SaveInstalledApp(ctx, installedApp); err != nil {
			return fmt.Errorf("保存安装记录失败: %w", err)
		}
	}

	return nil
}

// Uninstall 模拟卸载操作。
func (m *MockInstaller) Uninstall(ctx context.Context, name string, opts app.UninstallOptions) error {
	if m.UninstallError != nil {
		return m.UninstallError
	}

	m.mu.Lock()
	// 获取安装目录以便删除
	record, exists := m.InstalledApps[name]
	installDir := ""
	if exists {
		installDir = record.Opts.InstallDir
	}
	m.mu.Unlock()

	// 删除安装目录
	if installDir != "" {
		if err := os.RemoveAll(installDir); err != nil {
			return fmt.Errorf("删除安装目录失败: %w", err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.UninstalledApps[name] = &UninstallRecord{
		Name: name,
		Opts: opts,
	}

	// 从已安装列表中移除
	delete(m.InstalledApps, name)

	// 如果 storage 不为 nil，从 storage 中删除安装记录
	if m.storage != nil {
		if err := m.storage.DeleteInstalledApp(ctx, name); err != nil {
			return fmt.Errorf("删除安装记录失败: %w", err)
		}
	}

	return nil
}

// Refresh 模拟刷新操作。
func (m *MockInstaller) Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts app.RefreshOptions) error {
	if m.RefreshError != nil {
		return m.RefreshError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.RefreshedApps[app.Script.Name] = &RefreshRecord{
		App:       app,
		Installed: installed,
		Opts:      opts,
	}

	return nil
}

// Switch 模拟版本切换操作。
func (m *MockInstaller) Switch(ctx context.Context, name, version string) error {
	if m.SwitchError != nil {
		return m.SwitchError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.SwitchedApps[name] = &SwitchRecord{
		Name:    name,
		Version: version,
	}

	return nil
}

// IsInstalled 检查应用是否已安装。
func (m *MockInstaller) IsInstalled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.InstalledApps[name]
	return ok
}

// GetInstalledApp 获取已安装的应用记录。
func (m *MockInstaller) GetInstalledApp(name string) (*InstallRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.InstalledApps[name]
	return record, ok
}

// Reset 重置所有记录。
func (m *MockInstaller) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.InstalledApps = make(map[string]*InstallRecord)
	m.UninstalledApps = make(map[string]*UninstallRecord)
	m.RefreshedApps = make(map[string]*RefreshRecord)
	m.SwitchedApps = make(map[string]*SwitchRecord)
	m.InstallError = nil
	m.UninstallError = nil
	m.RefreshError = nil
	m.SwitchError = nil
}

// 编译时接口检查
var _ app.Installer = (*MockInstaller)(nil)
