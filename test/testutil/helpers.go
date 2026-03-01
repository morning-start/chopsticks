// Package testutil 提供集成测试的辅助工具。
package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/app"
	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

// TestComponents 集成测试组件集合。
type TestComponents struct {
	TmpDir    string
	Storage   store.Storage
	BucketMgr bucket.BucketManager
	AppMgr    app.AppManager
	Installer app.Installer
	MockGit   *MockGit
}

// SetupTestEnvironment 设置完整测试环境。
func SetupTestEnvironment(t *testing.T) *TestComponents {
	t.Helper()

	tmpDir := t.TempDir()

	// Setup storage
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// Setup bucket manager
	bucketsDir := filepath.Join(tmpDir, "buckets")
	mockGit := NewMockGit()
	bucketMgr := bucket.NewManager(storage, nil, bucketsDir)

	// Setup installer with mocks
	installer := NewMockInstaller()

	// Setup app manager
	appsDir := filepath.Join(tmpDir, "apps")
	appMgr := app.NewManager(bucketMgr, storage, installer, nil, appsDir)

	t.Cleanup(func() {
		storage.Close()
	})

	return &TestComponents{
		TmpDir:    tmpDir,
		Storage:   storage,
		BucketMgr: bucketMgr,
		AppMgr:    appMgr,
		Installer: installer,
		MockGit:   mockGit,
	}
}

// SetupTestStorage 创建测试存储。
func SetupTestStorage(t *testing.T, tmpDir string) store.Storage {
	t.Helper()

	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	return storage
}

// CreateTestBucket 创建测试 bucket 目录结构。
func CreateTestBucket(t *testing.T, path string) {
	t.Helper()

	// Create bucket structure
	appsDir := filepath.Join(path, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("创建 apps 目录失败: %v", err)
	}

	// Create bucket.yaml
	bucketConfig := `name: test-bucket
description: Test bucket
`
	if err := os.WriteFile(
		filepath.Join(path, "bucket.yaml"),
		[]byte(bucketConfig),
		0644,
	); err != nil {
		t.Fatalf("创建 bucket.yaml 失败: %v", err)
	}
}

// CreateTestApp 创建测试应用。
func CreateTestApp(t *testing.T, bucketPath, name string, deps []string) {
	t.Helper()

	appDir := filepath.Join(bucketPath, "apps", name)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("创建应用目录失败: %v", err)
	}

	// Create manifest
	manifestContent := fmt.Sprintf(`
name: %s
version: 1.0.0
description: Test app %s
`, name, name)

	if len(deps) > 0 {
		manifestContent += "dependencies:\n"
		for _, dep := range deps {
			manifestContent += fmt.Sprintf("  - %s\n", dep)
		}
	}

	if err := os.WriteFile(
		filepath.Join(appDir, "manifest.yaml"),
		[]byte(manifestContent),
		0644,
	); err != nil {
		t.Fatalf("创建 manifest 失败: %v", err)
	}
}

// CreateTestAppWithDeps 创建带依赖的测试应用。
func CreateTestAppWithDeps(t *testing.T, components *TestComponents, name string, deps []string) {
	t.Helper()

	// 获取第一个 bucket 路径
	bucketsDir := filepath.Join(components.TmpDir, "buckets")
	entries, err := os.ReadDir(bucketsDir)
	if err != nil || len(entries) == 0 {
		// 创建默认 bucket
		bucketPath := filepath.Join(bucketsDir, "main")
		CreateTestBucket(t, bucketPath)
		CreateTestApp(t, bucketPath, name, deps)
		return
	}

	bucketPath := filepath.Join(bucketsDir, entries[0].Name())
	CreateTestApp(t, bucketPath, name, deps)
}

// AddTestBucketWithApps 添加包含多个应用的测试 bucket。
func AddTestBucketWithApps(t *testing.T, components *TestComponents, bucketName string, appNames []string) {
	t.Helper()

	ctx := context.Background()

	// 创建 bucket 目录
	bucketPath := filepath.Join(components.TmpDir, "buckets", bucketName)
	CreateTestBucket(t, bucketPath)

	// 创建应用
	for _, appName := range appNames {
		CreateTestApp(t, bucketPath, appName, nil)
	}

	// 保存到数据库
	bucketConfig := &manifest.BucketConfig{
		ID:   bucketName,
		Name: bucketName,
		Repository: manifest.RepositoryInfo{
			URL:    "https://github.com/test/" + bucketName,
			Branch: "main",
		},
	}
	if err := components.Storage.SaveBucket(ctx, bucketConfig); err != nil {
		t.Fatalf("保存 bucket 配置失败: %v", err)
	}
}

// AssertAppInstalled 断言应用已安装。
func AssertAppInstalled(t *testing.T, storage store.Storage, name string) {
	t.Helper()

	ctx := context.Background()
	installed, err := storage.GetInstalledApp(ctx, name)
	if err != nil {
		t.Errorf("应用 %s 未安装: %v", name, err)
		return
	}
	if installed.Name != name {
		t.Errorf("应用名称不匹配: 期望 %s, 实际 %s", name, installed.Name)
	}
}

// AssertAppNotInstalled 断言应用未安装。
func AssertAppNotInstalled(t *testing.T, storage store.Storage, name string) {
	t.Helper()

	ctx := context.Background()
	_, err := storage.GetInstalledApp(ctx, name)
	if err == nil {
		t.Errorf("应用 %s 应该未安装", name)
	}
}
