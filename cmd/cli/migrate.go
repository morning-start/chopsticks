// Package cli 提供数据迁移命令实现。
package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	migrateFromSQLite string
	migrateRollback   bool
	migrateDryRun     bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [flags]",
	Short: "从 SQLite 迁移到文件系统存储",
	Long: `将 Chopsticks 的数据从 SQLite 数据库迁移到文件系统存储。

迁移内容包括：
  - 已安装应用信息 (manifest.json)
  - 应用操作记录 (operations.json)
  - 软件源配置 (bucket-index.json)
  - 运行时索引 (runtime-index.json)
  - 依赖索引 (deps-index.json)

示例:
  # 从 SQLite 迁移到文件系统
  chopsticks migrate --from-sqlite ~/.chopsticks/chopsticks.db

  # 预览迁移（不实际执行）
  chopsticks migrate --from-sqlite ~/.chopsticks/chopsticks.db --dry-run

  # 回滚迁移（从文件系统恢复到 SQLite）
  chopsticks migrate --rollback

注意：迁移前请备份重要数据！`,
	RunE: runMigrate,
}

func init() {
	migrateCmd.Flags().StringVar(&migrateFromSQLite, "from-sqlite", "", "SQLite 数据库文件路径")
	migrateCmd.Flags().BoolVar(&migrateRollback, "rollback", false, "回滚迁移（从文件系统恢复到 SQLite）")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "预览迁移（不实际执行）")
}

// runMigrate 执行迁移命令
func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if migrateRollback {
		return runRollback(ctx)
	}

	if migrateFromSQLite == "" {
		return fmt.Errorf("请指定 SQLite 数据库文件路径 (--from-sqlite)")
	}

	// 检查 SQLite 文件是否存在
	if _, err := os.Stat(migrateFromSQLite); os.IsNotExist(err) {
		return fmt.Errorf("SQLite 数据库文件不存在：%s", migrateFromSQLite)
	}

	// 获取目标存储目录
	app := getApp()
	if app == nil {
		return fmt.Errorf("应用未初始化")
	}

	// 从配置中获取目标存储目录
	cfg := app.Config()
	targetDir := cfg.StorageDir

	output.Info("开始从 SQLite 迁移到文件系统存储...\n")
	output.Info("源数据库：%s", migrateFromSQLite)
	output.Info("目标目录：%s\n", targetDir)

	if migrateDryRun {
		output.Warn("=== 预览模式：不会执行实际迁移 ===\n")
	}

	// 创建迁移器
	migrator, err := NewSQLiteToFSMigrator(migrateFromSQLite, targetDir)
	if err != nil {
		return fmt.Errorf("创建迁移器失败：%w", err)
	}

	// 执行迁移
	stats, err := migrator.Migrate(ctx, migrateDryRun)
	if err != nil {
		return fmt.Errorf("迁移失败：%w", err)
	}

	// 显示迁移统计
	printMigrationStats(stats)

	if !migrateDryRun {
		output.Success("\n迁移完成！")
		output.Info("备份位置：%s", migrator.GetBackupDir())
		output.Info("如需回滚，请运行：chopsticks migrate --rollback\n")
	}

	return nil
}

// runRollback 执行回滚
func runRollback(ctx context.Context) error {
	output.Warn("=== 执行回滚 ===\n")

	app := getApp()
	if app == nil {
		return fmt.Errorf("应用未初始化")
	}

	cfg := app.Config()
	sourceDir := cfg.StorageDir

	// 查找备份目录
	backupDir := findLatestBackup(sourceDir)
	if backupDir == "" {
		return fmt.Errorf("未找到备份数据，无法回滚")
	}

	output.Info("从备份恢复：%s\n", backupDir)

	// 创建回滚器
	rollbacker, err := NewFSRollbacker(sourceDir, backupDir)
	if err != nil {
		return fmt.Errorf("创建回滚器失败：%w", err)
	}

	// 执行回滚
	stats, err := rollbacker.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("回滚失败：%w", err)
	}

	printRollbackStats(stats)
	output.Success("\n回滚完成！\n")

	return nil
}

