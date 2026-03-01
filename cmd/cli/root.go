// Package cli 提供命令行界面实现。
package cli

import (
	"context"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	application app.Application
	rootContext context.Context
)

// rootCmd 表示 CLI 的根命令
var rootCmd = &cobra.Command{
	Use:   "chopsticks",
	Short: "Windows 包管理器 - 开发者友好的 Scoop 替代品",
	Long: `Chopsticks 是一个现代化的 Windows 包管理器，
提供快速、可靠的软件包安装和管理功能。`,
	Version: "0.5.0-alpha",
}

// Execute 执行 CLI
func Execute(ctx context.Context, appInstance app.Application) error {
	application = appInstance
	rootContext = ctx
	return rootCmd.ExecuteContext(ctx)
}

// getApp 获取应用实例
func getApp() app.Application {
	if application == nil {
		return nil
	}
	return application
}

// getContext 获取根上下文
func getContext() context.Context {
	return rootContext
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "指定配置文件路径")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "启用详细输出")
	rootCmd.PersistentFlags().Bool("no-color", false, "禁用彩色输出")

	// 处理 --no-color
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			output.DisableColor()
		}
	}
}
