// Package store 提供文件系统存储测试。
package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chopsticks/core/manifest"
)

// setupTestFSStorage 创建测试用的文件系统存储。
func setupTestFSStorage(t *testing.T) (*fsStorage, func()) {
	t.Helper()

	// 创建临时目录
	tmpDir := t.TempDir()

	storage, err := NewFSStorage(tmpDir)
	if err != nil {
		t.Fatalf("创建存储失败：%v", err)
	}

	cleanup := func() {
		storage.Close()
	}

	return storage.(*fsStorage), cleanup
}

// ============================================================================
// AppStorage 测试
// ============================================================================

func TestFSStorage_SaveAndGetApp(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试应用
	app := &AppManifest{
		Name:              "git",
		Bucket:            "main",
		CurrentVersion:    "2.43.0",
		InstalledVersions: []string{"2.43.0"},
		Dependencies: manifest.Dependencies{
			Runtime: []manifest.Dependency{
				{Name: "vcredist140", Version: ">=14.0"},
			},
			Tools: []manifest.Dependency{
				{Name: "7zip", Version: ">=19.0"},
			},
		},
		InstalledAt:        time.Now().UTC(),
		InstalledOnRequest: true,
		Isolated:           false,
	}

	// 保存应用
	err := storage.SaveApp(ctx, app)
	if err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 获取应用
	got, err := storage.GetApp(ctx, "git")
	if err != nil {
		t.Fatalf("GetApp() error = %v", err)
	}

	// 验证
	if got.Name != app.Name {
		t.Errorf("Name = %s, want %s", got.Name, app.Name)
	}
	if got.Bucket != app.Bucket {
		t.Errorf("Bucket = %s, want %s", got.Bucket, app.Bucket)
	}
	if got.CurrentVersion != app.CurrentVersion {
		t.Errorf("CurrentVersion = %s, want %s", got.CurrentVersion, app.CurrentVersion)
	}
}

func TestFSStorage_DeleteApp(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试应用
	app := &AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now().UTC(),
	}

	// 保存应用
	if err := storage.SaveApp(ctx, app); err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 删除应用
	err := storage.DeleteApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("DeleteApp() error = %v", err)
	}

	// 验证删除
	_, err = storage.GetApp(ctx, "test-app")
	if err == nil {
		t.Errorf("GetApp() should return error after delete")
	}
}

func TestFSStorage_ListApps(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建多个测试应用
	apps := []*AppManifest{
		{Name: "git", Bucket: "main", CurrentVersion: "2.43.0", InstalledVersions: []string{"2.43.0"}, InstalledAt: time.Now().UTC()},
		{Name: "nodejs", Bucket: "main", CurrentVersion: "20.0.0", InstalledVersions: []string{"20.0.0"}, InstalledAt: time.Now().UTC()},
		{Name: "python", Bucket: "main", CurrentVersion: "3.12.0", InstalledVersions: []string{"3.12.0"}, InstalledAt: time.Now().UTC()},
	}

	for _, app := range apps {
		if err := storage.SaveApp(ctx, app); err != nil {
			t.Fatalf("SaveApp() error = %v", err)
		}
	}

	// 列出所有应用
	got, err := storage.ListApps(ctx)
	if err != nil {
		t.Fatalf("ListApps() error = %v", err)
	}

	// 验证数量
	if len(got) != len(apps) {
		t.Errorf("len(ListApps()) = %d, want %d", len(got), len(apps))
	}
}

func TestFSStorage_IsInstalled(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试应用
	app := &AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now().UTC(),
	}

	// 保存应用
	if err := storage.SaveApp(ctx, app); err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 检查是否已安装
	installed, err := storage.IsInstalled(ctx, "test-app")
	if err != nil {
		t.Fatalf("IsInstalled() error = %v", err)
	}
	if !installed {
		t.Errorf("IsInstalled() = false, want true")
	}

	// 检查未安装的应用
	installed, err = storage.IsInstalled(ctx, "non-existent")
	if err != nil {
		t.Fatalf("IsInstalled() error = %v", err)
	}
	if installed {
		t.Errorf("IsInstalled() = true, want false")
	}
}

// ============================================================================
// BucketStorage 测试
// ============================================================================

