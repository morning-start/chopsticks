// Package app 提供应用管理功能。
package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

// Dependency 表示应用依赖
type Dependency struct {
	Name       string            // 依赖应用名称
	Version    string            // 版本约束（如 ">=1.0.0", "^2.0.0"）
	Optional   bool              // 是否为可选依赖
	Conditions map[string]string // 安装条件（如 {"os": "windows"}）
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

// DependencyResolver 依赖解析器
type DependencyResolver struct {
	bucketMgr bucket.Manager
	storage   store.Storage
}

// NewDependencyResolver 创建依赖解析器
func NewDependencyResolver(bucketMgr bucket.Manager, storage store.Storage) *DependencyResolver {
	return &DependencyResolver{
		bucketMgr: bucketMgr,
		storage:   storage,
	}
}

// Resolve 解析应用的依赖树
func (r *DependencyResolver) Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error) {
	if app == nil || app.Script == nil {
		return nil, errors.Newf(errors.KindInvalidInput, "invalid app")
	}

	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}

	// 从应用清单中获取依赖
	deps := r.extractDependencies(app)

	// 构建依赖树
	root := &DependencyNode{
		App:     app,
		Version: app.Meta.Version,
		Depth:   0,
	}
	graph.Nodes[app.Script.Name] = root

	// 递归解析依赖
	if err := r.resolveDependencies(ctx, root, deps, graph, make(map[string]bool)); err != nil {
		return nil, err
	}

	// 拓扑排序
	if err := r.topologicalSort(graph); err != nil {
		return nil, err
	}

	return graph, nil
}

// extractDependencies 从应用清单中提取依赖
func (r *DependencyResolver) extractDependencies(app *manifest.App) []Dependency {
	var deps []Dependency

	// 从脚本元数据中提取依赖
	if app.Script != nil && len(app.Script.Dependencies) > 0 {
		for _, dep := range app.Script.Dependencies {
			deps = append(deps, Dependency{
				Name:       dep.Name,
				Version:    dep.Version,
				Optional:   dep.Optional,
				Conditions: dep.Conditions,
			})
		}
	}

	return deps
}

// resolveDependencies 递归解析依赖
func (r *DependencyResolver) resolveDependencies(
	ctx context.Context,
	parent *DependencyNode,
	deps []Dependency,
	graph *DependencyGraph,
	visited map[string]bool,
) error {
	for _, dep := range deps {
		// 检查循环依赖 - 检查父链中是否已存在
		if r.isInParentChain(parent, dep.Name) {
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("检测到循环依赖: %s", r.buildDependencyChain(parent, dep.Name)),
			)
		}

		// 检查条件是否满足
		if !r.checkConditions(dep.Conditions) {
			if dep.Optional {
				continue // 可选依赖，条件不满足则跳过
			}
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("依赖 %s 的条件不满足", dep.Name),
			)
		}

		// 获取依赖应用
		depApp, err := r.findApp(ctx, dep.Name)
		if err != nil {
			if dep.Optional {
				continue // 可选依赖，找不到则跳过
			}
			return errors.Wrapf(err, "找不到依赖: %s", dep.Name)
		}

		// 检查版本约束
		if dep.Version != "" {
			if err := r.checkVersionConstraint(depApp.Meta.Version, dep.Version); err != nil {
				if dep.Optional {
					continue
				}
				return errors.NewDependencyConflict(
					dep.Name,
					fmt.Sprintf("版本不匹配: 需要 %s, 实际 %s", dep.Version, depApp.Meta.Version),
				)
			}
		}

		// 创建依赖节点
		node := &DependencyNode{
			App:     depApp,
			Version: depApp.Meta.Version,
			Parent:  parent,
			Depth:   parent.Depth + 1,
		}

		// 检查是否已存在
		if existing, ok := graph.Nodes[dep.Name]; ok {
			// 如果已存在，检查版本是否兼容
			if existing.Version != depApp.Meta.Version {
				// 版本冲突，尝试解决
				if err := r.resolveVersionConflict(existing, node); err != nil {
					return err
				}
			}
			parent.Dependencies = append(parent.Dependencies, existing)
			continue
		}

		graph.Nodes[dep.Name] = node
		parent.Dependencies = append(parent.Dependencies, node)

		// 递归解析子依赖
		childDeps := r.extractDependencies(depApp)
		visited[dep.Name] = true
		if err := r.resolveDependencies(ctx, node, childDeps, graph, visited); err != nil {
			return err
		}
		delete(visited, dep.Name)
	}

	return nil
}

// findApp 查找应用
func (r *DependencyResolver) findApp(ctx context.Context, name string) (*manifest.App, error) {
	// 先检查已安装的软件
	installed, err := r.storage.GetInstalledApp(ctx, name)
	if err == nil && installed != nil {
		// 已安装，返回已安装版本的信息
		return &manifest.App{
			Script: &manifest.AppScript{
				Name:   installed.Name,
				Bucket: installed.Bucket,
			},
			Meta: &manifest.AppMeta{
				Version: installed.Version,
			},
		}, nil
	}

	// 从软件源中查找
	buckets, err := r.bucketMgr.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	for _, bucketName := range buckets {
		app, err := r.bucketMgr.GetApp(ctx, bucketName, name)
		if err == nil && app != nil {
			return app, nil
		}
	}

	return nil, errors.NewAppNotFound(name)
}

