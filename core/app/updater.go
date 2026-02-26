package app

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/manifest"
)

// AppUpdater 定义应用更新器接口。
type AppUpdater interface {
	Update(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error
	Backup(ctx context.Context, name string) (string, error)
	Restore(ctx context.Context, name, backupPath string) error
}

// appUpdater 是 AppUpdater 的实现。
type appUpdater struct {
	installer Installer
}

// 编译时接口检查。
var _ AppUpdater = (*appUpdater)(nil)

// NewAppUpdater 创建新的 AppUpdater。
func NewAppUpdater(installer Installer) AppUpdater {
	return &appUpdater{
		installer: installer,
	}
}

// Update 执行更新流程。
func (u *appUpdater) Update(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	// 1. 备份当前版本
	backupPath, err := u.Backup(ctx, app.Script.Name)
	if err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 2. 获取新版本
	newVersion := app.Meta.Version
	if newVersion == installed.Version && !opts.Force {
		return fmt.Errorf("已经是最新版本: %s", newVersion)
	}

	// 3. 持久化数据目录
	persistDirs := u.getPersistDirs(installed)

	// 4. 执行更新（重新安装新版本）
	installOpts := InstallOptions{
		Arch:       "amd64",
		Force:      true,
		InstallDir: installed.InstallDir,
	}

	if err := u.installer.Install(ctx, app, installOpts); err != nil {
		// 安装失败，尝试恢复
		if restoreErr := u.Restore(ctx, app.Script.Name, backupPath); restoreErr != nil {
			return fmt.Errorf("更新失败且恢复失败: %v, 恢复错误: %v", err, restoreErr)
		}
		return fmt.Errorf("更新失败，已恢复到旧版本: %w", err)
	}

	// 5. 恢复持久化数据
	if err := u.restorePersistData(installed, persistDirs); err != nil {
		fmt.Fprintf(os.Stderr, "恢复持久化数据失败: %v\n", err)
	}

	// 6. 清理备份
	if err := os.RemoveAll(backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "清理备份失败: %v\n", err)
	}

	return nil
}

// Backup 备份当前版本。
func (u *appUpdater) Backup(ctx context.Context, name string) (string, error) {
	return "", nil
}

// Restore 从备份恢复。
func (u *appUpdater) Restore(ctx context.Context, name, backupPath string) error {
	// TODO: 从备份目录恢复到安装目录
	return nil
}

// getPersistDirs 获取持久化数据目录列表。
func (u *appUpdater) getPersistDirs(_ *manifest.InstalledApp) []string {
	return nil
}

// restorePersistData 恢复持久化数据。
func (u *appUpdater) restorePersistData(_ *manifest.InstalledApp, _ []string) error {
	return nil
}
