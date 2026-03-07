package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chopsticks/core/store"
)

func TestNeedsMigration(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	sqlitePath := filepath.Join(tempDir, "test.db")
	fsDataDir := filepath.Join(tempDir, "data")

	// 测试 1: SQLite 不存在，不需要迁移
	needs, err := store.NeedsMigration(sqlitePath, fsDataDir)
	if err != nil {
		t.Fatalf("检查迁移失败：%v", err)
	}
	if needs {
		t.Error("SQLite 不存在时不应该需要迁移")
	}

	// 测试 2: SQLite 存在，文件系统数据为空，需要迁移
	if err := os.WriteFile(sqlitePath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("创建 SQLite 文件失败：%v", err)
	}

	needs, err = store.NeedsMigration(sqlitePath, fsDataDir)
	if err != nil {
		t.Fatalf("检查迁移失败：%v", err)
	}
	if !needs {
		t.Error("SQLite 存在且文件系统数据为空时应该需要迁移")
	}

	// 测试 3: 文件系统数据已存在，不需要迁移
	if err := os.MkdirAll(filepath.Join(fsDataDir, "apps"), 0755); err != nil {
		t.Fatalf("创建 apps 目录失败：%v", err)
	}
	// 创建一个测试应用
	testAppDir := filepath.Join(fsDataDir, "apps", "test-app")
	if err := os.MkdirAll(testAppDir, 0755); err != nil {
		t.Fatalf("创建测试应用目录失败：%v", err)
	}
	if err := os.WriteFile(filepath.Join(testAppDir, "manifest.json"), []byte(`{"name":"test-app"}`), 0644); err != nil {
		t.Fatalf("创建 manifest.json 失败：%v", err)
	}

	needs, err = store.NeedsMigration(sqlitePath, fsDataDir)
	if err != nil {
		t.Fatalf("检查迁移失败：%v", err)
	}
	if needs {
		t.Error("文件系统数据已存在时不应该需要迁移")
	}
}

func TestHasExistingFSData(t *testing.T) {
	tempDir := t.TempDir()
	fsDataDir := filepath.Join(tempDir, "data")

	// 测试 1: 空目录
	if store.HasExistingFSData(fsDataDir) {
		t.Error("空目录不应该被认为有数据")
	}

	// 测试 2: 有 apps 目录（带应用）
	appsDir := filepath.Join(fsDataDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("创建 apps 目录失败：%v", err)
	}
	// 创建一个测试应用
	testAppDir := filepath.Join(appsDir, "test-app")
	if err := os.MkdirAll(testAppDir, 0755); err != nil {
		t.Fatalf("创建测试应用目录失败：%v", err)
	}
	if err := os.WriteFile(filepath.Join(testAppDir, "manifest.json"), []byte(`{"name":"test-app"}`), 0644); err != nil {
		t.Fatalf("创建 manifest.json 失败：%v", err)
	}

	if !store.HasExistingFSData(fsDataDir) {
		t.Error("有 apps 目录应该被认为有数据")
	}

	// 测试 3: 只有 bucket-index.json
	fsDataDir2 := filepath.Join(tempDir, "data2")
	bucketIndex := filepath.Join(fsDataDir2, "bucket-index.json")
	if err := os.MkdirAll(filepath.Dir(bucketIndex), 0755); err != nil {
		t.Fatalf("创建目录失败：%v", err)
	}
	if err := os.WriteFile(bucketIndex, []byte("{}"), 0644); err != nil {
		t.Fatalf("创建 bucket-index.json 失败：%v", err)
	}

	if !store.HasExistingFSData(fsDataDir2) {
		t.Error("有 bucket-index.json 应该被认为有数据")
	}
}

