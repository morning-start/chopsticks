// Package store 提供文件系统存储实现。
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chopsticks/core/manifest"
)

// ============================================================================
// 数据结构定义
// ============================================================================

// Operation 表示一个系统操作记录。
type Operation struct {
	Type          string    `json:"type"`                     // 操作类型：path, env, registry, shortcut, symlink
	Path          string    `json:"path,omitempty"`           // 路径（用于 path 类型）
	Key           string    `json:"key,omitempty"`            // 键名（用于 env, registry 类型）
	Name          string    `json:"name,omitempty"`           // 值名（用于 registry 类型）
	Value         string    `json:"value,omitempty"`          // 值（用于 env, registry 类型）
	Target        string    `json:"target,omitempty"`         // 目标（用于 shortcut, symlink 类型）
	Link          string    `json:"link,omitempty"`           // 符号链接路径（用于 symlink 类型）
	OriginalValue string    `json:"original_value,omitempty"` // 原始值（用于恢复）
	CreatedAt     time.Time `json:"created_at"`               // 创建时间
}

// OperationsFile 表示 operations.json 的文件结构。
type OperationsFile struct {
	Version    string      `json:"version"`    // 应用版本
	Operations []Operation `json:"operations"` // 操作列表
}

// AppManifest 表示已安装应用的 manifest.json 文件结构。
type AppManifest struct {
	Name               string                `json:"name"`                   // 软件名称
	Bucket             string                `json:"bucket"`                 // 来源软件源
	CurrentVersion     string                `json:"current_version"`        // 当前激活版本
	InstalledVersions  []string              `json:"installed_versions"`     // 已安装的所有版本
	Dependencies       manifest.Dependencies `json:"dependencies,omitempty"` // 依赖声明
	InstalledAt        time.Time             `json:"installed_at"`           // 安装时间
	InstalledOnRequest bool                  `json:"installed_on_request"`   // 是否用户主动安装
	Isolated           bool                  `json:"isolated"`               // 是否为隔离安装
}

// RuntimeIndex 表示 runtime-index.json 的文件结构。
type RuntimeIndex map[string]*manifest.RuntimeInfo

// DepsIndex 表示 deps-index.json 的文件结构。
type DepsIndex struct {
	GeneratedAt time.Time           `json:"generated_at"` // 索引生成时间
	Apps        map[string]*AppDeps `json:"apps"`         // 应用依赖关系字典
}

// AppDeps 表示单个应用的依赖关系。
type AppDeps struct {
	Dependencies []string `json:"dependencies"` // 该软件依赖的其他软件
	Dependents   []string `json:"dependents"`   // 依赖该软件的其他软件
}

// BucketIndex 表示软件源索引。
type BucketIndex struct {
	GeneratedAt time.Time                `json:"generated_at"` // 索引生成时间
	Buckets     map[string]*BucketConfig `json:"buckets"`      // 软件源配置字典
}

