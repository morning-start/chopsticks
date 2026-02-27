package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/engine"
	"chopsticks/engine/archive"
	"chopsticks/engine/checksum"
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
	storage     store.Storage
	config      interface{}
	jsEngine    *engine.JSEngine
	downloadDir string
	installBase string
}

var _ Installer = (*installer)(nil)

func NewInstaller(storage store.Storage, config interface{}, jsEngine *engine.JSEngine, installBase string) Installer {
	return &installer{
		storage:     storage,
		config:      config,
		jsEngine:    jsEngine,
		installBase: installBase,
		downloadDir: filepath.Join(installBase, "tmp"),
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
	if err := fetch.DownloadWithContext(ctx, info.URL, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

func (i *installer) verifyChecksum(filePath, expectedHash string) error {
	if expectedHash == "" {
		return nil
	}

	alg := "sha256"
	hash := expectedHash

	parts := strings.SplitN(expectedHash, ":", 2)
	if len(parts) == 2 {
		alg = strings.TrimSpace(strings.ToLower(parts[0]))
		hash = strings.TrimSpace(parts[1])
	}

	calc := checksum.New(checksum.SHA256)
	switch alg {
	case "md5":
		calc = checksum.New(checksum.MD5)
	case "sha256":
		calc = checksum.New(checksum.SHA256)
	case "sha512":
		calc = checksum.New(checksum.SHA512)
	default:
		calc = checksum.New(checksum.SHA256)
	}

	ok, err := calc.Verify(filePath, hash)
	if err != nil {
		return fmt.Errorf("校验和验证失败: %w", err)
	}
	if !ok {
		return fmt.Errorf("校验和不匹配")
	}
	return nil
}

func (i *installer) extractPackage(archivePath, destDir string) error {
	if archivePath == "" || destDir == "" {
		return nil
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("安装包不存在: %s", archivePath)
	}

	if !archive.IsArchive(archivePath) {
		return fmt.Errorf("不支持的压缩格式: %s", archivePath)
	}

	fmt.Printf("解压: %s\n", filepath.Base(archivePath))
	if err := archive.Extract(archivePath, destDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

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

// runScript 执行应用脚本中的生命周期钩子方法
func (i *installer) runScript(ctx context.Context, app *manifest.App, hookName string, env map[string]string) error {
	return i.executeScript(ctx, app, hookName, env)
}

func (i *installer) executeScript(_ context.Context, app *manifest.App, hookName string, env map[string]string) error {
	if i.jsEngine == nil {
		return nil
	}

	scriptPath := ""
	if app.Ref != nil {
		scriptPath = app.Ref.ScriptPath
	}

	if scriptPath == "" {
		return nil
	}

	if err := i.jsEngine.LoadFile(scriptPath); err != nil {
		return fmt.Errorf("加载脚本文件: %w", err)
	}

	ctxMap := make(map[string]interface{})
	for k, v := range env {
		ctxMap[k] = v
	}

	i.jsEngine.SetContext(ctxMap)

	if err := i.jsEngine.CallFunction(hookName, ctxMap); err != nil {
		return fmt.Errorf("执行 %s 钩子: %w", hookName, err)
	}

	return nil
}

// runInstallScript 执行应用脚本的 onInstall 方法
func (i *installer) runInstallScript(ctx context.Context, app *manifest.App, installDir string, env map[string]string) error {
	// 添加 installDir 到环境变量
	env["InstallDir"] = installDir

	// 调用 onInstall 钩子
	return i.runScript(ctx, app, "onInstall", env)
}

func (i *installer) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	installed, err := i.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return fmt.Errorf("获取安装信息: %w", err)
	}

	app, err := i.loadAppManifest(ctx, installed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载应用清单失败: %v\n", err)
	} else {
		if err := i.runScript(ctx, app, "preUninstall", map[string]string{
			"AppName":    name,
			"Version":    installed.Version,
			"InstallDir": installed.InstallDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "preUninstall 钩子失败: %v\n", err)
		}
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

	if app != nil {
		if err := i.runUninstallScript(ctx, app, installed); err != nil {
			fmt.Fprintf(os.Stderr, "卸载脚本执行失败: %v\n", err)
		}

		if err := i.runScript(ctx, app, "postUninstall", map[string]string{
			"AppName":    name,
			"Version":    installed.Version,
			"InstallDir": installed.InstallDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "postUninstall 钩子失败: %v\n", err)
		}
	}

	if err := i.storage.DeleteInstalledApp(ctx, name); err != nil {
		return fmt.Errorf("删除安装记录: %w", err)
	}

	fmt.Printf("✓ %s 卸载成功\n", name)
	return nil
}

func (i *installer) loadAppManifest(ctx context.Context, installed *manifest.InstalledApp) (*manifest.App, error) {
	return i.doLoadAppManifest(ctx, installed)
}

func (i *installer) doLoadAppManifest(ctx context.Context, installed *manifest.InstalledApp) (*manifest.App, error) {
	bucketName := installed.Bucket
	if bucketName == "" {
		bucketName = "main"
	}

	bucketPath := filepath.Join(i.installBase, "..", "buckets", bucketName)
	loader := bucket.NewLoader()
	b, err := loader.Load(ctx, bucketPath)
	if err != nil {
		return nil, err
	}

	ref, ok := b.Apps[installed.Name]
	if !ok {
		return nil, fmt.Errorf("应用信息不存在")
	}

	return &manifest.App{
		Script: &manifest.AppScript{
			Name:        ref.Name,
			Description: ref.Description,
			Bucket:      bucketName,
		},
		Meta: &manifest.AppMeta{
			Version: ref.Version,
		},
		Ref: ref,
	}, nil
}

// runUninstallScript 执行应用脚本的 onUninstall 方法
func (i *installer) runUninstallScript(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp) error {
	env := map[string]string{
		"AppName":    installed.Name,
		"Version":    installed.Version,
		"InstallDir": installed.InstallDir,
	}

	// 调用 onUninstall 钩子
	return i.runScript(ctx, app, "onUninstall", env)
}

func (i *installer) Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	backupDir := installed.InstallDir + ".backup"
	if _, err := os.Stat(backupDir); err == nil {
		os.RemoveAll(backupDir)
	}

	if err := os.Rename(installed.InstallDir, backupDir); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}

	err := i.Install(ctx, app, InstallOptions{
		Arch:       "amd64",
		Force:      opts.Force,
		InstallDir: installed.InstallDir,
	})
	if err != nil {
		os.Rename(backupDir, installed.InstallDir)
		return fmt.Errorf("更新失败，已回滚: %w", err)
	}

	os.RemoveAll(backupDir)
	return nil
}

func (i *installer) Switch(ctx context.Context, name, version string) error {
	installed, err := i.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return fmt.Errorf("应用未安装: %w", err)
	}

	currentVersion := installed.Version
	if currentVersion == version {
		return fmt.Errorf("当前已是版本 %s", version)
	}

	newVersionDir := filepath.Join(installed.InstallDir, version)
	if _, err := os.Stat(newVersionDir); err != nil {
		return fmt.Errorf("版本 %s 不存在", version)
	}

	currentVersionDir := filepath.Join(installed.InstallDir, currentVersion)
	backupDir := filepath.Join(installed.InstallDir, currentVersion+".old")

	if err := os.Rename(currentVersionDir, backupDir); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}

	if err := os.Rename(newVersionDir, currentVersionDir); err != nil {
		os.Rename(backupDir, currentVersionDir)
		return fmt.Errorf("切换版本失败: %w", err)
	}

	os.RemoveAll(backupDir)

	installed.Version = version
	installed.UpdatedAt = time.Now()
	if err := i.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("保存版本切换: %w", err)
	}

	fmt.Printf("✓ %s 切换到版本 %s\n", name, version)
	return nil
}
