// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"fmt"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/pkg/errors"
)

// Manager 依赖管理器接口
type Manager interface {
	// 依赖解析
	Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error)
	CheckConflicts(ctx context.Context, deps *manifest.Dependencies) ([]Conflict, error)

	// 运行时库管理
	InstallRuntime(ctx context.Context, dep, version, appName string) error
	UninstallRuntime(ctx context.Context, dep, appName string) error
	GetRuntimeInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error)
	CleanupRuntime(ctx context.Context) error

	// 反向依赖计算
	GetDependents(ctx context.Context, appName string) ([]string, error)

	// 孤儿依赖清理
	FindOrphans(ctx context.Context) (*manifest.Orphans, error)
	CleanupOrphans(ctx context.Context, orphans *manifest.Orphans) error
}

// DependencyManager 依赖管理器实现
type DependencyManager struct {
	bucketMgr   bucket.BucketManager
	runtimeIndex *RuntimeIndex
	depsIndex   *DepsIndex
}

// NewDependencyManager 创建依赖管理器
func NewDependencyManager(bucketMgr bucket.BucketManager, rootPath string) *DependencyManager {
	return &DependencyManager{
		bucketMgr:   bucketMgr,
		runtimeIndex: NewRuntimeIndex(rootPath),
		depsIndex:   NewDepsIndex(rootPath),
	}
}

// BucketManager 返回 bucket 管理器
func (m *DependencyManager) BucketManager() bucket.BucketManager {
	return m.bucketMgr
}

// Resolve 解析应用的依赖树
func (m *DependencyManager) Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error) {
	// 加载依赖索引
	if err := m.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	// 检查冲突
	if app.Script != nil && len(app.Script.Dependencies) > 0 {
		// 将 []Dependency 转换为 Dependencies 结构
		deps := &manifest.Dependencies{
			Runtime:    app.Script.Dependencies,
			Tools:      []manifest.Dependency{},
			Libraries:  []manifest.Dependency{},
			Conflicts:  []string{},
		}
		conflicts, err := m.CheckConflicts(ctx, deps)
		if err != nil {
			return nil, err
		}
		if len(conflicts) > 0 {
			return nil, errors.NewDependencyConflict(
				app.Script.Name,
				fmt.Sprintf("found %d conflicts", len(conflicts)),
			)
		}
	}

	// TODO: 实现依赖解析逻辑
	return &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}, nil
}

// CheckConflicts 检查依赖冲突
func (m *DependencyManager) CheckConflicts(ctx context.Context, deps *manifest.Dependencies) ([]Conflict, error) {
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

// InstallRuntime 安装运行时库
func (m *DependencyManager) InstallRuntime(ctx context.Context, dep, version, appName string) error {
	// 加载运行时索引
	if err := m.runtimeIndex.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 添加或更新运行时库，增加引用计数
	if err := m.runtimeIndex.Add(dep, version, 1); err != nil {
		return err
	}

	return nil
}

// UninstallRuntime 卸载运行时库
func (m *DependencyManager) UninstallRuntime(ctx context.Context, dep, appName string) error {
	// 加载运行时索引
	if err := m.runtimeIndex.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 减少引用计数
	if err := m.runtimeIndex.Remove(dep, appName); err != nil {
		return err
	}

	return nil
}

// GetRuntimeInfo 获取运行时库信息
func (m *DependencyManager) GetRuntimeInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error) {
	// 加载运行时索引
	if err := m.runtimeIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load runtime index: %w", err)
	}

	info, ok := m.runtimeIndex.Get(dep)
	if !ok {
		return nil, errors.Newf(errors.KindNotFound, "runtime %s not found", dep)
	}

	return info, nil
}

// GetDependents 获取反向依赖（谁依赖我）
func (m *DependencyManager) GetDependents(ctx context.Context, appName string) ([]string, error) {
	// 加载依赖索引
	if err := m.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	dependents := m.depsIndex.GetDependents(appName)
	return dependents, nil
}

// FindOrphans 查找孤儿依赖
func (m *DependencyManager) FindOrphans(ctx context.Context) (*manifest.Orphans, error) {
	// 加载运行时索引
	if err := m.runtimeIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 加载依赖索引
	if err := m.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	// 查找孤儿运行时库
	runtimeOrphans := m.runtimeIndex.FindOrphans()

	// 查找孤儿工具软件
	toolOrphans := m.depsIndex.FindOrphans()

	return &manifest.Orphans{
		Runtime: runtimeOrphans,
		Tools:   toolOrphans,
	}, nil
}

// CleanupOrphans 清理孤儿依赖
func (m *DependencyManager) CleanupOrphans(ctx context.Context, orphans *manifest.Orphans) error {
	// 清理孤儿运行时库
	if len(orphans.Runtime) > 0 {
		fmt.Printf("清理 %d 个孤儿运行时库：\n", len(orphans.Runtime))
		for _, runtime := range orphans.Runtime {
			if err := m.runtimeIndex.Remove(runtime, ""); err != nil {
				fmt.Printf("  ✗ 清理 %s 失败: %v\n", runtime, err)
				continue
			}
			fmt.Printf("  ✓ 已清理: %s\n", runtime)
		}
		fmt.Println()
	}

	// 清理孤儿工具软件
	if len(orphans.Tools) > 0 {
		fmt.Printf("清理 %d 个孤儿工具软件：\n", len(orphans.Tools))
		for _, tool := range orphans.Tools {
			if err := m.depsIndex.Remove(tool); err != nil {
				fmt.Printf("  ✗ 清理 %s 失败: %v\n", tool, err)
				continue
			}
			fmt.Printf("  ✓ 已清理: %s\n", tool)
		}
		fmt.Println()
	}

	fmt.Println("孤儿依赖清理完成")
	return nil
}

// CleanupRuntime 清理无用运行时库
func (m *DependencyManager) CleanupRuntime(ctx context.Context) error {
	// 加载运行时索引
	if err := m.runtimeIndex.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 获取所有孤儿运行时库
	orphans := m.runtimeIndex.FindOrphans()

	if len(orphans) == 0 {
		fmt.Println("没有需要清理的运行时库")
		return nil
	}

	fmt.Printf("找到 %d 个孤儿运行时库：\n", len(orphans))
	for _, orphan := range orphans {
		fmt.Printf("  - %s\n", orphan)
	}
	fmt.Println()

	// 清理孤儿运行时库
	for _, orphan := range orphans {
		if err := m.runtimeIndex.Remove(orphan, ""); err != nil {
			fmt.Printf("清理 %s 失败: %v\n", orphan, err)
			continue
		}
		fmt.Printf("✓ 已清理: %s\n", orphan)
	}

	fmt.Println("\n运行时库清理完成")
	return nil
}

// isAppInstalled 检查应用是否已安装
func (m *DependencyManager) isAppInstalled(ctx context.Context, name string) bool {
	// TODO: 实现检查逻辑
	return false
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
