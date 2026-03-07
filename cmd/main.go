package main

import (
	"context"
	"errors"
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

// 关闭超时常量
const shutdownTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "chopsticks: %v\n", err)
		os.Exit(1)
	}
}

// isHelpOrVersionOnly 检查是否只请求 help 或 version
func isHelpOrVersionOnly() bool {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help", "-v", "--version", "help", "version":
			return true
		}
	}
	return false
}

func run() error {
	// 如果只请求 help 或 version，跳过应用初始化
	if isHelpOrVersionOnly() {
		return cli.Execute(context.Background(), nil)
	}

	// 加载用户配置
	userCfg, err := config.LoadDefault()
	if err != nil {
		output.Warn("Failed to load config file, using default: %v", err)
		userCfg = config.DefaultConfig()
	}

	// 使用 pkg/config 配置直接创建应用
	application, err := coreapp.New(userCfg)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return cli.Execute(ctx, application)
}
