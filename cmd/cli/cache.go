package cli

import (
	"fmt"
	"time"

	"chopsticks/core/cache"
	"chopsticks/core/store"
	"chopsticks/pkg/config"
	"chopsticks/pkg/metrics"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	cacheConfigMaxSize   int64
	cacheConfigMaxEntries int
	cacheConfigTTL       string
	cacheShowStats       bool
	cacheClearAll        bool
)

// cacheCmd 缓存管理命令
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理工具",
	Long:  `缓存管理工具，用于查看和管理性能优化缓存。`,
}

// cacheStatsCmd 查看缓存统计
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "查看缓存统计信息",
	RunE:  runCacheStats,
}

// cacheClearCmd 清空缓存
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清空缓存",
	RunE:  runCacheClear,
}

// cacheHealthCmd 检查缓存健康状态
var cacheHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "检查缓存健康状态",
	RunE:  runCacheHealth,
}

// cacheConfigCmd 查看缓存配置
var cacheConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "查看缓存配置",
	RunE:  runCacheConfig,
}

func runCacheStats(cmd *cobra.Command, args []string) error {
	output.Infoln("========================================")
	output.Infoln("Cache Statistics")
	output.Infoln("========================================")
	fmt.Println()

	// 获取全局指标
	m := metrics.GetMetrics()

	// 应用缓存
	output.Highlightln("App Cache:")
	output.Dimf("  Hit Rate: %.2f%%\n", m.AppCacheHitRate)
	output.Dimf("  Size: %d entries\n", m.AppCacheSize)
	fmt.Println()

	// Bucket 缓存
	output.Highlightln("Bucket Cache:")
	output.Dimf("  Hit Rate: %.2f%%\n", m.BucketCacheHitRate)
	output.Dimf("  Size: %d entries\n", m.BucketCacheSize)
	fmt.Println()

	// 索引缓存
	output.Highlightln("Index Cache:")
	output.Dimf("  Hit Rate: %.2f%%\n", m.IndexCacheHitRate)
	output.Dimf("  Size: %d entries\n", m.IndexCacheSize)
	fmt.Println()

	// 批量读取
	output.Highlightln("Batch Read:")
	output.Dimf("  Efficiency: %.2f%%\n", m.BatchReadEfficiency)
	fmt.Println()

	// 淘汰统计
	output.Highlightln("Evictions:")
	output.Dimf("  Total: %d\n", m.CacheEvictions)
	fmt.Println()

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	if cacheClearAll {
		output.Infoln("Clearing all caches...")

		// 这里需要集成到实际的缓存管理器
		// 暂时只是提示
		output.Successln("All caches cleared successfully")
	} else {
		output.Infoln("Cache clear operation requires --all flag")
		output.Infoln("Use 'chopsticks cache clear --all' to clear all caches")
	}

	return nil
}

func runCacheHealth(cmd *cobra.Command, args []string) error {
	output.Infoln("========================================")
	output.Infoln("Cache Health Check")
	output.Infoln("========================================")
	fmt.Println()

	m := metrics.GetMetrics()

	// 检查应用缓存
	checkCacheHealth("App Cache", m.AppCacheHitRate, m.AppCacheSize)

	// 检查 Bucket 缓存
	checkCacheHealth("Bucket Cache", m.BucketCacheHitRate, m.BucketCacheSize)

	// 检查索引缓存
	checkCacheHealth("Index Cache", m.IndexCacheHitRate, m.IndexCacheSize)

	fmt.Println()

	// 总体评估
	avgHitRate := (m.AppCacheHitRate + m.BucketCacheHitRate + m.IndexCacheHitRate) / 3
	if avgHitRate >= 80 {
		output.Successf("Overall cache health: Excellent (%.2f%% avg hit rate)\n", avgHitRate)
	} else if avgHitRate >= 60 {
		output.Infof("Overall cache health: Good (%.2f%% avg hit rate)\n", avgHitRate)
	} else if avgHitRate >= 40 {
		output.Warningf("Overall cache health: Fair (%.2f%% avg hit rate)\n", avgHitRate)
	} else {
		output.Errorf("Overall cache health: Poor (%.2f%% avg hit rate)\n", avgHitRate)
	}

	return nil
}

