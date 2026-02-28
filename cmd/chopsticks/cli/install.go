package cli

import (
	"context"
	"fmt"
	"sync"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// installCommand 返回 install 命令定义。
func installCommand() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Aliases:   []string{"i"},
		Usage:     "安装软件包",
		ArgsUsage: "<package>[@version] ...",
		Description: `安装指定的软件包。可以指定版本号，格式为 package@version。
支持批量安装多个软件包。

示例:
  chopsticks install git
  chopsticks install nodejs@18.17.0
  chopsticks install git --force
  chopsticks install app1 app2 app3
  chopsticks install git@2.40 nodejs@18 python@3.11`,
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

// installAction 处理安装命令（支持批量安装）。
func installAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks install <package>[@version] ...")
		return cli.Exit("", 1)
	}

	force := c.Bool("force")
	arch := c.String("arch")
	bucket := c.String("bucket")

	ctx := getContext(c)
	application := getApp()

	// 获取所有要安装的包
	packages := make([]struct {
		name    string
		version string
		spec    string
	}, c.NArg())

	for i := 0; i < c.NArg(); i++ {
		spec := c.Args().Get(i)
		name, version := parseAppSpec(spec)
		packages[i] = struct {
			name    string
			version string
			spec    string
		}{name: name, version: version, spec: spec}
	}

	// 单个包直接安装
	if len(packages) == 1 {
		return installSingle(ctx, application.AppManager(), packages[0].name, packages[0].version, bucket, arch, force)
	}

	// 批量安装
	return installBatch(ctx, application.AppManager(), packages, bucket, arch, force)
}

// installSingle 安装单个软件包
func installSingle(ctx context.Context, mgr app.Manager, name, version, bucket, arch string, force bool) error {
	output.Info("正在安装 ")
	output.Highlight("%s", name)
	if version != "" {
		output.Dim("@")
		output.Info("%s", version)
	}
	if force {
		output.Warning(" (强制)")
	}
	fmt.Println()

	opts := app.InstallOptions{
		Arch:  arch,
		Force: force,
	}

	installSpec := app.InstallSpec{
		Bucket:  bucket,
		Name:    name,
		Version: version,
	}

	if err := mgr.Install(ctx, installSpec, opts); err != nil {
		output.ErrorCross(fmt.Sprintf("安装失败: %v", err))
		return cli.Exit("", 1)
	}

	output.SuccessCheck(fmt.Sprintf("%s 安装成功", name))
	return nil
}

// installBatch 批量安装软件包
func installBatch(ctx context.Context, mgr app.Manager, packages []struct {
	name    string
	version string
	spec    string
}, bucket, arch string, force bool) error {
	total := len(packages)

	output.Infoln("========================================")
	output.Infof("开始批量安装 %d 个软件包\n", total)
	output.Infoln("========================================")
	fmt.Println()

	results := make([]batchResult, total)
	var mu sync.Mutex

	for i, pkg := range packages {
		output.Infof("[%d/%d] ", i+1, total)
		output.Info("正在安装 ")
		output.Highlight("%s", pkg.name)
		if pkg.version != "" {
			output.Dim("@")
			output.Info("%s", pkg.version)
		}
		if force {
			output.Warning(" (强制)")
		}
		fmt.Println()

		opts := app.InstallOptions{
			Arch:  arch,
			Force: force,
		}

		installSpec := app.InstallSpec{
			Bucket:  bucket,
			Name:    pkg.name,
			Version: pkg.version,
		}

		err := mgr.Install(ctx, installSpec, opts)

		mu.Lock()
		results[i] = batchResult{
			name:    pkg.name,
			success: err == nil,
			err:     err,
		}
		mu.Unlock()

		if err != nil {
			output.ErrorCross(fmt.Sprintf("安装失败: %v", err))
		} else {
			output.SuccessCheck(fmt.Sprintf("%s 安装成功", pkg.name))
		}
		fmt.Println()
	}

	// 汇总结果
	return printBatchResults(results, "安装")
}

// batchResult 批量操作结果
type batchResult struct {
	name    string
	success bool
	err     error
}

// printBatchResults 打印批量操作结果汇总
func printBatchResults(results []batchResult, operation string) error {
	var successCount, failCount int
	var failedApps []string

	for _, r := range results {
		if r.success {
			successCount++
		} else {
			failCount++
			failedApps = append(failedApps, r.name)
		}
	}

	output.Infoln("========================================")
	output.Infof("批量%s完成\n", operation)
	output.Infoln("========================================")
	output.Successf("成功: %d\n", successCount)
	if failCount > 0 {
		output.Errorf("失败: %d\n", failCount)
		output.Errorln("失败的软件包:")
		for _, name := range failedApps {
			output.Errorf("  - %s\n", name)
		}
		return cli.Exit("", 1)
	}
	output.SuccessCheck("所有软件包处理完成")
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
