package cli

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// uninstallCommand 返回 uninstall 命令定义。
func uninstallCommand() *cli.Command {
	return &cli.Command{
		Name:      "uninstall",
		Aliases:   []string{"remove", "rm"},
		Usage:     "卸载软件包",
		ArgsUsage: "<package> ...",
		Description: `卸载指定的软件包。支持批量卸载多个软件包。

示例:
  chopsticks uninstall git
  chopsticks remove nodejs
  chopsticks rm python --purge
  chopsticks uninstall app1 app2 app3
  chopsticks rm --purge app1 app2`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "purge",
				Aliases: []string{"p"},
				Usage:   "彻底清除，包括配置文件和数据",
			},
		},
		Action: uninstallAction,
	}
}

// uninstallAction 处理卸载命令（支持批量卸载）。
func uninstallAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks uninstall <package> ...")
		return cli.Exit("", 1)
	}

	purge := c.Bool("purge")

	ctx := getContextFromCli(c)
	application := getApp()

	// 获取所有要卸载的包
	packages := make([]string, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		packages[i] = c.Args().Get(i)
	}

	// 单个包直接卸载
	if len(packages) == 1 {
		return uninstallSingle(ctx, application.AppManager(), packages[0], purge)
	}

	// 批量卸载
	return uninstallBatch(ctx, application.AppManager(), packages, purge)
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
		return cli.Exit("", 1)
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
		return cli.Exit("", 1)
	}
	output.SuccessCheck("所有软件包卸载完成")
	return nil
}
