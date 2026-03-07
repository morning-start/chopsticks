// Package config 提供配置管理功能。
//
// 该包负责加载、保存和管理应用程序的配置，支持 YAML 格式和
// 环境变量覆盖，包括全局配置、代理设置、日志配置等。
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
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
	ErrEmptyRootDir    = errors.New("root_dir cannot be empty")
	ErrInvalidParallel = errors.New("parallel must be greater than 0")
)

// Config 应用程序配置
// 使用 RootDir 作为根目录，其他路径自动推导
type Config struct {
	RootDir string `yaml:"root_dir" json:"root_dir"` // 根目录，其他目录基于此推导

	// 全局配置 - 简化字段命名
	AppsDir    string `yaml:"apps_dir" json:"apps_dir"`       // 应用安装目录
	BucketsDir string `yaml:"buckets_dir" json:"buckets_dir"` // 软件源目录
	CacheDir   string `yaml:"cache_dir" json:"cache_dir"`     // 缓存目录
	PersistDir string `yaml:"persist_dir" json:"persist_dir"` // 持久化数据目录
	ShimDir    string `yaml:"shim_dir" json:"shim_dir"`       // 可执行文件 shim 目录
	StorageDir string `yaml:"storage_dir" json:"storage_dir"` // 数据存储目录

	// 执行配置
	Parallel  int  `yaml:"parallel" json:"parallel"`
	Timeout   int  `yaml:"timeout" json:"timeout"`
	Retry     int  `yaml:"retry" json:"retry"`
	NoConfirm bool `yaml:"no_confirm" json:"no_confirm"`
	Color     bool `yaml:"color" json:"color"`
	Verbose   bool `yaml:"verbose" json:"verbose"`

	// Bucket 配置
	DefaultBucket string            `yaml:"default_bucket" json:"default_bucket"`
	AutoUpdate    bool              `yaml:"auto_update" json:"auto_update"`
	BucketMirrors map[string]string `yaml:"bucket_mirrors" json:"bucket_mirrors"`

	// 代理配置
	ProxyEnable  bool   `yaml:"proxy_enable" json:"proxy_enable"`
	ProxySystem  bool   `yaml:"proxy_system" json:"proxy_system"`
	ProxyHTTP    string `yaml:"proxy_http" json:"proxy_http"`
	ProxyHTTPS   string `yaml:"proxy_https" json:"proxy_https"`
	ProxyNoProxy string `yaml:"proxy_no_proxy" json:"proxy_no_proxy"`

	// 日志配置
	LogLevel      string `yaml:"log_level" json:"log_level"`
	LogFile       string `yaml:"log_file" json:"log_file"`
	LogMaxSize    int    `yaml:"log_max_size" json:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups" json:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age" json:"log_max_age"`
	LogCompress   bool   `yaml:"log_compress" json:"log_compress"`
}

// Option 配置选项函数类型
type Option func(*Config)

// WithRootDir 设置根目录
func WithRootDir(dir string) Option {
	return func(c *Config) {
		c.RootDir = dir
	}
}

// WithAppsDir 设置应用目录
func WithAppsDir(dir string) Option {
	return func(c *Config) {
		c.AppsDir = dir
	}
}

// WithBucketsDir 设置 Bucket 目录
func WithBucketsDir(dir string) Option {
	return func(c *Config) {
		c.BucketsDir = dir
	}
}

// WithCacheDir 设置缓存目录
func WithCacheDir(dir string) Option {
	return func(c *Config) {
		c.CacheDir = dir
	}
}

// WithStorageDir 设置存储目录
func WithStorageDir(dir string) Option {
	return func(c *Config) {
		c.StorageDir = dir
	}
}

// WithParallel 设置并行度
func WithParallel(n int) Option {
	return func(c *Config) {
		c.Parallel = n
	}
}

// WithTimeout 设置超时时间
func WithTimeout(seconds int) Option {
	return func(c *Config) {
		c.Timeout = seconds
	}
}

// WithRetry 设置重试次数
func WithRetry(n int) Option {
	return func(c *Config) {
		c.Retry = n
	}
}

// WithColor 设置是否启用彩色输出
func WithColor(enabled bool) Option {
	return func(c *Config) {
		c.Color = enabled
	}
}

