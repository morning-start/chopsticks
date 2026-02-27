package cli

import (
	"fmt"

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
		fmt.Println("已安装的软件包:")
		fmt.Println("--------------")

		apps, err := application.AppManager().ListInstalled()
		if err != nil {
			return cli.Exit(fmt.Sprintf("获取已安装列表失败: %v", err), 1)
		}

		if len(apps) == 0 {
			fmt.Println("  暂无已安装的软件包")
			return nil
		}

		for _, a := range apps {
			fmt.Printf("  %s@%s\n", a.Name, a.Version)
			fmt.Printf("    软件源: %s\n", a.Bucket)
			fmt.Printf("    安装目录: %s\n", a.InstallDir)
			fmt.Println()
		}
	} else {
		// 列出可用软件包
		fmt.Println("可用软件包:")
		fmt.Println("----------")
		fmt.Println("  使用 'chopsticks search <query>' 搜索特定软件包")
		fmt.Println()
		fmt.Println("  常用软件包:")
		fmt.Println("    git        - Distributed version control system")
		fmt.Println("    nodejs     - JavaScript runtime")
		fmt.Println("    python     - Python programming language")
		fmt.Println("    vscode     - Visual Studio Code editor")
	}

	return nil
}
