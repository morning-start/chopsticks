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

	"github.com/spf13/cobra"
)

// 性能阈值常量
const (
	// 高利用率阈值
	highUtilizationThreshold = 90.0
	// 低利用率阈值
	lowUtilizationThreshold = 10.0
)

// runUpdateAsync 异步更新命令
func runUpdateAsync(cmd *cobra.Command, args []string) error {
	maxWorkers := updateWorkers
	if maxWorkers <= 0 {
		maxWorkers = defaultWorkers
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		output.Warningln("\nReceived cancel signal, stopping...")
		cancel()
	}()

	application := getApp()
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

	// 更新所有
	if updateAll {
		output.Infoln("Updating all packages asynchronously...")
		if err := application.AppManager().UpdateAll(ctx, app.UpdateOptions{Force: updateForce}); err != nil {
			output.ErrorCrossf("Update failed: %v", err)
			return err
		}
		output.SuccessCheck("All packages updated successfully")
		return nil
	}

	// 没有参数时显示错误
	if len(args) < 1 {
		return fmt.Errorf("missing package name. Usage: chopsticks update [package ...] [--all] --async")
	}

	total := len(args)

	output.Infoln("========================================")
	output.Infof("Starting async update of %d packages (max concurrency: %d)\n", total, maxWorkers)
	output.Infoln("========================================")
	fmt.Println()

	// 创建任务池
	pool := parallel.NewPool(maxWorkers)
	results := make([]updateResult, total)
	var mu sync.Mutex

	for i, pkg := range args {
		pool.Add(func(idx int, name string) func() error {
			return func() error {
				result := updatePackage(ctx, application.AppManager(), name, updateForce)
				mu.Lock()
				results[idx] = result
				mu.Unlock()
				return result.err
			}
		}(i, pkg))
	}

	// 执行并行任务
	err := pool.Run(ctx)

	// 汇总结果
	return printAsyncUpdateResults(results, err)
}

// updatePackage 更新单个包
func updatePackage(ctx context.Context, mgr app.AppManager, name string, force bool) updateResult {
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

// updateResult 更新结果 - 优化内存布局（按字段大小排序）
type updateResult struct {
	duration time.Duration
	err      error
	name     string
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
			output.ErrorCross(fmt.Sprintf("%s update failed: %v", result.name, result.err))
			continue
		}
		successCount++
		totalDuration += result.duration
		output.SuccessCheck(fmt.Sprintf("%s updated successfully (%.2fs)", result.name, result.duration.Seconds()))
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("Async update completed")
	output.Infoln("========================================")
	output.Successf("Success: %d\n", successCount)
	if failCount > 0 {
		output.Errorf("Failed: %d\n", failCount)
		output.Errorln("Failed packages:")
		for _, name := range failedApps {
			output.Errorf("  - %s\n", name)
		}
	}
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		output.Dimf("Average duration: %.2fs\n", avgDuration.Seconds())
	}

	if failCount > 0 || poolErr != nil {
		return fmt.Errorf("some packages failed to update")
	}
	output.SuccessCheck("All packages updated")
	return nil
}
