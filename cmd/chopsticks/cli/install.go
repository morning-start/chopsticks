package cli

import (
	"fmt"
	"os"

	"chopsticks/core/app"

	"github.com/urfave/cli/v2"
)

// installCommand 返回 install 命令定义。
func installCommand() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Aliases:   []string{"i"},
		Usage:     "安装软件包",
		ArgsUsage: "<package>[@version]",
		Description: `安装指定的软件包。可以指定版本号，格式为 package@version。

示例:
  chopsticks install git
  chopsticks install nodejs@18.17.0
  chopsticks install git --force`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "强制重新安装，即使已存在",
			},
			&cli.StringFlag{
				Name:    "arch",
				Aliases: []string{"a"},
				Usage:   "指定架构 (amd64, x86, arm64)",
				Value:   "amd64",
			},
			&cli.StringFlag{
				Name:    "bucket",
				Aliases: []string{"b"},
				Usage:   "指定软件源",
				Value:   "main",
			},
		},
		Action: installAction,
	}
}

// installAction 处理安装命令。
func installAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少软件包名称\n用法: chopsticks install <package>[@version]", 1)
	}

	spec := c.Args().First()
	name, version := parseAppSpec(spec)

	force := c.Bool("force")
	arch := c.String("arch")
	bucket := c.String("bucket")

	fmt.Printf("正在安装 %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	if force {
		fmt.Print(" (强制)")
	}
	fmt.Println("...")

	ctx := getContext(c)
	application := getApp()

	opts := app.InstallOptions{
		Arch:  arch,
		Force: force,
	}

	installSpec := app.InstallSpec{
		Bucket:  bucket,
		Name:    name,
		Version: version,
	}

	if err := application.AppManager().Install(ctx, installSpec, opts); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 安装失败: %v\n", err)
		return cli.Exit("", 1)
	}

	fmt.Printf("✓ %s 安装成功\n", name)
	return nil
}

// parseAppSpec 解析软件包规格（name@version）。
func parseAppSpec(spec string) (name, version string) {
	// 从后向前查找 @，以支持名称中包含 @ 的情况
	for i := len(spec) - 1; i >= 0; i-- {
		if spec[i] == '@' {
			return spec[:i], spec[i+1:]
		}
	}
	return spec, ""
}
