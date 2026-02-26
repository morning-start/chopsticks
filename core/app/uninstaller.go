package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"
)

// Uninstaller 定义卸载器接口。
type Uninstaller interface {
	Uninstall(ctx context.Context, name string, opts UninstallOptions) error
}

// uninstaller 是 Uninstaller 的实现。
type uninstaller struct {
	installer *installer
}

// 编译时接口检查。
var _ Uninstaller = (*uninstaller)(nil)

// NewUninstaller 创建新的 Uninstaller。
func NewUninstaller(inst *installer) Uninstaller {
	return &uninstaller{
		installer: inst,
	}
}

// Uninstall 执行卸载流程。
func (u *uninstaller) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	// 1. 获取已安装的应用信息
	installed, err := u.installer.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return fmt.Errorf("获取安装信息: %w", err)
	}

	// 2. 调用 pre_uninstall 钩子
	if err := u.callPreUninstall(ctx, installed); err != nil {
		return fmt.Errorf("pre_uninstall 钩子失败: %w", err)
	}

	// 3. 删除安装目录
	installDir := installed.InstallDir
	if opts.Purge {
		// 彻底清除：删除整个应用目录
		if err := os.RemoveAll(installDir); err != nil {
			return fmt.Errorf("删除安装目录: %w", err)
		}
	} else {
		// 普通卸载：只删除当前版本
		versionDir := filepath.Join(installDir, installed.Version)
		if err := os.RemoveAll(versionDir); err != nil {
			return fmt.Errorf("删除版本目录: %w", err)
		}
	}

	// 4. 调用 post_uninstall 钩子
	if err := u.callPostUninstall(ctx, installed); err != nil {
		fmt.Fprintf(os.Stderr, "post_uninstall 钩子失败: %v\n", err)
	}

	// 5. 删除安装记录
	if err := u.installer.storage.DeleteInstalledApp(ctx, name); err != nil {
		return fmt.Errorf("删除安装记录: %w", err)
	}

	return nil
}

// callPreUninstall 调用 pre_uninstall 钩子。
func (u *uninstaller) callPreUninstall(_ context.Context, _ *manifest.InstalledApp) error {
	return nil
}

// callPostUninstall 调用 post_uninstall 钩子。
func (u *uninstaller) callPostUninstall(_ context.Context, _ *manifest.InstalledApp) error {
	return nil
}