// printMigrationStats 显示迁移统计
func printMigrationStats(stats *MigrationStats) {
	output.Info("\n=== 迁移统计 ===\n")
	output.Info("已安装应用：%d", stats.AppCount)
	output.Info("操作记录：%d 条", stats.OperationCount)
	output.Info("软件源：%d 个", stats.BucketCount)
	output.Info("运行时索引：%d 项", stats.RuntimeCount)
	output.Info("依赖索引：%d 项", stats.DepsCount)

	if stats.ErrorCount > 0 {
		output.Warn("错误数：%d", stats.ErrorCount)
		if len(stats.Errors) > 0 {
			output.Info("\n错误详情:")
			for _, err := range stats.Errors {
				output.Warn("  - %s", err)
			}
		}
	}

	output.Info("迁移耗时：%v", stats.Duration)
}

// printRollbackStats 显示回滚统计
func printRollbackStats(stats *RollbackStats) {
	output.Info("\n=== 回滚统计 ===\n")
	output.Info("恢复应用：%d", stats.AppCount)
	output.Info("恢复操作记录：%d 条", stats.OperationCount)
	output.Info("恢复软件源：%d 个", stats.BucketCount)
	output.Info("回滚耗时：%v", stats.Duration)
}

// findLatestBackup 查找最新的备份目录
func findLatestBackup(baseDir string) string {
	backupBase := filepath.Join(baseDir, "migrate-backups")

	entries, err := os.ReadDir(backupBase)
	if err != nil {
		return ""
	}

	var latestTime time.Time
	var latestDir string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 解析目录名中的时间戳 (格式：2006-01-02_15-04-05)
		dirName := entry.Name()
		t, err := time.ParseInLocation("2006-01-02_15-04-05", dirName, time.Local)
		if err != nil {
			continue
		}

		if t.After(latestTime) {
			latestTime = t
			latestDir = filepath.Join(backupBase, dirName)
		}
	}

	return latestDir
}

// ============================================================================
// SQLite 到文件系统迁移器
// ============================================================================

// MigrationStats 迁移统计
type MigrationStats struct {
	AppCount       int
	OperationCount int
	BucketCount    int
	RuntimeCount   int
	DepsCount      int
	ErrorCount     int
	Errors         []string
	Duration       time.Duration
}

// SQLiteToFSMigrator SQLite 到文件系统迁移器
type SQLiteToFSMigrator struct {
	sqlitePath string
	targetDir  string
	backupDir  string
	db         *sql.DB
}

// NewSQLiteToFSMigrator 创建新的迁移器
func NewSQLiteToFSMigrator(sqlitePath, targetDir string) (*SQLiteToFSMigrator, error) {
	// 创建备份目录
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDir := filepath.Join(targetDir, "migrate-backups", timestamp)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败：%w", err)
	}

	return &SQLiteToFSMigrator{
		sqlitePath: sqlitePath,
		targetDir:  targetDir,
		backupDir:  backupDir,
	}, nil
}

// GetBackupDir 返回备份目录
func (m *SQLiteToFSMigrator) GetBackupDir() string {
	return m.backupDir
}

// Migrate 执行迁移
func (m *SQLiteToFSMigrator) Migrate(ctx context.Context, dryRun bool) (*MigrationStats, error) {
	startTime := time.Now()
	stats := &MigrationStats{}

	// 打开 SQLite 数据库
	var err error
	m.db, err = sql.Open("sqlite3", m.sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 数据库失败：%w", err)
	}
	defer m.db.Close()

	// 备份现有数据
	if !dryRun {
		if err := m.backupExistingData(ctx); err != nil {
			return nil, fmt.Errorf("备份现有数据失败：%w", err)
		}
	}

	// 迁移应用
	if err := m.migrateApps(ctx, stats, dryRun); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("迁移应用失败：%v", err))
		stats.ErrorCount++
	}

	// 迁移操作记录
	if err := m.migrateOperations(ctx, stats, dryRun); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("迁移操作记录失败：%v", err))
		stats.ErrorCount++
	}

	// 迁移软件源
	if err := m.migrateBuckets(ctx, stats, dryRun); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("迁移软件源失败：%v", err))
		stats.ErrorCount++
	}

	// 迁移运行时索引
	if err := m.migrateRuntimeIndex(ctx, stats, dryRun); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("迁移运行时索引失败：%v", err))
		stats.ErrorCount++
	}

	// 迁移依赖索引
	if err := m.migrateDepsIndex(ctx, stats, dryRun); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("迁移依赖索引失败：%v", err))
		stats.ErrorCount++
	}

	stats.Duration = time.Since(startTime)

	return stats, nil
}

