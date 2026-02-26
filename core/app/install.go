package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/engine"
	"chopsticks/engine/fetch"
)

type Installer interface {
	Install(ctx context.Context, app *manifest.App, opts InstallOptions) error
	Uninstall(ctx context.Context, name string, opts UninstallOptions) error
	Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error
	Switch(ctx context.Context, name, version string) error
}

type InstallOptions struct {
	Arch       string
	Force      bool
	InstallDir string
}

type UninstallOptions struct {
	Purge bool
}

type RefreshOptions struct {
	Force bool
}

type installer struct {
	storage      store.Storage
	config       interface{}
	jsEngine     *engine.JSEngine
	downloadDir  string
	installBase string
}

var _ Installer = (*installer)(nil)

func NewInstaller(storage store.Storage, config interface{}, jsEngine *engine.JSEngine, installBase string) Installer {
	return &installer{
		storage:      storage,
		config:       config,
		jsEngine:     jsEngine,
		installBase:  installBase,
		downloadDir:  filepath.Join(installBase, "tmp"),
	}
}

func (i *installer) Install(ctx context.Context, app *manifest.App, opts InstallOptions) error {
	if app == nil || app.Script == nil {
		return fmt.Errorf("无效的应用信息")
	}

	appName := app.Script.Name

	installed, err := i.storage.GetInstalledApp(ctx, appName)
	if err == nil && installed != nil && !opts.Force {
		return fmt.Errorf("应用 %s 已安装 (版本: %s)", appName, installed.Version)
	}

	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(i.installBase, appName)
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录: %w", err)
	}

	if err := os.MkdirAll(i.downloadDir, 0755); err != nil {
		return fmt.Errorf("创建下载目录: %w", err)
	}

	version := "latest"
	if app.Meta != nil && app.Meta.Version != "" {
		version = app.Meta.Version
	}

	arch := opts.Arch
	if arch == "" {
		arch = "amd64"
	}

	downloadInfo, err := i.getDownloadInfo(app, version, arch)
	if err != nil {
		return fmt.Errorf("获取下载信息: %w", err)
	}

	downloadedPath, err := i.downloadPackage(ctx, downloadInfo)
	if err != nil {
		return fmt.Errorf("下载安装包: %w", err)
	}
	defer os.Remove(downloadedPath)

	if downloadInfo.Hash != "" {
		if err := i.verifyChecksum(downloadedPath, downloadInfo.Hash); err != nil {
			return fmt.Errorf("校验失败: %w", err)
		}
	}

	extractDir := filepath.Join(installDir, version)
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("创建版本目录: %w", err)
	}

	if err := i.extractPackage(downloadedPath, extractDir); err != nil {
		return fmt.Errorf("解压安装包: %w", err)
	}

	env := i.buildInstallEnv(appName, version, extractDir)

	if err := i.runScript(ctx, app, "preInstall", env); err != nil {
		return fmt.Errorf("preInstall 钩子失败: %w", err)
	}

	if err := i.runInstallScript(ctx, app, extractDir, env); err != nil {
		return fmt.Errorf("安装脚本执行失败: %w", err)
	}

	if err := i.runScript(ctx, app, "postInstall", env); err != nil {
		return fmt.Errorf("postInstall 钩子失败: %w", err)
	}

	installedApp := &manifest.InstalledApp{
		Name:        appName,
		Version:     version,
		Bucket:      app.Script.Bucket,
		InstallDir:  installDir,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.storage.SaveInstalledApp(ctx, installedApp); err != nil {
		return fmt.Errorf("保存安装记录: %w", err)
	}

	fmt.Printf("✓ %s (%s) 安装成功\n", appName, version)
	return nil
}

func (i *installer) getDownloadInfo(app *manifest.App, version, arch string) (*manifest.DownloadInfo, error) {
	if app.Meta != nil && app.Meta.Versions != nil {
		if versionInfo, ok := app.Meta.Versions[version]; ok {
			if downloadInfo, ok := versionInfo.Downloads[arch]; ok {
				return &downloadInfo, nil
			}
		}
	}
	return nil, fmt.Errorf("未找到 %s/%s 的下载信息", version, arch)
}

func (i *installer) downloadPackage(ctx context.Context, info *manifest.DownloadInfo) (string, error) {
	filename := filepath.Base(info.URL)
	destPath := filepath.Join(i.downloadDir, filename)

	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	fmt.Printf("下载: %s\n", info.URL)
	if err := fetch.Download(info.URL, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

func (i *installer) verifyChecksum(filePath, expectedHash string) error {
	return nil
}

func (i *installer) extractPackage(archivePath, destDir string) error {
	return nil
}

func (i *installer) buildInstallEnv(name, version, installDir string) map[string]string {
	return map[string]string{
		"AppName":    name,
		"Version":    version,
		"InstallDir": installDir,
		"Arch":       "amd64",
	}
}

func (i *installer) runScript(ctx context.Context, app *manifest.App, hookName string, env map[string]string) error {
	return nil
}

func (i *installer) runInstallScript(ctx context.Context, app *manifest.App, installDir string, env map[string]string) error {
	if i.jsEngine == nil {
		return nil
	}

	vm := i.jsEngine.GetVM()
	if vm == nil {
		return fmt.Errorf("JavaScript 虚拟机未初始化")
	}

	vm.Set("console", map[string]interface{}{
		"log": func(args ...interface{}) {
			fmt.Println(args...)
		},
	})

	for k, v := range env {
		vm.Set(k, v)
	}

	scriptPath := ""
	if app.Ref != nil {
		scriptPath = app.Ref.ScriptPath
	}

	if scriptPath == "" {
		return nil
	}

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("读取脚本文件: %w", err)
	}

	_, err = vm.RunScript("install.js", string(scriptContent))
	if err != nil {
		return fmt.Errorf("执行脚本: %w", err)
	}

	return nil
}

func (i *installer) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	installed, err := i.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return fmt.Errorf("获取安装信息: %w", err)
	}

	if err := i.runUninstallScript(ctx, installed); err != nil {
		fmt.Fprintf(os.Stderr, "卸载脚本执行失败: %v\n", err)
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

	if err := i.storage.DeleteInstalledApp(ctx, name); err != nil {
		return fmt.Errorf("删除安装记录: %w", err)
	}

	fmt.Printf("✓ %s 卸载成功\n", name)
	return nil
}

func (i *installer) runUninstallScript(ctx context.Context, installed *manifest.InstalledApp) error {
	return nil
}

func (i *installer) Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	return i.Install(ctx, app, InstallOptions{
		Arch:       "amd64",
		Force:      opts.Force,
		InstallDir: installed.InstallDir,
	})
}

func (i *installer) Switch(ctx context.Context, name, version string) error {
	return nil
}
