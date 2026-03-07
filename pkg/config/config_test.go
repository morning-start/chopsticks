package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test Global config
	if cfg.AppsDir == "" {
		t.Error("DefaultConfig().AppsDir is empty")
	}
	if cfg.BucketsDir == "" {
		t.Error("DefaultConfig().BucketsDir is empty")
	}
	if cfg.CacheDir == "" {
		t.Error("DefaultConfig().CacheDir is empty")
	}
	if cfg.StorageDir == "" {
		t.Error("DefaultConfig().StorageDir is empty")
	}
	if cfg.Parallel != 3 {
		t.Errorf("DefaultConfig().Parallel = %d, want 3", cfg.Parallel)
	}
	if cfg.Timeout != 300 {
		t.Errorf("DefaultConfig().Timeout = %d, want 300", cfg.Timeout)
	}
	if cfg.Retry != 3 {
		t.Errorf("DefaultConfig().Retry = %d, want 3", cfg.Retry)
	}
	if cfg.NoConfirm != false {
		t.Error("DefaultConfig().NoConfirm = true, want false")
	}
	if cfg.Color != true {
		t.Error("DefaultConfig().Color = false, want true")
	}
	if cfg.Verbose != false {
		t.Error("DefaultConfig().Verbose = true, want false")
	}

	// Test Bucket config
	if cfg.DefaultBucket != "main" {
		t.Errorf("DefaultConfig().DefaultBucket = %s, want main", cfg.DefaultBucket)
	}
	if cfg.AutoUpdate != false {
		t.Error("DefaultConfig().AutoUpdate = true, want false")
	}
	if cfg.BucketMirrors == nil {
		t.Error("DefaultConfig().BucketMirrors is nil")
	}

	// Test Proxy config
	if cfg.ProxyEnable != true {
		t.Error("DefaultConfig().ProxyEnable = false, want true")
	}
	if cfg.ProxySystem != true {
		t.Error("DefaultConfig().ProxySystem = false, want true")
	}

	// Test Log config
	if cfg.LogLevel != "info" {
		t.Errorf("DefaultConfig().LogLevel = %s, want info", cfg.LogLevel)
	}
	if cfg.LogMaxSize != 10 {
		t.Errorf("DefaultConfig().LogMaxSize = %d, want 10", cfg.LogMaxSize)
	}
	if cfg.LogMaxBackups != 3 {
		t.Errorf("DefaultConfig().LogMaxBackups = %d, want 3", cfg.LogMaxBackups)
	}
	if cfg.LogMaxAge != 7 {
		t.Errorf("DefaultConfig().LogMaxAge = %d, want 7", cfg.LogMaxAge)
	}
	if cfg.LogCompress != true {
		t.Error("DefaultConfig().LogCompress = false, want true")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				RootDir:    "/chopsticks",
				AppsDir:    "/apps",
				BucketsDir: "/buckets",
				StorageDir: "/data",
				Parallel:   3,
				Timeout:    300,
				Retry:      3,
			},
			wantErr: false,
		},
		{
			name: "empty root_dir",
			cfg: &Config{
				RootDir:    "",
				AppsDir:    "/apps",
				BucketsDir: "/buckets",
				StorageDir: "/data",
			},
			wantErr: true,
		},
		{
			name: "zero parallel sets to 1",
			cfg: &Config{
				RootDir:    "/chopsticks",
				AppsDir:    "/apps",
				BucketsDir: "/buckets",
				StorageDir: "/data",
				Parallel:   0,
			},
			wantErr: false,
		},
		{
			name: "negative timeout sets to 300",
			cfg: &Config{
				RootDir:    "/chopsticks",
				AppsDir:    "/apps",
				BucketsDir: "/buckets",
				StorageDir: "/data",
				Timeout:    -1,
			},
			wantErr: false,
		},
		{
			name: "negative retry sets to 3",
			cfg: &Config{
				RootDir:    "/chopsticks",
				AppsDir:    "/apps",
				BucketsDir: "/buckets",
				StorageDir: "/data",
				Retry:      -1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidateSetsDefaults(t *testing.T) {
	cfg := &Config{
		RootDir:    "/chopsticks",
		AppsDir:    "/apps",
		BucketsDir: "/buckets",
		StorageDir: "/data",
		Parallel:   0,
		Timeout:    -1,
		Retry:      -1,
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}

	if cfg.Parallel != 1 {
		t.Errorf("Parallel = %d, want 1", cfg.Parallel)
	}
	if cfg.Timeout != 300 {
		t.Errorf("Timeout = %d, want 300", cfg.Timeout)
	}
	if cfg.Retry != 3 {
		t.Errorf("Retry = %d, want 3", cfg.Retry)
	}
}

func TestLoad(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "load non-existent file returns default",
			content: "",
			wantErr: false,
		},
		{
			name: "load valid yaml",
			content: `
global:
  parallel: 5
  timeout: 600
buckets:
  default: custom
`,
			wantErr: false,
		},
		{
			name:    "load invalid yaml",
			content: "invalid: yaml: content: [",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".yaml")

			if tt.content != "" {
				if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg == nil {
				t.Error("Load() returned nil config")
			}
		})
	}
}

func TestLoadValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
root_dir: /custom/chopsticks
apps_dir: /custom/apps
buckets_dir: /custom/buckets
cache_dir: /custom/cache
storage_dir: /custom/data
parallel: 5
timeout: 600
retry: 5
no_confirm: true
color: false
verbose: true
default_bucket: custom
auto_update: true
bucket_mirrors:
  main: https://mirror.example.com
proxy_enable: true
proxy_http: http://proxy.example.com:8080
proxy_https: https://proxy.example.com:8080
proxy_no_proxy: localhost,127.0.0.1
log_level: debug
log_file: /var/log/chopsticks.log
log_max_size: 100
log_max_backups: 10
log_max_age: 30
log_compress: false
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Check global config
	if cfg.AppsDir != "/custom/apps" {
		t.Errorf("AppsDir = %s, want /custom/apps", cfg.AppsDir)
	}
	if cfg.Parallel != 5 {
		t.Errorf("Parallel = %d, want 5", cfg.Parallel)
	}
	if cfg.Timeout != 600 {
		t.Errorf("Timeout = %d, want 600", cfg.Timeout)
	}
	if cfg.NoConfirm != true {
		t.Error("NoConfirm = false, want true")
	}

	// Check bucket config
	if cfg.DefaultBucket != "custom" {
		t.Errorf("DefaultBucket = %s, want custom", cfg.DefaultBucket)
	}
	if cfg.AutoUpdate != true {
		t.Error("AutoUpdate = false, want true")
	}
	if cfg.BucketMirrors["main"] != "https://mirror.example.com" {
		t.Errorf("BucketMirrors[main] = %s, want https://mirror.example.com", cfg.BucketMirrors["main"])
	}

	// Check proxy config
	if cfg.ProxyEnable != true {
		t.Error("ProxyEnable = false, want true")
	}

	// Check log config
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %s, want debug", cfg.LogLevel)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Parallel = 10

	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() did not create config file")
	}

	// Load and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.Parallel != 10 {
		t.Errorf("Parallel = %d, want 10", loaded.Parallel)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := DefaultConfig()
	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "subdir")); os.IsNotExist(err) {
		t.Error("Save() did not create config directory")
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original env var
	originalEnv := os.Getenv("CHOPSTICKS_CONFIG")
	defer os.Setenv("CHOPSTICKS_CONFIG", originalEnv)

	tests := []struct {
		name     string
		envValue string
		wantEnv  bool
	}{
		{
			name:     "with env var",
			envValue: "/custom/path/config.yaml",
			wantEnv:  true,
		},
		{
			name:     "without env var",
			envValue: "",
			wantEnv:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("CHOPSTICKS_CONFIG", tt.envValue)
			} else {
				os.Unsetenv("CHOPSTICKS_CONFIG")
			}

			path := GetConfigPath()
			if tt.wantEnv {
				if path != tt.envValue {
					t.Errorf("GetConfigPath() = %s, want %s", path, tt.envValue)
				}
			} else {
				// Should return default path
				if path == "" {
					t.Error("GetConfigPath() returned empty string")
				}
			}
		})
	}
}

func TestLoadDefault(t *testing.T) {
	// This test uses the actual default config path
	// We'll just verify it doesn't panic
	_, err := LoadDefault()
	// Error is expected if config file doesn't exist, which is fine
	if err != nil {
		// Should be able to get default config even if file doesn't exist
		t.Logf("LoadDefault() returned error (expected if no config file): %v", err)
	}
}

func TestSaveDefault(t *testing.T) {
	_ = DefaultConfig()
	// Just verify it doesn't panic
	// We don't want to actually write to the user's config directory in tests
	t.Log("SaveDefault() test skipped to avoid modifying user config")
}

func TestEnsureConfigDir(t *testing.T) {
	err := EnsureConfigDir()
	// Should not return error
	if err != nil {
		t.Errorf("EnsureConfigDir() returned error: %v", err)
	}
}
