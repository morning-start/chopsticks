// Package cli 提供依赖清理命令
package cli

import (
	"fmt"
	"strings"

	"chopsticks/core/dep"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	autoremoveDryRun bool
)

// autoremoveCmd 表示 autoremove 命令
var autoremoveCmd = &cobra.Command{
	Use:   "autoremove",
	Short: "清理孤儿依赖",
	Long: `清理不再被任何软件需要的依赖（孤儿依赖）。

示例:
  chopsticks autoremove
  chopsticks autoremove --dry-run`,
	RunE: runAutoremove,
}

func init() {
	// autoremove 命令标志
	autoremoveCmd.Flags().BoolVarP(&autoremoveDryRun, "dry-run", "d", false, "预览可清理内容，不实际执行")
}

func runAutoremove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	// 创建依赖管理器
	depMgr := dep.NewDependencyManager(application.BucketManager(), application.Config().PersistPath)

	// 查找孤儿依赖
	orphans, err := depMgr.FindOrphans(ctx)
	if err != nil {
		return fmt.Errorf("查找孤儿依赖失败: %w", err)
	}

	// 显示孤儿依赖
	output.Infoln("========================================")
	output.Infoln("孤儿依赖分析")
	output.Infoln("========================================")
	fmt.Println()

	if len(orphans.Runtime) == 0 && len(orphans.Tools) == 0 {
		output.Success("没有发现孤儿依赖")
		return nil
	}

	// 显示运行时库
	if len(orphans.Runtime) > 0 {
		output.Infoln("孤儿运行时库：")
		for _, runtime := range orphans.Runtime {
			fmt.Printf("  - %s\n", runtime)
		}
		fmt.Println()
	}

	// 显示工具软件
	if len(orphans.Tools) > 0 {
		output.Infoln("孤儿工具软件：")
		for _, tool := range orphans.Tools {
			fmt.Printf("  - %s\n", tool)
		}
		fmt.Println()
	}

	// 预览模式
	if autoremoveDryRun {
		output.Info("预览模式：不会实际清理")
		return nil
	}

	// 确认清理
	output.Warning("即将清理上述孤儿依赖")
	if !confirm("是否继续？") {
		output.Info("已取消清理")
		return nil
	}

	// 执行清理
	if err := depMgr.CleanupOrphans(ctx, orphans); err != nil {
		return fmt.Errorf("清理孤儿依赖失败: %w", err)
	}

	output.Success("孤儿依赖清理完成")
	return nil
}

// confirm 确认操作
func confirm(message string) bool {
	fmt.Printf("%s [Y/n]: ", message)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}
