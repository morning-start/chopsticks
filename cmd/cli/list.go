package cli

import (
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// listCommand 返回 list 命令定义。
func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Usage:     "列出软件包",
		ArgsUsage: " ",
		Description: `列出已安装的软件包或可用软件包。

示例:
  chopsticks list
  chopsticks list --installed
  chopsticks ls`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "installed",
				Aliases: []string{"i"},
				Usage:   "只显示已安装的软件包",
			},
		},
		Action: listAction,
	}
}

// listAction 处理列表命令。
func listAction(c *cli.Context) error {
	installed := c.Bool("installed")
	application := getApp()

	if installed {
		// 列出已安装的软件包
		output.Highlightln("已安装的软件包:")
		output.Dimln("--------------")

		apps, err := application.AppManager().ListInstalled()
		if err != nil {
			output.ErrorCrossf("获取已安装列表失败: %v", err)
			return cli.Exit("", 1)
		}

		if len(apps) == 0 {
			output.Dimln("  (暂无已安装的软件包)")
			return nil
		}

		for _, a := range apps {
			output.Successf("  %s@%s\n", a.Name, a.Version)
			output.Dimf("    软件源: %s\n", a.Bucket)
			output.Dimf("    安装目录: %s\n", a.InstallDir)
		}
	} else {
		// 列出可用软件包
		output.Highlightln("可用软件包:")
		output.Dimln("----------")
		output.Infoln("  使用 'chopsticks search <query>' 搜索特定软件包")
		output.Highlightln("\n  常用软件包:")
		output.Item("git", "Distributed version control system")
		output.Item("nodejs", "JavaScript runtime")
		output.Item("python", "Python programming language")
		output.Item("vscode", "Visual Studio Code editor")
	}

	return nil
}
