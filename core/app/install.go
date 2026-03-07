package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chopsticks/core/bucket"
	"chopsticks/core/conflict"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/engine"
	"chopsticks/engine/archive"
	"chopsticks/engine/checksum"
	"chopsticks/engine/fetch"
	"chopsticks/pkg/errors"
)

const (
	// DefaultArch 默认架构
	DefaultArch = "amd64"
	// DefaultVersion 默认版本
	DefaultVersion = "latest"
	// DefaultBucket 默认软件源
	DefaultBucket = "main"

	// DefaultFilePerm 默认文件权限
	DefaultFilePerm = 0644
)

// InstallOptions 安装选项 - 字段按大小从大到小排列
type InstallOptions struct {
	InstallDir string  // 16 bytes (string header)
	Arch       string  // 16 bytes (string header)
	Force      bool    // 1 byte
	Isolate    bool    // 1 byte
	NoDeps     bool    // 1 byte
	_          [7]byte // padding for alignment
}

// UninstallOptions 卸载选项
type UninstallOptions struct {
	Purge bool // 1 byte
}

// RefreshOptions 刷新选项
type RefreshOptions struct {
	Force bool // 1 byte
}

type Installer interface {
	Install(ctx context.Context, app *manifest.App, opts InstallOptions) error
	Uninstall(ctx context.Context, name string, opts UninstallOptions) error
	Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error
	Switch(ctx context.Context, name, version string) error
}

// installer 安装器 - 字段按大小从大到小排列
type installer struct {
	jsEngine    *engine.JSEngine    // 8 bytes
	storage     store.LegacyStorage // 16 bytes (interface)
	config      interface{}         // 16 bytes (interface)
	downloadDir string              // 16 bytes (string header)
	installBase string              // 16 bytes (string header)
}

var _ Installer = (*installer)(nil)

func NewInstaller(storage store.LegacyStorage, config interface{}, jsEngine *engine.JSEngine, installBase string) Installer {
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
		return errors.Newf(errors.KindInvalidInput, "invalid app info")
	}

	appName := app.Script.Name

	// Conflict detection
	detector := conflict.NewDetector(i.storage, i.installBase)
	conflictResult, err := detector.Detect(ctx, app)
	if err != nil {
		// Conflict detection failed, log warning but don't block installation
		fmt.Fprintf(os.Stderr, "Warning: conflict detection failed: %v\n", err)
	} else if conflictResult != nil && len(conflictResult.Conflicts) > 0 {
		// Format and display conflict report
		formatter := conflict.NewFormatter(true)
		fmt.Println(formatter.Format(conflictResult))

		// If serious conflicts exist and force is not used, block installation
		if conflict.ShouldBlockInstall(conflictResult, opts.Force) {
			return errors.Newf(errors.KindConflict, "serious conflicts detected, please resolve before installing or use --force to force install")
		}

		// If warning level conflicts exist and force is not used, ask user
		if conflictResult.HasWarning && !opts.Force {
			fmt.Print("Continue installation? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if !isYesResponse(response) {
				return errors.Newf(errors.KindCancelled, "user cancelled installation")
			}
		}
	}

	installed, err := i.storage.GetInstalledApp(ctx, appName)
	if err == nil && installed != nil && !opts.Force {
		return errors.NewAppAlreadyInstalled(appName, installed.Version)
	}

	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(i.installBase, appName)
	}

	if err := os.MkdirAll(installDir, DefaultDirPerm); err != nil {
		return errors.Wrapf(err, "create install directory %s", installDir)
	}

	if err := os.MkdirAll(i.downloadDir, DefaultDirPerm); err != nil {
		return errors.Wrapf(err, "create download directory %s", i.downloadDir)
	}

	version := DefaultVersion
	if app.Meta != nil && app.Meta.Version != "" {
		version = app.Meta.Version
	}

	arch := opts.Arch
	if arch == "" {
		arch = DefaultArch
	}

	downloadInfo, err := i.getDownloadInfo(app, version, arch)
	if err != nil {
		return errors.Wrap(err, "get download info")
	}

	downloadedPath, err := i.downloadPackage(ctx, downloadInfo)
	if err != nil {
		return errors.Wrap(err, "download package")
	}
	defer os.Remove(downloadedPath)

	if downloadInfo.Hash != "" {
		if err := i.verifyChecksum(downloadedPath, downloadInfo.Hash); err != nil {
			return errors.Wrap(err, "verify checksum")
		}
	}

	extractDir := filepath.Join(installDir, version)
	if err := os.MkdirAll(extractDir, DefaultDirPerm); err != nil {
		return errors.Wrapf(err, "create version directory %s", extractDir)
	}

	if err := i.extractPackage(downloadedPath, extractDir); err != nil {
		return errors.Wrap(err, "extract package")
	}

	hookEnv := i.buildHookEnv(appName, version, extractDir, downloadedPath)

	if err := i.runScript(ctx, app, "preInstall", hookEnv); err != nil {
		return errors.NewHookFailed("preInstall", err)
	}

	if err := i.runInstallScript(ctx, app, extractDir, hookEnv); err != nil {
		return errors.NewScriptFailed("onInstall", err)
	}

	if err := i.runScript(ctx, app, "postInstall", hookEnv); err != nil {
		return errors.NewHookFailed("postInstall", err)
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
		return errors.Wrap(err, "save install record")
	}

	fmt.Printf("✓ %s (%s) installed successfully\n", appName, version)
	return nil
}

