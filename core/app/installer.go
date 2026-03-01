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
	// Create progress manager
	pm := output.NewProgressManager()
	defer pm.Wait()

	installDir := opts.InstallDir
	if installDir == "" {
		installDir = filepath.Join(i.installer.installBase, app.Script.Name)
	}

	if err := os.MkdirAll(installDir, DefaultDirPerm); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}

	arch := opts.Arch
	if arch == "" {
		arch = DefaultArch
	}

	version := app.Meta.Version
	if version == "" {
		version = DefaultVersion
	}

	// Define installation stages
	stages := []string{"Prepare", "Download", "Verify", "Extract", "Complete"}
	progressBar := pm.AddInstallBar(app.Script.Name, stages)

	// Stage 1: Prepare
	progressBar.SetStage(0)
	downloadInfo, err := i.getDownloadInfo(app, version, arch)
	if err != nil {
		return fmt.Errorf("get download info: %w", err)
	}
	progressBar.CompleteStage()

	// Stage 2: Download
	progressBar.SetStage(1)
	cacheFile := filepath.Join(i.installer.downloadDir, fmt.Sprintf("%s-%s-%s", app.Script.Name, version, arch))
	if err := i.DownloadWithProgress(ctx, downloadInfo.URL, cacheFile, pm); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	progressBar.CompleteStage()

	// Stage 3: Verify
	progressBar.SetStage(2)
	if downloadInfo.Hash != "" {
		alg := checksum.AutoDetectAlgorithm(downloadInfo.Hash)
		if err := i.Verify(cacheFile, downloadInfo.Hash, alg); err != nil {
			return fmt.Errorf("verify failed: %w", err)
		}
	}
	progressBar.CompleteStage()

	// Stage 4: Extract
	progressBar.SetStage(3)
	extractDir := filepath.Join(installDir, version)
	if err := i.Extract(cacheFile, extractDir); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}
	progressBar.CompleteStage()

	// Stage 5: Complete
	progressBar.SetStage(4)
	installed := &manifest.InstalledApp{
		Name:        app.Script.Name,
		Version:     version,
		Bucket:      app.Script.Bucket,
		InstallDir:  installDir,
		InstalledAt: time.Now(),
	}

	if err := i.installer.storage.SaveInstalledApp(ctx, installed); err != nil {
		return fmt.Errorf("save install record: %w", err)
	}
	progressBar.Complete()

	fmt.Printf("✓ %s (%s) installed successfully\n", app.Script.Name, version)
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
	return nil, fmt.Errorf("download info not found for %s/%s", version, arch)
}

func (i *appInstaller) Download(url, dest string) error {
	return fetch.Download(url, dest)
}

// DownloadWithProgress downloads file with progress bar
func (i *appInstaller) DownloadWithProgress(ctx context.Context, url, dest string, pm *output.ProgressManager) error {
	return fetch.DownloadWithProgress(ctx, url, dest, pm)
}

func (i *appInstaller) Verify(path, hash string, alg checksum.Algorithm) error {
	ok, err := checksum.VerifyFile(path, hash, alg)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func (i *appInstaller) Extract(src, dest string) error {
	return archive.Extract(src, dest)
}