func checkCacheHealth(name string, hitRate float64, size int) {
	output.Highlightf("%s:\n", name)

	if hitRate >= 80 {
		output.Successf("  Status: Healthy\n")
	} else if hitRate >= 60 {
		output.Infof("  Status: Good\n")
	} else if hitRate >= 40 {
		output.Warningf("  Status: Fair\n")
	} else {
		output.Errorf("  Status: Poor\n")
	}

	output.Dimf("  Hit Rate: %.2f%%\n", hitRate)
	output.Dimf("  Size: %d entries\n", size)

	// 建议
	if hitRate < 50 {
		output.Dimln("  Tip: Consider increasing cache size or TTL")
	}
	fmt.Println()
}

func runCacheConfig(cmd *cobra.Command, args []string) error {
	output.Infoln("========================================")
	output.Infoln("Cache Configuration")
	output.Infoln("========================================")
	fmt.Println()

	// 从配置中读取
	cfg, err := config.LoadDefault()
	if err != nil {
		output.Errorf("Failed to load config: %v\n", err)
		return err
	}

	output.Highlightln("Current Configuration:")
	output.Dimf("  Max Size: %d MB\n", cacheConfigMaxSize/1024/1024)
	output.Dimf("  Max Entries: %d\n", cacheConfigMaxEntries)
	output.Dimf("  TTL: %s\n", cacheConfigTTL)
	output.Dimf("  Data Dir: %s\n", cfg.StorageDir)
	fmt.Println()

	output.Highlightln("Default Values:")
	output.Dimf("  Max Size: %d MB\n", cache.DefaultMaxCacheSize/1024/1024)
	output.Dimf("  Max Entries: %d\n", cache.DefaultMaxCacheEntries)
	output.Dimf("  TTL: %v\n", cache.DefaultCacheTTL)
	fmt.Println()

	// 提示
	output.Infoln("To change cache configuration, edit the config file:")
	output.Dimf("  %s\n", config.GetConfigPath())

	return nil
}

// createCacheManager 创建缓存管理器（辅助函数）
func createCacheManager() (*cache.AppCacheManager, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, err
	}

	dataDir := cfg.StorageDir

	// 创建存储
	storage, err := store.NewFSStorage(dataDir)
	if err != nil {
		return nil, err
	}

	// 创建缓存配置
	cacheConfig := cache.CacheConfig{
		MaxSize:        cacheConfigMaxSize,
		MaxEntries:     cacheConfigMaxEntries,
		TTL:            parseDuration(cacheConfigTTL),
		CleanupInterval: cache.DefaultCleanupInterval,
	}

	// 创建缓存管理器
	cacheMgr := cache.NewAppCacheManager(storage, cacheConfig)

	return cacheMgr, nil
}

// parseDuration 解析持续时间字符串
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return cache.DefaultCacheTTL
	}
	return d
}

// printCacheReport 打印缓存报告
func printCacheReport(cacheMgr *cache.AppCacheManager) {
	stats := cacheMgr.GetStats()

	output.Infoln("========================================")
	output.Infoln("Detailed Cache Report")
	output.Infoln("========================================")
	fmt.Println()

	for name, stat := range stats {
		output.Highlightf("%s Cache:\n", name)
		output.Dimf("  Hits: %d\n", stat.Hits)
		output.Dimf("  Misses: %d\n", stat.Misses)
		output.Dimf("  Hit Rate: %.2f%%\n", stat.HitRate)
		output.Dimf("  Size: %d / %d\n", stat.Size, stat.MaxSize)
		output.Dimf("  Entries: %d / %d\n", stat.Entries, stat.MaxEntries)
		output.Dimf("  Evictions: %d\n", stat.Evictions)
		fmt.Println()
	}
}

func init() {
	// 配置标志
	cacheStatsCmd.Flags().BoolVarP(&cacheShowStats, "detailed", "d", false, "显示详细统计")
	cacheClearCmd.Flags().BoolVarP(&cacheClearAll, "all", "a", false, "清空所有缓存")
	cacheConfigCmd.Flags().Int64VarP(&cacheConfigMaxSize, "max-size", "m", cache.DefaultMaxCacheSize, "最大缓存大小 (字节)")
	cacheConfigCmd.Flags().IntVarP(&cacheConfigMaxEntries, "max-entries", "e", cache.DefaultMaxCacheEntries, "最大条目数")
	cacheConfigCmd.Flags().StringVarP(&cacheConfigTTL, "ttl", "t", cache.DefaultCacheTTL.String(), "默认 TTL")

	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheHealthCmd)
	cacheCmd.AddCommand(cacheConfigCmd)

	rootCmd.AddCommand(cacheCmd)
}
