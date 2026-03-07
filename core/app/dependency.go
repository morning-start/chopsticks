// Package app 提供应用管理功能。
package app

import (
	"cmp"
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
	bucketMgr bucket.BucketManager
	storage   store.LegacyStorage
}

// NewDependencyResolver 创建依赖解析器
func NewDependencyResolver(bucketMgr bucket.BucketManager, storage store.LegacyStorage) *DependencyResolver {
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

	// Get dependencies from app manifest
	deps := r.extractDependencies(app)

	// Build dependency tree
	root := &DependencyNode{
		App:     app,
		Version: app.Meta.Version,
		Depth:   0,
	}
	graph.Nodes[app.Script.Name] = root

	// Recursively resolve dependencies
	if err := r.resolveDependencies(ctx, root, deps, graph, make(map[string]bool)); err != nil {
		return nil, err
	}

	// Topological sort
	if err := r.topologicalSort(graph); err != nil {
		return nil, err
	}

	return graph, nil
}

// extractDependencies 从应用清单中提取依赖
func (r *DependencyResolver) extractDependencies(app *manifest.App) []Dependency {
	var deps []Dependency

	// Extract dependencies from script metadata
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
		// Check for circular dependency - check if exists in parent chain
		if r.isInParentChain(parent, dep.Name) {
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("circular dependency detected: %s", r.buildDependencyChain(parent, dep.Name)),
			)
		}

		// Check if conditions are satisfied
		if !r.checkConditions(dep.Conditions) {
			if dep.Optional {
				continue // Optional dependency, skip if conditions not met
			}
			return errors.NewDependencyConflict(
				dep.Name,
				fmt.Sprintf("conditions not satisfied for dependency %s", dep.Name),
			)
		}

		// Get dependency app
		depApp, err := r.findApp(ctx, dep.Name)
		if err != nil {
			if dep.Optional {
				continue // Optional dependency, skip if not found
			}
			return errors.Wrapf(err, "find dependency app %q", dep.Name)
		}

		// Check version constraint
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

		// Create dependency node
		node := &DependencyNode{
			App:     depApp,
			Version: depApp.Meta.Version,
			Parent:  parent,
			Depth:   parent.Depth + 1,
		}

		// Check if already exists
		if existing, ok := graph.Nodes[dep.Name]; ok {
			// If exists, check if versions are compatible
			if existing.Version != depApp.Meta.Version {
				// Version conflict, try to resolve
				if err := r.resolveVersionConflict(ctx, existing, node); err != nil {
					return err
				}
			}
			parent.Dependencies = append(parent.Dependencies, existing)
			continue
		}

		graph.Nodes[dep.Name] = node
		parent.Dependencies = append(parent.Dependencies, node)

		// Recursively resolve child dependencies
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
	// First check installed apps
	installed, err := r.storage.GetInstalledApp(ctx, name)
	if err == nil && installed != nil {
		// Installed, return installed version info
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

	// Search in buckets
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

	// TODO: Implement condition checking logic
	// For example: {"os": "windows"}, {"arch": "amd64"}, etc.

	return true
}

// checkVersionConstraint 检查版本约束
func (r *DependencyResolver) checkVersionConstraint(version, constraint string) error {
	if constraint == "" || constraint == "*" {
		return nil
	}

	// Parse version constraint
	// Supports: >=1.0.0, ^1.0.0, ~1.0.0, 1.0.0, >1.0.0, <1.0.0, etc.

	constraint = strings.TrimSpace(constraint)

	// Exact version match
	if !strings.ContainsAny(constraint, ">=<^~") {
		if version == constraint {
			return nil
		}
		return fmt.Errorf("version mismatch: required %s, actual %s", constraint, version)
	}

	// Use cmp.Compare for version comparison
	// >=1.0.0 means 1.0.0 and above
	if strings.HasPrefix(constraint, ">=") {
		requiredVersion := strings.TrimPrefix(constraint, ">=")
		if compareVersions(version, requiredVersion) >= 0 {
			return nil
		}
		return fmt.Errorf("version too low: required >= %s, actual %s", requiredVersion, version)
	}

	// >1.0.0 means above 1.0.0
	if strings.HasPrefix(constraint, ">") {
		requiredVersion := strings.TrimPrefix(constraint, ">")
		if compareVersions(version, requiredVersion) > 0 {
			return nil
		}
		return fmt.Errorf("version too low: required > %s, actual %s", requiredVersion, version)
	}

	// <=1.0.0 means 1.0.0 and below
	if strings.HasPrefix(constraint, "<=") {
		requiredVersion := strings.TrimPrefix(constraint, "<=")
		if compareVersions(version, requiredVersion) <= 0 {
			return nil
		}
		return fmt.Errorf("version too high: required <= %s, actual %s", requiredVersion, version)
	}

	// <1.0.0 means below 1.0.0
	if strings.HasPrefix(constraint, "<") {
		requiredVersion := strings.TrimPrefix(constraint, "<")
		if compareVersions(version, requiredVersion) < 0 {
			return nil
		}
		return fmt.Errorf("version too high: required < %s, actual %s", requiredVersion, version)
	}

	// TODO: Implement more complex version constraint parsing
	// For example: semver comparison, ^, ~, etc.

	return nil
}

// compareVersions compares two version strings using cmp.Compare
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Simple string comparison for basic semantic versioning
	// For more complex semver, consider using a proper semver library
	return cmp.Compare(v1, v2)
}

// resolveVersionConflict 解决版本冲突
func (r *DependencyResolver) resolveVersionConflict(ctx context.Context, existing, new *DependencyNode) error {
	// Simple strategy: choose newer version
	// TODO: Implement more complex version conflict resolution strategy

	// If existing version is installed, prefer keeping it
	if r.isAppInstalled(ctx, existing.App.Script.Name) {
		return nil
	}

	// Otherwise choose the higher version number using cmp.Compare
	if compareVersions(new.Version, existing.Version) > 0 {
		// Update to new version
		existing.App = new.App
		existing.Version = new.Version
	}

	return nil
}

// isAppInstalled 检查应用是否已安装
func (r *DependencyResolver) isAppInstalled(ctx context.Context, name string) bool {
	_, err := r.storage.GetInstalledApp(ctx, name)
	return err == nil
}

// topologicalSort 拓扑排序
func (r *DependencyResolver) topologicalSort(graph *DependencyGraph) error {
	// Use Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)

	// Calculate in-degrees
	for name, node := range graph.Nodes {
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
		for _, dep := range node.Dependencies {
			inDegree[dep.App.Script.Name]++
		}
	}

	// Find all nodes with in-degree 0
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort
	result := make([]string, 0, len(graph.Nodes))
	for len(queue) > 0 {
		// Sort by name for determinism
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

	// Check for cycles
	if len(result) != len(graph.Nodes) {
		return errors.Newf(errors.KindInvalidInput, "circular dependency detected in dependency graph")
	}

	// Reverse result to get correct installation order (dependencies first)
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
