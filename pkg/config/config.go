package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// 默认配置常量
const (
	DefaultParallel      = 3
	DefaultTimeout       = 300
	DefaultRetry         = 3
	DefaultLogLevel      = "info"
	DefaultLogMaxSize    = 10
	DefaultLogMaxBackups = 3
	DefaultLogMaxAge     = 7
	DefaultBucket        = "main"
)

// 环境变量名
const (
	EnvRootDir    = "CHOPSTICKS_ROOT"
	EnvConfigPath = "CHOPSTICKS_CONFIG"
)

// 错误定义
var (
	ErrEmptyAppsPath    = errors.New("apps_path cannot be empty")
	ErrEmptyBucketsPath = errors.New("buckets_path cannot be empty")
	ErrEmptyStoragePath = errors.New("storage_path cannot be empty")
	ErrInvalidParallel  = errors.New("parallel must be greater than 0")
)

// Config 应用程序配置
type Config struct {
	Global  GlobalConfig `yaml:"global" json:"global"`
	Buckets BucketConfig `yaml:"buckets" json:"buckets"`
	Proxy   ProxyConfig  `yaml:"proxy" json:"proxy"`
	Log     LogConfig    `yaml:"log" json:"log"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	AppsPath    string `yaml:"apps_path" json:"apps_path"`
	BucketsPath string `yaml:"buckets_path" json:"buckets_path"`
	CachePath   string `yaml:"cache_path" json:"cache_path"`
	StoragePath string `yaml:"storage_path" json:"storage_path"`
	Parallel    int    `yaml:"parallel" json:"parallel"`
	Timeout     int    `yaml:"timeout" json:"timeout"`
	Retry       int    `yaml:"retry" json:"retry"`
	NoConfirm   bool   `yaml:"no_confirm" json:"no_confirm"`
	Color       bool   `yaml:"color" json:"color"`
	Verbose     bool   `yaml:"verbose" json:"verbose"`
}

// BucketConfig Bucket 配置
type BucketConfig struct {
	Default    string            `yaml:"default" json:"default"`
	AutoUpdate bool              `yaml:"auto_update" json:"auto_update"`
	Mirrors    map[string]string `yaml:"mirrors" json:"mirrors"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enable  bool   `yaml:"enable" json:"enable"`
	System  bool   `yaml:"system" json:"system"`
	HTTP    string `yaml:"http" json:"http"`
	HTTPS   string `yaml:"https" json:"https"`
	NoProxy string `yaml:"no_proxy" json:"no_proxy"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level" json:"level"`
	File       string `yaml:"file" json:"file"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// Option 配置选项函数类型
type Option func(*Config)

// WithAppsPath 设置应用路径
func WithAppsPath(path string) Option {
	return func(c *Config) {
		c.Global.AppsPath = path
	}
}

// WithBucketsPath 设置 Bucket 路径
func WithBucketsPath(path string) Option {
	return func(c *Config) {
		c.Global.BucketsPath = path
	}
}

// WithCachePath 设置缓存路径
func WithCachePath(path string) Option {
	return func(c *Config) {
		c.Global.CachePath = path
	}
}

// WithStoragePath 设置存储路径
func WithStoragePath(path string) Option {
	return func(c *Config) {
		c.Global.StoragePath = path
	}
}

// WithParallel 设置并行度
func WithParallel(n int) Option {
	return func(c *Config) {
		c.Global.Parallel = n
	}
}

// WithTimeout 设置超时时间
func WithTimeout(seconds int) Option {
	return func(c *Config) {
		c.Global.Timeout = seconds
	}
}

// WithRetry 设置重试次数
func WithRetry(n int) Option {
	return func(c *Config) {
		c.Global.Retry = n
	}
}

// WithColor 设置是否启用彩色输出
func WithColor(enabled bool) Option {
	return func(c *Config) {
		c.Global.Color = enabled
	}
}

// WithVerbose 设置是否启用详细输出
func WithVerbose(enabled bool) Option {
	return func(c *Config) {
		c.Global.Verbose = enabled
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level string) Option {
	return func(c *Config) {
		c.Log.Level = level
	}
}

// getRootDir 获取 Chopsticks 根目录
func getRootDir() string {
	if dir := os.Getenv(EnvRootDir); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".chopsticks")
}

// DefaultConfig 创建默认配置
func DefaultConfig() *Config {
	return NewConfig()
}

// NewConfig 使用选项模式创建配置
func NewConfig(opts ...Option) *Config {
	rootDir := getRootDir()

	cfg := &Config{
		Global: GlobalConfig{
			AppsPath:    filepath.Join(rootDir, "apps"),
			BucketsPath: filepath.Join(rootDir, "buckets"),
			CachePath:   filepath.Join(rootDir, "cache"),
			StoragePath: filepath.Join(rootDir, "data.db"),
			Parallel:    DefaultParallel,
			Timeout:     DefaultTimeout,
			Retry:       DefaultRetry,
			NoConfirm:   false,
			Color:       true,
			Verbose:     false,
		},
		Buckets: BucketConfig{
			Default:    DefaultBucket,
			AutoUpdate: false,
			Mirrors:    make(map[string]string),
		},
		Proxy: ProxyConfig{
			Enable:  true,
			System:  true,
			HTTP:    "",
			HTTPS:   "",
			NoProxy: "",
		},
		Log: LogConfig{
			Level:      DefaultLogLevel,
			File:       "",
			MaxSize:    DefaultLogMaxSize,
			MaxBackups: DefaultLogMaxBackups,
			MaxAge:     DefaultLogMaxAge,
			Compress:   true,
		},
	}

	// 应用选项
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Validate 验证配置有效性，并为无效值设置默认值
func (c *Config) Validate() error {
	if c.Global.AppsPath == "" {
		return ErrEmptyAppsPath
	}
	if c.Global.BucketsPath == "" {
		return ErrEmptyBucketsPath
	}
	if c.Global.StoragePath == "" {
		return ErrEmptyStoragePath
	}

	// 设置默认值
	if c.Global.Parallel <= 0 {
		c.Global.Parallel = 1
	}
	if c.Global.Timeout <= 0 {
		c.Global.Timeout = DefaultTimeout
	}
	if c.Global.Retry <= 0 {
		c.Global.Retry = DefaultRetry
	}

	return nil
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Save 保存配置到文件
func Save(cfg *Config, path string) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	if envPath := os.Getenv(EnvConfigPath); envPath != "" {
		return envPath
	}
	return filepath.Join(getRootDir(), "config.yaml")
}

// LoadDefault 加载默认配置文件
func LoadDefault() (*Config, error) {
	return Load(GetConfigPath())
}

// SaveDefault 保存到默认配置文件
func SaveDefault(cfg *Config) error {
	return Save(cfg, GetConfigPath())
}

// EnsureConfigDir 确保配置目录存在
func EnsureConfigDir() error {
	return os.MkdirAll(getRootDir(), 0755)
}

// Builder 配置构建器
type Builder struct {
	config *Config
}

// NewBuilder 创建配置构建器
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultConfig(),
	}
}

// SetAppsPath 设置应用路径
func (b *Builder) SetAppsPath(path string) *Builder {
	b.config.Global.AppsPath = path
	return b
}

// SetBucketsPath 设置 Bucket 路径
func (b *Builder) SetBucketsPath(path string) *Builder {
	b.config.Global.BucketsPath = path
	return b
}

// SetCachePath 设置缓存路径
func (b *Builder) SetCachePath(path string) *Builder {
	b.config.Global.CachePath = path
	return b
}

// SetStoragePath 设置存储路径
func (b *Builder) SetStoragePath(path string) *Builder {
	b.config.Global.StoragePath = path
	return b
}

// SetParallel 设置并行度
func (b *Builder) SetParallel(n int) *Builder {
	b.config.Global.Parallel = n
	return b
}

// SetTimeout 设置超时时间
func (b *Builder) SetTimeout(seconds int) *Builder {
	b.config.Global.Timeout = seconds
	return b
}

// SetRetry 设置重试次数
func (b *Builder) SetRetry(n int) *Builder {
	b.config.Global.Retry = n
	return b
}

// SetColor 设置是否启用彩色输出
func (b *Builder) SetColor(enabled bool) *Builder {
	b.config.Global.Color = enabled
	return b
}

// SetVerbose 设置是否启用详细输出
func (b *Builder) SetVerbose(enabled bool) *Builder {
	b.config.Global.Verbose = enabled
	return b
}

// SetLogLevel 设置日志级别
func (b *Builder) SetLogLevel(level string) *Builder {
	b.config.Log.Level = level
	return b
}

// Build 构建配置
func (b *Builder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// GetEffectiveProxy 获取有效的代理配置
// 如果启用了系统代理，会从环境变量读取代理设置
func (c *ProxyConfig) GetEffectiveProxy() (httpProxy, httpsProxy, noProxy string) {
	// 如果未启用代理，返回空
	if !c.Enable {
		return "", "", ""
	}

	// 如果使用系统代理，从环境变量读取
	if c.System {
		httpProxy = c.getSystemHTTPProxy()
		httpsProxy = c.getSystemHTTPSProxy()
		noProxy = c.getSystemNoProxy()
		return httpProxy, httpsProxy, noProxy
	}

	// 使用手动配置的代理
	return c.HTTP, c.HTTPS, c.NoProxy
}

// getSystemHTTPProxy 从环境变量获取 HTTP 代理
func (c *ProxyConfig) getSystemHTTPProxy() string {
	// 按优先级检查环境变量
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("http_proxy"); proxy != "" {
		return proxy
	}
	return c.HTTP
}

// getSystemHTTPSProxy 从环境变量获取 HTTPS 代理
func (c *ProxyConfig) getSystemHTTPSProxy() string {
	// 按优先级检查环境变量
	if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("https_proxy"); proxy != "" {
		return proxy
	}
	return c.HTTPS
}

// getSystemNoProxy 从环境变量获取不代理的地址列表
func (c *ProxyConfig) getSystemNoProxy() string {
	// 按优先级检查环境变量
	if noProxy := os.Getenv("NO_PROXY"); noProxy != "" {
		return noProxy
	}
	if noProxy := os.Getenv("no_proxy"); noProxy != "" {
		return noProxy
	}
	return c.NoProxy
}
