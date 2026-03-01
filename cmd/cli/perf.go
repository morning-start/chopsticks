package cli

import (
	"fmt"
	"time"

	"chopsticks/pkg/metrics"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

// ANSI 转义序列常量
const (
	ansiClearScreen   = "\033[H\033[2J"
	ansiMoveCursorFmt = "\033[%d;%dH"
	ansiClearBelow    = "\033[J"
)

var (
	perfMonitorInterval int
	perfReportDuration  int
)

// perfCmd 性能监控命令
var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "性能监控和诊断工具",
	Long:  `性能监控和诊断工具，用于实时监控和生成性能报告。`,
}

// perfMonitorCmd 实时监控
var perfMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "实时监控性能指标",
	RunE:  runPerfMonitor,
}

// perfReportCmd 生成性能报告
var perfReportCmd = &cobra.Command{
	Use:   "report",
	Short: "生成性能报告",
	RunE:  runPerfReport,
}

// perfStatusCmd 查看当前状态
var perfStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前性能状态",
	RunE:  runPerfStatus,
}

// perfJSPoolCmd 查看 JS 池状态
var perfJSPoolCmd = &cobra.Command{
	Use:   "js-pool",
	Short: "查看 JS 引擎池状态",
	RunE:  runPerfJSPool,
}

func runPerfMonitor(cmd *cobra.Command, args []string) error {
	interval := perfMonitorInterval
	if interval <= 0 {
		interval = 2
	}

	output.Infoln("========================================")
	output.Infoln("Performance Monitor - Press Ctrl+C to exit")
	output.Infoln("========================================")
	fmt.Println()

	// 初始化指标收集
	metrics.Init()
	defer metrics.Shutdown()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 清屏并显示表头
	clearScreen()
	printMonitorHeader()

	for range ticker.C {
		// 移动光标到表头下方
		moveCursor(1, 4)
		clearBelow()
		printMetrics()
	}
	return nil
}

