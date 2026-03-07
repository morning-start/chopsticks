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
	"chopsticks/pkg/config"
	"chopsticks/pkg/output"
)

const (
	// DefaultDirPerm 默认目录权限
	DefaultDirPerm = 0755
)

type Application interface {
	BucketManager() bucket.BucketManager
	AppManager() AppManager
	Installer() Installer
	Storage() store.LegacyStorage
	Config() *config.Config
	Logger() *logx.Logger
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// app 应用结构体
type app struct {
	jsEngine  *engine.JSEngine
	config    *config.Config
	bucketMgr bucket.BucketManager
	appMgr    AppManager
	installer Installer
	storage   store.LegacyStorage
	logger    *logx.Logger
}

func New(cfg *config.Config) (*app, error) {
	// 输入验证
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.RootDir == "" {
		return nil, fmt.Errorf("config.RootDir cannot be empty")
	}

	a := &app{
		config: cfg,
	}

	// 创建必要的目录
	dirs := []string{
		cfg.AppsDir,
		cfg.BucketsDir,
		cfg.CacheDir,
		cfg.PersistDir,
		cfg.ShimDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DefaultDirPerm); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Initialize logging system
	logCfg := logx.DefaultConfig()
	logCfg.Filename = filepath.Join(cfg.RootDir, "logs", "chopsticks.log")
	if err := logx.InitDefault(logCfg); err != nil {
		return nil, fmt.Errorf("initialize logging system: %w", err)
	}
	a.logger = logx.GetDefault()

	// 使用文件系统存储
	// cfg.StorageDir 现在作为数据存储的根目录
	fsStorage, err := store.NewFSStorage(cfg.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("initialize storage: %w", err)
	}

	// 检查是否有现有的文件系统数据，如果没有，提示用户可能需要迁移
	if !store.HasExistingFSData(cfg.StorageDir) {
		// 检查是否有旧的 SQLite 数据库
		oldDBPath := filepath.Join(cfg.RootDir, "chopsticks.db")
		if _, err := os.Stat(oldDBPath); err == nil {
			output.Warn("检测到旧的 SQLite 数据库：%s", oldDBPath)
			output.Info("请运行以下命令迁移数据到文件系统存储：")
			output.Info("  chopsticks migrate --from-sqlite %s\n", oldDBPath)
		}
	}

	// 创建存储适配器以向后兼容
	a.storage = store.NewStorageAdapter(fsStorage, cfg.AppsDir)

	a.bucketMgr = bucket.NewManager(a.storage, cfg, cfg.BucketsDir, nil)

	a.jsEngine = engine.NewJSEngine()

	var createErr error
	a.installer, createErr = NewInstaller(a.storage, cfg, a.jsEngine, cfg.AppsDir)
	if createErr != nil {
		return nil, fmt.Errorf("create installer: %w", createErr)
	}

	a.appMgr, createErr = NewManager(a.bucketMgr, a.storage, a.installer, cfg, cfg.AppsDir)
	if createErr != nil {
		return nil, fmt.Errorf("create app manager: %w", createErr)
	}

	return a, nil
}

func DefaultConfig() *config.Config {
	return config.DefaultConfig()
}

func (a *app) BucketManager() bucket.BucketManager {
	return a.bucketMgr
}

func (a *app) AppManager() AppManager {
	return a.appMgr
}

func (a *app) Installer() Installer {
	return a.installer
}

func (a *app) Storage() store.LegacyStorage {
	return a.storage
}

func (a *app) Config() *config.Config {
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
