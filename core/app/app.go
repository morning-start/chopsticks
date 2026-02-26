package app

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/bucket"
	"chopsticks/core/store"
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
	config    *Config
	bucketMgr bucket.Manager
	appMgr    Manager
	installer Installer
	storage   store.Storage
}

func New(cfg *Config) (*app, error) {
	a := &app{
		config: cfg,
	}

	storage, err := store.New(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("初始化存储: %w", err)
	}
	a.storage = storage

	a.bucketMgr = bucket.NewManager(storage, cfg)
	a.installer = NewInstaller(storage, cfg, nil)
	a.appMgr = NewManager(a.bucketMgr, storage, a.installer, cfg)

	return a, nil
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		AppsPath:    home + "/.chopsticks/apps",
		CachePath:   home + "/.chopsticks/cache",
		StoragePath: home + "/.chopsticks/data.db",
	}
}

type Config struct {
	AppsPath    string
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
	return nil
}
