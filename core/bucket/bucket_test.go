package bucket

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	adapter := store.NewStorageAdapter(storage, tmpDir)
	mgr := NewManager(adapter, nil, tmpDir, nil)
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestAddOptions(t *testing.T) {
	opts := AddOptions{
		Branch: "main",
		Depth:  1,
	}

	if opts.Branch != "main" {
		t.Error("Branch mismatch")
	}

	if opts.Depth != 1 {
		t.Error("Depth mismatch")
	}
}

func TestSearchOptions(t *testing.T) {
	opts := SearchOptions{
		Bucket:   "main",
		Category: "dev",
		Tags:     []string{"cli", "tool"},
	}

	if opts.Bucket != "main" {
		t.Error("Bucket mismatch")
	}

	if opts.Category != "dev" {
		t.Error("Category mismatch")
	}

	if len(opts.Tags) != 2 {
		t.Error("Tags length mismatch")
	}
}

func TestSearchResult(t *testing.T) {
	result := SearchResult{
		Bucket: "main",
		App: &manifest.AppRef{
			Name: "test-app",
		},
	}

	if result.Bucket != "main" {
		t.Error("Bucket mismatch")
	}

	if result.App == nil {
		t.Fatal("App should not be nil")
	}

	if result.App.Name != "test-app" {
		t.Error("App.Name mismatch")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"hello", "hello world", false},
		{"", "", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello world"},
		{"HELLO", "hello"},
		{"hello", "hello"},
		{"", ""},
		{"ABC123", "abc123"},
	}

	for _, tt := range tests {
		result := lower(tt.input)
		if result != tt.expected {
			t.Errorf("lower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMatchesSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		app      *manifest.AppRef
		query    string
		opts     SearchOptions
		expected bool
	}{
		{
			name: "match by name",
			app: &manifest.AppRef{
				Name:        "test-app",
				Description: "A test application",
			},
			query:    "test",
			opts:     SearchOptions{},
			expected: true,
		},
		{
			name: "match by description",
			app: &manifest.AppRef{
				Name:        "myapp",
				Description: "A testing tool",
			},
			query:    "testing",
			opts:     SearchOptions{},
			expected: true,
		},
		{
			name: "no match",
			app: &manifest.AppRef{
				Name:        "myapp",
				Description: "A tool",
			},
			query:    "xyz",
			opts:     SearchOptions{},
			expected: false,
		},
		{
			name: "match with category filter",
			app: &manifest.AppRef{
				Name:     "test-app",
				Category: "dev",
			},
			query: "test",
			opts: SearchOptions{
				Category: "dev",
			},
			expected: true,
		},
		{
			name: "no match with wrong category",
			app: &manifest.AppRef{
				Name:     "test-app",
				Category: "dev",
			},
			query: "test",
			opts: SearchOptions{
				Category: "games",
			},
			expected: false,
		},
		{
			name: "match with tags filter",
			app: &manifest.AppRef{
				Name: "test-app",
				Tags: []string{"cli", "tool"},
			},
			query: "test",
			opts: SearchOptions{
				Tags: []string{"cli"},
			},
			expected: true,
		},
		{
			name: "no match with wrong tags",
			app: &manifest.AppRef{
				Name: "test-app",
				Tags: []string{"cli", "tool"},
			},
			query: "test",
			opts: SearchOptions{
				Tags: []string{"gui"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSearchQuery(tt.app, tt.query, tt.opts)
			if result != tt.expected {
				t.Errorf("matchesSearchQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestManagerListBuckets(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	adapter := store.NewStorageAdapter(storage, tmpDir)
	mgr := NewManager(adapter, nil, tmpDir, nil)

	ctx := context.Background()
	buckets, err := mgr.ListBuckets(ctx)
	if err != nil {
		t.Errorf("ListBuckets() failed: %v", err)
	}

	if buckets == nil {
		t.Error("ListBuckets() returned nil")
	}

	if len(buckets) == 0 {
		t.Error("ListBuckets() should return at least 'main'")
	}
}

func TestManagerGetBucketNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	adapter := store.NewStorageAdapter(storage, tmpDir)
	mgr := NewManager(adapter, nil, tmpDir, nil)

	ctx := context.Background()
	_, err = mgr.GetBucket(ctx, "nonexistent")
	if err == nil {
		t.Error("GetBucket() should return error for nonexistent bucket")
	}
}

func TestManagerGetAppBucketNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	adapter := store.NewStorageAdapter(storage, tmpDir)
	mgr := NewManager(adapter, nil, tmpDir, nil)

	ctx := context.Background()
	_, err = mgr.GetApp(ctx, "nonexistent", "app")
	if err == nil {
		t.Error("GetApp() should return error for nonexistent bucket")
	}
}

func TestManagerListAppsBucketNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "data")
	storage, err := store.NewFSStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	adapter := store.NewStorageAdapter(storage, tmpDir)
	mgr := NewManager(adapter, nil, tmpDir, nil)

	ctx := context.Background()
	_, err = mgr.ListApps(ctx, "nonexistent")
	if err == nil {
		t.Error("ListApps() should return error for nonexistent bucket")
	}
}

func TestExtractBucketName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"https://example.com/path/to/bucket.git", "bucket"},
		{"", "custom"},
	}

	for _, tt := range tests {
		result := extractBucketName(tt.url)
		if result != tt.expected {
			t.Errorf("extractBucketName(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestLoaderNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
}

func TestLoaderScanAppsEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader()

	ctx := context.Background()
	apps, err := loader.ScanApps(ctx, tmpDir)
	if err != nil {
		t.Errorf("ScanApps() failed: %v", err)
	}

	if apps == nil {
		t.Error("ScanApps() returned nil")
	}

	if len(apps) != 0 {
		t.Errorf("ScanApps() returned %d apps, want 0", len(apps))
	}
}

func TestLoaderScanAppsWithScripts(t *testing.T) {
	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create a test JS script
	jsScript := filepath.Join(appsDir, "test-app.js")
	if err := os.WriteFile(jsScript, []byte("// Test script"), 0644); err != nil {
		t.Fatalf("Failed to create JS script: %v", err)
	}

	// Create a test Lua script
	luaScript := filepath.Join(appsDir, "test-app2.lua")
	if err := os.WriteFile(luaScript, []byte("-- Test script"), 0644); err != nil {
		t.Fatalf("Failed to create Lua script: %v", err)
	}

	// Create a non-script file (should be ignored)
	txtFile := filepath.Join(appsDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	loader := NewLoader()

	ctx := context.Background()
	apps, err := loader.ScanApps(ctx, tmpDir)
	if err != nil {
		t.Errorf("ScanApps() failed: %v", err)
	}

	if apps == nil {
		t.Fatal("ScanApps() returned nil")
	}

	if len(apps) != 2 {
		t.Errorf("ScanApps() returned %d apps, want 2", len(apps))
	}

	if _, ok := apps["test-app"]; !ok {
		t.Error("test-app not found in apps")
	}

	if _, ok := apps["test-app2"]; !ok {
		t.Error("test-app2 not found in apps")
	}
}

func TestExtractFromJS(t *testing.T) {
	content := `
/**
 * @description A test application
 * @version 1.0.0
 * category: "dev"
 */
function install() {}
`

	info := extractFromJS(content)
	if info == nil {
		t.Fatal("extractFromJS() returned nil")
	}

	if info.Description != "A test application" {
		t.Errorf("Description = %q, want %q", info.Description, "A test application")
	}

	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}

	if info.Category != "dev" {
		t.Errorf("Category = %q, want %q", info.Category, "dev")
	}
}

func TestExtractFromLua(t *testing.T) {
	content := `
-- description: A test application
-- version: 1.0.0
-- category: dev
function install() end
`

	info := extractFromLua(content)
	if info == nil {
		t.Fatal("extractFromLua() returned nil")
	}

	if info.Description != "A test application" {
		t.Errorf("Description = %q, want %q", info.Description, "A test application")
	}

	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}

	if info.Category != "dev" {
		t.Errorf("Category = %q, want %q", info.Category, "dev")
	}
}

func TestScriptInfo(t *testing.T) {
	info := &scriptInfo{
		Description: "Test app",
		Version:     "1.0.0",
		Category:    "dev",
		Tags:        []string{"cli"},
	}

	if info.Description != "Test app" {
		t.Error("Description mismatch")
	}

	if info.Version != "1.0.0" {
		t.Error("Version mismatch")
	}

	if info.Category != "dev" {
		t.Error("Category mismatch")
	}

	if len(info.Tags) != 1 || info.Tags[0] != "cli" {
		t.Error("Tags mismatch")
	}
}

func TestNewUpdater(t *testing.T) {
	loader := NewLoader()
	updater := NewUpdater(loader)
	if updater == nil {
		t.Fatal("NewUpdater() returned nil")
	}
}

func TestUpdaterLoadAppRef(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.js")
	if err := os.WriteFile(scriptPath, []byte("// test"), 0644); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	loader := NewLoader()
	upd := NewUpdater(loader)

	// Test that updater was created successfully
	if upd == nil {
		t.Fatal("NewUpdater() returned nil")
	}
}
