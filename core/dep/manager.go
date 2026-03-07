// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

// Manager 依赖管理器接口
type Manager interface {
	// 依赖解析
	Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error)
	CheckConflicts(ctx context.Context, deps *manifest.Dependencies) ([]Conflict, error)
	CheckCircular(ctx context.Context, deps []string) error

	// 运行时库管理
	InstallRuntime(ctx context.Context, dep, version, appName string, size int64) error
	UninstallRuntime(ctx context.Context, dep, appName string) error
	GetRuntimeInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error)
	CleanupRuntime(ctx context.Context) error
	ListRuntimes(ctx context.Context) map[string]*manifest.RuntimeInfo

	// 反向依赖计算
	GetDependents(ctx context.Context, appName string) ([]string, error)
	GetAllDependents(ctx context.Context, appName string) ([]string, error)
	GetDependentsTree(ctx context.Context, appName string) *DependentTree

	// 孤儿依赖清理
	FindOrphans(ctx context.Context) (*manifest.Orphans, error)
	CleanupOrphans(ctx context.Context, orphans *manifest.Orphans) error
	DryRunCleanup(ctx context.Context, orphans *manifest.Orphans) error

	// 依赖索引管理
	RebuildIndex(ctx context.Context) error
	UpdateDepsIndex(ctx context.Context, appName string, deps []string) error
}

// DependencyManager 依赖管理器实现
type DependencyManager struct {
	mu              sync.RWMutex
	bucketMgr       bucket.BucketManager
	storage         store.LegacyStorage
	resolver        Resolver
	runtimeMgr      RuntimeManager
	reverseDepsCalc ReverseDepsCalculator
	orphanDetector  OrphanDetector
	rootPath        string
	depsIndex       *DepsIndex
}

// NewDependencyManager 创建依赖管理器
func NewDependencyManager(bucketMgr bucket.BucketManager, storage store.LegacyStorage, rootPath string) (*DependencyManager, error) {
	// 创建运行时管理器
	runtimeMgr, err := NewRuntimeManager(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime manager: %w", err)
	}

	// 创建依赖索引
	depsIndex := NewDepsIndex(rootPath)

	// 创建依赖解析器
	resolver := NewResolver(bucketMgr, storage)

	// 创建反向依赖计算器
	reverseDepsCalc := NewReverseDepsCalculator(rootPath)

	// 创建孤儿依赖检测器
	orphanDetector := NewOrphanDetector(rootPath, runtimeMgr, depsIndex)

	return &DependencyManager{
		bucketMgr:       bucketMgr,
		storage:         storage,
		resolver:        resolver,
		runtimeMgr:      runtimeMgr,
		reverseDepsCalc: reverseDepsCalc,
		orphanDetector:  orphanDetector,
		rootPath:        rootPath,
		depsIndex:       depsIndex,
	}, nil
}

// BucketManager 返回 bucket 管理器
func (m *DependencyManager) BucketManager() bucket.BucketManager {
	return m.bucketMgr
}

// Resolve 解析应用的依赖树
func (m *DependencyManager) Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if app == nil || app.Script == nil {
		return nil, errors.Newf(errors.KindInvalidInput, "invalid app")
	}

	// 使用解析器解析依赖
	graph, err := m.resolver.Resolve(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	return graph, nil
}

// CheckConflicts 检查依赖冲突
func (m *DependencyManager) CheckConflicts(ctx context.Context, deps *manifest.Dependencies) ([]Conflict, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var conflicts []Conflict

	// 检查 conflicts 字段
	if deps != nil {
		for _, name := range deps.Conflicts {
			// 检查是否已安装
			if m.isAppInstalled(ctx, name) {
				conflicts = append(conflicts, Conflict{
					Name:   name,
					Reason: "conflict with installed app",
				})
			}
		}
	}

	return conflicts, nil
}

// CheckCircular 检测循环依赖
func (m *DependencyManager) CheckCircular(ctx context.Context, deps []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.resolver.CheckCircular(ctx, deps)
}

