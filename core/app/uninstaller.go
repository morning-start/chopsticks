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
		return fmt.Errorf("get install info: %w", err)
	}

	installDir := installed.InstallDir
	if opts.Purge {
		if err := os.RemoveAll(installDir); err != nil {
			return fmt.Errorf("remove install directory: %w", err)
		}
	} else {
		versionDir := filepath.Join(installDir, installed.Version)
		if err := os.RemoveAll(versionDir); err != nil {
			return fmt.Errorf("remove version directory: %w", err)
		}
	}

	if err := u.installer.storage.DeleteInstalledApp(ctx, name); err != nil {
		return fmt.Errorf("delete install record: %w", err)
	}

	fmt.Printf("✓ %s uninstalled successfully\n", name)
	return nil
}
