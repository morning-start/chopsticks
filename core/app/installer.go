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
	"chopsticks/pkg/output"
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
	// 创建进度管理器
	pm := output.NewProgressManager()
	defer pm.Wait()

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

	// 定义安装阶段
	stages := []string{"准备", "下载", "校验", "解压", "完成"}
	progressBar := pm.AddInstallBar(app.Script.Name, stages)

	// 阶段 1: 准备
	progressBar.SetStage(0)
	downloadInfo, err := i.getDownloadInfo(app, version, arch)
	if err != nil {
		return fmt.Errorf("获取下载信息: %w", err)
	}
	progressBar.CompleteStage()

	// 阶段 2: 下载
	progressBar.SetStage(1)
	cacheFile := filepath.Join(i.installer.downloadDir, fmt.Sprintf("%s-%s-%s", app.Script.Name, version, arch))
	if err := i.DownloadWithProgress(ctx, downloadInfo.URL, cacheFile, pm); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	progressBar.CompleteStage()

	// 阶段 3: 校验
	progressBar.SetStage(2)
	if downloadInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(downloadInfo.Hash)
		if err := i.Verify(cacheFile, downloadInfo.Hash, alg); err != nil {
			return fmt.Errorf("校验失败: %w", err)
		}
	}
	progressBar.CompleteStage()

	// 阶段 4: 解压
	progressBar.SetStage(3)
	extractDir := filepath.Join(installDir, version)
	if err := i.Extract(cacheFile, extractDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}
	progressBar.CompleteStage()

	// 阶段 5: 完成
	progressBar.SetStage(4)
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
	progressBar.Complete()

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

// DownloadWithProgress 使用进度条下载文件
func (i *appInstaller) DownloadWithProgress(ctx context.Context, url, dest string, pm *output.ProgressManager) error {
	return fetch.DownloadWithProgress(ctx, url, dest, pm)
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