// backupExistingData 备份现有文件系统数据
func (m *SQLiteToFSMigrator) backupExistingData(ctx context.Context) error {
	output.Info("正在备份现有数据...")

	// 检查是否存在现有的文件系统数据
	appsDir := filepath.Join(m.targetDir, "apps")
	bucketsDir := filepath.Join(m.targetDir, "buckets")
	indexFiles := []string{
		filepath.Join(m.targetDir, "bucket-index.json"),
		filepath.Join(m.targetDir, "runtime-index.json"),
		filepath.Join(m.targetDir, "deps-index.json"),
	}

	// 备份 apps 目录
	if _, err := os.Stat(appsDir); err == nil {
		if err := m.copyDir(appsDir, filepath.Join(m.backupDir, "apps")); err != nil {
			return fmt.Errorf("备份 apps 目录失败：%w", err)
		}
	}

	// 备份 buckets 目录
	if _, err := os.Stat(bucketsDir); err == nil {
		if err := m.copyDir(bucketsDir, filepath.Join(m.backupDir, "buckets")); err != nil {
			return fmt.Errorf("备份 buckets 目录失败：%w", err)
		}
	}

	// 备份索引文件
	for _, indexFile := range indexFiles {
		if _, err := os.Stat(indexFile); err == nil {
			destFile := filepath.Join(m.backupDir, filepath.Base(indexFile))
			if err := m.copyFile(indexFile, destFile); err != nil {
				return fmt.Errorf("备份索引文件失败：%w", err)
			}
		}
	}

	output.Success(" 完成")
	return nil
}

// migrateApps 迁移已安装应用
func (m *SQLiteToFSMigrator) migrateApps(ctx context.Context, stats *MigrationStats, dryRun bool) error {
	output.Info("正在迁移已安装应用...")

	// 查询 SQLite 中的应用
	rows, err := m.db.QueryContext(ctx, `
		SELECT name, bucket, version, installed_at, install_dir,
		       current_version, installed_versions, dependencies, isolated
		FROM installed_apps
	`)
	if err != nil {
		// 表不存在，跳过
		if isTableNotExist(err, "installed_apps") {
			output.Warn(" 跳过 (表不存在)")
			return nil
		}
		return err
	}
	defer rows.Close()

	appsDir := filepath.Join(m.targetDir, "apps")
	count := 0

	for rows.Next() {
		var (
			name, bucket, version, installDir string
			installedAt                       time.Time
			currentVersion, installedVersions string
			dependencies                      string
			isolated                          bool
		)

		if err := rows.Scan(&name, &bucket, &version, &installedAt, &installDir,
			&currentVersion, &installedVersions, &dependencies, &isolated); err != nil {
			return fmt.Errorf("扫描应用记录失败：%w", err)
		}

		if dryRun {
			count++
			continue
		}

		// 创建应用目录
		appDir := filepath.Join(appsDir, name)
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return fmt.Errorf("创建应用目录失败：%w", err)
		}

		// 解析依赖
		var deps manifest.Dependencies
		if dependencies != "" {
			json.Unmarshal([]byte(dependencies), &deps)
		}

		// 解析已安装版本列表
		var versions []string
		if installedVersions != "" {
			json.Unmarshal([]byte(installedVersions), &versions)
		}

		// 创建 manifest.json
		appManifest := &store.AppManifest{
			Name:               name,
			Bucket:             bucket,
			CurrentVersion:     currentVersion,
			InstalledVersions:  versions,
			Dependencies:       deps,
			InstalledAt:        installedAt,
			InstalledOnRequest: true, // 默认标记为用户主动安装
			Isolated:           isolated,
		}

		manifestData, err := json.MarshalIndent(appManifest, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化 manifest 失败：%w", err)
		}

		manifestPath := filepath.Join(appDir, "manifest.json")
		if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
			return fmt.Errorf("写入 manifest 文件失败：%w", err)
		}

		count++
	}

	stats.AppCount = count
	if count > 0 {
		output.Success(" 完成 (%d 个应用)", count)
	} else {
		output.Info(" 跳过 (无数据)")
	}

	return nil
}

