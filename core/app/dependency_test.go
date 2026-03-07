// Package app 提供应用管理功能。
package app

import (
	"context"
	"testing"

	"chopsticks/core/bucket"
	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

// mockBucketManager 模拟 Bucket Manager
type mockBucketManager struct {
	apps map[string]*manifest.App
}

func newMockBucketManager() *mockBucketManager {
	return &mockBucketManager{
		apps: make(map[string]*manifest.App),
	}
}

func (m *mockBucketManager) Add(ctx context.Context, name, url string, opts bucket.AddOptions) error {
	return nil
}

func (m *mockBucketManager) Remove(ctx context.Context, name string, purge bool) error {
	return nil
}

func (m *mockBucketManager) Update(ctx context.Context, name string) error {
	return nil
}

func (m *mockBucketManager) UpdateAll(ctx context.Context) error {
	return nil
}

func (m *mockBucketManager) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	return nil, nil
}

func (m *mockBucketManager) GetApp(ctx context.Context, bucket, name string) (*manifest.App, error) {
	if app, ok := m.apps[name]; ok {
		return app, nil
	}
	return nil, nil
}

func (m *mockBucketManager) ListApps(ctx context.Context, bucket string) (map[string]*manifest.AppRef, error) {
	return nil, nil
}

func (m *mockBucketManager) ListBuckets(ctx context.Context) ([]string, error) {
	return []string{"main"}, nil
}

func (m *mockBucketManager) Search(ctx context.Context, query string, opts bucket.SearchOptions) ([]bucket.SearchResult, error) {
	return nil, nil
}

func (m *mockBucketManager) AddApp(name string, app *manifest.App) {
	m.apps[name] = app
}

// mockStorage 模拟 Storage
type mockStorage struct {
	installed map[string]*manifest.InstalledApp
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		installed: make(map[string]*manifest.InstalledApp),
	}
}

func (s *mockStorage) SaveInstalledApp(ctx context.Context, a *manifest.InstalledApp) error {
	s.installed[a.Name] = a
	return nil
}

func (s *mockStorage) GetInstalledApp(ctx context.Context, name string) (*manifest.InstalledApp, error) {
	if app, ok := s.installed[name]; ok {
		return app, nil
	}
	return nil, nil
}

func (s *mockStorage) DeleteInstalledApp(ctx context.Context, name string) error {
	delete(s.installed, name)
	return nil
}

func (s *mockStorage) ListInstalledApps(ctx context.Context) ([]*manifest.InstalledApp, error) {
	var apps []*manifest.InstalledApp
	for _, app := range s.installed {
		apps = append(apps, app)
	}
	return apps, nil
}

func (s *mockStorage) IsInstalled(ctx context.Context, name string) (bool, error) {
	_, ok := s.installed[name]
	return ok, nil
}

func (s *mockStorage) SaveBucket(ctx context.Context, b *manifest.BucketConfig) error {
	return nil
}

func (s *mockStorage) GetBucket(ctx context.Context, name string) (*manifest.BucketConfig, error) {
	return nil, nil
}

func (s *mockStorage) DeleteBucket(ctx context.Context, name string) error {
	return nil
}

func (s *mockStorage) ListBuckets(ctx context.Context) ([]*manifest.BucketConfig, error) {
	return nil, nil
}

func (s *mockStorage) SaveOperation(ctx context.Context, appName string, op *store.Operation) error {
	return nil
}

func (s *mockStorage) GetOperations(ctx context.Context, appName string) ([]store.Operation, error) {
	return nil, nil
}

func (s *mockStorage) DeleteOperations(ctx context.Context, appName string) error {
	return nil
}

func (s *mockStorage) Close() error {
	return nil
}

func TestDependencyResolver_Resolve_NoDependencies(t *testing.T) {
	bucketMgr := newMockBucketManager()
	storage := newMockStorage()
	resolver := NewDependencyResolver(bucketMgr, storage)

	app := &manifest.App{
		Script: &manifest.AppScript{
			Name: "test-app",
		},
		Meta: &manifest.AppMeta{
			Version: "1.0.0",
		},
	}

	graph, err := resolver.Resolve(context.Background(), app)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if len(graph.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(graph.Nodes))
	}

	if len(graph.Order) != 1 {
		t.Errorf("Expected order length 1, got %d", len(graph.Order))
	}
}

