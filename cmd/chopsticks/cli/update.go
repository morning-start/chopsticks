package cli

import (
	"fmt"
	"os"

	"chopsticks/core/app"

	"github.com/urfave/cli/v2"
)

// updateCommand 返回 update 命令定义。
func updateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Aliases:   []string{"upgrade", "up"},
		Usage:     "更新软件包",
		ArgsUsage: "[package]",
		Description: `更新指定的软件包，或使用 --all 更新所有软件包。

示例:
  chopsticks update git
  chopsticks upgrade nodejs --force
  chopsticks update --all`,
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

// updateAction 处理更新命令。
func updateAction(c *cli.Context) error {
	force := c.Bool("force")
	updateAll := c.Bool("all")

	ctx := getContext(c)
	application := getApp()

	opts := app.UpdateOptions{
		Force: force,
	}

	if updateAll {
		fmt.Println("正在更新所有软件包...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			fmt.Fprintf(os.Stderr, "✗ 更新失败: %v\n", err)
			return cli.Exit("", 1)
		}
		fmt.Println("✓ 所有软件包更新成功")
		return nil
	}

	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少软件包名称\n用法: chopsticks update [package] [--all]", 1)
	}

	pkgName := c.Args().First()

	fmt.Printf("正在更新 %s...\n", pkgName)
	if err := application.AppManager().Update(ctx, pkgName, opts); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 更新失败: %v\n", err)
		return cli.Exit("", 1)
	}

	fmt.Printf("✓ %s 更新成功\n", pkgName)
	return nil
}