// migrateOperations 迁移操作记录
func (m *SQLiteToFSMigrator) migrateOperations(ctx context.Context, stats *MigrationStats, dryRun bool) error {
	output.Info("正在迁移操作记录...")

	rows, err := m.db.QueryContext(ctx, `
		SELECT app_name, type, path, key, name, value, target, link, original_value, created_at
		FROM operations
		ORDER BY app_name, created_at
	`)
	if err != nil {
		if isTableNotExist(err, "operations") {
			output.Warn(" 跳过 (表不存在)")
			return nil
		}
		return err
	}
	defer rows.Close()

	// 按应用分组操作记录
	opsByApp := make(map[string][]store.Operation)
	count := 0

	for rows.Next() {
		var (
			appName, opType                                     string
			path, key, name, value, target, link, originalValue string
			createdAt                                           time.Time
		)

		if err := rows.Scan(&appName, &opType, &path, &key, &name, &value, &target, &link, &originalValue, &createdAt); err != nil {
			return fmt.Errorf("扫描操作记录失败：%w", err)
		}

		opsByApp[appName] = append(opsByApp[appName], store.Operation{
			Type:          opType,
			Path:          path,
			Key:           key,
			Name:          name,
			Value:         value,
			Target:        target,
			Link:          link,
			OriginalValue: originalValue,
			CreatedAt:     createdAt,
		})

		count++
	}

	if dryRun {
		stats.OperationCount = count
		if count > 0 {
			output.Success(" 完成 (%d 条记录)", count)
		} else {
			output.Info(" 跳过 (无数据)")
		}
		return nil
	}

	// 写入每个应用的操作记录
	appsDir := filepath.Join(m.targetDir, "apps")
	for appName, ops := range opsByApp {
		appDir := filepath.Join(appsDir, appName)
		if err := os.MkdirAll(appDir, 0755); err != nil {
			continue
		}

		opsFile := store.OperationsFile{
			Version:    "",
			Operations: ops,
		}

		opsData, err := json.MarshalIndent(opsFile, "", "  ")
		if err != nil {
			continue
		}

		opsPath := filepath.Join(appDir, "operations.json")
		if err := os.WriteFile(opsPath, opsData, 0644); err != nil {
			continue
		}
	}

	stats.OperationCount = count
	if count > 0 {
		output.Success(" 完成 (%d 条记录)", count)
	} else {
		output.Info(" 跳过 (无数据)")
	}

	return nil
}

// migrateBuckets 迁移软件源配置
func (m *SQLiteToFSMigrator) migrateBuckets(ctx context.Context, stats *MigrationStats, dryRun bool) error {
	output.Info("正在迁移软件源配置...")

	rows, err := m.db.QueryContext(ctx, `
		SELECT id, name, author, description, homepage, license,
		       repository_url, repository_branch, added_at, updated_at
		FROM buckets
	`)
	if err != nil {
		if isTableNotExist(err, "buckets") {
			output.Warn(" 跳过 (表不存在)")
			return nil
		}
		return err
	}
	defer rows.Close()

	count := 0
	buckets := make(map[string]*store.BucketConfig)

	for rows.Next() {
		var (
			id, name, author, description, homepage, license string
			repoURL, repoBranch                              string
			addedAt, updatedAt                               time.Time
		)

		if err := rows.Scan(&id, &name, &author, &description, &homepage, &license,
			&repoURL, &repoBranch, &addedAt, &updatedAt); err != nil {
			return fmt.Errorf("扫描软件源记录失败：%w", err)
		}

		buckets[id] = &store.BucketConfig{
			ID:          id,
			Name:        name,
			Author:      author,
			Description: description,
			Homepage:    homepage,
			License:     license,
			Repository: manifest.RepositoryInfo{
				URL:    repoURL,
				Branch: repoBranch,
			},
			AddedAt:   addedAt,
			UpdatedAt: updatedAt,
		}

		count++
	}

	if dryRun {
		stats.BucketCount = count
		if count > 0 {
			output.Success(" 完成 (%d 个软件源)", count)
		} else {
			output.Info(" 跳过 (无数据)")
		}
		return nil
	}

	// 写入 bucket-index.json
	if len(buckets) > 0 {
		bucketIndex := store.BucketIndex{
			GeneratedAt: time.Now().UTC(),
			Buckets:     buckets,
		}

		bucketData, err := json.MarshalIndent(bucketIndex, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化 bucket 索引失败：%w", err)
		}

		bucketIndexPath := filepath.Join(m.targetDir, "bucket-index.json")
		if err := os.WriteFile(bucketIndexPath, bucketData, 0644); err != nil {
			return fmt.Errorf("写入 bucket 索引文件失败：%w", err)
		}

		stats.BucketCount = count
		output.Success(" 完成 (%d 个软件源)", count)
	} else {
		output.Info(" 跳过 (无数据)")
	}

	return nil
}