// checkConditions 检查依赖条件是否满足
func (r *DependencyResolver) checkConditions(conditions map[string]string) bool {
	if len(conditions) == 0 {
		return true
	}

	// TODO: 实现条件检查逻辑
	// 例如: {"os": "windows"}, {"arch": "amd64"} 等

	return true
}

// checkVersionConstraint 检查版本约束
func (r *DependencyResolver) checkVersionConstraint(version, constraint string) error {
	if constraint == "" || constraint == "*" {
		return nil
	}

	// 解析版本约束
	// 支持: >=1.0.0, ^1.0.0, ~1.0.0, 1.0.0, >1.0.0, <1.0.0 等

	constraint = strings.TrimSpace(constraint)

	// 精确版本匹配
	if !strings.ContainsAny(constraint, ">=<^~") {
		if version == constraint {
			return nil
		}
		return fmt.Errorf("版本不匹配: 需要 %s, 实际 %s", constraint, version)
	}

	// 简单实现：检查前缀匹配
	// >=1.0.0 表示 1.0.0 及以上版本
	if strings.HasPrefix(constraint, ">=") {
		requiredVersion := strings.TrimPrefix(constraint, ">=")
		if version >= requiredVersion {
			return nil
		}
		return fmt.Errorf("版本过低: 需要 >= %s, 实际 %s", requiredVersion, version)
	}

	// TODO: 实现更复杂的版本约束解析
	// 例如: semver 版本比较, ^, ~ 等

	return nil
}

// resolveVersionConflict 解决版本冲突
func (r *DependencyResolver) resolveVersionConflict(existing, new *DependencyNode) error {
	// 简单策略：选择较新的版本
	// TODO: 实现更复杂的版本冲突解决策略

	// 如果现有版本是已安装的，优先保留
	if r.isAppInstalled(existing.App.Script.Name) {
		return nil
	}

	// 否则选择版本号较大的
	if new.Version > existing.Version {
		// 更新为新版本
		existing.App = new.App
		existing.Version = new.Version
	}

	return nil
}

// isAppInstalled 检查应用是否已安装
func (r *DependencyResolver) isAppInstalled(name string) bool {
	_, err := r.storage.GetInstalledApp(context.Background(), name)
	return err == nil
}

// topologicalSort 拓扑排序
func (r *DependencyResolver) topologicalSort(graph *DependencyGraph) error {
	// 使用 Kahn 算法进行拓扑排序
	inDegree := make(map[string]int)

	// 计算入度
	for name, node := range graph.Nodes {
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
		for _, dep := range node.Dependencies {
			inDegree[dep.App.Script.Name]++
		}
	}

	// 找到所有入度为 0 的节点
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 排序
	result := make([]string, 0, len(graph.Nodes))
	for len(queue) > 0 {
		// 按名称排序，确保确定性
		sort.Strings(queue)
		name := queue[0]
		queue = queue[1:]

		result = append(result, name)

		node := graph.Nodes[name]
		for _, dep := range node.Dependencies {
			depName := dep.App.Script.Name
			inDegree[depName]--
			if inDegree[depName] == 0 {
				queue = append(queue, depName)
			}
		}
	}

	// 检查是否有环
	if len(result) != len(graph.Nodes) {
		return errors.Newf(errors.KindInvalidInput, "依赖图中存在循环依赖")
	}

	// 反转结果，得到正确的安装顺序（依赖在前）
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	graph.Order = result
	return nil
}

// buildDependencyChain 构建依赖链字符串
func (r *DependencyResolver) buildDependencyChain(node *DependencyNode, target string) string {
	var chain []string
	current := node

	for current != nil {
		chain = append([]string{current.App.Script.Name}, chain...)
		if current.App.Script.Name == target {
			break
		}
		current = current.Parent
	}

	chain = append(chain, target)
	return strings.Join(chain, " -> ")
}

// isInParentChain 检查目标是否在父链中
func (r *DependencyResolver) isInParentChain(node *DependencyNode, target string) bool {
	current := node
	for current != nil {
		if current.App.Script.Name == target {
			return true
		}
		current = current.Parent
	}
	return false
}

// GetInstallOrder 获取安装顺序
func (g *DependencyGraph) GetInstallOrder() []string {
	return g.Order
}

// GetDependencies 获取应用的所有依赖
func (g *DependencyGraph) GetDependencies(appName string) []*DependencyNode {
	node, ok := g.Nodes[appName]
	if !ok {
		return nil
	}
	return node.Dependencies
}

// HasDependency 检查应用是否有依赖
func (g *DependencyGraph) HasDependency(appName string) bool {
	node, ok := g.Nodes[appName]
	if !ok {
		return false
	}
	return len(node.Dependencies) > 0
}
