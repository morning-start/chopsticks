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
	bucketMgr := bucket.NewManager(storage, nil, bucketsDir, mockGit)

	// Setup installer with mocks
	installer := NewMockInstaller(storage)

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

	// Create bucket.json
	bucketConfig := `{
  "id": "test-bucket",
  "name": "test-bucket",
  "description": "Test bucket"
}
`
	if err := os.WriteFile(
		filepath.Join(path, "bucket.json"),
		[]byte(bucketConfig),
		0644,
	); err != nil {
		t.Fatalf("创建 bucket.json 失败: %v", err)
	}
}

// CreateTestApp 创建测试应用。
func CreateTestApp(t *testing.T, bucketPath, name string, deps []string) {
	t.Helper()

	// 构建依赖字段
	depsStr := ""
	if len(deps) > 0 {
		depItems := make([]string, len(deps))
		for i, dep := range deps {
			depItems[i] = fmt.Sprintf(`"%s"`, dep)
		}
		depsStr = fmt.Sprintf(`,
  depends: [%s]`, joinStrings(depItems, ", "))
	}

	// Create app script file directly in apps directory
	scriptContent := fmt.Sprintf(`/**
 * @description Test app %s
 * @version 1.0.0
 */

const app = {
  name: "%s",
  version: "1.0.0",%s
  architecture: {
    "64bit": {
      url: "https://example.com/%s-1.0.0.zip",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["%s.exe"]
};

module.exports = app;
`, name, name, depsStr, name, name)

	if err := os.WriteFile(
		filepath.Join(bucketPath, "apps", name+".js"),
		[]byte(scriptContent),
		0644,
	); err != nil {
		t.Fatalf("创建应用脚本失败: %v", err)
	}
}

// joinStrings 将字符串切片用分隔符连接
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// CreateTestAppWithDeps 创建带依赖的测试应用。
func CreateTestAppWithDeps(t *testing.T, components *TestComponents, name string, deps []string) {
	t.Helper()

	ctx := context.Background()
	bucketName := "main"

	// 获取第一个 bucket 路径
	bucketsDir := filepath.Join(components.TmpDir, "buckets")
	entries, err := os.ReadDir(bucketsDir)
	if err != nil || len(entries) == 0 {
		// 创建默认 bucket
		bucketPath := filepath.Join(bucketsDir, bucketName)
		CreateTestBucket(t, bucketPath)
		CreateTestApp(t, bucketPath, name, deps)

		// 保存 bucket 配置到数据库
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
		return
	}

	bucketPath := filepath.Join(bucketsDir, entries[0].Name())
	CreateTestApp(t, bucketPath, name, deps)
}

// AddTestBucketWithApps 添加包含多个应用的测试 bucket。
// 当传入 ["app-a", "app-b", "app-c"] 时，会创建依赖链: app-a -> app-b -> app-c
func AddTestBucketWithApps(t *testing.T, components *TestComponents, bucketName string, appNames []string) {
	t.Helper()

	ctx := context.Background()

	// 创建 bucket 目录
	bucketPath := filepath.Join(components.TmpDir, "buckets", bucketName)
	CreateTestBucket(t, bucketPath)

	// 检查是否是特定的依赖测试应用组合
	// 如果是 ["app-a", "app-b", "app-c"] 或类似组合，创建依赖链
	hasAppB := contains(appNames, "app-b")
	hasAppC := contains(appNames, "app-c")

	// 创建应用
	for _, appName := range appNames {
		var deps []string

		// 根据 fixtures 中的定义设置依赖
		switch appName {
		case "app-a":
			if hasAppB {
				deps = []string{"app-b"}
			}
		case "app-b":
			if hasAppC {
				deps = []string{"app-c"}
			}
		case "app-c":
			// app-c 没有依赖
			deps = []string{}
		}

		CreateTestApp(t, bucketPath, appName, deps)
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

// contains 检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