// isYesResponse checks if the response is a yes answer
func isYesResponse(response string) bool {
	lower := strings.ToLower(strings.TrimSpace(response))
	return lower == "y" || lower == "yes"
}

func (i *installer) getDownloadInfo(app *manifest.App, version, arch string) (*manifest.DownloadInfo, error) {
	if app.Meta != nil && app.Meta.Versions != nil {
		if versionInfo, ok := app.Meta.Versions[version]; ok {
			if downloadInfo, ok := versionInfo.Downloads[arch]; ok {
				return &downloadInfo, nil
			}
		}
	}
	return nil, errors.NewVersionNotFound(app.Script.Name, version)
}

func (i *installer) downloadPackage(ctx context.Context, info *manifest.DownloadInfo) (string, error) {
	filename := filepath.Base(info.URL)
	destPath := filepath.Join(i.downloadDir, filename)

	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	fmt.Printf("downloading: %s\n", info.URL)
	if err := fetch.DownloadWithContext(ctx, info.URL, destPath); err != nil {
		return "", errors.NewDownloadFailed(info.URL, err)
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

	var calc checksum.Calculator
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
		return errors.Wrap(err, "checksum verification")
	}
	if !ok {
		return errors.NewChecksumMismatch(hash, "")
	}
	return nil
}

func (i *installer) extractPackage(archivePath, destDir string) error {
	if archivePath == "" || destDir == "" {
		return nil
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return errors.Newf(errors.KindNotFound, "archive not found: %s", archivePath)
	}

	if !archive.IsArchive(archivePath) {
		return errors.Newf(errors.KindInvalidInput, "unsupported archive format: %s", archivePath)
	}

	fmt.Printf("extracting: %s\n", filepath.Base(archivePath))
	if err := archive.Extract(archivePath, destDir); err != nil {
		return errors.NewArchiveExtractFailed(archivePath, err)
	}

	return nil
}

func (i *installer) buildInstallEnv(name, version, installDir string) map[string]string {
	return map[string]string{
		"name":       name,
		"version":    version,
		"installDir": installDir,
		"cookDir":    installDir,
		"arch":       DefaultArch,
		"bucket":     DefaultBucket,
	}
}

// buildHookEnv builds hook environment variables, including downloadPath
func (i *installer) buildHookEnv(name, version, installDir, downloadPath string) map[string]string {
	env := i.buildInstallEnv(name, version, installDir)
	env["downloadPath"] = downloadPath
	return env
}

// runScript executes lifecycle hook methods in the app script
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
		return errors.Wrap(err, "load script file")
	}

	ctxMap := make(map[string]interface{})
	for k, v := range env {
		ctxMap[k] = v
	}

	i.jsEngine.SetContext(ctxMap)

	if err := i.jsEngine.CallFunction(hookName, ctxMap); err != nil {
		return errors.NewHookFailed(hookName, err)
	}

	return nil
}

// runInstallScript executes the onInstall method of the app script
func (i *installer) runInstallScript(ctx context.Context, app *manifest.App, installDir string, env map[string]string) error {
	// Add installDir to environment variables
	env["InstallDir"] = installDir

	// Call onInstall hook
	return i.runScript(ctx, app, "onInstall", env)
}