// migrateRuntimeIndex 迁移运行时索引
func (m *SQLiteToFSMigrator) migrateRuntimeIndex(ctx context.Context, stats *MigrationStats, dryRun bool) error {
	output.Info("正在迁移运行时索引...")

	rows, err := m.db.QueryContext(ctx, `
		SELECT app_name, runtime_name, version, path
		FROM runtime_index
	`)
	if err != nil {
		if isTableNotExist(err, "runtime_index") {
			output.Warn(" 跳过 (表不存在)")
			return nil
		}
		return err
	}
	defer rows.Close()

	count := 0
	runtimeIndex := make(store.RuntimeIndex)

	for rows.Next() {
		var appName, runtimeName, version, path string

		if err := rows.Scan(&appName, &runtimeName, &version, &path); err != nil {
			return fmt.Errorf("扫描运行时记录失败：%w", err)
		}

		runtimeIndex[runtimeName] = &manifest.RuntimeInfo{
			Version:     version,
			InstalledAt: time.Now().UTC(),
			RequiredBy:  []string{appName},
			RefCount:    1,
		}

		count++
	}

	if dryRun {
		stats.RuntimeCount = count
		if count > 0 {
			output.Success(" 完成 (%d 项)", count)
		} else {
			output.Info(" 跳过 (无数据)")
		}
		return nil
	}

	// 写入 runtime-index.json
	if len(runtimeIndex) > 0 {
		runtimeData, err := json.MarshalIndent(runtimeIndex, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化 runtime 索引失败：%w", err)
		}

		runtimeIndexPath := filepath.Join(m.targetDir, "runtime-index.json")
		if err := os.WriteFile(runtimeIndexPath, runtimeData, 0644); err != nil {
			return fmt.Errorf("写入 runtime 索引文件失败：%w", err)
		}

		stats.RuntimeCount = count
		output.Success(" 完成 (%d 项)", count)
	} else {
		output.Info(" 跳过 (无数据)")
	}

	return nil
}

// migrateDepsIndex 迁移依赖索引
func (m *SQLiteToFSMigrator) migrateDepsIndex(ctx context.Context, stats *MigrationStats, dryRun bool) error {
	output.Info("正在迁移依赖索引...")

	rows, err := m.db.QueryContext(ctx, `
		SELECT app_name, dependencies, dependents
		FROM deps_index
	`)
	if err != nil {
		if isTableNotExist(err, "deps_index") {
			output.Warn(" 跳过 (表不存在)")
			return nil
		}
		return err
	}
	defer rows.Close()

	count := 0
	appsDeps := make(map[string]*store.AppDeps)

	for rows.Next() {
		var appName, depsJSON, dependentsJSON string

		if err := rows.Scan(&appName, &depsJSON, &dependentsJSON); err != nil {
			return fmt.Errorf("扫描依赖记录失败：%w", err)
		}

		appDeps := &store.AppDeps{
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
		}

		if depsJSON != "" {
			json.Unmarshal([]byte(depsJSON), &appDeps.Dependencies)
		}
		if dependentsJSON != "" {
			json.Unmarshal([]byte(dependentsJSON), &appDeps.Dependents)
		}

		appsDeps[appName] = appDeps
		count++
	}

	if dryRun {
		stats.DepsCount = count
		if count > 0 {
			output.Success(" 完成 (%d 项)", count)
		} else {
			output.Info(" 跳过 (无数据)")
		}
		return nil
	}

	// 写入 deps-index.json
	if len(appsDeps) > 0 {
		depsIndex := &store.DepsIndex{
			GeneratedAt: time.Now().UTC(),
			Apps:        appsDeps,
		}

		depsData, err := json.MarshalIndent(depsIndex, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化依赖索引失败：%w", err)
		}

		depsIndexPath := filepath.Join(m.targetDir, "deps-index.json")
		if err := os.WriteFile(depsIndexPath, depsData, 0644); err != nil {
			return fmt.Errorf("写入依赖索引文件失败：%w", err)
		}

		stats.DepsCount = count
		output.Success(" 完成 (%d 项)", count)
	} else {
		output.Info(" 跳过 (无数据)")
	}

	return nil
}

