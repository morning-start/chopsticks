// Package storage 提供存储功能。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"

	_ "modernc.org/sqlite"
)

// InstallOperation 表示一个安装操作记录。
type InstallOperation struct {
	ID          string
	InstalledID string
	Operation   string
	TargetPath  string
	TargetValue string
	CreatedAt   string
}

// SystemOperation 表示一个系统操作记录。
type SystemOperation struct {
	ID            string
	InstalledID   string
	Operation     string
	TargetType    string
	TargetPath    string
	TargetKey     string
	TargetValue   string
	OriginalValue string
	CreatedAt     string
}

// Storage 定义存储接口。
type Storage interface {
	// 已安装的应用
	SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error
	GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error)
	DeleteInstalledApp(ctx context.Context, name string) error
	ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error)
	IsInstalled(ctx context.Context, name string) (bool, error)

	// 软件源
	SaveBucket(ctx context.Context, b *manifest.BucketConfig) error
	GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error)

	// 安装操作追踪
	SaveInstallOperation(ctx context.Context, op *InstallOperation) error
	GetInstallOperations(ctx context.Context, installedID string) ([]*InstallOperation, error)
	DeleteInstallOperations(ctx context.Context, installedID string) error

	// 系统操作追踪
	SaveSystemOperation(ctx context.Context, op *SystemOperation) error
	GetSystemOperations(ctx context.Context, installedID string) ([]*SystemOperation, error)
	DeleteSystemOperations(ctx context.Context, installedID string) error

	Close() error
}

// sqliteStorage 是 Storage 的 SQLite 实现。
type sqliteStorage struct {
	db   *sql.DB
	path string
}

// 编译时接口检查。
var _ Storage = (*sqliteStorage)(nil)

// New 创建新的 Storage（使用 SQLite）。
func New(path string) (Storage, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_fk=on")
	if err != nil {
		return nil, fmt.Errorf("打开 sqlite 数据库: %w", err)
	}

	if err := initSQLiteTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表: %w", err)
	}

	return &sqliteStorage{
		db:   db,
		path: path,
	}, nil
}

func initSQLiteTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS buckets (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		author TEXT,
		description TEXT,
		homepage TEXT,
		license TEXT,
		repo_type TEXT,
		repo_url TEXT,
		repo_branch TEXT DEFAULT 'main',
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		local_path TEXT
	);

	CREATE TABLE IF NOT EXISTS apps (
		id TEXT PRIMARY KEY,
		bucket_id TEXT NOT NULL,
		name TEXT NOT NULL,
		version TEXT,
		description TEXT,
		homepage TEXT,
		license TEXT,
		category TEXT,
		tags TEXT,
		maintainer TEXT,
		script_path TEXT,
		meta_path TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS app_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_id TEXT NOT NULL,
		version TEXT NOT NULL,
		released_at DATETIME,
		downloads TEXT,
		FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
		UNIQUE(app_id, version)
	);

	CREATE TABLE IF NOT EXISTS installed (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		version TEXT NOT NULL,
		bucket_id TEXT NOT NULL,
		install_dir TEXT NOT NULL,
		installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (bucket_id) REFERENCES buckets(id)
	);

	CREATE TABLE IF NOT EXISTS install_operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		installed_id TEXT NOT NULL,
		operation_type TEXT NOT NULL,
		target_path TEXT,
		target_value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (installed_id) REFERENCES installed(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS system_operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		installed_id TEXT NOT NULL,
		operation_type TEXT NOT NULL,
		target_type TEXT NOT NULL,
		target_path TEXT,
		target_key TEXT,
		target_value TEXT,
		original_value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (installed_id) REFERENCES installed(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_apps_bucket_id ON apps(bucket_id);
	CREATE INDEX IF NOT EXISTS idx_app_versions_app_id ON app_versions(app_id);
	CREATE INDEX IF NOT EXISTS idx_installed_name ON installed(name);
	CREATE INDEX IF NOT EXISTS idx_install_operations_installed_id ON install_operations(installed_id);
	CREATE INDEX IF NOT EXISTS idx_system_operations_installed_id ON system_operations(installed_id);
	`

	_, err := db.Exec(schema)
	return err
}

func (s *sqliteStorage) SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error {
	query := `
	INSERT OR REPLACE INTO installed (id, name, version, bucket_id, install_dir, installed_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.db.ExecContext(ctx, query, a.Name, a.Name, a.Version, a.Bucket, a.InstallDir, a.InstalledAt)
	return err
}

func (s *sqliteStorage) GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error) {
	query := `SELECT name, version, bucket_id, install_dir, installed_at, updated_at FROM installed WHERE name = ?`
	row := s.db.QueryRowContext(ctx, query, name)

	var a manifest.InstalledApp
	err := row.Scan(&a.Name, &a.Version, &a.Bucket, &a.InstallDir, &a.InstalledAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("应用不存在: %s", name)
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *sqliteStorage) DeleteInstalledApp(ctx context.Context, name string) error {
	query := `DELETE FROM installed WHERE name = ?`
	_, err := s.db.ExecContext(ctx, query, name)
	return err
}

func (s *sqliteStorage) ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error) {
	query := `SELECT name, version, bucket_id, install_dir, installed_at, updated_at FROM installed`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*manifest.InstalledApp
	for rows.Next() {
		var a manifest.InstalledApp
		if err := rows.Scan(&a.Name, &a.Version, &a.Bucket, &a.InstallDir, &a.InstalledAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, rows.Err()
}

func (s *sqliteStorage) IsInstalled(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM installed WHERE name = ?)`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, name).Scan(&exists)
	return exists, err
}

func (s *sqliteStorage) SaveBucket(ctx context.Context, b *manifest.BucketConfig) error {
	query := `
	INSERT OR REPLACE INTO buckets (id, name, author, description, homepage, license, repo_type, repo_url, repo_branch, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.db.ExecContext(ctx, query,
		b.ID, b.Name, b.Author, b.Description, b.Homepage, b.License,
		b.Repository.Type, b.Repository.URL, b.Repository.Branch)
	return err
}

func (s *sqliteStorage) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	query := `SELECT id, name, author, description, homepage, license, repo_type, repo_url, repo_branch FROM buckets WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, name)

	var b manifest.BucketConfig
	err := row.Scan(&b.ID, &b.Name, &b.Author, &b.Description, &b.Homepage, &b.License,
		&b.Repository.Type, &b.Repository.URL, &b.Repository.Branch)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("软件源不存在: %s", name)
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *sqliteStorage) DeleteBucket(ctx context.Context, name string) error {
	query := `DELETE FROM buckets WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, name)
	return err
}

func (s *sqliteStorage) ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error) {
	query := `SELECT id, name, author, description, homepage, license, repo_type, repo_url, repo_branch FROM buckets`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []*manifest.BucketConfig
	for rows.Next() {
		var b manifest.BucketConfig
		if err := rows.Scan(&b.ID, &b.Name, &b.Author, &b.Description, &b.Homepage, &b.License,
			&b.Repository.Type, &b.Repository.URL, &b.Repository.Branch); err != nil {
			return nil, err
		}
		buckets = append(buckets, &b)
	}
	return buckets, rows.Err()
}

