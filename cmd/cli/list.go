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
			output.ErrorCrossf("获取已安装列表失败：%v", err)
			return fmt.Errorf("获取已安装软件包列表失败：%w", err)
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

		ctx := cmd.Context()

		// 获取所有 bucket
		buckets, err := application.BucketManager().ListBuckets(ctx)
		if err != nil {
			output.ErrorCrossf("获取软件源列表失败：%v", err)
			return fmt.Errorf("获取软件源列表失败：%w", err)
		}

		if len(buckets) == 0 {
			output.Dimln("  (暂无软件源)")
			output.Infoln("\n使用 'chopsticks bucket add <name> <url>' 添加软件源")
			return nil
		}

		// 从所有 bucket 中获取应用列表
		totalApps := 0
		for _, bucketName := range buckets {
			apps, err := application.BucketManager().ListApps(ctx, bucketName)
			if err != nil {
				continue
			}

			if len(apps) > 0 {
				output.Successf("\n  %s:\n", bucketName)
				for name, app := range apps {
					if app != nil {
						output.Item(name, app.Description)
						totalApps++
					}
				}
			}
		}

		if totalApps == 0 {
			output.Dimln("  (暂无可用软件包)")
		}

		output.Dimln("\n  使用 'chopsticks search <query>' 搜索特定软件包")
	}

	return nil
}

func init() {
	listCmd.Flags().BoolVarP(&listInstalled, "installed", "i", false, "只显示已安装的软件包")
	rootCmd.AddCommand(listCmd)
}
