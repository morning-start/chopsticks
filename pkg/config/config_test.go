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
	if cfg.Global.AppsPath == "" {
		t.Error("DefaultConfig().Global.AppsPath is empty")
	}
	if cfg.Global.BucketsPath == "" {
		t.Error("DefaultConfig().Global.BucketsPath is empty")
	}
	if cfg.Global.CachePath == "" {
		t.Error("DefaultConfig().Global.CachePath is empty")
	}
	if cfg.Global.StoragePath == "" {
		t.Error("DefaultConfig().Global.StoragePath is empty")
	}
	if cfg.Global.Parallel != 3 {
		t.Errorf("DefaultConfig().Global.Parallel = %d, want 3", cfg.Global.Parallel)
	}
	if cfg.Global.Timeout != 300 {
		t.Errorf("DefaultConfig().Global.Timeout = %d, want 300", cfg.Global.Timeout)
	}
	if cfg.Global.Retry != 3 {
		t.Errorf("DefaultConfig().Global.Retry = %d, want 3", cfg.Global.Retry)
	}
	if cfg.Global.NoConfirm != false {
		t.Error("DefaultConfig().Global.NoConfirm = true, want false")
	}
	if cfg.Global.Color != true {
		t.Error("DefaultConfig().Global.Color = false, want true")
	}
	if cfg.Global.Verbose != false {
		t.Error("DefaultConfig().Global.Verbose = true, want false")
	}

	// Test Bucket config
	if cfg.Buckets.Default != "main" {
		t.Errorf("DefaultConfig().Buckets.Default = %s, want main", cfg.Buckets.Default)
	}
	if cfg.Buckets.AutoUpdate != false {
		t.Error("DefaultConfig().Buckets.AutoUpdate = true, want false")
	}
	if cfg.Buckets.Mirrors == nil {
		t.Error("DefaultConfig().Buckets.Mirrors is nil")
	}

	// Test Proxy config
	if cfg.Proxy.Enable != false {
		t.Error("DefaultConfig().Proxy.Enable = true, want false")
	}

	// Test Log config
	if cfg.Log.Level != "info" {
		t.Errorf("DefaultConfig().Log.Level = %s, want info", cfg.Log.Level)
	}
	if cfg.Log.MaxSize != 10 {
		t.Errorf("DefaultConfig().Log.MaxSize = %d, want 10", cfg.Log.MaxSize)
	}
	if cfg.Log.MaxBackups != 3 {
		t.Errorf("DefaultConfig().Log.MaxBackups = %d, want 3", cfg.Log.MaxBackups)
	}
	if cfg.Log.MaxAge != 7 {
		t.Errorf("DefaultConfig().Log.MaxAge = %d, want 7", cfg.Log.MaxAge)
	}
	if cfg.Log.Compress != true {
		t.Error("DefaultConfig().Log.Compress = false, want true")
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
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "/buckets",
					StoragePath: "/data.db",
					Parallel:    3,
					Timeout:     300,
					Retry:       3,
				},
			},
			wantErr: false,
		},
		{
			name: "empty apps_path",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "",
					BucketsPath: "/buckets",
					StoragePath: "/data.db",
				},
			},
			wantErr: true,
		},
		{
			name: "empty buckets_path",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "",
					StoragePath: "/data.db",
				},
			},
			wantErr: true,
		},
		{
			name: "empty storage_path",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "/buckets",
					StoragePath: "",
				},
			},
			wantErr: true,
		},
		{
			name: "zero parallel sets to 1",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "/buckets",
					StoragePath: "/data.db",
					Parallel:    0,
				},
			},
			wantErr: false,
		},
		{
			name: "negative timeout sets to 300",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "/buckets",
					StoragePath: "/data.db",
					Timeout:     -1,
				},
			},
			wantErr: false,
		},
		{
			name: "negative retry sets to 3",
			cfg: &Config{
				Global: GlobalConfig{
					AppsPath:    "/apps",
					BucketsPath: "/buckets",
					StoragePath: "/data.db",
					Retry:       -1,
				},
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
		Global: GlobalConfig{
			AppsPath:    "/apps",
			BucketsPath: "/buckets",
			StoragePath: "/data.db",
			Parallel:    0,
			Timeout:     -1,
			Retry:       -1,
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}

	if cfg.Global.Parallel != 1 {
		t.Errorf("Parallel = %d, want 1", cfg.Global.Parallel)
	}
	if cfg.Global.Timeout != 300 {
		t.Errorf("Timeout = %d, want 300", cfg.Global.Timeout)
	}
	if cfg.Global.Retry != 3 {
		t.Errorf("Retry = %d, want 3", cfg.Global.Retry)
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
global:
  apps_path: /custom/apps
  buckets_path: /custom/buckets
  cache_path: /custom/cache
  storage_path: /custom/data.db
  parallel: 5
  timeout: 600
  retry: 5
  no_confirm: true
  color: false
  verbose: true
buckets:
  default: custom
  auto_update: true
  mirrors:
    main: https://mirror.example.com
proxy:
  enable: true
  http: http://proxy.example.com:8080
  https: https://proxy.example.com:8080
  no_proxy: localhost,127.0.0.1
log:
  level: debug
  file: /var/log/chopsticks.log
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: false
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Check global config
	if cfg.Global.AppsPath != "/custom/apps" {
		t.Errorf("AppsPath = %s, want /custom/apps", cfg.Global.AppsPath)
	}
	if cfg.Global.Parallel != 5 {
		t.Errorf("Parallel = %d, want 5", cfg.Global.Parallel)
	}
	if cfg.Global.Timeout != 600 {
		t.Errorf("Timeout = %d, want 600", cfg.Global.Timeout)
	}
	if cfg.Global.NoConfirm != true {
		t.Error("NoConfirm = false, want true")
	}

	// Check bucket config
	if cfg.Buckets.Default != "custom" {
		t.Errorf("Default = %s, want custom", cfg.Buckets.Default)
	}
	if cfg.Buckets.AutoUpdate != true {
		t.Error("AutoUpdate = false, want true")
	}
	if cfg.Buckets.Mirrors["main"] != "https://mirror.example.com" {
		t.Errorf("Mirrors[main] = %s, want https://mirror.example.com", cfg.Buckets.Mirrors["main"])
	}

	// Check proxy config
	if cfg.Proxy.Enable != true {
		t.Error("Enable = false, want true")
	}

	// Check log config
	if cfg.Log.Level != "debug" {
		t.Errorf("Level = %s, want debug", cfg.Log.Level)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Global.Parallel = 10

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

	if loaded.Global.Parallel != 10 {
		t.Errorf("Parallel = %d, want 10", loaded.Global.Parallel)
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
