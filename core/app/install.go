// Package install 提供安装功能。
package app

import (
	"context"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	engine "chopsticks/engine"
)

// Installer 定义安装接口。
type Installer interface {
	Install(ctx context.Context, app *manifest.App, opts InstallOptions) error
	Uninstall(ctx context.Context, name string, opts UninstallOptions) error
	Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error
	Switch(ctx context.Context, name, version string) error
}

// InstallOptions 包含安装选项。
type InstallOptions struct {
	Arch       string // 架构
	Force      bool   // 强制安装
	InstallDir string // 安装目录
}

// UninstallOptions 包含卸载选项。
type UninstallOptions struct {
	Purge bool // 彻底清除
}

// RefreshOptions 包含刷新选项。
type RefreshOptions struct {
	Force bool // 强制刷新
}

// installer 是 Installer 的实现。
type installer struct {
	storage store.Storage
	config  interface{}
	engines map[string]engine.Engine
}

// 编译时接口检查。
var _ Installer = (*installer)(nil)

// NewInstaller 创建新的 Installer。
func NewInstaller(storage store.Storage, config interface{}, engines map[string]engine.Engine) Installer {
	return &installer{
		storage: storage,
		config:  config,
		engines: engines,
	}
}

func (i *installer) Install(ctx context.Context, app *manifest.App, opts InstallOptions) error {
	// TODO: 实现
	return nil
}

func (i *installer) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	// TODO: 实现
	return nil
}

func (i *installer) Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	// TODO: 实现
	return nil
}

func (i *installer) Switch(ctx context.Context, name, version string) error {
	// TODO: 实现
	return nil
}