func TestDependencyResolver_Resolve_WithDependencies(t *testing.T) {
	bucketMgr := newMockBucketManager()
	storage := newMockStorage()

	// 添加依赖应用
	depApp := &manifest.App{
		Script: &manifest.AppScript{
			Name: "dep-app",
		},
		Meta: &manifest.AppMeta{
			Version: "1.0.0",
		},
	}
	bucketMgr.AddApp("dep-app", depApp)

	// 添加主应用（带依赖）
	mainApp := &manifest.App{
		Script: &manifest.AppScript{
			Name: "main-app",
			Dependencies: []manifest.Dependency{
				{
					Name:    "dep-app",
					Version: ">=1.0.0",
				},
			},
		},
		Meta: &manifest.AppMeta{
			Version: "1.0.0",
		},
	}

	resolver := NewDependencyResolver(bucketMgr, storage)
	graph, err := resolver.Resolve(context.Background(), mainApp)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Order) != 2 {
		t.Errorf("Expected order length 2, got %d", len(graph.Order))
	}

	// 检查安装顺序（依赖应该在前）
	if graph.Order[0] != "dep-app" || graph.Order[1] != "main-app" {
		t.Errorf("Expected order [dep-app, main-app], got %v", graph.Order)
	}
}

func TestDependencyResolver_Resolve_CircularDependency(t *testing.T) {
	bucketMgr := newMockBucketManager()
	storage := newMockStorage()

	// 创建循环依赖: A -> B -> A
	appA := &manifest.App{
		Script: &manifest.AppScript{
			Name: "app-a",
			Dependencies: []manifest.Dependency{
				{Name: "app-b"},
			},
		},
		Meta: &manifest.AppMeta{Version: "1.0.0"},
	}

	appB := &manifest.App{
		Script: &manifest.AppScript{
			Name: "app-b",
			Dependencies: []manifest.Dependency{
				{Name: "app-a"},
			},
		},
		Meta: &manifest.AppMeta{Version: "1.0.0"},
	}

	bucketMgr.AddApp("app-a", appA)
	bucketMgr.AddApp("app-b", appB)

	resolver := NewDependencyResolver(bucketMgr, storage)
	_, err := resolver.Resolve(context.Background(), appA)
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}
}

func TestDependencyResolver_Resolve_VersionConstraint(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		wantErr    bool
	}{
		{
			name:       "exact match",
			version:    "1.0.0",
			constraint: "1.0.0",
			wantErr:    false,
		},
		{
			name:       "exact mismatch",
			version:    "1.0.0",
			constraint: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "wildcard",
			version:    "1.0.0",
			constraint: "*",
			wantErr:    false,
		},
		{
			name:       "empty constraint",
			version:    "1.0.0",
			constraint: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketMgr := newMockBucketManager()
			storage := newMockStorage()

			depApp := &manifest.App{
				Script: &manifest.AppScript{
					Name: "dep-app",
				},
				Meta: &manifest.AppMeta{
					Version: tt.version,
				},
			}
			bucketMgr.AddApp("dep-app", depApp)

			mainApp := &manifest.App{
				Script: &manifest.AppScript{
					Name: "main-app",
					Dependencies: []manifest.Dependency{
						{
							Name:    "dep-app",
							Version: tt.constraint,
						},
					},
				},
				Meta: &manifest.AppMeta{Version: "1.0.0"},
			}

			resolver := NewDependencyResolver(bucketMgr, storage)
			_, err := resolver.Resolve(context.Background(), mainApp)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDependencyGraph_GetInstallOrder(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"app-a": {},
			"app-b": {},
		},
		Order: []string{"app-b", "app-a"},
	}

	order := graph.GetInstallOrder()
	if len(order) != 2 {
		t.Errorf("Expected 2 items, got %d", len(order))
	}

	if order[0] != "app-b" || order[1] != "app-a" {
		t.Errorf("Expected [app-b, app-a], got %v", order)
	}
}

func TestDependencyGraph_HasDependency(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"app-a": {
				Dependencies: []*DependencyNode{
					{App: &manifest.App{Script: &manifest.AppScript{Name: "app-b"}}},
				},
			},
			"app-b": {
				Dependencies: []*DependencyNode{},
			},
		},
	}

	if !graph.HasDependency("app-a") {
		t.Error("Expected app-a to have dependencies")
	}

	if graph.HasDependency("app-b") {
		t.Error("Expected app-b to have no dependencies")
	}
}
