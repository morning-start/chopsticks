package cli

import (
	"chopsticks/pkg/output"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listInstalled bool
)

// listCmd 表示 list 命令
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出软件包",
	Long: `列出已安装的软件包或可用软件包。

示例:
  chopsticks list
  chopsticks list --installed
  chopsticks ls`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	if listInstalled {
		// 列出已安装的软件包
		output.Highlightln("已安装的软件包:")
		output.Dimln("--------------")

		apps, err := application.AppManager().ListInstalled()
		if err != nil {
			output.ErrorCrossf("获取已安装列表失败: %v", err)
			return err
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

func init() {
	listCmd.Flags().BoolVarP(&listInstalled, "installed", "i", false, "只显示已安装的软件包")
	rootCmd.AddCommand(listCmd)
}
