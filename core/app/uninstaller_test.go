package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/core/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUninstaller(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建一个 installer 实例
	inst := &installer{storage: storage, installBase: tmpDir}
	uninstaller := NewUninstaller(inst)
	assert.NotNil(t, uninstaller)
}

func TestUninstaller_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建一个 installer 实例
	inst := &installer{storage: storage, installBase: tmpDir}
	uninstaller := NewUninstaller(inst)
	ctx := context.Background()

	// 测试卸载未安装的应用
	err = uninstaller.Uninstall(ctx, "non-existent-app", UninstallOptions{})
	assert.Error(t, err)
}

func TestUninstaller_UninstallWithPurge(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建一个 installer 实例
	inst := &installer{storage: storage, installBase: tmpDir}
	uninstaller := NewUninstaller(inst)
	ctx := context.Background()

	// 测试带 purge 选项卸载未安装的应用
	err = uninstaller.Uninstall(ctx, "non-existent-app", UninstallOptions{Purge: true})
	assert.Error(t, err)
}

func TestUninstaller_UninstallInstalledApp(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建一个已安装的应用记录
	installedApp := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  filepath.Join(tmpDir, "apps", "test-app"),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	err = storage.SaveInstalledApp(ctx, installedApp)
	require.NoError(t, err)

	// 创建安装目录结构
	appDir := filepath.Join(tmpDir, "apps", "test-app")
	versionDir := filepath.Join(appDir, "1.0.0")
	err = os.MkdirAll(versionDir, DefaultDirPerm)
	require.NoError(t, err)

	// 创建一个 installer 实例
	inst := &installer{storage: storage, installBase: tmpDir}
	uninstaller := NewUninstaller(inst)

	// 测试卸载已安装的应用（不 purge）
	err = uninstaller.Uninstall(ctx, "test-app", UninstallOptions{})
	assert.NoError(t, err)

	// 验证应用已被移除
	_, err = storage.GetInstalledApp(ctx, "test-app")
	assert.Error(t, err)
}

func TestUninstaller_UninstallWithPurgeInstalledApp(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建一个已安装的应用记录
	installedApp := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  filepath.Join(tmpDir, "apps", "test-app"),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	err = storage.SaveInstalledApp(ctx, installedApp)
	require.NoError(t, err)

	// 创建安装目录和文件
	appDir := filepath.Join(tmpDir, "apps", "test-app")
	versionDir := filepath.Join(appDir, "1.0.0")
	err = os.MkdirAll(versionDir, DefaultDirPerm)
	require.NoError(t, err)
	testFile := filepath.Join(versionDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// 创建一个 installer 实例
	inst := &installer{storage: storage, installBase: tmpDir}
	uninstaller := NewUninstaller(inst)

	// 测试带 purge 选项卸载已安装的应用
	err = uninstaller.Uninstall(ctx, "test-app", UninstallOptions{Purge: true})
	assert.NoError(t, err)

	// 验证目录已被删除
	_, err = os.Stat(appDir)
	assert.True(t, os.IsNotExist(err))
}

func TestUninstallOptions_Validation(t *testing.T) {
	tests := []struct {
		name  string
		opts  UninstallOptions
		purge bool
	}{
		{
			name:  "purge true",
			opts:  UninstallOptions{Purge: true},
			purge: true,
		},
		{
			name:  "purge false",
			opts:  UninstallOptions{Purge: false},
			purge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.purge, tt.opts.Purge)
		})
	}
}
