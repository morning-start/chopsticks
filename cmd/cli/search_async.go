package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chopsticks/core/bucket"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

// runSearchAsync 异步搜索命令
func runSearchAsync(cmd *cobra.Command, args []string) error {
	query := args[0]
	maxWorkers := searchWorkers
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	ctx, cancel := context.WithCancel(cmd.Context())
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

	output.Info("异步搜索: ")
	output.Highlightln(query)
	if searchBucket != "" {
		output.Dim("软件源: ")
		output.Infoln(searchBucket)
	}
	output.Dimf("并发数: %d\n", maxWorkers)
	fmt.Println()

	startTime := time.Now()

	// 使用并行搜索器
	searcher := bucket.NewParallelSearcher(application.BucketManager(), maxWorkers)

	opts := bucket.SearchOptions{
		Bucket: searchBucket,
	}

	results, err := searcher.SearchWithCache(ctx, query, opts)
	if err != nil {
		output.ErrorCrossf("搜索失败: %v", err)
		return fmt.Errorf("搜索失败: %w", err)
	}

	duration := time.Since(startTime)

	// 显示结果
	output.Highlightln("\n搜索结果:")
	output.Dimln("-----------")

	if len(results) == 0 {
		output.Warningln("未找到匹配的应用")
		return nil
	}

	for _, result := range results {
		output.Success("%s", result.App.Name)
		if result.App.Description != "" {
			output.Dimf("    描述: %s\n", result.App.Description)
		}
		output.Dimf("    版本: %s\n", result.App.Version)
		output.Dimf("    软件源: %s\n", result.Bucket)
	}

	// 显示统计
	fmt.Println()
	output.Dimf("找到 %d 个结果 (耗时: %.2fs)\n", len(results), duration.Seconds())

	// 显示缓存统计
	stats := searcher.GetCacheStats()
	if stats.Hits > 0 || stats.Misses > 0 {
		output.Dimf("缓存命中率: %.1f%% (命中: %d, 未命中: %d)\n",
			stats.HitRate*100, stats.Hits, stats.Misses)
	}

	return nil
}
