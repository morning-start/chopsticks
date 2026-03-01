package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/bucket"
	"chopsticks/core/store"
	"chopsticks/engine"
	"chopsticks/engine/logx"
)

const (
	// DefaultDirPerm 默认目录权限
	DefaultDirPerm = 0755
)

type Application interface {
	BucketManager() bucket.Manager
	AppManager() Manager
	Installer() Installer
	Storage() store.Storage
	Config() *Config
	Logger() *logx.Logger
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// app 结构体字段按大小从大到小排列以优化内存布局
// 指针/引用类型 (8字节): 64位系统上
// 接口类型 (16字节): 包含类型和值指针
type app struct {
	jsEngine   *engine.JSEngine  // 8 bytes
	config     *Config           // 8 bytes
	bucketMgr  bucket.Manager    // 16 bytes (interface)
	appMgr     Manager           // 16 bytes (interface)
	installer  Installer         // 16 bytes (interface)
	storage    store.Storage     // 16 bytes (interface)
	logger     *logx.Logger      // 8 bytes
}

func New(cfg *Config) (*app, error) {
	a := &app{
		config: cfg,
	}

	if err := os.MkdirAll(cfg.AppsPath, DefaultDirPerm); err != nil {
		return nil, fmt.Errorf("create apps directory: %w", err)
	}

	if err := os.MkdirAll(cfg.BucketsPath, DefaultDirPerm); err != nil {
		return nil, fmt.Errorf("create buckets directory: %w", err)
	}

	// Initialize logging system
	logCfg := logx.DefaultConfig()
	logCfg.Filename = filepath.Join(filepath.Dir(cfg.StoragePath), "logs", "chopsticks.log")
	if err := logx.InitDefault(logCfg); err != nil {
		return nil, fmt.Errorf("initialize logging system: %w", err)
	}
	a.logger = logx.GetDefault()

	storage, err := store.New(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("initialize storage: %w", err)
	}
	a.storage = storage

	a.bucketMgr = bucket.NewManager(storage, cfg, cfg.BucketsPath)

	a.jsEngine = engine.NewJSEngine()

	a.installer = NewInstaller(storage, cfg, a.jsEngine, cfg.AppsPath)
	a.appMgr = NewManager(a.bucketMgr, storage, a.installer, cfg, cfg.AppsPath)

	return a, nil
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	chopsticksDir := filepath.Join(home, ".chopsticks")
	return &Config{
		AppsPath:    filepath.Join(chopsticksDir, "apps"),
		BucketsPath: filepath.Join(chopsticksDir, "buckets"),
		CachePath:   filepath.Join(chopsticksDir, "cache"),
		StoragePath: filepath.Join(chopsticksDir, "data.db"),
	}
}

type Config struct {
	AppsPath    string
	BucketsPath string
	CachePath   string
	StoragePath string
}

func (a *app) BucketManager() bucket.Manager {
	return a.bucketMgr
}

func (a *app) AppManager() Manager {
	return a.appMgr
}

func (a *app) Installer() Installer {
	return a.installer
}

func (a *app) Storage() store.Storage {
	return a.storage
}

func (a *app) Config() *Config {
	return a.config
}

func (a *app) Logger() *logx.Logger {
	return a.logger
}

func (a *app) Run(ctx context.Context) error {
	fmt.Println("Chopsticks is running...")
	return nil
}

func (a *app) Shutdown(ctx context.Context) error {
	fmt.Println("Chopsticks is shutting down...")
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			return err
		}
	}
	if a.logger != nil {
		if err := a.logger.Close(); err != nil {
			return err
		}
	}
	return nil
}
