package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Global  GlobalConfig  `yaml:"global" json:"global"`
	Buckets BucketConfig `yaml:"buckets" json:"buckets"`
	Proxy   ProxyConfig  `yaml:"proxy" json:"proxy"`
	Log     LogConfig    `yaml:"log" json:"log"`
}

type GlobalConfig struct {
	AppsPath    string `yaml:"apps_path" json:"apps_path"`
	BucketsPath string `yaml:"buckets_path" json:"buckets_path"`
	CachePath   string `yaml:"cache_path" json:"cache_path"`
	StoragePath string `yaml:"storage_path" json:"storage_path"`
	Parallel   int    `yaml:"parallel" json:"parallel"`
	Timeout    int    `yaml:"timeout" json:"timeout"`
	Retry      int    `yaml:"retry" json:"retry"`
	NoConfirm  bool   `yaml:"no_confirm" json:"no_confirm"`
	Color      bool   `yaml:"color" json:"color"`
	Verbose    bool   `yaml:"verbose" json:"verbose"`
}

type BucketConfig struct {
	Default   string            `yaml:"default" json:"default"`
	AutoUpdate bool            `yaml:"auto_update" json:"auto_update"`
	Mirrors  map[string]string `yaml:"mirrors" json:"mirrors"`
}

type ProxyConfig struct {
	Enable   bool   `yaml:"enable" json:"enable"`
	HTTP     string `yaml:"http" json:"http"`
	HTTPS    string `yaml:"https" json:"https"`
	NoProxy  string `yaml:"no_proxy" json:"no_proxy"`
}

type LogConfig struct {
	Level    string `yaml:"level" json:"level"`
	File     string `yaml:"file" json:"file"`
	MaxSize  int    `yaml:"max_size" json:"max_size"`
	MaxBackups int  `yaml:"max_backups" json:"max_backups"`
	MaxAge   int    `yaml:"max_age" json:"max_age"`
	Compress bool   `yaml:"compress" json:"compress"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	chopsticksDir := filepath.Join(home, ".chopsticks")

	return &Config{
		Global: GlobalConfig{
			AppsPath:    filepath.Join(chopsticksDir, "apps"),
			BucketsPath: filepath.Join(chopsticksDir, "buckets"),
			CachePath:   filepath.Join(chopsticksDir, "cache"),
			StoragePath: filepath.Join(chopsticksDir, "data.db"),
			Parallel:    3,
			Timeout:     300,
			Retry:      3,
			NoConfirm:  false,
			Color:      true,
			Verbose:    false,
		},
		Buckets: BucketConfig{
			Default:    "main",
			AutoUpdate: false,
			Mirrors:   make(map[string]string),
		},
		Proxy: ProxyConfig{
			Enable:  false,
			HTTP:    "",
			HTTPS:   "",
			NoProxy: "",
		},
		Log: LogConfig{
			Level:    "info",
			File:     "",
			MaxSize:  10,
			MaxBackups: 3,
			MaxAge:   7,
			Compress: true,
		},
	}
}

func (c *Config) Validate() error {
	if c.Global.AppsPath == "" {
		return fmt.Errorf("apps_path 不能为空")
	}
	if c.Global.BucketsPath == "" {
		return fmt.Errorf("buckets_path 不能为空")
	}
	if c.Global.StoragePath == "" {
		return fmt.Errorf("storage_path 不能为空")
	}
	if c.Global.Parallel <= 0 {
		c.Global.Parallel = 1
	}
	if c.Global.Timeout <= 0 {
		c.Global.Timeout = 300
	}
	if c.Global.Retry <= 0 {
		c.Global.Retry = 3
	}
	return nil
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	if envPath := os.Getenv("CHOPSTICKS_CONFIG"); envPath != "" {
		return envPath
	}
	return filepath.Join(home, ".chopsticks", "config.yaml")
}

func LoadDefault() (*Config, error) {
	configPath := GetConfigPath()
	return Load(configPath)
}

func SaveDefault(cfg *Config) error {
	configPath := GetConfigPath()
	return Save(cfg, configPath)
}

func EnsureConfigDir() error {
	home, _ := os.UserHomeDir()
	chopsticksDir := filepath.Join(home, ".chopsticks")
	return os.MkdirAll(chopsticksDir, 0755)
}
