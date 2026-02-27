package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.AppsPath == "" {
		t.Error("AppsPath should not be empty")
	}

	if cfg.BucketsPath == "" {
		t.Error("BucketsPath should not be empty")
	}

	if cfg.CachePath == "" {
		t.Error("CachePath should not be empty")
	}

	if cfg.StoragePath == "" {
		t.Error("StoragePath should not be empty")
	}
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		AppsPath:    filepath.Join(tmpDir, "apps"),
		BucketsPath: filepath.Join(tmpDir, "buckets"),
		CachePath:   filepath.Join(tmpDir, "cache"),
		StoragePath: filepath.Join(tmpDir, "data.db"),
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.Config() != cfg {
		t.Error("Config() should return the same config")
	}

	if app.BucketManager() == nil {
		t.Error("BucketManager() should not be nil")
	}

	if app.AppManager() == nil {
		t.Error("AppManager() should not be nil")
	}

	if app.Installer() == nil {
		t.Error("Installer() should not be nil")
	}

	if app.Storage() == nil {
		t.Error("Storage() should not be nil")
	}

	ctx := context.Background()
	if err := app.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestAppRun(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		AppsPath:    filepath.Join(tmpDir, "apps"),
		BucketsPath: filepath.Join(tmpDir, "buckets"),
		CachePath:   filepath.Join(tmpDir, "cache"),
		StoragePath: filepath.Join(tmpDir, "data.db"),
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	ctx := context.Background()
	err = app.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	app.Shutdown(ctx)
}

func TestAppCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		AppsPath:    filepath.Join(tmpDir, "apps"),
		BucketsPath: filepath.Join(tmpDir, "buckets"),
		CachePath:   filepath.Join(tmpDir, "cache"),
		StoragePath: filepath.Join(tmpDir, "data.db"),
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Shutdown(context.Background())

	if _, err := os.Stat(cfg.AppsPath); os.IsNotExist(err) {
		t.Error("AppsPath directory was not created")
	}

	if _, err := os.Stat(cfg.BucketsPath); os.IsNotExist(err) {
		t.Error("BucketsPath directory was not created")
	}
}

func TestMatchesQuery(t *testing.T) {
	tests := []struct {
		name     string
		app      *manifest.AppRef
		query    string
		expected bool
	}{
		{
			name: "match by name",
			app: &manifest.AppRef{
				Name:        "test-app",
				Description: "A test application",
			},
			query:    "test",
			expected: true,
		},
		{
			name: "match by description",
			app: &manifest.AppRef{
				Name:        "myapp",
				Description: "A testing tool",
			},
			query:    "testing",
			expected: true,
		},
		{
			name: "no match",
			app: &manifest.AppRef{
				Name:        "myapp",
				Description: "A tool",
			},
			query:    "xyz",
			expected: false,
		},
		{
			name: "case insensitive match",
			app: &manifest.AppRef{
				Name:        "TestApp",
				Description: "Description",
			},
			query:    "testapp",
			expected: true,
		},
		{
			name: "empty query",
			app: &manifest.AppRef{
				Name:        "app",
				Description: "desc",
			},
			query:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesQuery(tt.app, tt.query)
			if result != tt.expected {
				t.Errorf("matchesQuery() = %v, want %v", result, tt.expected)
			}
		})
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

func TestToLower(t *testing.T) {
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
		result := toLower(tt.input)
		if result != tt.expected {
			t.Errorf("toLower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestManagerListInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	mgr := NewManager(nil, storage, nil, nil, tmpDir)

	apps, err := mgr.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() failed: %v", err)
	}

	if apps == nil {
		t.Error("ListInstalled() returned nil")
	}
}

func TestInstallSpec(t *testing.T) {
	spec := InstallSpec{
		Bucket:  "main",
		Name:    "test-app",
		Version: "1.0.0",
	}

	if spec.Bucket != "main" {
		t.Error("Bucket mismatch")
	}

	if spec.Name != "test-app" {
		t.Error("Name mismatch")
	}

	if spec.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
}

func TestInstallOptions(t *testing.T) {
	opts := InstallOptions{
		Arch:       "amd64",
		Force:      true,
		InstallDir: "/apps/test",
	}

	if opts.Arch != "amd64" {
		t.Error("Arch mismatch")
	}

	if !opts.Force {
		t.Error("Force should be true")
	}

	if opts.InstallDir != "/apps/test" {
		t.Error("InstallDir mismatch")
	}
}

func TestRemoveOptions(t *testing.T) {
	opts := RemoveOptions{
		Purge: true,
	}

	if !opts.Purge {
		t.Error("Purge should be true")
	}
}

func TestUpdateOptions(t *testing.T) {
	opts := UpdateOptions{
		Force: true,
	}

	if !opts.Force {
		t.Error("Force should be true")
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
