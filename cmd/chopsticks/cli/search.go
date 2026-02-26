package cli

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/app"
)

// SearchCommand 处理搜索命令。
func SearchCommand(ctx context.Context, application app.Application, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks search <query> [--bucket <bucket>]")
		return fmt.Errorf("缺少搜索关键词")
	}

	query := args[0]
	bucketName := ""

	// 解析 --bucket 参数
	for i, arg := range args[1:] {
		if arg == "--bucket" && i+1 < len(args[1:]) {
			bucketName = args[1:][i+1]
			break
		}
	}

	fmt.Printf("搜索: %s\n", query)
	if bucketName != "" {
		fmt.Printf("软件源: %s\n", bucketName)
	}
	fmt.Println()

	// TODO: 调用 app manager 搜索
	// results, err := application.AppManager().Search(ctx, query, bucketName)
	// if err != nil {
	//     return fmt.Errorf("搜索失败: %w", err)
	// }

	// 显示结果
	fmt.Println("搜索结果:")
	fmt.Println("-----------")

	// 模拟搜索结果
	fmt.Printf("  git\n")
	fmt.Printf("    描述: Distributed version control system\n")
	fmt.Printf("    版本: 2.43.0\n")
	fmt.Printf("    软件源: main\n")
	fmt.Println()

	return nil
}
