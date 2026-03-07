package dep

import (
	"context"
	"strings"
	"testing"

	"chopsticks/core/manifest"
)

// mockResolver 用于测试的模拟解析器
type mockResolver struct {
	apps map[string]*manifest.App
}

func (r *mockResolver) Resolve(ctx context.Context, app *manifest.App) (*DependencyGraph, error) {
	return nil, nil
}

func (r *mockResolver) CheckCircular(ctx context.Context, deps []string) error {
	return checkCircularWithApps(ctx, deps, r.apps)
}

func (r *mockResolver) TopologicalSort(graph *DependencyGraph) error {
	return nil
}

// checkCircularWithApps 使用 DFS 检测循环依赖
func checkCircularWithApps(ctx context.Context, deps []string, apps map[string]*manifest.App) error {
	if len(deps) == 0 {
		return nil
	}

	// 构建邻接表
	adjList := make(map[string][]string)

	for _, depName := range deps {
		app, exists := apps[depName]
		if !exists {
			continue
		}

		adjList[depName] = []string{}
		if app.Script != nil && len(app.Script.Dependencies) > 0 {
			for _, dep := range app.Script.Dependencies {
				adjList[depName] = append(adjList[depName], dep.Name)
			}
		}
	}

	// DFS 检测循环
	visited := make(map[string]bool)
	stack := make(map[string]bool)
	path := []string{}

	var dfs func(node string) error
	dfs = func(node string) error {
		visited[node] = true
		stack[node] = true
		path = append(path, node)

		for _, neighbor := range adjList[node] {
			if stack[neighbor] {
				// 找到循环
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := append(path[cycleStart:], neighbor)
					return &mockError{
						msg: "circular dependency detected: " + strings.Join(cycle, " -> "),
					}
				}
			}

			if !visited[neighbor] {
				if err := dfs(neighbor); err != nil {
					return err
				}
			}
		}

		path = path[:len(path)-1]
		delete(stack, node)
		return nil
	}

	for node := range adjList {
		if !visited[node] {
			path = []string{}
			if err := dfs(node); err != nil {
				return err
			}
		}
	}

	return nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func createMockApp(name, version string, deps []manifest.Dependency) *manifest.App {
	return &manifest.App{
		Script: &manifest.AppScript{
			Name:         name,
			Bucket:       "main",
			Dependencies: deps,
		},
		Meta: &manifest.AppMeta{
			Version: version,
		},
	}
}

func TestCheckCircular_NoCircular(t *testing.T) {
	ctx := context.Background()

	// 无循环依赖：A -> B -> C
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{{Name: "C", Version: "1.0.0"}}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckCircular_SimpleCircular(t *testing.T) {
	ctx := context.Background()

	// 简单循环：A -> B -> A
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{{Name: "A", Version: "1.0.0"}}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B"})
	if err == nil {
		t.Errorf("expected circular dependency error, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("expected 'circular dependency detected' in error message, got: %v", err)
	}

	if err != nil && !strings.Contains(err.Error(), "->") {
		t.Errorf("expected dependency chain in error message, got: %v", err)
	}
}

func TestCheckCircular_ComplexCircular(t *testing.T) {
	ctx := context.Background()

	// 复杂循环：A -> B -> C -> A
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{{Name: "C", Version: "1.0.0"}}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{{Name: "A", Version: "1.0.0"}}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C"})
	if err == nil {
		t.Errorf("expected circular dependency error, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("expected 'circular dependency detected' in error message, got: %v", err)
	}
}

func TestCheckCircular_EmptyDeps(t *testing.T) {
	ctx := context.Background()

	apps := map[string]*manifest.App{}
	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{})
	if err != nil {
		t.Errorf("expected no error for empty deps, got %v", err)
	}
}

func TestCheckCircular_SelfDependency(t *testing.T) {
	ctx := context.Background()

	// 自依赖：A -> A
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "A", Version: "1.0.0"}}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A"})
	if err == nil {
		t.Errorf("expected self-dependency error, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("expected 'circular dependency detected' in error message, got: %v", err)
	}
}

func TestCheckCircular_MultipleChainsNoCircular(t *testing.T) {
	ctx := context.Background()

	// 多个依赖链，无循环：A -> B, C -> D
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{{Name: "D", Version: "1.0.0"}}),
		"D": createMockApp("D", "1.0.0", []manifest.Dependency{}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C", "D"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckCircular_DisconnectedGraph(t *testing.T) {
	ctx := context.Background()

	// 不连通图：A -> B, C (独立)
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckCircular_LongChain(t *testing.T) {
	ctx := context.Background()

	// 长链：A -> B -> C -> D -> E
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{{Name: "C", Version: "1.0.0"}}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{{Name: "D", Version: "1.0.0"}}),
		"D": createMockApp("D", "1.0.0", []manifest.Dependency{{Name: "E", Version: "1.0.0"}}),
		"E": createMockApp("E", "1.0.0", []manifest.Dependency{}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C", "D", "E"})
	if err != nil {
		t.Errorf("expected no error for long chain, got %v", err)
	}
}

func TestCheckCircular_LongChainWithCycle(t *testing.T) {
	ctx := context.Background()

	// 长链带循环：A -> B -> C -> D -> E -> C
	apps := map[string]*manifest.App{
		"A": createMockApp("A", "1.0.0", []manifest.Dependency{{Name: "B", Version: "1.0.0"}}),
		"B": createMockApp("B", "1.0.0", []manifest.Dependency{{Name: "C", Version: "1.0.0"}}),
		"C": createMockApp("C", "1.0.0", []manifest.Dependency{{Name: "D", Version: "1.0.0"}}),
		"D": createMockApp("D", "1.0.0", []manifest.Dependency{{Name: "E", Version: "1.0.0"}}),
		"E": createMockApp("E", "1.0.0", []manifest.Dependency{{Name: "C", Version: "1.0.0"}}),
	}

	resolver := &mockResolver{apps: apps}

	err := resolver.CheckCircular(ctx, []string{"A", "B", "C", "D", "E"})
	if err == nil {
		t.Errorf("expected circular dependency error, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("expected 'circular dependency detected' in error message, got: %v", err)
	}
}
