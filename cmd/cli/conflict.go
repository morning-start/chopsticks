package cli

import (
	"fmt"

	"chopsticks/core/conflict"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	conflictBucket  string
	conflictVerbose bool
)

// conflictCmd 表示 conflict 命令
var conflictCmd = &cobra.Command{
	Use:     "conflict <package>",
	Aliases: []string{"check"},
	Short:   "检查应用安装冲突",
	Long: `检查指定应用安装时可能产生的冲突。

支持检测的冲突类型:
  - 文件冲突: 安装目录、shim 文件冲突
  - 端口冲突: 应用常用端口被占用
  - 环境变量冲突: 环境变量被其他应用设置
  - 注册表冲突: 注册表项已存在

示例:
  chopsticks conflict git
  chopsticks check nodejs
  chopsticks conflict nginx --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runConflict,
}

func runConflict(cmd *cobra.Command, args []string) error {
	name := args[0]
	application := getApp()

	ctx := cmd.Context()

	output.Info("正在检查 ")
	output.Highlight("%s", name)
	output.Dimln(" 的冲突...")
	fmt.Println()

	// 获取应用信息
	appInfo, err := application.AppManager().Info(ctx, conflictBucket, name)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("获取应用信息失败: %v", err))
		return err
	}

	// 显示应用基本信息
	if conflictVerbose {
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
	app, err := bucketMgr.GetApp(ctx, conflictBucket, name)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("获取应用详情失败: %v", err))
		return err
	}

	// 创建冲突检测器
	storage := application.Storage()
	installDir := application.Config().AppsPath
	detector := conflict.NewDetector(storage, installDir)

	// 执行检测
	result, err := detector.Detect(ctx, app)
	if err != nil {
		output.ErrorCross(fmt.Sprintf("冲突检测失败: %v", err))
		return err
	}

	// 显示结果
	formatter := conflict.NewFormatter(true)
	fmt.Println(formatter.Format(result))

	// 返回退出码
	if result.HasCritical {
		return fmt.Errorf("检测到严重冲突")
	}
	if result.HasWarning {
		return fmt.Errorf("检测到警告级别冲突")
	}

	output.SuccessCheck("未检测到冲突，可以安全安装")
	return nil
}

func init() {
	conflictCmd.Flags().StringVarP(&conflictBucket, "bucket", "b", "main", "指定软件源")
	conflictCmd.Flags().BoolVarP(&conflictVerbose, "verbose", "v", false, "显示详细信息")
	rootCmd.AddCommand(conflictCmd)
}