func TestFSStorage_SaveAndGetBucket(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试软件源
	bucket := &BucketConfig{
		ID:          "main",
		Name:        "Main Bucket",
		Author:      "Chopsticks Team",
		Description: "Main bucket for chopsticks",
		Homepage:    "https://github.com/chopsticks-sh/main",
		License:     "MIT",
		Repository: manifest.RepositoryInfo{
			Type:   "git",
			URL:    "https://github.com/chopsticks-sh/main",
			Branch: "main",
		},
		AddedAt:   time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// 保存软件源
	err := storage.SaveBucket(ctx, bucket)
	if err != nil {
		t.Fatalf("SaveBucket() error = %v", err)
	}

	// 获取软件源
	got, err := storage.GetBucket(ctx, "main")
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	// 验证
	if got.ID != bucket.ID {
		t.Errorf("ID = %s, want %s", got.ID, bucket.ID)
	}
	if got.Name != bucket.Name {
		t.Errorf("Name = %s, want %s", got.Name, bucket.Name)
	}
	if got.Repository.URL != bucket.Repository.URL {
		t.Errorf("Repository.URL = %s, want %s", got.Repository.URL, bucket.Repository.URL)
	}
}

func TestFSStorage_DeleteBucket(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试软件源
	bucket := &BucketConfig{
		ID:          "test-bucket",
		Name:        "Test Bucket",
		Description: "Test bucket",
		AddedAt:     time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// 保存软件源
	if err := storage.SaveBucket(ctx, bucket); err != nil {
		t.Fatalf("SaveBucket() error = %v", err)
	}

	// 删除软件源
	err := storage.DeleteBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("DeleteBucket() error = %v", err)
	}

	// 验证删除
	_, err = storage.GetBucket(ctx, "test-bucket")
	if err == nil {
		t.Errorf("GetBucket() should return error after delete")
	}
}

func TestFSStorage_ListBuckets(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建多个测试软件源
	buckets := []*BucketConfig{
		{ID: "main", Name: "Main", Description: "Main bucket", AddedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: "extras", Name: "Extras", Description: "Extras bucket", AddedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		{ID: "versions", Name: "Versions", Description: "Versions bucket", AddedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}

	for _, bucket := range buckets {
		if err := storage.SaveBucket(ctx, bucket); err != nil {
			t.Fatalf("SaveBucket() error = %v", err)
		}
	}

	// 列出所有软件源
	got, err := storage.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}

	// 验证数量
	if len(got) != len(buckets) {
		t.Errorf("len(ListBuckets()) = %d, want %d", len(got), len(buckets))
	}
}

// ============================================================================
// OperationStorage 测试
// ============================================================================

func TestFSStorage_SaveAndGetOperation(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 先创建应用
	app := &AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now().UTC(),
	}
	if err := storage.SaveApp(ctx, app); err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 创建测试操作
	ops := []*Operation{
		{
			Type:      "path",
			Path:      "bin",
			CreatedAt: time.Now().UTC(),
		},
		{
			Type:      "env",
			Key:       "TEST_APP_HOME",
			Value:     "C:\\Users\\test\\.chopsticks\\apps\\test-app\\1.0.0",
			CreatedAt: time.Now().UTC(),
		},
		{
			Type:      "symlink",
			Link:      "C:\\Users\\test\\.chopsticks\\shims\\test-app.exe",
			Target:    "C:\\Users\\test\\.chopsticks\\apps\\test-app\\1.0.0\\bin\\test-app.exe",
			CreatedAt: time.Now().UTC(),
		},
	}

	// 保存操作
	for _, op := range ops {
		if err := storage.SaveOperation(ctx, "test-app", op); err != nil {
			t.Fatalf("SaveOperation() error = %v", err)
		}
	}

	// 获取操作
	got, err := storage.GetOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetOperations() error = %v", err)
	}

	// 验证数量
	if len(got) != len(ops) {
		t.Errorf("len(GetOperations()) = %d, want %d", len(got), len(ops))
	}

	// 验证操作类型
	for i, op := range ops {
		if got[i].Type != op.Type {
			t.Errorf("Operation[%d].Type = %s, want %s", i, got[i].Type, op.Type)
		}
	}
}