func (s *sqliteStorage) SaveInstallOperation(ctx context.Context, op *InstallOperation) error {
	query := `
	INSERT INTO install_operations (installed_id, operation_type, target_path, target_value)
	VALUES (?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, op.InstalledID, op.Operation, op.TargetPath, op.TargetValue)
	return err
}

func (s *sqliteStorage) GetInstallOperations(ctx context.Context, installedID string) ([]*InstallOperation, error) {
	query := `
	SELECT id, installed_id, operation_type, target_path, target_value, created_at
	FROM install_operations WHERE installed_id = ?
	`
	rows, err := s.db.QueryContext(ctx, query, installedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []*InstallOperation
	for rows.Next() {
		var op InstallOperation
		if err := rows.Scan(&op.ID, &op.InstalledID, &op.Operation, &op.TargetPath, &op.TargetValue, &op.CreatedAt); err != nil {
			return nil, err
		}
		ops = append(ops, &op)
	}
	return ops, rows.Err()
}

func (s *sqliteStorage) DeleteInstallOperations(ctx context.Context, installedID string) error {
	query := `DELETE FROM install_operations WHERE installed_id = ?`
	_, err := s.db.ExecContext(ctx, query, installedID)
	return err
}

func (s *sqliteStorage) SaveSystemOperation(ctx context.Context, op *SystemOperation) error {
	query := `
	INSERT INTO system_operations (installed_id, operation_type, target_type, target_path, target_key, target_value, original_value)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		op.InstalledID, op.Operation, op.TargetType, op.TargetPath, op.TargetKey, op.TargetValue, op.OriginalValue)
	return err
}

func (s *sqliteStorage) GetSystemOperations(ctx context.Context, installedID string) ([]*SystemOperation, error) {
	query := `
	SELECT id, installed_id, operation_type, target_type, target_path, target_key, target_value, original_value, created_at
	FROM system_operations WHERE installed_id = ?
	`
	rows, err := s.db.QueryContext(ctx, query, installedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []*SystemOperation
	for rows.Next() {
		var op SystemOperation
		if err := rows.Scan(&op.ID, &op.InstalledID, &op.Operation, &op.TargetType,
			&op.TargetPath, &op.TargetKey, &op.TargetValue, &op.OriginalValue, &op.CreatedAt); err != nil {
			return nil, err
		}
		ops = append(ops, &op)
	}
	return ops, rows.Err()
}

func (s *sqliteStorage) DeleteSystemOperations(ctx context.Context, installedID string) error {
	query := `DELETE FROM system_operations WHERE installed_id = ?`
	_, err := s.db.ExecContext(ctx, query, installedID)
	return err
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}
