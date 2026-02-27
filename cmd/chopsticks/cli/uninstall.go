package cli

import (
	"fmt"

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
		ArgsUsage: "<package>",
		Description: `卸载指定的软件包。

示例:
  chopsticks uninstall git
  chopsticks remove nodejs
  chopsticks rm python --purge`,
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

// uninstallAction 处理卸载命令。
func uninstallAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks uninstall <package>")
		return cli.Exit("", 1)
	}

	name := c.Args().First()
	purge := c.Bool("purge")

	output.Info("正在卸载 ")
	output.Highlight(name)
	if purge {
		output.Warning(" (彻底清除)")
	}
	fmt.Println()

	ctx := getContext(c)
	application := getApp()

	opts := app.RemoveOptions{
		Purge: purge,
	}

	if err := application.AppManager().Remove(ctx, name, opts); err != nil {
		output.ErrorCross(fmt.Sprintf("卸载失败: %v", err))
		return cli.Exit("", 1)
	}

	output.SuccessCheck(fmt.Sprintf("%s 卸载成功", name))
	return nil
}
