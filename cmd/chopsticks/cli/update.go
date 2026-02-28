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
		},
		Action: updateAction,
	}
}

// updateAction 处理更新命令（支持批量更新）。
func updateAction(c *cli.Context) error {
	force := c.Bool("force")
	updateAll := c.Bool("all")

	ctx := getContext(c)
	application := getApp()

	opts := app.UpdateOptions{
		Force: force,
	}

	// 更新所有
	if updateAll {
		output.Infoln("正在更新所有软件包...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			output.ErrorCrossf("更新失败: %v", err)
			return cli.Exit("", 1)
		}
		output.SuccessCheck("所有软件包更新成功")
		return nil
	}

	// 没有参数时显示错误
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks update [package ...] [--all]")
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
func updateSingle(ctx context.Context, mgr app.Manager, pkgName string, opts app.UpdateOptions) error {
	output.Infof("正在更新 %s...\n", pkgName)
	if err := mgr.Update(ctx, pkgName, opts); err != nil {
		output.ErrorCrossf("更新失败: %v", err)
		return cli.Exit("", 1)
	}

	output.SuccessCheckf("%s 更新成功", pkgName)
	return nil
}

// updateBatch 批量更新软件包
func updateBatch(ctx context.Context, mgr app.Manager, packages []string, opts app.UpdateOptions) error {
	total := len(packages)

	output.Infoln("========================================")
	output.Infof("开始批量更新 %d 个软件包\n", total)
	output.Infoln("========================================")
	fmt.Println()

	results := make([]batchResult, total)
	var mu sync.Mutex

	for i, name := range packages {
		output.Infof("[%d/%d] ", i+1, total)
		output.Infof("正在更新 %s...\n", name)

		err := mgr.Update(ctx, name, opts)

		mu.Lock()
		results[i] = batchResult{
			name:    name,
			success: err == nil,
			err:     err,
		}
		mu.Unlock()

		if err != nil {
			output.ErrorCrossf("更新失败: %v", err)
		} else {
			output.SuccessCheckf("%s 更新成功", name)
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
		} else {
			failCount++
			failedApps = append(failedApps, r.name)
		}
	}

	output.Infoln("========================================")
	output.Infoln("批量更新完成")
	output.Infoln("========================================")
	output.Successf("成功: %d\n", successCount)
	if failCount > 0 {
		output.Errorf("失败: %d\n", failCount)
		output.Errorln("失败的软件包:")
		for _, name := range failedApps {
			output.Errorf("  - %s\n", name)
		}
		return cli.Exit("", 1)
	}
	output.SuccessCheck("所有软件包更新完成")
	return nil
}
