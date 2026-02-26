package cli

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/app"
)

// ServeCommand 处理安装命令。
func ServeCommand(ctx context.Context, application app.Application, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks install <package>[@version]")
		return fmt.Errorf("缺少软件包名称")
	}

	spec := args[0]

	// 解析软件包名称和版本
	name, version := parseAppSpec(spec)

	fmt.Printf("正在安装 %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	fmt.Println("...")

	// 调用 app manager 安装
	opts := app.InstallOptions{
		Arch:  "amd64", // 默认架构
		Force: false,
	}

	installSpec := app.InstallSpec{
		Bucket:  "main", // 默认软件源
		Name:    name,
		Version: version,
	}

	if err := application.AppManager().Install(ctx, installSpec, opts); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	fmt.Printf("✓ %s 安装成功\n", name)
	return nil
}

// parseAppSpec 解析软件包规格（name@version）。
func parseAppSpec(spec string) (name, version string) {
	for i := len(spec) - 1; i >= 0; i-- {
		if spec[i] == '@' {
			return spec[:i], spec[i+1:]
		}
	}
	return spec, ""
}
