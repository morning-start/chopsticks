// Package cli 提供依赖管理命令
package cli

import (
	"context"
	"fmt"

	"chopsticks/core/dep"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	depsTree    bool
	depsReverse bool
)

// depsCmd 表示 deps 命令
var depsCmd = &cobra.Command{
	Use:   "deps <app>",
	Short: "查看应用依赖",
	Long: `查看指定应用的依赖关系，包括依赖列表、依赖树和反向依赖。

示例:
  chopsticks deps git
  chopsticks deps git --tree
  chopsticks deps git --reverse`,
	Args: cobra.ExactArgs(1),
	RunE: runDeps,
}

func init() {
	// deps 命令标志
	depsCmd.Flags().BoolVarP(&depsTree, "tree", "t", false, "以树形结构显示依赖")
	depsCmd.Flags().BoolVarP(&depsReverse, "reverse", "r", false, "显示反向依赖（谁依赖我）")
}

func runDeps(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	appName := args[0]

	// 创建依赖管理器
	depMgr, err := dep.NewDependencyManager(
		application.BucketManager(),
		application.Storage(),
		application.Config().PersistDir,
	)
	if err != nil {
		return fmt.Errorf("创建依赖管理器失败：%w", err)
	}

	// 查看依赖树
	if depsTree {
		return showDepsTree(ctx, depMgr, appName)
	}

	// 查看反向依赖
	if depsReverse {
		return showReverseDeps(ctx, depMgr, appName)
	}

	// 查看依赖列表
	return showDeps(ctx, depMgr, appName)
}

// showDeps 显示依赖列表
func showDeps(ctx context.Context, depMgr dep.Manager, appName string) error {
	output.Info("查看依赖：")
	output.Highlight("%s\n", appName)
	fmt.Println()

	// 获取应用信息
	bucketMgr := depMgr.(*dep.DependencyManager).BucketManager()
	app, err := bucketMgr.GetApp(ctx, "main", appName)
	if err != nil {
		return fmt.Errorf("获取应用信息失败: %w", err)
	}

	if app.Script == nil || len(app.Script.Dependencies) == 0 {
		output.Info("此应用没有依赖")
		return nil
	}

	output.Infoln("依赖列表：")
	for i, dep := range app.Script.Dependencies {
		fmt.Printf("  %d. %s", i+1, dep.Name)
		if dep.Version != "" {
			fmt.Printf(" (%s)", dep.Version)
		}
		if dep.Optional {
			fmt.Printf(" [可选]")
		}
		if len(dep.Conditions) > 0 {
			fmt.Printf(" [条件: %v]", dep.Conditions)
		}
		fmt.Println()
	}

	return nil
}

// showReverseDeps 显示反向依赖
func showReverseDeps(ctx context.Context, depMgr dep.Manager, appName string) error {
	output.Info("查看反向依赖（谁依赖我）：")
	output.Highlight("%s\n", appName)
	fmt.Println()

	// 获取反向依赖
	dependents, err := depMgr.GetDependents(ctx, appName)
	if err != nil {
		return fmt.Errorf("获取反向依赖失败: %w", err)
	}

	if len(dependents) == 0 {
		output.Info("没有应用依赖此软件")
		return nil
	}

	output.Infoln("以下应用依赖此软件：")
	for _, dep := range dependents {
		fmt.Printf("  - %s\n", dep)
	}

	return nil
}

// showDepsTree 显示依赖树
func showDepsTree(ctx context.Context, depMgr dep.Manager, appName string) error {
	output.Info("查看依赖树：")
	output.Highlight("%s\n", appName)
	fmt.Println()

	// 获取应用信息
	bucketMgr := depMgr.(*dep.DependencyManager).BucketManager()
	app, err := bucketMgr.GetApp(ctx, "main", appName)
	if err != nil {
		return fmt.Errorf("获取应用信息失败: %w", err)
	}

	if app.Script == nil || len(app.Script.Dependencies) == 0 {
		output.Info("此应用没有依赖")
		return nil
	}

	// 解析依赖图
	graph, err := depMgr.Resolve(ctx, app)
	if err != nil {
		return fmt.Errorf("解析依赖图失败: %w", err)
	}

	// 显示依赖树
	printDependencyTree(graph, appName, 0)

	return nil
}

// printDependencyTree 递归打印依赖树
func printDependencyTree(graph *dep.DependencyGraph, appName string, depth int) {
	node := graph.Nodes[appName]
	if node == nil {
		return
	}

	// 打印当前节点
	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "  "
	}
	fmt.Printf("%s├─ %s", prefix, appName)
	if node.Version != "" {
		fmt.Printf(" (%s)", node.Version)
	}
	fmt.Println()

	// 递归打印子依赖
	for _, dep := range node.Dependencies {
		printDependencyTree(graph, dep.App.Script.Name, depth+1)
	}
}