func TestMigrator_BackupAndRestore(t *testing.T) {
	tempDir := t.TempDir()
	sqlitePath := filepath.Join(tempDir, "test.db")
	targetDir := filepath.Join(tempDir, "data")

	// 创建模拟 SQLite 文件（实际上是空文件，用于测试备份逻辑）
	if err := os.WriteFile(sqlitePath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("创建 SQLite 文件失败：%v", err)
	}

	// 创建一些现有的文件系统数据
	appsDir := filepath.Join(targetDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("创建 apps 目录失败：%v", err)
	}

	testAppDir := filepath.Join(appsDir, "test-app")
	if err := os.MkdirAll(testAppDir, 0755); err != nil {
		t.Fatalf("创建测试应用目录失败：%v", err)
	}

	testManifest := filepath.Join(testAppDir, "manifest.json")
	if err := os.WriteFile(testManifest, []byte(`{"name":"test-app"}`), 0644); err != nil {
		t.Fatalf("创建 manifest.json 失败：%v", err)
	}

	// 创建迁移器
	migrator, err := NewSQLiteToFSMigrator(sqlitePath, targetDir)
	if err != nil {
		t.Fatalf("创建迁移器失败：%v", err)
	}

	// 测试备份
	ctx := context.Background()
	if err := migrator.backupExistingData(ctx); err != nil {
		t.Fatalf("备份现有数据失败：%v", err)
	}

	// 检查备份是否创建
	backupDir := migrator.GetBackupDir()
	if backupDir == "" {
		t.Fatal("备份目录为空")
	}

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Fatalf("备份目录不存在：%s", backupDir)
	}

	// 检查备份文件是否存在
	backedUpManifest := filepath.Join(backupDir, "apps", "test-app", "manifest.json")
	if _, err := os.Stat(backedUpManifest); os.IsNotExist(err) {
		t.Error("备份文件中缺少 manifest.json")
	}

	t.Logf("备份目录：%s", backupDir)
}

func TestRollbacker(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backup")

	// 创建备份数据
	backupAppsDir := filepath.Join(backupDir, "apps")
	if err := os.MkdirAll(backupAppsDir, 0755); err != nil {
		t.Fatalf("创建备份 apps 目录失败：%v", err)
	}

	testAppDir := filepath.Join(backupAppsDir, "test-app")
	if err := os.MkdirAll(testAppDir, 0755); err != nil {
		t.Fatalf("创建测试应用目录失败：%v", err)
	}

	testManifest := filepath.Join(testAppDir, "manifest.json")
	if err := os.WriteFile(testManifest, []byte(`{"name":"test-app","version":"1.0.0"}`), 0644); err != nil {
		t.Fatalf("创建 manifest.json 失败：%v", err)
	}

	// 创建备份索引文件
	bucketIndex := filepath.Join(backupDir, "bucket-index.json")
	if err := os.WriteFile(bucketIndex, []byte(`{"generated_at":"2026-01-01T00:00:00Z","buckets":{}}`), 0644); err != nil {
		t.Fatalf("创建 bucket-index.json 失败：%v", err)
	}

	// 创建回滚器
	rollbacker, err := NewFSRollbacker(sourceDir, backupDir)
	if err != nil {
		t.Fatalf("创建回滚器失败：%v", err)
	}

	// 执行回滚
	ctx := context.Background()
	stats, err := rollbacker.Rollback(ctx)
	if err != nil {
		t.Fatalf("回滚失败：%v", err)
	}

	// 验证回滚结果
	if stats.AppCount != 1 {
		t.Errorf("期望恢复 1 个应用，实际恢复 %d 个", stats.AppCount)
	}

	// 检查恢复的文件是否存在
	restoredManifest := filepath.Join(sourceDir, "apps", "test-app", "manifest.json")
	if _, err := os.Stat(restoredManifest); os.IsNotExist(err) {
		t.Error("恢复的 manifest.json 不存在")
	}

	restoredBucketIndex := filepath.Join(sourceDir, "bucket-index.json")
	if _, err := os.Stat(restoredBucketIndex); os.IsNotExist(err) {
		t.Error("恢复的 bucket-index.json 不存在")
	}
}

func TestMigrationStats(t *testing.T) {
	stats := &MigrationStats{
		AppCount:       5,
		OperationCount: 10,
		BucketCount:    2,
		RuntimeCount:   3,
		DepsCount:      4,
		ErrorCount:     0,
		Duration:       2 * time.Second,
	}

	if stats.AppCount != 5 {
		t.Errorf("期望 5 个应用，实际 %d", stats.AppCount)
	}

	if stats.OperationCount != 10 {
		t.Errorf("期望 10 条操作记录，实际 %d", stats.OperationCount)
	}

	if stats.BucketCount != 2 {
		t.Errorf("期望 2 个软件源，实际 %d", stats.BucketCount)
	}
}

func TestRollbackStats(t *testing.T) {
	stats := &RollbackStats{
		AppCount:       3,
		OperationCount: 5,
		BucketCount:    1,
		Duration:       1 * time.Second,
	}

	if stats.AppCount != 3 {
		t.Errorf("期望恢复 3 个应用，实际 %d", stats.AppCount)
	}

	if stats.Duration != 1*time.Second {
		t.Errorf("期望耗时 1 秒，实际 %v", stats.Duration)
	}
}
