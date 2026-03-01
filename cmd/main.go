package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chopsticks/cmd/cli"
	coreapp "chopsticks/core/app"
	"chopsticks/pkg/config"
	"chopsticks/pkg/output"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "chopsticks: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 加载用户配置
	userCfg, err := config.LoadDefault()
	if err != nil {
		output.Warn("加载配置文件失败，使用默认配置: %v", err)
		userCfg = config.DefaultConfig()
	}

	// 将 pkg/config 配置转换为 core/app 配置
	cfg := &coreapp.Config{
		AppsPath:    userCfg.Global.AppsPath,
		BucketsPath: userCfg.Global.BucketsPath,
		CachePath:   userCfg.Global.CachePath,
		StoragePath: userCfg.Global.StoragePath,
	}

	application, err := coreapp.New(cfg)
	if err != nil {
		return fmt.Errorf("创建应用: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "关闭错误: %v\n", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return cli.Execute(ctx, application)
}
