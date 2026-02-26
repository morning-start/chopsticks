package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"
	"chopsticks/engine/archive"
	"chopsticks/engine/checksum"
	"chopsticks/engine/fetch"
)

// AppInstaller 定义应用安装器接口。
type AppInstaller interface {
	Install(ctx context.Context, app *manifest.App, opts InstallOptions) error
	Download(url, dest string) error
	Verify(path, hash string, alg checksum.Algorithm) error
	Extract(src, dest string) error
}

// appInstaller 是 AppInstaller 的实现。
type appInstaller struct {
	installer *installer
}

// 编译时接口检查。
var _ AppInstaller = (*appInstaller)(nil)

// NewAppInstaller 创建新的 AppInstaller。
func NewAppInstaller(inst *installer) AppInstaller {
	return &appInstaller{
		installer: inst,
	}
}

// Install 执行完整的安装流程。
func (i *appInstaller) Install(ctx context.Context, app *manifest.App, opts InstallOptions) error {
	// 1. 创建安装目录
	installDir := opts.InstallDir
	if installDir == "" {
		// 使用配置中的 AppsPath
		if cfg, ok := i.installer.config.(interface{ GetAppsPath() string }); ok {
			installDir = filepath.Join(cfg.GetAppsPath(), app.Script.Name)
		} else {
			return fmt.Errorf("无法获取应用路径")
		}
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录: %w", err)
	}

	// 2. 调用 pre_install 钩子
	if err := i.callPreInstall(ctx, app, opts); err != nil {
		return fmt.Errorf("pre_install 钩子失败: %w", err)
	}

	// 3. 获取下载信息
	arch := opts.Arch
	if arch == "" {
		arch = "amd64"
	}

	version := app.Meta.Version
	downloadInfo, ok := app.Meta.Versions[version]
	if !ok {
		return fmt.Errorf("版本信息不存在: %s", version)
	}

	archInfo, ok := downloadInfo.Downloads[arch]
	if !ok {
		return fmt.Errorf("架构不支持: %s", arch)
	}

	// 4. 下载文件
	cacheDir := ""
	if cfg, ok := i.installer.config.(interface{ GetCachePath() string }); ok {
		cacheDir = cfg.GetCachePath()
	} else {
		return fmt.Errorf("无法获取缓存路径")
	}

	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s-%s-%s.%s",
		app.Script.Name, version, arch, archInfo.Type))

	if err := i.Download(archInfo.URL, cacheFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 5. 验证校验和
	shouldVerify := false
	if cfg, ok := i.installer.config.(interface{ GetVerifyHash() bool }); ok {
		shouldVerify = cfg.GetVerifyHash()
	}

	if shouldVerify && archInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(archInfo.Hash)
		if err := i.Verify(cacheFile, archInfo.Hash, alg); err != nil {
			return fmt.Errorf("校验和验证失败: %w", err)
		}
	}

	// 6. 解压文件
	extractDir := filepath.Join(installDir, version)
	if err := i.Extract(cacheFile, extractDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 7. 调用 post_install 钩子
	if err := i.callPostInstall(ctx, app, opts); err != nil {
		// post_install 失败不中断安装，只记录警告
		fmt.Fprintf(os.Stderr, "post_install 钩子失败: %v\n", err)
	}

	// 8. 保存安装记录
	installed := &manifest.InstalledApp{
		Name:        app.Script.Name,
		Version:     version,
		Bucket:      app.Script.Bucket,
		InstallDir:  installDir,
		InstalledAt: downloadInfo.ReleasedAt,
	}

	if err := i.installer.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("保存安装记录: %w", err)
	}

	return nil
}

// Download 下载文件。
func (i *appInstaller) Download(url, dest string) error {
	return fetch.Download(url, dest)
}

// Verify 验证文件校验和。
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

// Extract 解压文件。
func (i *appInstaller) Extract(src, dest string) error {
	return archive.Extract(src, dest)
}

// callPreInstall 调用 pre_install 钩子。
func (i *appInstaller) callPreInstall(_ context.Context, _ *manifest.App, _ InstallOptions) error {
	return nil
}

// callPostInstall 调用 post_install 钩子。
func (i *appInstaller) callPostInstall(_ context.Context, _ *manifest.App, _ InstallOptions) error {
	return nil
}
