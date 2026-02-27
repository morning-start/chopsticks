package cli

import (
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
		output.Infoln("正在更新所有软件包...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			output.ErrorCrossf("更新失败: %v", err)
			return cli.Exit("", 1)
		}
		output.SuccessCheck("所有软件包更新成功")
		return nil
	}

	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks update [package] [--all]")
		return cli.Exit("", 1)
	}

	pkgName := c.Args().First()

	output.Infof("正在更新 %s...\n", pkgName)
	if err := application.AppManager().Update(ctx, pkgName, opts); err != nil {
		output.ErrorCrossf("更新失败: %v", err)
		return cli.Exit("", 1)
	}

	output.SuccessCheckf("%s 更新成功", pkgName)
	return nil
}
