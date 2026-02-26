package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Uninstaller interface {
	Uninstall(ctx context.Context, name string, opts UninstallOptions) error
}

type uninstaller struct {
	installer *installer
}

var _ Uninstaller = (*uninstaller)(nil)

func NewUninstaller(inst *installer) Uninstaller {
	return &uninstaller{
		installer: inst,
	}
}

func (u *uninstaller) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	installed, err := u.installer.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return fmt.Errorf("获取安装信息: %w", err)
	}

	installDir := installed.InstallDir
	if opts.Purge {
		if err := os.RemoveAll(installDir); err != nil {
			return fmt.Errorf("删除安装目录: %w", err)
		}
	} else {
		versionDir := filepath.Join(installDir, installed.Version)
		if err := os.RemoveAll(versionDir); err != nil {
			return fmt.Errorf("删除版本目录: %w", err)
		}
	}

	if err := u.installer.storage.DeleteInstalledApp(ctx, name); err != nil {
		return fmt.Errorf("删除安装记录: %w", err)
	}

	fmt.Printf("✓ %s 卸载成功\n", name)
	return nil
}