func (i *installer) Uninstall(ctx context.Context, name string, opts UninstallOptions) error {
	installed, err := i.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	app, err := i.loadAppManifest(ctx, installed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load app manifest failed: %v\n", err)
	} else {
		if err := i.runScript(ctx, app, "preUninstall", map[string]string{
			"name":       name,
			"version":    installed.Version,
			"installDir": installed.InstallDir,
			"cookDir":    installed.InstallDir,
			"arch":       DefaultArch,
			"bucket":     installed.Bucket,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "preUninstall hook failed: %v\n", err)
		}
	}

	installDir := installed.InstallDir
	if opts.Purge {
		if err := os.RemoveAll(installDir); err != nil {
			return errors.Wrapf(err, "remove install directory %s", installDir)
		}
	} else {
		versionDir := filepath.Join(installDir, installed.Version)
		if err := os.RemoveAll(versionDir); err != nil {
			return errors.Wrapf(err, "remove version directory %s", versionDir)
		}
	}

	if app != nil {
		if err := i.runUninstallScript(ctx, app, installed); err != nil {
			fmt.Fprintf(os.Stderr, "uninstall script failed: %v\n", err)
		}

		if err := i.runScript(ctx, app, "postUninstall", map[string]string{
			"name":       name,
			"version":    installed.Version,
			"installDir": installed.InstallDir,
			"cookDir":    installed.InstallDir,
			"arch":       DefaultArch,
			"bucket":     installed.Bucket,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "postUninstall hook failed: %v\n", err)
		}
	}

	if err := i.storage.DeleteInstalledApp(ctx, name); err != nil {
		return errors.Wrap(err, "delete install record")
	}

	fmt.Printf("✓ %s uninstalled successfully\n", name)
	return nil
}

func (i *installer) loadAppManifest(ctx context.Context, installed *manifest.InstalledApp) (*manifest.App, error) {
	return i.doLoadAppManifest(ctx, installed)
}

func (i *installer) doLoadAppManifest(ctx context.Context, installed *manifest.InstalledApp) (*manifest.App, error) {
	bucketName := installed.Bucket
	if bucketName == "" {
		bucketName = DefaultBucket
	}

	bucketPath := filepath.Join(i.installBase, "..", "buckets", bucketName)
	loader := bucket.NewLoader()
	b, err := loader.Load(ctx, bucketPath)
	if err != nil {
		return nil, err
	}

	ref, ok := b.Apps[installed.Name]
	if !ok {
		return nil, errors.NewAppNotFound(installed.Name)
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

// runUninstallScript executes the onUninstall method of the app script
func (i *installer) runUninstallScript(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp) error {
	env := map[string]string{
		"name":       installed.Name,
		"version":    installed.Version,
		"installDir": installed.InstallDir,
		"cookDir":    installed.InstallDir,
		"arch":       DefaultArch,
		"bucket":     installed.Bucket,
	}

	// Call onUninstall hook
	return i.runScript(ctx, app, "onUninstall", env)
}

func (i *installer) Refresh(ctx context.Context, app *manifest.App, installed *manifest.InstalledApp, opts RefreshOptions) error {
	backupDir := installed.InstallDir + ".backup"
	if _, err := os.Stat(backupDir); err == nil {
		os.RemoveAll(backupDir)
	}

	if err := os.Rename(installed.InstallDir, backupDir); err != nil {
		return errors.Wrap(err, "backup current version")
	}

	err := i.Install(ctx, app, InstallOptions{
		Arch:       DefaultArch,
		Force:      opts.Force,
		InstallDir: installed.InstallDir,
	})
	if err != nil {
		os.Rename(backupDir, installed.InstallDir)
		return errors.NewUpdateFailed(app.Script.Name, err)
	}

	os.RemoveAll(backupDir)
	return nil
}

func (i *installer) Switch(ctx context.Context, name, version string) error {
	installed, err := i.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return errors.NewAppNotInstalled(name)
	}

	currentVersion := installed.Version
	if currentVersion == version {
		return errors.Newf(errors.KindInvalidInput, "already at version %s", version)
	}

	newVersionDir := filepath.Join(installed.InstallDir, version)
	if _, err := os.Stat(newVersionDir); err != nil {
		return errors.NewVersionNotFound(name, version)
	}

	currentVersionDir := filepath.Join(installed.InstallDir, currentVersion)
	backupDir := filepath.Join(installed.InstallDir, currentVersion+".old")

	if err := os.Rename(currentVersionDir, backupDir); err != nil {
		return errors.Wrap(err, "backup current version")
	}

	if err := os.Rename(newVersionDir, currentVersionDir); err != nil {
		os.Rename(backupDir, currentVersionDir)
		return errors.Wrap(err, "switch version")
	}

	os.RemoveAll(backupDir)

	installed.Version = version
	installed.UpdatedAt = time.Now()
	if err := i.storage.SaveInstalledApp(ctx, installed); err != nil {
		return errors.Wrap(err, "save version switch")
	}

	fmt.Printf("✓ %s switched to version %s\n", name, version)
	return nil
}