func runPerfReport(cmd *cobra.Command, args []string) error {
	duration := perfReportDuration
	if duration <= 0 {
		duration = 10
	}

	output.Infoln("========================================")
	output.Infof("Collecting performance data (%d seconds)...\n", duration)
	output.Infoln("========================================")

	// 初始化指标收集
	metrics.Init()
	defer metrics.Shutdown()

	// 等待收集数据
	time.Sleep(time.Duration(duration) * time.Second)

	// 生成报告
	history := metrics.GetMetricsHistory()
	snapshots := history.GetAll()

	if len(snapshots) == 0 {
		output.Warningln("No performance data collected")
		return nil
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("Performance Report")
	output.Infoln("========================================")

	// 计算统计数据
	var totalTasks, completedTasks, failedTasks int64
	var totalMemory float64
	var maxGoroutines int
	var totalDownloadSpeed float64

	for _, snap := range snapshots {
		totalTasks += snap.Metrics.TotalTasks
		completedTasks += snap.Metrics.CompletedTasks
		failedTasks += snap.Metrics.FailedTasks
		totalMemory += snap.Metrics.MemoryUsage
		if snap.Metrics.GoroutineCount > maxGoroutines {
			maxGoroutines = snap.Metrics.GoroutineCount
		}
		totalDownloadSpeed += snap.Metrics.DownloadSpeed
	}

	count := int64(len(snapshots))
	avgMemory := totalMemory / float64(count)
	avgDownloadSpeed := totalDownloadSpeed / float64(count)

	// 显示统计
	output.Successf("Monitor duration: %d seconds\n", duration)
	output.Successf("Data points: %d\n", count)
	fmt.Println()

	output.Highlightln("Task Statistics:")
	output.Dimf("  Total tasks: %d\n", totalTasks)
	output.Dimf("  Completed: %d\n", completedTasks)
	output.Dimf("  Failed: %d\n", failedTasks)
	if totalTasks > 0 {
		successRate := float64(completedTasks) / float64(totalTasks) * 100
		output.Dimf("  Success rate: %.1f%%\n", successRate)
	}
	fmt.Println()

	output.Highlightln("Resource Usage:")
	output.Dimf("  Avg memory: %.2f MB\n", avgMemory)
	output.Dimf("  Max goroutines: %d\n", maxGoroutines)
	fmt.Println()

	output.Highlightln("Download Statistics:")
	output.Dimf("  Avg download speed: %.2f MB/s\n", avgDownloadSpeed)
	fmt.Println()

	// 显示最新指标
	latest := snapshots[len(snapshots)-1].Metrics
	output.Highlightln("Current Metrics:")
	printMetricsDetail(latest)

	return nil
}

func runPerfStatus(cmd *cobra.Command, args []string) error {
	m := metrics.GetMetrics()

	output.Infoln("========================================")
	output.Infoln("Current Performance Status")
	output.Infoln("========================================")
	fmt.Println()

	printMetricsDetail(m)

	return nil
}

func runPerfJSPool(cmd *cobra.Command, args []string) error {
	m := metrics.GetMetrics()

	output.Infoln("========================================")
	output.Infoln("JS Engine Pool Status")
	output.Infoln("========================================")
	fmt.Println()

	output.Highlightln("Pool Info:")
	output.Dimf("  Pool size: %d\n", m.JSPoolSize)
	output.Dimf("  Active engines: %d\n", m.JSPoolActive)
	output.Dimf("  Idle engines: %d\n", m.JSPoolSize-m.JSPoolActive)
	output.Dimf("  Utilization: %.1f%%\n", m.JSPoolUtilization)
	fmt.Println()

	output.Highlightln("Cache Info:")
	output.Dimf("  Cache size: %d\n", m.JSCacheSize)
	output.Dimf("  Cache hit rate: %.1f%%\n", m.JSCacheHitRate)
	fmt.Println()

	// 健康状态
	switch {
	case m.JSPoolUtilization > highUtilizationThreshold:
		output.Warningln("Warning: JS pool utilization is too high")
	case m.JSPoolUtilization < lowUtilizationThreshold && m.JSPoolSize > 0:
		output.Dimln("Tip: JS pool utilization is low, consider reducing pool size")
	default:
		output.Successln("JS pool status is healthy")
	}

	return nil
}

// printMonitorHeader 打印监控表头
func printMonitorHeader() {
	fmt.Println("Time       Tasks   Memory(MB) Goroutines  Download(MB/s)  JSPool%")
	fmt.Println("---------- ------  ---------- ----------  --------------  --------")
}

// printMetrics 打印指标
func printMetrics() {
	m := metrics.GetMetrics()
	timestamp := m.Timestamp.Format("15:04:05")

	fmt.Printf("%s  %6d  %8.1f  %10d  %14.2f  %7.1f%%\n",
		timestamp,
		m.TaskInProgress,
		m.MemoryUsage,
		m.GoroutineCount,
		m.DownloadSpeed,
		m.JSPoolUtilization,
	)
}

// printMetricsDetail 打印详细指标
func printMetricsDetail(m metrics.PerformanceMetrics) {
	output.Highlightln("Task Metrics:")
	output.Dimf("  In progress: %d\n", m.TaskInProgress)
	output.Dimf("  Completed: %d\n", m.CompletedTasks)
	output.Dimf("  Failed: %d\n", m.FailedTasks)
	output.Dimf("  Submit rate: %.2f tasks/sec\n", m.TaskSubmitRate)
	output.Dimf("  Complete rate: %.2f tasks/sec\n", m.TaskCompleteRate)
	fmt.Println()

	output.Highlightln("Resource Usage:")
	output.Dimf("  Memory: %.2f MB\n", m.MemoryUsage)
	output.Dimf("  Goroutines: %d\n", m.GoroutineCount)
	output.Dimf("  GC count: %d\n", m.NumGC)
	fmt.Println()

	output.Highlightln("Download Metrics:")
	output.Dimf("  Active downloads: %d\n", m.ActiveDownloads)
	output.Dimf("  Download speed: %.2f MB/s\n", m.DownloadSpeed)
	output.Dimf("  Download errors: %d\n", m.DownloadErrors)
	fmt.Println()

	output.Highlightln("Search Metrics:")
	output.Dimf("  Active searches: %d\n", m.ActiveSearches)
	output.Dimf("  Cache hit rate: %.1f%%\n", m.SearchCacheHitRate)
	fmt.Println()

	output.Highlightln("Install Metrics:")
	output.Dimf("  Active installs: %d\n", m.ActiveInstalls)
	output.Dimf("  Queue size: %d\n", m.InstallQueueSize)
}

// clearScreen 清屏
func clearScreen() {
	fmt.Print(ansiClearScreen)
}

// moveCursor 移动光标
func moveCursor(row, col int) {
	fmt.Printf(ansiMoveCursorFmt, row, col)
}

// clearBelow 清除光标下方内容
func clearBelow() {
	fmt.Print(ansiClearBelow)
}

func init() {
	perfMonitorCmd.Flags().IntVarP(&perfMonitorInterval, "interval", "i", 2, "刷新间隔(秒)")
	perfReportCmd.Flags().IntVarP(&perfReportDuration, "duration", "d", 10, "监控时长(秒)")

	perfCmd.AddCommand(perfMonitorCmd)
	perfCmd.AddCommand(perfReportCmd)
	perfCmd.AddCommand(perfStatusCmd)
	perfCmd.AddCommand(perfJSPoolCmd)

	rootCmd.AddCommand(perfCmd)
}
