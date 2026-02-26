package cli

import (
	"context"
	"fmt"

	"chopsticks/core/app"
)

// ListCommand 处理列表命令。
func ListCommand(ctx context.Context, application app.Application, args []string) error {
	installed := false

	// 解析 --installed 参数
	for _, arg := range args {
		if arg == "--installed" {
			installed = true
			break
		}
	}

	if installed {
		// 列出已安装的软件包
		fmt.Println("已安装的软件包:")
		fmt.Println("--------------")

		apps, err := application.AppManager().ListInstalled()
		if err != nil {
			return fmt.Errorf("获取已安装列表失败: %w", err)
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
