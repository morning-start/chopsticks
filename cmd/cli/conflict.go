package cli

import (
	"fmt"

	"chopsticks/core/conflict"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// conflictCommand 返回 conflict 命令定义。
func conflictCommand() *cli.Command {
	return &cli.Command{
		Name:      "conflict",
		Aliases:   []string{"check"},
		Usage:     "检查应用安装冲突",
		ArgsUsage: "<package>",
		Description: `检查指定应用安装时可能产生的冲突。

支持检测的冲突类型:
  - 文件冲突: 安装目录、shim 文件冲突
  - 端口冲突: 应用常用端口被占用
  - 环境变量冲突: 环境变量被其他应用设置
  - 注册表冲突: 注册表项已存在

示例:
  chopsticks conflict git
  chopsticks check nodejs
  chopsticks conflict nginx --verbose`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "bucket",
				Aliases: []string{"b"},
				Usage:   "指定软件源",
				Value:   "main",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "显示详细信息",
			},
		},
		Action: conflictAction,
	}
}

// conflictAction 处理冲突检查命令。
func conflictAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks conflict <package>")
		return cli.Exit("", 1)
	}

	name := c.Args().Get(0)
	bucket := c.String("bucket")
	verbose := c.Bool("verbose")

	ctx := getContext(c)
	application := getApp()

	output.Info("正在检查 ")
	output.Highlight("%s", name)
	output.Dimln(" 的冲突...")
	fmt.Println()

	// 获取应用信息
	appInfo, err := application.AppManager().Info(ctx, bucket, name)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("获取应用信息失败: %v", err))
		return cli.Exit("", 1)
	}

	// 显示应用基本信息
	if verbose {
		output.Infoln("应用信息:")
		output.Infof("  名称: %s\n", appInfo.Name)
		output.Infof("  版本: %s\n", appInfo.Version)
		output.Infof("  软件源: %s\n", appInfo.Bucket)
		if appInfo.Installed {
			output.Warningf("  状态: 已安装 (%s)\n", appInfo.InstalledVersion)
		} else {
			output.Successln("  状态: 未安装")
		}
		fmt.Println()
	}

	// 执行冲突检测
	// 注意：这里我们需要获取完整的 App 对象来进行冲突检测
	// 由于 Manager 接口没有直接暴露获取 App 的方法，我们需要通过 bucket.Manager 获取
	bucketMgr := application.BucketManager()
	app, err := bucketMgr.GetApp(ctx, bucket, name)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("获取应用详情失败: %v", err))
		return cli.Exit("", 1)
	}

	// 创建冲突检测器
	storage := application.Storage()
	installDir := application.Config().AppsPath
	detector := conflict.NewDetector(storage, installDir)

	// 执行检测
	result, err := detector.Detect(ctx, app)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("冲突检测失败: %v", err))
		return cli.Exit("", 1)
	}

	// 显示结果
	formatter := conflict.NewFormatter(true)
	fmt.Println(formatter.Format(result))

	// 返回退出码
	if result.HasCritical {
		return cli.Exit("检测到严重冲突", 2)
	}
	if result.HasWarning {
		return cli.Exit("检测到警告级别冲突", 1)
	}

	output.SuccessCheck("未检测到冲突，可以安全安装")
	return nil
}
