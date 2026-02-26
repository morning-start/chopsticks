package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chopsticks/cmd/chopsticks/cli"
	"chopsticks/core/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "chopsticks: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := app.DefaultConfig()

	application, err := app.New(cfg)
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