// InstallRuntime 安装运行时库
func (m *DependencyManager) InstallRuntime(ctx context.Context, dep, version, appName string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runtimeMgr.Install(ctx, dep, version, appName, size)
}

// UninstallRuntime 卸载运行时库
func (m *DependencyManager) UninstallRuntime(ctx context.Context, dep, appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runtimeMgr.Uninstall(ctx, dep, appName)
}

// GetRuntimeInfo 获取运行时库信息
func (m *DependencyManager) GetRuntimeInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.runtimeMgr.GetInfo(ctx, dep)
}

// ListRuntimes 列出所有运行时库
func (m *DependencyManager) ListRuntimes(ctx context.Context) map[string]*manifest.RuntimeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.runtimeMgr.List(ctx)
}

// GetDependents 获取反向依赖（谁依赖我）
func (m *DependencyManager) GetDependents(ctx context.Context, appName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 加载依赖索引
	if err := m.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	dependents := m.depsIndex.GetDependents(appName)
	return dependents, nil
}

// GetAllDependents 获取所有反向依赖（包括间接依赖）
func (m *DependencyManager) GetAllDependents(ctx context.Context, appName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.reverseDepsCalc.GetAllDependents(appName), nil
}

// GetDependentsTree 获取反向依赖树
func (m *DependencyManager) GetDependentsTree(ctx context.Context, appName string) *DependentTree {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.reverseDepsCalc.GetDependentsTree(appName)
}

// FindOrphans 查找孤儿依赖
func (m *DependencyManager) FindOrphans(ctx context.Context) (*manifest.Orphans, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.orphanDetector.Detect(ctx)
}

// CleanupOrphans 清理孤儿依赖
func (m *DependencyManager) CleanupOrphans(ctx context.Context, orphans *manifest.Orphans) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orphanDetector.Cleanup(ctx, orphans)
}

// DryRunCleanup 预演清理孤儿依赖
func (m *DependencyManager) DryRunCleanup(ctx context.Context, orphans *manifest.Orphans) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.orphanDetector.DryRun(ctx, orphans)
}

// RebuildIndex 重建依赖索引
func (m *DependencyManager) RebuildIndex(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重建依赖索引
	if err := m.depsIndex.Rebuild(ctx, m.rootPath); err != nil {
		return fmt.Errorf("failed to rebuild deps index: %w", err)
	}

	// 计算反向依赖
	if err := m.reverseDepsCalc.Calculate(ctx); err != nil {
		return fmt.Errorf("failed to calculate reverse deps: %w", err)
	}

	return nil
}

// UpdateDepsIndex 更新依赖索引
func (m *DependencyManager) UpdateDepsIndex(ctx context.Context, appName string, deps []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载依赖索引
	if err := m.depsIndex.Load(); err != nil {
		return fmt.Errorf("failed to load deps index: %w", err)
	}

	// 更新应用依赖信息
	m.depsIndex.apps[appName] = &AppDeps{
		Dependencies: deps,
		Dependents:   []string{},
	}

	// 重新计算反向依赖
	m.depsIndex.calculateDependents()

	// 保存索引
	return m.depsIndex.Save()
}

// CleanupRuntime 清理无用运行时库
func (m *DependencyManager) CleanupRuntime(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runtimeMgr.Cleanup(ctx)
}

// isAppInstalled 检查应用是否已安装
func (m *DependencyManager) isAppInstalled(ctx context.Context, name string) bool {
	_, err := m.storage.GetInstalledApp(ctx, name)
	return err == nil
}

// Conflict 表示依赖冲突
type Conflict struct {
	Name   string // 冲突的软件名称
	Reason string // 冲突原因
}

// DependencyNode 表示依赖树节点
type DependencyNode struct {
	App          *manifest.App
	Version      string
	Dependencies []*DependencyNode
	Parent       *DependencyNode
	Depth        int
}

// DependencyGraph 表示依赖图
type DependencyGraph struct {
	Nodes map[string]*DependencyNode // 应用名称 -> 节点
	Order []string                   // 拓扑排序后的安装顺序
}