func TestFSStorage_DeleteOperations(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 先创建应用
	app := &AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now().UTC(),
	}
	if err := storage.SaveApp(ctx, app); err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 创建测试操作
	op := &Operation{
		Type:      "path",
		Path:      "bin",
		CreatedAt: time.Now().UTC(),
	}
	if err := storage.SaveOperation(ctx, "test-app", op); err != nil {
		t.Fatalf("SaveOperation() error = %v", err)
	}

	// 删除操作
	err := storage.DeleteOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("DeleteOperations() error = %v", err)
	}

	// 验证删除
	got, err := storage.GetOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetOperations() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(GetOperations()) = %d, want 0", len(got))
	}
}

// ============================================================================
// DependencyStorage 测试
// ============================================================================

func TestFSStorage_SaveAndGetRuntimeIndex(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试运行时索引
	index := RuntimeIndex{
		"vcredist140": &manifest.RuntimeInfo{
			Version:     "14.38.33135",
			InstalledAt: time.Now().UTC(),
			RequiredBy:  []string{"git", "nodejs"},
			RefCount:    2,
			Size:        23500000,
		},
		"dotnet6": &manifest.RuntimeInfo{
			Version:     "6.0.25",
			InstalledAt: time.Now().UTC(),
			RequiredBy:  []string{"powershell"},
			RefCount:    1,
		},
	}

	// 保存索引
	err := storage.SaveRuntimeIndex(ctx, index)
	if err != nil {
		t.Fatalf("SaveRuntimeIndex() error = %v", err)
	}

	// 获取索引
	got, err := storage.GetRuntimeIndex(ctx)
	if err != nil {
		t.Fatalf("GetRuntimeIndex() error = %v", err)
	}

	// 验证
	if len(got) != len(index) {
		t.Errorf("len(GetRuntimeIndex()) = %d, want %d", len(got), len(index))
	}

	if got["vcredist140"] == nil {
		t.Errorf("vcredist140 not found in runtime index")
	} else if got["vcredist140"].RefCount != 2 {
		t.Errorf("vcredist140.RefCount = %d, want 2", got["vcredist140"].RefCount)
	}
}

func TestFSStorage_SaveAndGetDepsIndex(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试依赖索引
	index := &DepsIndex{
		GeneratedAt: time.Now().UTC(),
		Apps: map[string]*AppDeps{
			"git": {
				Dependencies: []string{"vcredist140", "7zip"},
				Dependents:   []string{"git-lfs", "hub"},
			},
			"7zip": {
				Dependencies: []string{},
				Dependents:   []string{"git", "nodejs"},
			},
		},
	}

	// 保存索引
	err := storage.SaveDepsIndex(ctx, index)
	if err != nil {
		t.Fatalf("SaveDepsIndex() error = %v", err)
	}

	// 获取索引
	got, err := storage.GetDepsIndex(ctx)
	if err != nil {
		t.Fatalf("GetDepsIndex() error = %v", err)
	}

	// 验证
	if len(got.Apps) != len(index.Apps) {
		t.Errorf("len(GetDepsIndex().Apps) = %d, want %d", len(got.Apps), len(index.Apps))
	}

	if got.Apps["git"] == nil {
		t.Errorf("git not found in deps index")
	} else if len(got.Apps["git"].Dependencies) != 2 {
		t.Errorf("git.Dependencies length = %d, want 2", len(got.Apps["git"].Dependencies))
	}
}

