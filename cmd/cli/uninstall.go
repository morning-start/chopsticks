package cli

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	uninstallPurge bool
)

// uninstallCmd 表示 uninstall 命令
var uninstallCmd = &cobra.Command{
	Use:     "uninstall <package> ...",
	Aliases: []string{"remove", "rm"},
	Short:   "卸载软件包",
	Long: `卸载指定的软件包。支持批量卸载多个软件包。

示例:
  chopsticks uninstall git
  chopsticks remove nodejs
  chopsticks rm python --purge
  chopsticks uninstall app1 app2 app3
  chopsticks rm --purge app1 app2`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	// 单个包直接卸载
	if len(args) == 1 {
		return uninstallSingle(ctx, application.AppManager(), args[0], uninstallPurge)
	}

	// 批量卸载
	return uninstallBatch(ctx, application.AppManager(), args, uninstallPurge)
}

// uninstallSingle 卸载单个软件包
func uninstallSingle(ctx context.Context, mgr app.AppManager, name string, purge bool) error {
	output.Info("正在卸载 ")
	output.Highlight("%s", name)
	if purge {
		output.Warning(" (彻底清除)")
	}
	fmt.Println()

	opts := app.RemoveOptions{
		Purge: purge,
	}

	if err := mgr.Remove(ctx, name, opts); err != nil {
		output.ErrorCross(fmt.Sprintf("卸载失败: %v", err))
		return fmt.Errorf("卸载失败: %w", err)
	}

	output.SuccessCheck(fmt.Sprintf("%s 卸载成功", name))
	return nil
}

// uninstallBatch 批量卸载软件包
func uninstallBatch(ctx context.Context, mgr app.AppManager, packages []string, purge bool) error {
	total := len(packages)

	output.Infoln("========================================")
	output.Infof("开始批量卸载 %d 个软件包\n", total)
	output.Infoln("========================================")
	fmt.Println()

	results := make([]batchResult, total)
	var mu sync.Mutex

	for i, name := range packages {
		output.Infof("[%d/%d] ", i+1, total)
		output.Info("正在卸载 ")
		output.Highlight("%s", name)
		if purge {
			output.Warning(" (彻底清除)")
		}
		fmt.Println()

		opts := app.RemoveOptions{
			Purge: purge,
		}

		err := mgr.Remove(ctx, name, opts)

		mu.Lock()
		results[i] = batchResult{
			name:    name,
			success: err == nil,
			err:     err,
		}
		mu.Unlock()

		if err != nil {
			output.ErrorCross(fmt.Sprintf("卸载失败: %v", err))
		} else {
			output.SuccessCheck(fmt.Sprintf("%s 卸载成功", name))
		}
		fmt.Println()
	}

	// 汇总结果
	return printUninstallResults(results)
}

// printUninstallResults 打印批量卸载结果汇总
func printUninstallResults(results []batchResult) error {
	var successCount, failCount int
	var failedApps []string

	for _, r := range results {
		if r.success {
			successCount++
		} else {
			failCount++
			failedApps = append(failedApps, r.name)
		}
	}

	output.Infoln("========================================")
	output.Infoln("批量卸载完成")
	output.Infoln("========================================")
	output.Successf("成功: %d\n", successCount)
	if failCount > 0 {
		output.Errorf("失败: %d\n", failCount)
		output.Errorln("失败的软件包:")
		for _, name := range failedApps {
			output.Errorf("  - %s\n", name)
		}
		return fmt.Errorf("部分软件包卸载失败")
	}
	output.SuccessCheck("所有软件包卸载完成")
	return nil
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallPurge, "purge", "p", false, "彻底清除，包括配置文件和数据")
	rootCmd.AddCommand(uninstallCmd)
}
