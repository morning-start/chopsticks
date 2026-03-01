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
	updateForce   bool
	updateAll     bool
	updateAsync   bool
	updateWorkers int
)

// updateCmd 表示 update 命令
var updateCmd = &cobra.Command{
	Use:     "update [package] ...",
	Aliases: []string{"upgrade", "up"},
	Short:   "更新软件包",
	Long: `更新指定的软件包，或使用 --all 更新所有软件包。
支持批量更新多个指定软件包。

示例:
  chopsticks update git
  chopsticks upgrade nodejs --force
  chopsticks update --all
  chopsticks update app1 app2 app3
  chopsticks upgrade git nodejs python`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// 异步模式
	if updateAsync {
		return runUpdateAsync(cmd, args)
	}

	ctx := cmd.Context()
	application := getApp()

	opts := app.UpdateOptions{
		Force: updateForce,
	}

	// 更新所有
	if updateAll {
		output.Infoln("Updating all packages...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			output.ErrorCrossf("Update failed: %v", err)
			return err
		}
		output.SuccessCheck("All packages updated successfully")
		return nil
	}

	// 没有参数时显示错误
	if len(args) < 1 {
		return fmt.Errorf("missing package name. Usage: chopsticks update [package ...] [--all]")
	}

	// 单个包直接更新
	if len(args) == 1 {
		return updateSingle(ctx, application.AppManager(), args[0], opts)
	}

	// 批量更新
	return updateBatch(ctx, application.AppManager(), args, opts)
}

// updateSingle 更新单个软件包
func updateSingle(ctx context.Context, mgr app.AppManager, pkgName string, opts app.UpdateOptions) error {
	output.Infof("Updating %s...\n", pkgName)
	if err := mgr.Update(ctx, pkgName, opts); err != nil {
		output.ErrorCrossf("Update failed: %v", err)
		return err
	}

	output.SuccessCheckf("%s updated successfully", pkgName)
	return nil
}

// updateBatch 批量更新软件包
func updateBatch(ctx context.Context, mgr app.AppManager, packages []string, opts app.UpdateOptions) error {
	total := len(packages)

	output.Infoln("========================================")
	output.Infof("Starting batch update of %d packages\n", total)
	output.Infoln("========================================")
	fmt.Println()

	results := make([]batchResult, total)
	var mu sync.Mutex

	for i, name := range packages {
		output.Infof("[%d/%d] ", i+1, total)
		output.Infof("Updating %s...\n", name)

		err := mgr.Update(ctx, name, opts)

		mu.Lock()
		results[i] = batchResult{
			name:    name,
			success: err == nil,
			err:     err,
		}
		mu.Unlock()

		if err != nil {
			output.ErrorCrossf("Update failed: %v", err)
		} else {
			output.SuccessCheckf("%s updated successfully", name)
		}
		fmt.Println()
	}

	// 汇总结果
	return printUpdateResults(results)
}

// printUpdateResults 打印批量更新结果汇总
func printUpdateResults(results []batchResult) error {
	var successCount, failCount int
	var failedApps []string

	for _, r := range results {
		if r.success {
			successCount++
			continue
		}
		failCount++
		failedApps = append(failedApps, r.name)
	}

	output.Infoln("========================================")
	output.Infoln("Batch update completed")
	output.Infoln("========================================")
	output.Successf("Success: %d\n", successCount)
	if failCount == 0 {
		output.SuccessCheck("All packages updated")
		return nil
	}
	output.Errorf("Failed: %d\n", failCount)
	output.Errorln("Failed packages:")
	for _, name := range failedApps {
		output.Errorf("  - %s\n", name)
	}
	return fmt.Errorf("some packages failed to update")
}

func init() {
	updateCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "更新所有已安装的软件包")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "强制更新，即使版本相同")
	updateCmd.Flags().BoolVar(&updateAsync, "async", false, "使用异步模式更新（并行更新多个包）")
	updateCmd.Flags().IntVarP(&updateWorkers, "workers", "w", defaultWorkers, "异步模式下的最大并发数")

	rootCmd.AddCommand(updateCmd)
}
