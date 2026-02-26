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
		return fmt.Errorf("已经是最新版本: %s", newVersion)
	}

	backupDir := filepath.Join(os.TempDir(), "chopsticks-backup", app.Script.Name+"-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录: %w", err)
	}

	currentVersionDir := filepath.Join(installed.InstallDir, installed.Version)
	if err := copyDir(currentVersionDir, backupDir); err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("备份失败: %w", err)
	}

	arch := "amd64"
	downloadInfo, err := u.getDownloadInfo(app, newVersion, arch)
	if err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("获取下载信息: %w", err)
	}

	cacheFile := filepath.Join(u.installer.downloadDir, fmt.Sprintf("%s-%s-%s", app.Script.Name, newVersion, arch))
	if err := fetch.Download(downloadInfo.URL, cacheFile); err != nil {
		os.RemoveAll(backupDir)
		return fmt.Errorf("下载失败: %w", err)
	}

	if downloadInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(downloadInfo.Hash)
		ok, err := checksum.VerifyFile(cacheFile, downloadInfo.Hash, alg)
		if err != nil || !ok {
			os.RemoveAll(backupDir)
			return fmt.Errorf("校验失败: %w", err)
		}
	}

	extractDir := filepath.Join(installed.InstallDir, newVersion)
	if err := archive.Extract(cacheFile, extractDir); err != nil {
		os.Rename(backupDir, currentVersionDir)
		os.RemoveAll(backupDir)
		return fmt.Errorf("解压失败: %w", err)
	}

	os.RemoveAll(currentVersionDir)
	os.RemoveAll(backupDir)

	installed.Version = newVersion
	installed.UpdatedAt = time.Now()
	if err := u.installer.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("保存安装记录: %w", err)
	}

	fmt.Printf("✓ %s 更新成功 (%s -> %s)\n", app.Script.Name, installed.Version, newVersion)
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
	return nil, fmt.Errorf("未找到 %s/%s 的下载信息", version, arch)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
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
