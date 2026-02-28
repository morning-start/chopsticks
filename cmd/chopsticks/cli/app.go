// Package cli 提供命令行界面功能。
package cli

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// application 全局应用实例
var application app.Application

// Execute 执行 CLI 应用程序。
func Execute(ctx context.Context, appInstance app.Application) error {
	application = appInstance
	
	cliApp := &cli.App{
		Name:    "chopsticks",
		Usage:   "Windows 包管理器 - 开发者友好的 Scoop 替代品",
		Version: "0.5.0-alpha",
		Authors: []*cli.Author{
			{
				Name:  "Chopsticks Team",
				Email: "team@chopsticks.dev",
			},
		},
		Commands: []*cli.Command{
			installCommand(),
			uninstallCommand(),
			updateCommand(),
			searchCommand(),
			listCommand(),
			bucketCommand(),
			configCommand(),
			conflictCommand(),
			completionCommand(),
			PerfCommand,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "指定配置文件路径",
				EnvVars: []string{"CHOPSTICKS_CONFIG"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "启用详细输出",
				EnvVars: []string{"CHOPSTICKS_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:    "no-color",
				Usage:   "禁用彩色输出",
				EnvVars: []string{"NO_COLOR"},
			},
		},
		Before: func(c *cli.Context) error {
			// 设置上下文
			c.App.Metadata = map[string]interface{}{
				"context": ctx,
			}
			// 处理 --no-color 选项
			if c.Bool("no-color") {
				output.DisableColor()
			}
			return nil
		},
		CommandNotFound: func(c *cli.Context, command string) {
			fmt.Fprintf(os.Stderr, "错误: 未知命令 '%s'\n", command)
			fmt.Fprintf(os.Stderr, "使用 'chopsticks --help' 查看可用命令\n")
			os.Exit(1)
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			if isSubcommand {
				fmt.Fprintf(os.Stderr, "使用 'chopsticks %s --help' 查看用法\n", c.Command.Name)
			} else {
				fmt.Fprintf(os.Stderr, "使用 'chopsticks --help' 查看用法\n")
			}
			return err
		},
	}

	return cliApp.RunContext(ctx, os.Args)
}

// getContext 从 cli.Context 获取应用上下文。
func getContext(c *cli.Context) context.Context {
	if ctx, ok := c.App.Metadata["context"].(context.Context); ok {
		return ctx
	}
	return context.Background()
}

// getApp 获取应用实例。
func getApp() app.Application {
	return application
}