// copyFile 复制文件
func (m *SQLiteToFSMigrator) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// copyDir 复制目录
func (m *SQLiteToFSMigrator) copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := m.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := m.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// isTableNotExist 检查错误是否表示表不存在
func isTableNotExist(err error, tableName string) bool {
	return err != nil && (err.Error() == fmt.Sprintf("no such table: %s", tableName) ||
		err.Error() == fmt.Sprintf("table %s does not exist", tableName))
}

// ============================================================================
// 文件系统回滚器
// ============================================================================

// RollbackStats 回滚统计
type RollbackStats struct {
	AppCount       int
	OperationCount int
	BucketCount    int
	Duration       time.Duration
}

// FSRollbacker 文件系统回滚器
type FSRollbacker struct {
	sourceDir string
	backupDir string
}

// NewFSRollbacker 创建新的回滚器
func NewFSRollbacker(sourceDir, backupDir string) (*FSRollbacker, error) {
	return &FSRollbacker{
		sourceDir: sourceDir,
		backupDir: backupDir,
	}, nil
}

// Rollback 执行回滚
func (r *FSRollbacker) Rollback(ctx context.Context) (*RollbackStats, error) {
	startTime := time.Now()
	stats := &RollbackStats{}

	// 恢复 apps 目录
	output.Info("正在恢复应用数据...")
	if err := r.restoreApps(ctx, stats); err != nil {
		return nil, fmt.Errorf("恢复应用数据失败：%w", err)
	}

	// 恢复索引文件
	output.Info("正在恢复索引文件...")
	if err := r.restoreIndexes(ctx, stats); err != nil {
		return nil, fmt.Errorf("恢复索引文件失败：%w", err)
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

// restoreApps 恢复应用数据
func (r *FSRollbacker) restoreApps(ctx context.Context, stats *RollbackStats) error {
	appsBackupDir := filepath.Join(r.backupDir, "apps")
	appsTargetDir := filepath.Join(r.sourceDir, "apps")

	if _, err := os.Stat(appsBackupDir); os.IsNotExist(err) {
		output.Info(" 跳过 (无备份数据)")
		return nil
	}

	// 删除现有的 apps 目录
	if err := os.RemoveAll(appsTargetDir); err != nil {
		return fmt.Errorf("删除现有 apps 目录失败：%w", err)
	}

	// 复制备份的 apps 目录
	if err := copyDir(appsBackupDir, appsTargetDir); err != nil {
		return fmt.Errorf("复制 apps 目录失败：%w", err)
	}

	// 统计应用数量
	entries, _ := os.ReadDir(appsTargetDir)
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}

	stats.AppCount = count
	output.Success(" 完成 (%d 个应用)", count)
	return nil
}

// restoreIndexes 恢复索引文件
func (r *FSRollbacker) restoreIndexes(ctx context.Context, stats *RollbackStats) error {
	indexFiles := []string{"bucket-index.json", "runtime-index.json", "deps-index.json"}
	count := 0

	for _, indexFile := range indexFiles {
		backupPath := filepath.Join(r.backupDir, indexFile)
		targetPath := filepath.Join(r.sourceDir, indexFile)

		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			continue
		}

		// 复制文件
		if err := copyFile(backupPath, targetPath); err != nil {
			return fmt.Errorf("复制 %s 失败：%w", indexFile, err)
		}

		count++
	}

	if count > 0 {
		output.Success(" 完成 (%d 个文件)", count)
	} else {
		output.Info(" 跳过 (无备份数据)")
	}

	return nil
}

// copyFile 复制文件（包级别辅助函数）
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// copyDir 复制目录（包级别辅助函数）
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
