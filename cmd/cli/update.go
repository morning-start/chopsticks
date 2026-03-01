package cli

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// updateCommand 返回 update 命令定义。
func updateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Aliases:   []string{"upgrade", "up"},
		Usage:     "更新软件包",
		ArgsUsage: "[package] ...",
		Description: `更新指定的软件包，或使用 --all 更新所有软件包。
支持批量更新多个指定软件包。

示例:
  chopsticks update git
  chopsticks upgrade nodejs --force
  chopsticks update --all
  chopsticks update app1 app2 app3
  chopsticks upgrade git nodejs python`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "更新所有已安装的软件包",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "强制更新，即使版本相同",
			},
			&cli.BoolFlag{
				Name:  "async",
				Usage: "使用异步模式更新（并行更新多个包）",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "Max concurrency for async mode",
				Value:   defaultWorkers,
			},
		},
		Action: updateAction,
	}
}

// updateAction 处理更新命令（支持批量更新）。
func updateAction(c *cli.Context) error {
	// 异步模式
	if c.Bool("async") {
		return updateAsyncAction(c)
	}

	force := c.Bool("force")
	updateAll := c.Bool("all")

	ctx := getContext(c)
	application := getApp()

	opts := app.UpdateOptions{
		Force: force,
	}

	// 更新所有
	if updateAll {
		output.Infoln("Updating all packages...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			output.ErrorCrossf("Update failed: %v", err)
			return cli.Exit("", 1)
		}
		output.SuccessCheck("All packages updated successfully")
		return nil
	}

	// 没有参数时显示错误
	if c.NArg() < 1 {
		output.Errorln("Error: missing package name")
		output.Dimln("Usage: chopsticks update [package ...] [--all]")
		return cli.Exit("", 1)
	}

	// 获取所有要更新的包
	packages := make([]string, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		packages[i] = c.Args().Get(i)
	}

	// 单个包直接更新
	if len(packages) == 1 {
		return updateSingle(ctx, application.AppManager(), packages[0], opts)
	}

	// 批量更新
	return updateBatch(ctx, application.AppManager(), packages, opts)
}

// updateSingle 更新单个软件包
func updateSingle(ctx context.Context, mgr app.AppManager, pkgName string, opts app.UpdateOptions) error {
	output.Infof("Updating %s...\n", pkgName)
	if err := mgr.Update(ctx, pkgName, opts); err != nil {
		output.ErrorCrossf("Update failed: %v", err)
		return cli.Exit("", 1)
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
	return cli.Exit("", 1)
}
