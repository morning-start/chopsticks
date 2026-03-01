package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/engine/archive"
	"chopsticks/engine/checksum"
	"chopsticks/engine/fetch"
)

type AppUpdater interface {
	Update(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error
}

type appUpdater struct {
	installer *installer
}

var _ AppUpdater = (*appUpdater)(nil)

func NewAppUpdater(inst *installer) AppUpdater {
	return &appUpdater{
		installer: inst,
	}
}

func (u *appUpdater) Update(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	newVersion := app.Meta.Version
	if newVersion == installed.Version && !opts.Force {
		return fmt.Errorf("already at latest version: %s", newVersion)
	}

	backupDir := filepath.Join(os.TempDir(), "chopsticks-backup", app.Script.Name+"-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(backupDir, DefaultDirPerm); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}

	currentVersionDir := filepath.Join(installed.InstallDir, installed.Version)
	if err := copyDir(currentVersionDir, backupDir); err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("backup failed: %w", err)
	}

	arch := DefaultArch
	downloadInfo, err := u.getDownloadInfo(app, newVersion, arch)
	if err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("get download info: %w", err)
	}

	cacheFile := filepath.Join(u.installer.downloadDir, fmt.Sprintf("%s-%s-%s", app.Script.Name, newVersion, arch))
	if err := fetch.Download(downloadInfo.URL, cacheFile); err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("download failed: %w", err)
	}

	if downloadInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(downloadInfo.Hash)
		ok, err := checksum.VerifyFile(cacheFile, downloadInfo.Hash, alg)
		if err != nil {
			os.RemoveAll(backupDir)
			return fmt.Errorf("verify failed: %w", err)
		}
		if !ok {
			os.RemoveAll(backupDir)
			return fmt.Errorf("checksum mismatch")
		}
	}

	extractDir := filepath.Join(installed.InstallDir, newVersion)
	if err := archive.Extract(cacheFile, extractDir); err != nil {
		os.Rename(backupDir, currentVersionDir)
		os.RemoveAll(backupDir)
		return fmt.Errorf("extract failed: %w", err)
	}

	os.RemoveAll(currentVersionDir)
	os.RemoveAll(backupDir)

	installed.Version = newVersion
	installed.UpdatedAt = time.Now()
	if err := u.installer.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("save install record: %w", err)
	}

	fmt.Printf("✓ %s updated successfully (%s -> %s)\n", app.Script.Name, installed.Version, newVersion)
	return nil
}

func (u *appUpdater) getDownloadInfo(app *manifest.App, version, arch string) (*manifest.DownloadInfo, error) {
	if app.Meta != nil && app.Meta.Versions != nil {
		if versionInfo, ok := app.Meta.Versions[version]; ok {
			if downloadInfo, ok := versionInfo.Downloads[arch]; ok {
				return &downloadInfo, nil
			}
		}
	}
	return nil, fmt.Errorf("download info not found for %s/%s", version, arch)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}
