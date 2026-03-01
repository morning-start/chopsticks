package app

import (
	"context"
	"path/filepath"
	"testing"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewManager 测试管理器创建
func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)
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
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)

	// 测试空列表
	apps, err := mgr.ListInstalled()
	require.NoError(t, err)
	assert.Empty(t, apps)
	assert.NotNil(t, apps) // 应该返回空切片而不是 nil

	// 添加已安装应用
	ctx := context.Background()
	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: filepath.Join(tmpDir, "test-app"),
	}
	err = storage.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	// 再次测试
	apps, err = mgr.ListInstalled()
	require.NoError(t, err)
	assert.Len(t, apps, 1)
	assert.Equal(t, "test-app", apps[0].Name)
}

// TestManager_Info 测试获取应用信息
func TestManager_Info(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := NewMockBucketManager()
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

	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	// 测试获取应用信息
	info, err := mgr.Info(ctx, "main", "test-app")
	require.NoError(t, err)
	assert.Equal(t, "test-app", info.Name)
	assert.Equal(t, "Test application", info.Description)
	assert.Equal(t, "https://example.com", info.Homepage)
	assert.Equal(t, "MIT", info.License)
	assert.False(t, info.Installed)

	// 测试不存在的应用
	_, err = mgr.Info(ctx, "main", "nonexistent")
	assert.Error(t, err)
}

// TestManager_Info_Installed 测试获取已安装应用信息
func TestManager_Info_Installed(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := NewMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "test-app",
			Description: "Test application",
		},
		Meta: &manifest.AppMeta{
			Version: "2.0.0",
		},
	})

	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	// 添加已安装应用（旧版本）
	installed := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: filepath.Join(tmpDir, "test-app"),
	}
	err = storage.SaveInstalledApp(ctx, installed)
	require.NoError(t, err)

	// 测试获取信息
	info, err := mgr.Info(ctx, "main", "test-app")
	require.NoError(t, err)
	assert.True(t, info.Installed)
	assert.Equal(t, "1.0.0", info.InstalledVersion)
	assert.Equal(t, "2.0.0", info.Version) // bucket 中的最新版本
}

// TestManager_Search 测试搜索应用
func TestManager_Search(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// 创建 mock bucket manager
	mockBucketMgr := NewMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "git",
			Description: "Version control system",
		},
	})
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "github-cli",
			Description: "GitHub command line tool",
		},
	})
	mockBucketMgr.AddApp("extras", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "vscode",
			Description: "Code editor",
		},
	})

	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	// 测试搜索所有 bucket
	results, err := mgr.Search(ctx, "git", "")
	require.NoError(t, err)
	assert.Len(t, results, 2) // git 和 github-cli

	// 测试搜索特定 bucket
	results, err = mgr.Search(ctx, "code", "extras")
	require.NoError(t, err)
	assert.Len(t, results, 1) // vscode

	// 测试无结果搜索
	results, err = mgr.Search(ctx, "nonexistent", "")
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestManager_Search_EdgeCases 测试搜索边界情况
func TestManager_Search_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mockBucketMgr := NewMockBucketManager()
	mockBucketMgr.AddApp("main", &manifest.App{
		Script: &manifest.AppScript{
			Name:        "test-app",
			Description: "A test application",
			Tags:        []string{"dev", "test"},
		},
	})

	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"空查询", "", 0},
		{"大小写不敏感", "TEST", 1},
		{"部分匹配", "app", 1},
		{"描述匹配", "application", 1},
		{"无匹配", "xyz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := mgr.Search(ctx, tt.query, "")
			require.NoError(t, err)
			assert.Len(t, results, tt.expected)
		})
	}
}

