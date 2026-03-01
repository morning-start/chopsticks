package cli

import (
	"fmt"

	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	searchBucket  string
	searchAsync   bool
	searchWorkers int
)

// searchCmd 表示 search 命令
var searchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"find", "s"},
	Short:   "搜索软件包",
	Long: `在软件源中搜索软件包。

示例:
  chopsticks search git
  chopsticks find editor
  chopsticks search node --bucket main`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	// 异步模式
	if searchAsync {
		return runSearchAsync(cmd, args)
	}

	query := args[0]

	output.Info("搜索: ")
	output.Highlightln(query)
	if searchBucket != "" {
		output.Dim("软件源: ")
		output.Infoln(searchBucket)
	}

	ctx := cmd.Context()
	application := getApp()

	// 调用 app manager 搜索
	results, err := application.AppManager().Search(ctx, query, searchBucket)
	if err != nil {
		output.ErrorCrossf("搜索失败: %v", err)
		return fmt.Errorf("搜索失败: %w", err)
	}

	// 显示结果
	output.Highlightln("\n搜索结果:")
	output.Dimln("-----------")

	if len(results) == 0 {
		output.Warningln("未找到匹配的应用")
		return nil
	}

	for _, result := range results {
		output.Success("%s", result.App.Name)
		if result.App.Description != "" {
			output.Dimf("    描述: %s\n", result.App.Description)
		}
		output.Dimf("    版本: %s\n", result.App.Version)
		output.Dimf("    软件源: %s\n", result.Bucket)
	}

	return nil
}

func init() {
	searchCmd.Flags().StringVarP(&searchBucket, "bucket", "b", "", "指定软件源进行搜索")
	searchCmd.Flags().BoolVar(&searchAsync, "async", false, "使用异步模式搜索（并行搜索多个软件源）")
	searchCmd.Flags().IntVarP(&searchWorkers, "workers", "w", 10, "异步模式下的最大并发数")
	rootCmd.AddCommand(searchCmd)
}
