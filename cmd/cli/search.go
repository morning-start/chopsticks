package cli

import (
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// searchCommand 返回 search 命令定义。
func searchCommand() *cli.Command {
	return &cli.Command{
		Name:      "search",
		Aliases:   []string{"find", "s"},
		Usage:     "搜索软件包",
		ArgsUsage: "<query>",
		Description: `在软件源中搜索软件包。

示例:
  chopsticks search git
  chopsticks find editor
  chopsticks search node --bucket main`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "bucket",
				Aliases: []string{"b"},
				Usage:   "指定软件源进行搜索",
			},
			&cli.BoolFlag{
				Name:  "async",
				Usage: "使用异步模式搜索（并行搜索多个软件源）",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "异步模式下的最大并发数",
				Value:   10,
			},
		},
		Action: searchAction,
	}
}

// searchAction 处理搜索命令。
func searchAction(c *cli.Context) error {
	// 异步模式
	if c.Bool("async") {
		return searchAsyncAction(c)
	}

	if c.NArg() < 1 {
		output.Errorln("错误: 缺少搜索关键词")
		output.Dimln("用法: chopsticks search <query>")
		return cli.Exit("", 1)
	}

	query := c.Args().First()
	bucketName := c.String("bucket")

	output.Info("搜索: ")
	output.Highlightln(query)
	if bucketName != "" {
		output.Dim("软件源: ")
		output.Infoln(bucketName)
	}

	ctx := getContext(c)
	application := getApp()

	// 调用 app manager 搜索
	results, err := application.AppManager().Search(ctx, query, bucketName)
	if err != nil {
		output.ErrorCrossf("搜索失败: %v", err)
		return cli.Exit("", 1)
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
