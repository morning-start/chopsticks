package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"chopsticks/core/app"
	"chopsticks/pkg/output"
	"chopsticks/pkg/parallel"

	"github.com/urfave/cli/v2"
)

// updateAsyncAction 异步更新命令
func updateAsyncAction(c *cli.Context) error {
	force := c.Bool("force")
	updateAll := c.Bool("all")
	maxWorkers := c.Int("workers")
	if maxWorkers <= 0 {
		maxWorkers = 4
	}

	ctx, cancel := context.WithCancel(getContext(c))
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		output.Warningln("\n收到取消信号，正在停止...")
		cancel()
	}()

	application := getApp()

	// 更新所有
	if updateAll {
		output.Infoln("正在异步更新所有软件包...")
		if err := application.AppManager().UpdateAll(ctx, app.UpdateOptions{Force: force}); err != nil {
			output.ErrorCrossf("更新失败: %v", err)
			return cli.Exit("", 1)
		}
		output.SuccessCheck("所有软件包更新成功")
		return nil
	}

	// 没有参数时显示错误
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks update [package ...] [--all] --async")
		return cli.Exit("", 1)
	}

	// 获取所有要更新的包
	packages := make([]string, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		packages[i] = c.Args().Get(i)
	}

	total := len(packages)

	output.Infoln("========================================")
	output.Infof("开始异步更新 %d 个软件包 (最大并发: %d)\n", total, maxWorkers)
	output.Infoln("========================================")
	fmt.Println()

	// 创建任务池
	pool := parallel.NewPool(maxWorkers)
	results := make([]updateResult, total)
	var mu sync.Mutex

	for i, pkg := range packages {
		idx := i
		pkg := pkg
		pool.Add(func() error {
			result := updatePackage(ctx, application.AppManager(), pkg, force)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
			return result.err
		})
	}

	// 执行并行任务
	err := pool.Run(ctx)

	// 汇总结果
	return printAsyncUpdateResults(results, err)
}

// updatePackage 更新单个包
func updatePackage(ctx context.Context, mgr app.Manager, name string, force bool) updateResult {
	opts := app.UpdateOptions{
		Force: force,
	}

	startTime := time.Now()
	err := mgr.Update(ctx, name, opts)
	duration := time.Since(startTime)

	return updateResult{
		name:     name,
		duration: duration,
		err:      err,
	}
}

// updateResult 更新结果
type updateResult struct {
	name     string
	duration time.Duration
	err      error
}

// printAsyncUpdateResults 打印异步更新结果
func printAsyncUpdateResults(results []updateResult, poolErr error) error {
	var successCount, failCount int
	var failedApps []string
	var totalDuration time.Duration

	for _, result := range results {
		if result.err != nil {
			failCount++
			failedApps = append(failedApps, result.name)
			output.ErrorCross(fmt.Sprintf("%s 更新失败: %v", result.name, result.err))
		} else {
			successCount++
			totalDuration += result.duration
			output.SuccessCheck(fmt.Sprintf("%s 更新成功 (%.2fs)", result.name, result.duration.Seconds()))
		}
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("异步更新完成")
	output.Infoln("========================================")
	output.Successf("成功: %d\n", successCount)
	if failCount > 0 {
		output.Errorf("失败: %d\n", failCount)
		output.Errorln("失败的软件包:")
		for _, name := range failedApps {
			output.Errorf("  - %s\n", name)
		}
	}
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		output.Dimf("平均耗时: %.2fs\n", avgDuration.Seconds())
	}

	if failCount > 0 || poolErr != nil {
		return cli.Exit("", 1)
	}
	output.SuccessCheck("所有软件包更新完成")
	return nil
}
