package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"chopsticks/core/app"
)

// RefreshCommand 处理更新命令。
func RefreshCommand(ctx context.Context, application app.Application, args []string) error {
	force := false
	updateAll := false
	var pkgName string

	// 解析参数
	for _, arg := range args {
		switch arg {
		case "--all":
			updateAll = true
		case "--force":
			force = true
		default:
			if pkgName == "" && !strings.HasPrefix(arg, "-") {
				pkgName = arg
			}
		}
	}

	opts := app.UpdateOptions{
		Force: force,
	}

	if updateAll {
		fmt.Println("正在更新所有软件包...")
		if err := application.AppManager().UpdateAll(ctx, opts); err != nil {
			return fmt.Errorf("更新失败: %w", err)
		}
		fmt.Println("✓ 所有软件包更新成功")
		return nil
	}

	if pkgName == "" {
		fmt.Fprintln(os.Stderr, "用法: chopsticks update [package] [--all] [--force]")
		return fmt.Errorf("缺少软件包名称或 --all 标志")
	}

	fmt.Printf("正在更新 %s...\n", pkgName)
	if err := application.AppManager().Update(ctx, pkgName, opts); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	fmt.Printf("✓ %s 更新成功\n", pkgName)
	return nil
}