// BucketConfig 是 BucketIndex 中存储的软件源配置。
type BucketConfig struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Author      string                  `json:"author"`
	Description string                  `json:"description"`
	Homepage    string                  `json:"homepage"`
	License     string                  `json:"license"`
	Repository  manifest.RepositoryInfo `json:"repository"`
	AddedAt     time.Time               `json:"added_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// ============================================================================
// 接口定义
// ============================================================================

// AppStorage 定义已安装应用存储接口。
type AppStorage interface {
	SaveApp(ctx context.Context, app *AppManifest) error
	GetApp(ctx context.Context, name string) (*AppManifest, error)
	DeleteApp(ctx context.Context, name string) error
	ListApps(ctx context.Context) ([]*AppManifest, error)
	IsInstalled(ctx context.Context, name string) (bool, error)
}

// BucketStorage 定义软件源存储接口。
type BucketStorage interface {
	SaveBucket(ctx context.Context, bucket *BucketConfig) error
	GetBucket(ctx context.Context, name string) (*BucketConfig, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]*BucketConfig, error)
}

// OperationStorage 定义操作记录存储接口。
type OperationStorage interface {
	SaveOperation(ctx context.Context, appName string, op *Operation) error
	GetOperations(ctx context.Context, appName string) ([]Operation, error)
	DeleteOperations(ctx context.Context, appName string) error
}

// DependencyStorage 定义依赖索引存储接口。
type DependencyStorage interface {
	SaveRuntimeIndex(ctx context.Context, index RuntimeIndex) error
	GetRuntimeIndex(ctx context.Context) (RuntimeIndex, error)
	SaveDepsIndex(ctx context.Context, index *DepsIndex) error
	GetDepsIndex(ctx context.Context) (*DepsIndex, error)
}

// Storage 组合所有存储接口。
type Storage interface {
	AppStorage
	BucketStorage
	OperationStorage
	DependencyStorage
	Close() error
}

// ============================================================================
// 文件系统存储实现
// ============================================================================

// fsStorage 是 Storage 的文件系统实现。
type fsStorage struct {
	rootDir      string                   // 根目录
	appsDir      string                   // 应用目录
	bucketsDir   string                   // 软件源目录
	runtimeIndex string                   // runtime-index.json 路径
	depsIndex    string                   // deps-index.json 路径
	bucketIndex  string                   // bucket-index.json 路径
	mu           sync.RWMutex             // 并发控制
	bucketCache  map[string]*BucketConfig // Bucket 配置缓存
	runtimeCache RuntimeIndex             // Runtime 索引缓存
	depsCache    *DepsIndex               // 依赖索引缓存
}

// 编译时接口检查。
var _ Storage = (*fsStorage)(nil)

// NewFSStorage 创建新的文件系统 Storage。
func NewFSStorage(rootDir string) (Storage, error) {
	// 创建目录结构
	appsDir := filepath.Join(rootDir, "apps")
	bucketsDir := filepath.Join(rootDir, "buckets")
	cacheDir := filepath.Join(rootDir, "cache")

	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建应用目录：%w", err)
	}
	if err := os.MkdirAll(bucketsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建软件源目录：%w", err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录：%w", err)
	}

	s := &fsStorage{
		rootDir:      rootDir,
		appsDir:      appsDir,
		bucketsDir:   bucketsDir,
		runtimeIndex: filepath.Join(rootDir, "runtime-index.json"),
		depsIndex:    filepath.Join(rootDir, "deps-index.json"),
		bucketIndex:  filepath.Join(rootDir, "bucket-index.json"),
		bucketCache:  make(map[string]*BucketConfig),
	}

	// 加载缓存
	if err := s.loadCaches(); err != nil {
		return nil, fmt.Errorf("加载缓存：%w", err)
	}

	return s, nil
}

// loadCaches 从文件系统加载所有缓存。
func (s *fsStorage) loadCaches() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 加载 bucket 索引
	if err := s.loadBucketIndexLocked(); err != nil {
		return fmt.Errorf("加载 bucket 索引：%w", err)
	}

	// 加载 runtime 索引
	if err := s.loadRuntimeIndexLocked(); err != nil {
		return fmt.Errorf("加载 runtime 索引：%w", err)
	}

	// 加载依赖索引
	if err := s.loadDepsIndexLocked(); err != nil {
		return fmt.Errorf("加载依赖索引：%w", err)
	}

	return nil
}

// loadBucketIndexLocked 加载 bucket 索引（需要持有锁）。
func (s *fsStorage) loadBucketIndexLocked() error {
	data, err := os.ReadFile(s.bucketIndex)
	if err != nil {
		if os.IsNotExist(err) {
			// 初始化空索引
			s.bucketCache = make(map[string]*BucketConfig)
			return nil
		}
		return err
	}

	var index BucketIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("解析 bucket 索引：%w", err)
	}

	s.bucketCache = index.Buckets
	return nil
}

// loadRuntimeIndexLocked 加载 runtime 索引（需要持有锁）。
func (s *fsStorage) loadRuntimeIndexLocked() error {
	data, err := os.ReadFile(s.runtimeIndex)
	if err != nil {
		if os.IsNotExist(err) {
			// 初始化空索引
			s.runtimeCache = make(RuntimeIndex)
			return nil
		}
		return err
	}

	var index RuntimeIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("解析 runtime 索引：%w", err)
	}

	s.runtimeCache = index
	return nil
}

// loadDepsIndexLocked 加载依赖索引（需要持有锁）。
func (s *fsStorage) loadDepsIndexLocked() error {
	data, err := os.ReadFile(s.depsIndex)
	if err != nil {
		if os.IsNotExist(err) {
			// 初始化空索引
			s.depsCache = &DepsIndex{
				GeneratedAt: time.Time{},
				Apps:        make(map[string]*AppDeps),
			}
			return nil
		}
		return err
	}

	var index DepsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("解析依赖索引：%w", err)
	}

	s.depsCache = &index
	return nil
}

// ============================================================================
// AppStorage 实现
// ============================================================================

// SaveApp 保存已安装应用信息。
func (s *fsStorage) SaveApp(ctx context.Context, app *AppManifest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app.Name == "" {
		return fmt.Errorf("应用名称不能为空")
	}

	// 创建应用目录
	appDir := filepath.Join(s.appsDir, app.Name)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("创建应用目录：%w", err)
	}

	// 序列化并保存 manifest.json
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 manifest：%w", err)
	}

	manifestPath := filepath.Join(appDir, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("写入 manifest 文件：%w", err)
	}

	return nil
}

// GetApp 获取已安装应用信息。
func (s *fsStorage) GetApp(ctx context.Context, name string) (*AppManifest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	manifestPath := filepath.Join(s.appsDir, name, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("应用不存在：%s", name)
		}
		return nil, fmt.Errorf("读取 manifest 文件：%w", err)
	}

	var app AppManifest
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("解析 manifest 文件：%w", err)
	}

	return &app, nil
}

// DeleteApp 删除已安装应用。
func (s *fsStorage) DeleteApp(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appDir := filepath.Join(s.appsDir, name)

	// 检查应用是否存在
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("应用不存在：%s", name)
	}

	// 删除整个应用目录（包括 manifest.json 和 operations.json）
	if err := os.RemoveAll(appDir); err != nil {
		return fmt.Errorf("删除应用目录：%w", err)
	}

	return nil
}

// ListApps 列出所有已安装应用。
func (s *fsStorage) ListApps(ctx context.Context) ([]*AppManifest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.appsDir)
	if err != nil {
		return nil, fmt.Errorf("读取应用目录：%w", err)
	}

	var apps []*AppManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查是否存在 manifest.json
		manifestPath := filepath.Join(s.appsDir, entry.Name(), "manifest.json")
		if _, err := os.Stat(manifestPath); err != nil {
			continue // 跳过没有 manifest.json 的目录
		}

		// 读取并解析 manifest
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // 跳过读取失败的应用
		}

		var app AppManifest
		if err := json.Unmarshal(data, &app); err != nil {
			continue // 跳过解析失败的应用
		}

		apps = append(apps, &app)
	}

	return apps, nil
}

// IsInstalled 检查应用是否已安装。
func (s *fsStorage) IsInstalled(ctx context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	manifestPath := filepath.Join(s.appsDir, name, "manifest.json")
	_, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ============================================================================
// BucketStorage 实现
// ============================================================================

// SaveBucket 保存软件源配置。
func (s *fsStorage) SaveBucket(ctx context.Context, bucket *BucketConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bucket.ID == "" {
		return fmt.Errorf("软件源 ID 不能为空")
	}

	// 更新缓存
	s.bucketCache[bucket.ID] = bucket

	// 保存 bucket-index.json
	return s.saveBucketIndexLocked()
}

// saveBucketIndexLocked 保存 bucket 索引（需要持有锁）。
func (s *fsStorage) saveBucketIndexLocked() error {
	index := BucketIndex{
		GeneratedAt: time.Now().UTC(),
		Buckets:     s.bucketCache,
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 bucket 索引：%w", err)
	}

	if err := os.WriteFile(s.bucketIndex, data, 0644); err != nil {
		return fmt.Errorf("写入 bucket 索引文件：%w", err)
	}

	return nil
}

// GetBucket 获取软件源配置。
func (s *fsStorage) GetBucket(ctx context.Context, name string) (*BucketConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, exists := s.bucketCache[name]
	if !exists {
		return nil, fmt.Errorf("软件源不存在：%s", name)
	}

	// 返回副本
	bucketCopy := *bucket
	return &bucketCopy, nil
}

// DeleteBucket 删除软件源配置。
func (s *fsStorage) DeleteBucket(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.bucketCache[name]; !exists {
		return fmt.Errorf("软件源不存在：%s", name)
	}

	// 从缓存中删除
	delete(s.bucketCache, name)

	// 保存 bucket-index.json
	return s.saveBucketIndexLocked()
}

// ListBuckets 列出所有软件源。
func (s *fsStorage) ListBuckets(ctx context.Context) ([]*BucketConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets := make([]*BucketConfig, 0, len(s.bucketCache))
	for _, bucket := range s.bucketCache {
		// 返回副本
		bucketCopy := *bucket
		buckets = append(buckets, &bucketCopy)
	}

	return buckets, nil
}

// ============================================================================
// OperationStorage 实现
// ============================================================================

// SaveOperation 保存操作记录。
func (s *fsStorage) SaveOperation(ctx context.Context, appName string, op *Operation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	opsPath := filepath.Join(s.appsDir, appName, "operations.json")

	// 读取现有操作记录
	var opsFile OperationsFile
	data, err := os.ReadFile(opsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("读取操作记录文件：%w", err)
		}
		// 文件不存在，初始化空记录
		opsFile = OperationsFile{
			Version:    "",
			Operations: make([]Operation, 0),
		}
	} else {
		if err := json.Unmarshal(data, &opsFile); err != nil {
			return fmt.Errorf("解析操作记录文件：%w", err)
		}
	}

	// 添加新操作
	if op.CreatedAt.IsZero() {
		op.CreatedAt = time.Now().UTC()
	}
	opsFile.Operations = append(opsFile.Operations, *op)

	// 保存回文件
	data, err = json.MarshalIndent(opsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化操作记录：%w", err)
	}

	if err := os.WriteFile(opsPath, data, 0644); err != nil {
		return fmt.Errorf("写入操作记录文件：%w", err)
	}

	return nil
}

// GetOperations 获取操作记录。
func (s *fsStorage) GetOperations(ctx context.Context, appName string) ([]Operation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	opsPath := filepath.Join(s.appsDir, appName, "operations.json")

	data, err := os.ReadFile(opsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Operation{}, nil // 返回空列表
		}
		return nil, fmt.Errorf("读取操作记录文件：%w", err)
	}

	var opsFile OperationsFile
	if err := json.Unmarshal(data, &opsFile); err != nil {
		return nil, fmt.Errorf("解析操作记录文件：%w", err)
	}

	return opsFile.Operations, nil
}

// DeleteOperations 删除操作记录。
func (s *fsStorage) DeleteOperations(ctx context.Context, appName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	opsPath := filepath.Join(s.appsDir, appName, "operations.json")

	// 检查文件是否存在
	if _, err := os.Stat(opsPath); os.IsNotExist(err) {
		return nil // 文件不存在，直接返回成功
	}

	// 删除文件
	if err := os.Remove(opsPath); err != nil {
		return fmt.Errorf("删除操作记录文件：%w", err)
	}

	return nil
}

// ============================================================================
// DependencyStorage 实现
// ============================================================================

// SaveRuntimeIndex 保存运行时库索引。
func (s *fsStorage) SaveRuntimeIndex(ctx context.Context, index RuntimeIndex) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runtimeCache = index

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 runtime 索引：%w", err)
	}

	if err := os.WriteFile(s.runtimeIndex, data, 0644); err != nil {
		return fmt.Errorf("写入 runtime 索引文件：%w", err)
	}

	return nil
}

// GetRuntimeIndex 获取运行时库索引。
func (s *fsStorage) GetRuntimeIndex(ctx context.Context) (RuntimeIndex, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本
	result := make(RuntimeIndex)
	for k, v := range s.runtimeCache {
		if v != nil {
			vCopy := *v
			result[k] = &vCopy
		}
	}

	return result, nil
}

// SaveDepsIndex 保存依赖索引。
func (s *fsStorage) SaveDepsIndex(ctx context.Context, index *DepsIndex) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.depsCache = index

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化依赖索引：%w", err)
	}

	if err := os.WriteFile(s.depsIndex, data, 0644); err != nil {
		return fmt.Errorf("写入依赖索引文件：%w", err)
	}

	return nil
}

// GetDepsIndex 获取依赖索引。
func (s *fsStorage) GetDepsIndex(ctx context.Context) (*DepsIndex, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.depsCache == nil {
		return &DepsIndex{
			GeneratedAt: time.Time{},
			Apps:        make(map[string]*AppDeps),
		}, nil
	}

	// 返回深拷贝
	result := &DepsIndex{
		GeneratedAt: s.depsCache.GeneratedAt,
		Apps:        make(map[string]*AppDeps),
	}
	for k, v := range s.depsCache.Apps {
		if v != nil {
			vCopy := *v
			depsCopy := make([]string, len(v.Dependencies))
			copy(depsCopy, v.Dependencies)
			dependentsCopy := make([]string, len(v.Dependents))
			copy(dependentsCopy, v.Dependents)
			vCopy.Dependencies = depsCopy
			vCopy.Dependents = dependentsCopy
			result.Apps[k] = &vCopy
		}
	}

	return result, nil
}

// ============================================================================
// 辅助方法
// ============================================================================

// Close 关闭存储（文件系统不需要特殊清理）。
func (s *fsStorage) Close() error {
	// 文件系统存储不需要显式关闭
	return nil
}

// GetAppsDir 返回应用目录路径。
func (s *fsStorage) GetAppsDir() string {
	return s.appsDir
}

// GetBucketsDir 返回软件源目录路径。
func (s *fsStorage) GetBucketsDir() string {
	return s.bucketsDir
}

// GetRootDir 返回根目录路径。
func (s *fsStorage) GetRootDir() string {
	return s.rootDir
}

// ============================================================================
// 迁移辅助方法
// ============================================================================

// NeedsMigration 检查是否需要从 SQLite 迁移数据。
// 如果检测到 SQLite 数据库文件且文件系统数据为空，则需要迁移。
func NeedsMigration(sqlitePath, fsDataDir string) (bool, error) {
	// 检查 SQLite 文件是否存在
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		return false, nil // SQLite 不存在，不需要迁移
	}

	// 检查文件系统数据是否存在
	appsDir := filepath.Join(fsDataDir, "apps")
	bucketIndex := filepath.Join(fsDataDir, "bucket-index.json")

	// 如果 apps 目录不存在或为空，且 bucket-index.json 不存在，则需要迁移
	appsEmpty := true
	if entries, err := os.ReadDir(appsDir); err == nil {
		if len(entries) > 0 {
			appsEmpty = false
		}
	}

	bucketIndexExists := true
	if _, err := os.Stat(bucketIndex); os.IsNotExist(err) {
		bucketIndexExists = false
	}

	// 如果文件系统数据为空，则需要迁移
	return appsEmpty && !bucketIndexExists, nil
}

// HasExistingFSData 检查文件系统存储是否已有数据。
func HasExistingFSData(fsDataDir string) bool {
	appsDir := filepath.Join(fsDataDir, "apps")
	bucketIndex := filepath.Join(fsDataDir, "bucket-index.json")

	// 检查 apps 目录
	if entries, err := os.ReadDir(appsDir); err == nil && len(entries) > 0 {
		return true
	}

	// 检查 bucket-index.json
	if _, err := os.Stat(bucketIndex); err == nil {
		return true
	}

	return false
}

// GetMigrationBackupDir 获取迁移备份目录。
func GetMigrationBackupDir(fsDataDir string) (string, error) {
	backupBase := filepath.Join(fsDataDir, "migrate-backups")

	entries, err := os.ReadDir(backupBase)
	if err != nil {
		return "", err
	}

	var latestTime time.Time
	var latestDir string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

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

	return latestDir, nil
}

// RebuildDepsIndex 从所有 manifest.json 重建依赖索引。
func (s *fsStorage) RebuildDepsIndex(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 读取所有应用
	apps, err := s.listAppsLocked()
	if err != nil {
		return fmt.Errorf("读取应用列表：%w", err)
	}

	// 构建依赖关系
	index := &DepsIndex{
		GeneratedAt: time.Now().UTC(),
		Apps:        make(map[string]*AppDeps),
	}

	// 第一步：收集所有依赖关系
	for _, app := range apps {
		appDeps := &AppDeps{
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
		}

		// 收集所有类型的依赖
		for _, dep := range app.Dependencies.Runtime {
			appDeps.Dependencies = append(appDeps.Dependencies, dep.Name)
		}
		for _, dep := range app.Dependencies.Tools {
			appDeps.Dependencies = append(appDeps.Dependencies, dep.Name)
		}
		for _, dep := range app.Dependencies.Libraries {
			appDeps.Dependencies = append(appDeps.Dependencies, dep.Name)
		}

		index.Apps[app.Name] = appDeps
	}

	// 第二步：计算反向依赖
	for appName, appDeps := range index.Apps {
		for _, depName := range appDeps.Dependencies {
			if depAppDeps, exists := index.Apps[depName]; exists {
				depAppDeps.Dependents = append(depAppDeps.Dependents, appName)
			}
		}
	}

	s.depsCache = index

	// 保存索引
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化依赖索引：%w", err)
	}

	if err := os.WriteFile(s.depsIndex, data, 0644); err != nil {
		return fmt.Errorf("写入依赖索引文件：%w", err)
	}

	return nil
}

// listAppsLocked 列出所有应用（需要持有锁）。
func (s *fsStorage) listAppsLocked() ([]*AppManifest, error) {
	entries, err := os.ReadDir(s.appsDir)
	if err != nil {
		return nil, fmt.Errorf("读取应用目录：%w", err)
	}

	var apps []*AppManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(s.appsDir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var app AppManifest
		if err := json.Unmarshal(data, &app); err != nil {
			continue
		}

		apps = append(apps, &app)
	}

	return apps, nil
}
