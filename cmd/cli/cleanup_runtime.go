// Package cli 提供运行时库清理命令
package cli

import (
	"fmt"

	"chopsticks/core/dep"

	"github.com/spf13/cobra"
)

var (
	cleanupRuntimeDryRun bool
)

// cleanupRuntimeCmd 表示 cleanup-runtime 命令
var cleanupRuntimeCmd = &cobra.Command{
	Use:   "cleanup-runtime",
	Short: "清理无用运行时库",
	Long: `清理不再被任何软件需要的运行时库（引用计数为 0）。

示例:
  chopsticks cleanup-runtime
  chopsticks cleanup-runtime --dry-run`,
	RunE: runCleanupRuntime,
}

func init() {
	// cleanup-runtime 命令标志
	cleanupRuntimeCmd.Flags().BoolVarP(&cleanupRuntimeDryRun, "dry-run", "d", false, "预览可清理内容，不实际执行")
}

func runCleanupRuntime(cmd *cobra.Command, args []string) error {
	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	// 创建依赖管理器
	depMgr, err := dep.NewDependencyManager(
		application.BucketManager(),
		application.Storage(),
		application.Config().PersistDir,
	)
	if err != nil {
		return fmt.Errorf("创建依赖管理器失败：%w", err)
	}

	// 清理运行时库
	if err := depMgr.CleanupRuntime(cmd.Context()); err != nil {
		return fmt.Errorf("清理运行时库失败：%w", err)
	}

	return nil
}
