// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/engine/semver"
	"chopsticks/pkg/errors"
)

// Resolver 依赖解析器接口
type Resolver interface {
	// Resolve 解析应用的依赖树
	Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error)
	// CheckCircular 检测循环依赖
	CheckCircular(ctx context.Context, deps []string) error
	// TopologicalSort 拓扑排序
	TopologicalSort(graph *DependencyGraph) error
}

// resolver 依赖解析器实现
type resolver struct {
	bucketMgr bucket.BucketManager
	storage   store.LegacyStorage
}

// NewResolver 创建依赖解析器
func NewResolver(bucketMgr bucket.BucketManager, storage store.LegacyStorage) Resolver {
	return &resolver{
		bucketMgr: bucketMgr,
		storage:   storage,
	}
}

// Resolve 解析应用的依赖树
func (r *resolver) Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error) {
	if app == nil || app.Script == nil {
		return nil, errors.Newf(errors.KindInvalidInput, "invalid app")
	}

	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}

	// 创建根节点
	root := &DependencyNode{
		App:     app,
		Version: app.Meta.Version,
		Depth:   0,
	}
	graph.Nodes[app.Script.Name] = root

	// 提取依赖
	deps := r.extractDependencies(app)

	// 递归解析依赖
	visited := make(map[string]bool)
	resolving := make(map[string]bool) // 用于检测循环依赖
	if err := r.resolveDependencies(ctx, root, deps, graph, visited, resolving); err != nil {
		return nil, err
	}

	// 拓扑排序
	if err := r.TopologicalSort(graph); err != nil {
		return nil, err
	}

	return graph, nil
}

// extractDependencies 从应用清单中提取依赖
func (r *resolver) extractDependencies(app *manifest.App) []manifest.Dependency {
	var deps []manifest.Dependency

	// 从脚本元数据中提取依赖
	if app.Script != nil && len(app.Script.Dependencies) > 0 {
		deps = append(deps, app.Script.Dependencies...)
	}

	return deps
}

// resolveDependencies 递归解析依赖
func (r *resolver) resolveDependencies(
	ctx context.Context,
	parent *DependencyNode,
	deps []manifest.Dependency,
	graph *DependencyGraph,
	visited map[string]bool,
	resolving map[string]bool,
) error {
	for _, dep := range deps {
		// 检测循环依赖
		if resolving[dep.Name] {
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("circular dependency detected: %s", r.buildDependencyChain(parent, dep.Name)),
			)
		}

		// 检查条件是否满足
		if !r.checkConditions(dep.Conditions) {
			if dep.Optional {
				continue // 可选依赖，条件不满足则跳过
			}
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("conditions not satisfied for dependency %s", dep.Name),
			)
		}

		// 查找依赖应用
		depApp, err := r.findApp(ctx, dep.Name)
		if err != nil {
			if dep.Optional {
				continue // 可选依赖，未找到则跳过
			}
			return errors.Wrapf(err, "dependency not found: %s", dep.Name)
		}

		// 检查版本约束
		if dep.Version != "" {
			if err := r.checkVersionConstraint(depApp.Meta.Version, dep.Version); err != nil {
				if dep.Optional {
					continue
				}
				return errors.NewDependencyConflict(
					dep.Name,
					fmt.Sprintf("version mismatch: required %s, actual %s", dep.Version, depApp.Meta.Version),
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
		resolving[dep.Name] = true
		childDeps := r.extractDependencies(depApp)
		if err := r.resolveDependencies(ctx, node, childDeps, graph, visited, resolving); err != nil {
			return err
		}
		resolving[dep.Name] = false
	}

	return nil
}

// findApp 查找应用
func (r *resolver) findApp(ctx context.Context, name string) (*manifest.App, error) {
	// 首先检查已安装应用
	installed, err := r.storage.GetInstalledApp(ctx, name)
	if err == nil && installed != nil {
		// 已安装，返回已安装版本信息
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

	// 在软件源中搜索
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
func (r *resolver) checkConditions(conditions map[string]string) bool {
	if len(conditions) == 0 {
		return true
	}

	// TODO: 实现条件检查逻辑
	// 例如：{"os": "windows"}, {"arch": "amd64"} 等

	return true
}

// checkVersionConstraint 检查版本约束
func (r *resolver) checkVersionConstraint(version, constraint string) error {
	if constraint == "" || constraint == "*" {
		return nil
	}

	constraint = strings.TrimSpace(constraint)

	// 使用 semver 库进行版本比较
	matched, err := semver.Satisfies(version, constraint)
	if err != nil {
		return fmt.Errorf("version constraint not satisfied: %s %s: %w", version, constraint, err)
	}

	if !matched {
		return fmt.Errorf("version constraint not satisfied: %s %s", version, constraint)
	}

	return nil
}

// resolveVersionConflict 解决版本冲突
func (r *resolver) resolveVersionConflict(existing, new *DependencyNode) error {
	// 简单策略：选择较新的版本
	// TODO: 实现更复杂的版本冲突解决策略

	// 如果已安装现有版本，优先保留
	if r.isAppInstalled(existing.App.Script.Name) {
		return nil
	}

	// 否则使用 semver 比较选择较高版本
	cmp, err := semver.CompareStrings(new.Version, existing.Version)
	if err != nil {
		// 比较失败，保留现有版本
		return nil
	}

	if cmp > 0 {
		// 更新为新版本
		existing.App = new.App
		existing.Version = new.Version
	}

	return nil
}

// isAppInstalled 检查应用是否已安装
func (r *resolver) isAppInstalled(name string) bool {
	_, err := r.storage.GetInstalledApp(context.Background(), name)
	return err == nil
}

// TopologicalSort 拓扑排序（使用 Kahn 算法）
func (r *resolver) TopologicalSort(graph *DependencyGraph) error {
	inDegree := make(map[string]int)

	// 计算入度
	for name := range graph.Nodes {
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
	}

	for _, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			depName := dep.App.Script.Name
			inDegree[depName]++
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
		// 按名称排序以保证确定性
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

	// 检查是否有循环依赖
	if len(result) != len(graph.Nodes) {
		return errors.Newf(errors.KindInvalidInput, "circular dependency detected in dependency graph")
	}

	// 反转结果以获得正确的安装顺序（依赖优先）
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	graph.Order = result
	return nil
}

// CheckCircular 检测循环依赖
func (r *resolver) CheckCircular(ctx context.Context, deps []string) error {
	// 构建依赖图
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}

	// 加载所有依赖
	for _, depName := range deps {
		depApp, err := r.findApp(ctx, depName)
		if err != nil {
			return errors.Wrapf(err, "dependency not found: %s", depName)
		}

		graph.Nodes[depName] = &DependencyNode{
			App:     depApp,
			Version: depApp.Meta.Version,
			Depth:   0,
		}
	}

	// 解析依赖关系
	resolving := make(map[string]bool)
	visited := make(map[string]bool)

	for name, node := range graph.Nodes {
		if visited[name] {
			continue
		}

		deps := r.extractDependencies(node.App)
		resolving[name] = true
		if err := r.resolveDependencies(ctx, node, deps, graph, visited, resolving); err != nil {
			return err
		}
		resolving[name] = false
		visited[name] = true
	}

	return nil
}

// buildDependencyChain 构建依赖链字符串
func (r *resolver) buildDependencyChain(node *DependencyNode, target string) string {
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
