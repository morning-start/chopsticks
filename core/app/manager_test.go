package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestManager 创建测试管理器
func createTestManager(t *testing.T) (store.Storage, store.LegacyStorage, AppManager, string) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	require.NoError(t, err)
	adapter := store.NewStorageAdapter(storage, tmpDir)

	bucketsDir := filepath.Join(tmpDir, "buckets")
	require.NoError(t, os.MkdirAll(bucketsDir, 0755))
	bucketMgr := bucket.NewManager(adapter, nil, bucketsDir, nil)

	mgr := NewManager(bucketMgr, adapter, nil, nil, tmpDir)
	return storage, adapter, mgr, tmpDir
}

// TestNewManager 测试管理器创建
func TestNewManager(t *testing.T) {
	storage, _, mgr, _ := createTestManager(t)
	defer storage.Close()
	assert.NotNil(t, mgr)
}

// TestManagerErrors 测试错误变量
func TestManagerErrors(t *testing.T) {
	assert.NotNil(t, ErrAppNotFound)
	assert.NotNil(t, ErrAppAlreadyInstalled)
	assert.NotNil(t, ErrVersionNotFound)
	assert.NotNil(t, ErrDependencyConflict)
}

// TestManager_ListInstalled 测试列出已安装应用
func TestManager_ListInstalled(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	// 测试空列表
	apps, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Empty(t, apps)
	assert.NotNil(t, apps)

	// 添加已安装应用
	ctx := context.Background()
	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: filepath.Join(tmpDir, "test-app"),
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	// 再次测试
	apps, err = mgr.ListInstalled()
	require.NoError(t, err)
	assert.Len(t, apps, 1)
	assert.Equal(t, "test-app", apps[0].Name)
}

// TestManager_Info 测试获取应用信息
func TestManager_Info(t *testing.T) {
	storage, _, mgr, _ := createTestManager(t)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := newMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "test-app",
			Description: "Test application",
			Homepage:    "https://example.com",
			License:     "MIT",
			Category:    "dev",
			Tags:        []string{"test"},
		},
		Meta: &manifest.AppMeta{
			Version: "1.0.0",
		},
	})

	ctx := context.Background()

	// 测试获取应用信息
	info, err := mgr.Info(ctx, "main", "test-app")
	require.NoError(t, err)
	assert.Equal(t, "test-app", info.Name)
	assert.Equal(t, "Test application", info.Description)

	// 测试不存在的应用
	_, err = mgr.Info(ctx, "main", "nonexistent")
	assert.Error(t, err)
}

// TestManager_Info_Installed 测试获取已安装应用信息
func TestManager_Info_Installed(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := newMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "test-app",
			Description: "Test application",
		},
		Meta: &manifest.AppMeta{
			Version: "2.0.0",
		},
	})

	ctx := context.Background()

	// 添加已安装应用（旧版本）
	installed := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: filepath.Join(tmpDir, "test-app"),
	}
	err := adapter.SaveInstalledApp(ctx, installed)
	require.NoError(t, err)

	// 测试获取信息
	info, err := mgr.Info(ctx, "main", "test-app")
	require.NoError(t, err)
	assert.True(t, info.Installed)
	assert.Equal(t, "1.0.0", info.InstalledVersion)
	assert.Equal(t, "2.0.0", info.Version)
}

// TestManager_Search 测试搜索应用
func TestManager_Search(t *testing.T) {
	storage, _, mgr, _ := createTestManager(t)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := newMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "git",
			Description: "Git version control",
		},
		Meta: &manifest.AppMeta{
			Version: "2.30.0",
		},
	})

	ctx := context.Background()

	// 测试搜索
	results, err := mgr.Search(ctx, "main", "git")
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}
