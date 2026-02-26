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

type AppInstaller interface {
	Install(ctx context.Context, app *manifest.App, opts InstallOptions) error
	Download(url, dest string) error
	Verify(path, hash string, alg checksum.Algorithm) error
	Extract(src, dest string) error
}

type appInstaller struct {
	installer *installer
}

var _ AppInstaller = (*appInstaller)(nil)

func NewAppInstaller(inst *installer) AppInstaller {
	return &appInstaller{
		installer: inst,
	}
}

func (i *appInstaller) Install(ctx context.Context, app *manifest.App, opts InstallOptions) error {
	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(i.installer.installBase, app.Script.Name)
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录: %w", err)
	}

	arch := opts.Arch
	if arch == "" {
		arch = "amd64"
	}

	version := app.Meta.Version
	if version == "" {
		version = "latest"
	}

	downloadInfo, err := i.getDownloadInfo(app, version, arch)
	if err != nil {
		return fmt.Errorf("获取下载信息: %w", err)
	}

	cacheFile := filepath.Join(i.installer.downloadDir, fmt.Sprintf("%s-%s-%s", app.Script.Name, version, arch))
	if err := i.Download(downloadInfo.URL, cacheFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	if downloadInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(downloadInfo.Hash)
		if err := i.Verify(cacheFile, downloadInfo.Hash, alg); err != nil {
			return fmt.Errorf("校验失败: %w", err)
		}
	}

	extractDir := filepath.Join(installDir, version)
	if err := i.Extract(cacheFile, extractDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	installed := &manifest.InstalledApp{
		Name:        app.Script.Name,
		Version:     version,
		Bucket:      app.Script.Bucket,
		InstallDir:  installDir,
		InstalledAt: time.Now(),
	}

	if err := i.installer.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("保存安装记录: %w", err)
	}

	fmt.Printf("✓ %s (%s) 安装成功\n", app.Script.Name, version)
	return nil
}

func (i *appInstaller) getDownloadInfo(app *manifest.App, version, arch string) (*manifest.DownloadInfo, error) {
	if app.Meta != nil && app.Meta.Versions != nil {
		if versionInfo, ok := app.Meta.Versions[version]; ok {
			if downloadInfo, ok := versionInfo.Downloads[arch]; ok {
				return &downloadInfo, nil
			}
		}
	}
	return nil, fmt.Errorf("未找到 %s/%s 的下载信息", version, arch)
}

func (i *appInstaller) Download(url, dest string) error {
	return fetch.Download(url, dest)
}

func (i *appInstaller) Verify(path, hash string, alg checksum.Algorithm) error {
	ok, err := checksum.VerifyFile(path, hash, alg)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("校验和不匹配")
	}
	return nil
}

func (i *appInstaller) Extract(src, dest string) error {
	return archive.Extract(src, dest)
}
