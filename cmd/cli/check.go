package cli

import (
	"context"
	"fmt"
	"os"

	"chopsticks/core/app"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	checkQuiet   bool
	checkVerbose bool
)

var checkCmd = &cobra.Command{
	Use:   "check [app]",
	Short: "检查已安装应用的完整性",
	Long: `检查已安装应用的完整性，验证：
- PATH 环境变量中的路径是否存在
- 环境变量是否正确设置
- 符号链接是否有效
- 文件是否存在

如果不指定应用名称，则检查所有已安装应用。`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVarP(&checkQuiet, "quiet", "q", false, "静默模式，仅返回退出码")
	checkCmd.Flags().BoolVarP(&checkVerbose, "verbose", "v", false, "详细输出")
}

func runCheck(_ *cobra.Command, args []string) error {
	ctx := getContext()
	application := getApp()
	if application == nil {
		return fmt.Errorf("application not initialized")
	}

	appMgr := application.AppManager()
	if appMgr == nil {
		return fmt.Errorf("app manager not available")
	}

	opts := app.CheckOptions{
		Verbose: checkVerbose,
	}

	if len(args) == 0 {
		return checkAllApps(ctx, appMgr, opts)
	}

	appName := args[0]
	return checkSingleApp(ctx, appMgr, appName, opts)
}

func checkAllApps(ctx context.Context, appMgr app.AppManager, opts app.CheckOptions) error {
	installedApps, err := appMgr.ListInstalled(ctx)
	if err != nil {
		return fmt.Errorf("列出已安装应用失败: %w", err)
	}

	if len(installedApps) == 0 {
		if !checkQuiet {
			output.Infoln("没有已安装的应用")
		}
		return nil
	}

	successCount := 0
	failCount := 0

	for _, installedApp := range installedApps {
		result, err := appMgr.Check(ctx, installedApp.Name, opts)
		if err != nil {
			failCount++
			if !checkQuiet {
				output.ErrorCrossf("%s - 检查失败: %v", installedApp.Name, err)
			}
			continue
		}

		if result.Status == app.CheckStatusPassed {
			successCount++
			if !checkQuiet {
				output.SuccessCheckf("%s - 所有检查通过", result.Name)
			}
		} else {
			failCount++
			if !checkQuiet {
				output.ErrorCrossf("%s - 检查失败", result.Name)
				for _, issue := range result.Issues {
					fmt.Printf("  - [%s] %s: %s\n", issue.Type, issue.Message, issue.Target)
				}
			}
		}
	}

	if !checkQuiet {
		fmt.Println()
		fmt.Printf("检查完成: %d 成功, %d 失败\n", successCount, failCount)
	}

	if failCount > 0 {
		return os.ErrPermission
	}

	return nil
}

func checkSingleApp(ctx context.Context, appMgr app.AppManager, appName string, opts app.CheckOptions) error {
	result, err := appMgr.Check(ctx, appName, opts)
	if err != nil {
		return fmt.Errorf("检查应用 %s 失败: %w", appName, err)
	}

	if result.Status == app.CheckStatusPassed {
		if !checkQuiet {
			output.SuccessCheckf("%s - 所有检查通过", result.Name)
		}
		return nil
	}

	if !checkQuiet {
		output.ErrorCrossf("%s - 检查失败", result.Name)
		for _, issue := range result.Issues {
			switch issue.Type {
			case app.IssueTypePath:
				fmt.Printf("  - [PATH] %s: %s\n", issue.Message, issue.Target)
			case app.IssueTypeEnv:
				fmt.Printf("  - [ENV] %s: %s\n", issue.Message, issue.Target)
			case app.IssueTypeSymlink:
				fmt.Printf("  - [SYMLINK] %s: %s\n", issue.Message, issue.Target)
			case app.IssueTypeFile:
				fmt.Printf("  - [FILE] %s: %s\n", issue.Message, issue.Target)
			case app.IssueTypeRegistry:
				fmt.Printf("  - [REGISTRY] %s: %s\n", issue.Message, issue.Target)
			default:
				fmt.Printf("  - [%s] %s: %s\n", issue.Type, issue.Message, issue.Target)
			}
		}
	}

	return os.ErrPermission
}
