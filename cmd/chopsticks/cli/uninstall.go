package cli

import (
	"fmt"
	"os"

	"chopsticks/core/app"

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
		return cli.Exit("错误: 缺少软件包名称\n用法: chopsticks uninstall <package>", 1)
	}

	name := c.Args().First()
	purge := c.Bool("purge")

	fmt.Printf("正在卸载 %s", name)
	if purge {
		fmt.Print(" (彻底清除)")
	}
	fmt.Println("...")

	ctx := getContext(c)
	application := getApp()

	opts := app.RemoveOptions{
		Purge: purge,
	}

	if err := application.AppManager().Remove(ctx, name, opts); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 卸载失败: %v\n", err)
		return cli.Exit("", 1)
	}

	fmt.Printf("✓ %s 卸载成功\n", name)
	return nil
}
