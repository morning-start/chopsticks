package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/bucket"
	"chopsticks/core/store"
	"chopsticks/engine"
)

type Application interface {
	BucketManager() bucket.Manager
	AppManager() Manager
	Installer() Installer
	Storage() store.Storage
	Config() *Config
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type app struct {
	config     *Config
	bucketMgr  bucket.Manager
	appMgr     Manager
	installer  Installer
	storage    store.Storage
	jsEngine   *engine.JSEngine
}

func New(cfg *Config) (*app, error) {
	a := &app{
		config: cfg,
	}

	if err := os.MkdirAll(cfg.AppsPath, 0755); err != nil {
		return nil, fmt.Errorf("创建应用目录: %w", err)
	}

	if err := os.MkdirAll(cfg.BucketsPath, 0755); err != nil {
		return nil, fmt.Errorf("创建软件源目录: %w", err)
	}

	storage, err := store.New(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("初始化存储: %w", err)
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

func (a *app) Run(ctx context.Context) error {
	fmt.Println("Chopsticks is running...")
	return nil
}

func (a *app) Shutdown(ctx context.Context) error {
	fmt.Println("Chopsticks is shutting down...")
	if a.storage != nil {
		return a.storage.Close()
	}
	return nil
}