// TestManager_Install_NoBucket 测试安装时 bucket 不存在
func TestManager_Install_NoBucket(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mockBucketMgr := NewMockBucketManager()
	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	spec := InstallSpec{
		Bucket: "nonexistent",
		Name:   "test-app",
	}
	opts := InstallOptions{}

	err = mgr.Install(ctx, spec, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket")
}

// TestManager_Install_AppNotFound 测试安装时应用不存在
func TestManager_Install_AppNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mockBucketMgr := NewMockBucketManager()
	mockBucketMgr.CreateBucket("main")
	mgr := NewManager(mockBucketMgr, storage, nil, nil, tmpDir)
	ctx := context.Background()

	spec := InstallSpec{
		Bucket: "main",
		Name:   "nonexistent",
	}
	opts := InstallOptions{}

	err = mgr.Install(ctx, spec, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app")
}

// TestManager_Remove_NotInstalled 测试卸载未安装的应用
func TestManager_Remove_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)
	ctx := context.Background()

	opts := RemoveOptions{}
	err = mgr.Remove(ctx, "not-installed", opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAppNotFound)
}

// TestManager_Update_NotInstalled 测试更新未安装的应用
func TestManager_Update_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)
	ctx := context.Background()

	opts := UpdateOptions{}
	err = mgr.Update(ctx, "not-installed", opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAppNotFound)
}

// TestManager_UpdateAll_Empty 测试更新所有应用（空列表）
func TestManager_UpdateAll_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)
	ctx := context.Background()

	opts := UpdateOptions{}
	err = mgr.UpdateAll(ctx, opts)
	assert.NoError(t, err) // 空列表应该返回 nil
}

// MockBucketManager 用于测试的 mock bucket manager
type MockBucketManager struct {
	buckets map[string]*manifest.BucketConfig
	apps    map[string]map[string]*manifest.App
}

func NewMockBucketManager() *MockBucketManager {
	return &MockBucketManager{
		buckets: make(map[string]*manifest.BucketConfig),
		apps:    make(map[string]map[string]*manifest.App),
	}
}

func (m *MockBucketManager) CreateBucket(name string) {
	m.buckets[name] = &manifest.BucketConfig{Name: name}
	m.apps[name] = make(map[string]*manifest.App)
}

func (m *MockBucketManager) AddApp(bucketName string, app *manifest.App) {
	if m.apps[bucketName] == nil {
		m.apps[bucketName] = make(map[string]*manifest.App)
	}
	m.apps[bucketName][app.Script.Name] = app
}

func (m *MockBucketManager) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	if b, ok := m.buckets[name]; ok {
		return b, nil
	}
	return nil, bucket.ErrBucketNotFound
}

func (m *MockBucketManager) ListBuckets(ctx context.Context) ([]string, error) {
	var result []string
	for name := range m.buckets {
		result = append(result, name)
	}
	return result, nil
}

func (m *MockBucketManager) Add(ctx context.Context, name, url string, opts bucket.AddOptions) error {
	m.CreateBucket(name)
	return nil
}

func (m *MockBucketManager) Remove(ctx context.Context, name string, purge bool) error {
	delete(m.buckets, name)
	delete(m.apps, name)
	return nil
}

func (m *MockBucketManager) Update(ctx context.Context, name string) error {
	return nil
}

func (m *MockBucketManager) UpdateAll(ctx context.Context) error {
	return nil
}

func (m *MockBucketManager) GetApp(ctx context.Context, bucketName, appName string) (*manifest.App, error) {
	if apps, ok := m.apps[bucketName]; ok {
		if app, ok := apps[appName]; ok {
			return app, nil
		}
	}
	return nil, ErrAppNotFound
}

func (m *MockBucketManager) ListApps(ctx context.Context, bucketName string) (map[string]*manifest.AppRef, error) {
	result := make(map[string]*manifest.AppRef)
	if apps, ok := m.apps[bucketName]; ok {
		for name, app := range apps {
			result[name] = &manifest.AppRef{
				Name:        app.Script.Name,
				Description: app.Script.Description,
			}
		}
	}
	return result, nil
}

func (m *MockBucketManager) Search(ctx context.Context, query string, opts bucket.SearchOptions) ([]bucket.SearchResult, error) {
	return nil, nil
}