// WithVerbose 设置是否启用详细输出
func WithVerbose(enabled bool) Option {
	return func(c *Config) {
		c.Verbose = enabled
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level string) Option {
	return func(c *Config) {
		c.LogLevel = level
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
		RootDir:       rootDir,
		AppsDir:       filepath.Join(rootDir, "apps"),
		BucketsDir:    filepath.Join(rootDir, "buckets"),
		CacheDir:      filepath.Join(rootDir, "cache"),
		PersistDir:    filepath.Join(rootDir, "persist"),
		ShimDir:       filepath.Join(rootDir, "shims"),
		StorageDir:    filepath.Join(rootDir, "data"),
		Parallel:      DefaultParallel,
		Timeout:       DefaultTimeout,
		Retry:         DefaultRetry,
		NoConfirm:     false,
		Color:         true,
		Verbose:       false,
		DefaultBucket: DefaultBucket,
		AutoUpdate:    false,
		BucketMirrors: make(map[string]string),
		ProxyEnable:   true,
		ProxySystem:   true,
		ProxyHTTP:     "",
		ProxyHTTPS:    "",
		ProxyNoProxy:  "",
		LogLevel:      DefaultLogLevel,
		LogFile:       "",
		LogMaxSize:    DefaultLogMaxSize,
		LogMaxBackups: DefaultLogMaxBackups,
		LogMaxAge:     DefaultLogMaxAge,
		LogCompress:   true,
	}

	// 应用选项
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Validate 验证配置有效性，并为无效值设置默认值
func (c *Config) Validate() error {
	if c.RootDir == "" {
		return ErrEmptyRootDir
	}

	// 如果目录未设置，基于 RootDir 自动推导
	if c.AppsDir == "" {
		c.AppsDir = filepath.Join(c.RootDir, "apps")
	}
	if c.BucketsDir == "" {
		c.BucketsDir = filepath.Join(c.RootDir, "buckets")
	}
	if c.CacheDir == "" {
		c.CacheDir = filepath.Join(c.RootDir, "cache")
	}
	if c.PersistDir == "" {
		c.PersistDir = filepath.Join(c.RootDir, "persist")
	}
	if c.ShimDir == "" {
		c.ShimDir = filepath.Join(c.RootDir, "shims")
	}
	if c.StorageDir == "" {
		c.StorageDir = filepath.Join(c.RootDir, "data")
	}

	// 设置默认值
	if c.Parallel <= 0 {
		c.Parallel = 1
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.Retry <= 0 {
		c.Retry = DefaultRetry
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

	// 如果 RootDir 未设置，使用默认值
	if cfg.RootDir == "" {
		cfg.RootDir = getRootDir()
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

// SetRootDir 设置根目录
func (b *Builder) SetRootDir(dir string) *Builder {
	b.config.RootDir = dir
	return b
}

// SetAppsDir 设置应用目录
func (b *Builder) SetAppsDir(dir string) *Builder {
	b.config.AppsDir = dir
	return b
}

// SetBucketsDir 设置 Bucket 目录
func (b *Builder) SetBucketsDir(dir string) *Builder {
	b.config.BucketsDir = dir
	return b
}

// SetCacheDir 设置缓存目录
func (b *Builder) SetCacheDir(dir string) *Builder {
	b.config.CacheDir = dir
	return b
}

// SetStorageDir 设置存储目录
func (b *Builder) SetStorageDir(dir string) *Builder {
	b.config.StorageDir = dir
	return b
}

// SetParallel 设置并行度
func (b *Builder) SetParallel(n int) *Builder {
	b.config.Parallel = n
	return b
}

// SetTimeout 设置超时时间
func (b *Builder) SetTimeout(seconds int) *Builder {
	b.config.Timeout = seconds
	return b
}

// SetRetry 设置重试次数
func (b *Builder) SetRetry(n int) *Builder {
	b.config.Retry = n
	return b
}

// SetColor 设置是否启用彩色输出
func (b *Builder) SetColor(enabled bool) *Builder {
	b.config.Color = enabled
	return b
}

// SetVerbose 设置是否启用详细输出
func (b *Builder) SetVerbose(enabled bool) *Builder {
	b.config.Verbose = enabled
	return b
}

// SetLogLevel 设置日志级别
func (b *Builder) SetLogLevel(level string) *Builder {
	b.config.LogLevel = level
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
func (c *Config) GetEffectiveProxy() (httpProxy, httpsProxy, noProxy string) {
	// 如果未启用代理，返回空
	if !c.ProxyEnable {
		return "", "", ""
	}

	// 如果使用系统代理，从环境变量读取
	if c.ProxySystem {
		httpProxy = c.getSystemHTTPProxy()
		httpsProxy = c.getSystemHTTPSProxy()
		noProxy = c.getSystemNoProxy()
		return httpProxy, httpsProxy, noProxy
	}

	// 使用手动配置的代理
	return c.ProxyHTTP, c.ProxyHTTPS, c.ProxyNoProxy
}

// getSystemHTTPProxy 从环境变量获取 HTTP 代理
func (c *Config) getSystemHTTPProxy() string {
	// 按优先级检查环境变量
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("http_proxy"); proxy != "" {
		return proxy
	}
	return c.ProxyHTTP
}

// getSystemHTTPSProxy 从环境变量获取 HTTPS 代理
func (c *Config) getSystemHTTPSProxy() string {
	// 按优先级检查环境变量
	if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("https_proxy"); proxy != "" {
		return proxy
	}
	return c.ProxyHTTPS
}

// getSystemNoProxy 从环境变量获取不代理的地址列表
func (c *Config) getSystemNoProxy() string {
	// 按优先级检查环境变量
	if noProxy := os.Getenv("NO_PROXY"); noProxy != "" {
		return noProxy
	}
	if noProxy := os.Getenv("no_proxy"); noProxy != "" {
		return noProxy
	}
	return c.ProxyNoProxy
}