func TestFSStorage_RebuildDepsIndex(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试应用
	apps := []*AppManifest{
		{
			Name:              "git",
			Bucket:            "main",
			CurrentVersion:    "2.43.0",
			InstalledVersions: []string{"2.43.0"},
			Dependencies: manifest.Dependencies{
				Runtime: []manifest.Dependency{{Name: "vcredist140", Version: ">=14.0"}},
				Tools:   []manifest.Dependency{{Name: "7zip", Version: ">=19.0"}},
			},
			InstalledAt: time.Now().UTC(),
		},
		{
			Name:              "nodejs",
			Bucket:            "main",
			CurrentVersion:    "20.0.0",
			InstalledVersions: []string{"20.0.0"},
			Dependencies: manifest.Dependencies{
				Tools: []manifest.Dependency{{Name: "7zip", Version: ">=19.0"}},
			},
			InstalledAt: time.Now().UTC(),
		},
		{
			Name:              "7zip",
			Bucket:            "main",
			CurrentVersion:    "19.0.0",
			InstalledVersions: []string{"19.0.0"},
			Dependencies:      manifest.Dependencies{},
			InstalledAt:       time.Now().UTC(),
		},
	}

	for _, app := range apps {
		if err := storage.SaveApp(ctx, app); err != nil {
			t.Fatalf("SaveApp() error = %v", err)
		}
	}

	// 重建依赖索引
	err := storage.RebuildDepsIndex(ctx)
	if err != nil {
		t.Fatalf("RebuildDepsIndex() error = %v", err)
	}

	// 获取重建的索引
	got, err := storage.GetDepsIndex(ctx)
	if err != nil {
		t.Fatalf("GetDepsIndex() error = %v", err)
	}

	// 验证依赖关系
	if got.Apps["git"] == nil {
		t.Errorf("git not found in deps index")
	} else {
		// 检查 git 的依赖
		if len(got.Apps["git"].Dependencies) != 2 {
			t.Errorf("git.Dependencies length = %d, want 2", len(got.Apps["git"].Dependencies))
		}

		// 检查 7zip 的反向依赖
		if got.Apps["7zip"] == nil {
			t.Errorf("7zip not found in deps index")
		} else if len(got.Apps["7zip"].Dependents) != 2 {
			t.Errorf("7zip.Dependents length = %d, want 2", len(got.Apps["7zip"].Dependents))
		}
	}
}

// ============================================================================
// JSON 格式测试
// ============================================================================

func TestFSStorage_JSONFormat(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试应用
	app := &AppManifest{
		Name:              "test-app",
		Bucket:            "main",
		CurrentVersion:    "1.0.0",
		InstalledVersions: []string{"1.0.0"},
		InstalledAt:       time.Now().UTC(),
	}

	// 保存应用
	if err := storage.SaveApp(ctx, app); err != nil {
		t.Fatalf("SaveApp() error = %v", err)
	}

	// 读取 manifest.json 文件
	manifestPath := filepath.Join(storage.GetAppsDir(), "test-app", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("读取 manifest.json 失败：%v", err)
	}

	// 验证 JSON 格式（应该包含缩进）
	content := string(data)
	if len(content) == 0 {
		t.Errorf("manifest.json 为空")
	}

	// 检查是否包含缩进（人类可读）
	// 使用 json.MarshalIndent 应该会包含 "  " 缩进
	if len(content) > 0 {
		// 简单检查：至少应该有一些换行和缩进
		hasNewline := false
		hasIndent := false
		for i := 0; i < len(content); i++ {
			if content[i] == '\n' {
				hasNewline = true
			}
			if i < len(content)-1 && content[i] == ' ' && content[i+1] == ' ' {
				hasIndent = true
			}
		}
		if !hasNewline {
			t.Errorf("manifest.json 应该包含换行符以提高可读性")
		}
		if !hasIndent {
			t.Errorf("manifest.json 应该包含缩进以提高可读性")
		}
	}
}

// ============================================================================
// 并发安全测试
// ============================================================================

func TestFSStorage_ConcurrentAccess(t *testing.T) {
	storage, cleanup := setupTestFSStorage(t)
	defer cleanup()

	ctx := context.Background()

	// 并发读写测试
	done := make(chan bool, 10)

	// 启动 10 个 goroutine 并发访问
	for i := 0; i < 10; i++ {
		go func(id int) {
			appName := "test-app-" + string(rune('A'+id))
			app := &AppManifest{
				Name:              appName,
				Bucket:            "main",
				CurrentVersion:    "1.0.0",
				InstalledVersions: []string{"1.0.0"},
				InstalledAt:       time.Now().UTC(),
			}

			// 保存应用
			if err := storage.SaveApp(ctx, app); err != nil {
				t.Errorf("SaveApp() error = %v", err)
				done <- false
				return
			}

			// 读取应用
			_, err := storage.GetApp(ctx, appName)
			if err != nil {
				t.Errorf("GetApp() error = %v", err)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	success := true
	for i := 0; i < 10; i++ {
		if !<-done {
			success = false
		}
	}

	if !success {
		t.Errorf("并发访问测试失败")
	}
}
