package cli

import (
	"fmt"

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
		},
		Action: searchAction,
	}
}

// searchAction 处理搜索命令。
func searchAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少搜索关键词\n用法: chopsticks search <query>", 1)
	}

	query := c.Args().First()
	bucketName := c.String("bucket")

	fmt.Printf("搜索: %s\n", query)
	if bucketName != "" {
		fmt.Printf("软件源: %s\n", bucketName)
	}
	fmt.Println()

	ctx := getContext(c)
	application := getApp()

	// 调用 app manager 搜索
	results, err := application.AppManager().Search(ctx, query, bucketName)
	if err != nil {
		return cli.Exit(fmt.Sprintf("搜索失败: %v", err), 1)
	}

	// 显示结果
	fmt.Println("搜索结果:")
	fmt.Println("-----------")

	if len(results) == 0 {
		fmt.Println("未找到匹配的应用")
		return nil
	}

	for _, result := range results {
		fmt.Printf("  %s\n", result.App.Name)
		if result.App.Description != "" {
			fmt.Printf("    描述: %s\n", result.App.Description)
		}
		fmt.Printf("    版本: %s\n", result.App.Version)
		fmt.Printf("    软件源: %s\n", result.Bucket)
		fmt.Println()
	}

	return nil
}
