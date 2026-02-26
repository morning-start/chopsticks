package cli

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/app"
)

// ClearCommand 处理卸载命令。
func ClearCommand(ctx context.Context, application app.Application, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks uninstall <package> [--purge]")
		return fmt.Errorf("缺少软件包名称")
	}

	name := args[0]
	purge := false

	// 检查 --purge 标志
	for _, arg := range args[1:] {
		if arg == "--purge" {
			purge = true
		}
	}

	fmt.Printf("正在卸载 %s", name)
	if purge {
		fmt.Print(" (彻底清除)")
	}
	fmt.Println("...")

	// 使用 app.RemoveOptions
	opts := app.RemoveOptions{
		Purge: purge,
	}

	if err := application.AppManager().Remove(ctx, name, opts); err != nil {
		return fmt.Errorf("卸载失败: %w", err)
	}

	fmt.Printf("✓ %s 卸载成功\n", name)
	return nil
}
