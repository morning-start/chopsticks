package cli

import (
	"fmt"
	"time"

	"chopsticks/pkg/metrics"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// PerfCommand 性能监控命令
var PerfCommand = &cli.Command{
	Name:  "perf",
	Usage: "性能监控和诊断工具",
	Subcommands: []*cli.Command{
		{
			Name:   "monitor",
			Usage:  "实时监控性能指标",
			Action: perfMonitorAction,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "interval",
					Aliases: []string{"i"},
					Usage:   "刷新间隔(秒)",
					Value:   2,
				},
			},
		},
		{
			Name:   "report",
			Usage:  "生成性能报告",
			Action: perfReportAction,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "duration",
					Aliases: []string{"d"},
					Usage:   "监控时长(秒)",
					Value:   10,
				},
			},
		},
		{
			Name:   "status",
			Usage:  "查看当前性能状态",
			Action: perfStatusAction,
		},
		{
			Name:   "js-pool",
			Usage:  "查看 JS 引擎池状态",
			Action: perfJSPoolAction,
		},
	},
}

// perfMonitorAction 实时监控
func perfMonitorAction(c *cli.Context) error {
	interval := c.Int("interval")
	if interval <= 0 {
		interval = 2
	}

	output.Infoln("========================================")
	output.Infoln("性能监控 - 按 Ctrl+C 退出")
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

	for {
		select {
		case <-ticker.C:
			// 移动光标到表头下方
			moveCursor(1, 4)
			clearBelow()
			printMetrics()
		}
	}
}

// perfReportAction 生成性能报告
func perfReportAction(c *cli.Context) error {
	duration := c.Int("duration")
	if duration <= 0 {
		duration = 10
	}

	output.Infoln("========================================")
	output.Infof("开始收集性能数据 (%d 秒)...\n", duration)
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
		output.Warningln("没有收集到性能数据")
		return nil
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("性能报告")
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
	output.Successf("监控时长: %d 秒\n", duration)
	output.Successf("数据点: %d\n", count)
	fmt.Println()

	output.Highlightln("任务统计:")
	output.Dimf("  总任务数: %d\n", totalTasks)
	output.Dimf("  完成任务: %d\n", completedTasks)
	output.Dimf("  失败任务: %d\n", failedTasks)
	if totalTasks > 0 {
		successRate := float64(completedTasks) / float64(totalTasks) * 100
		output.Dimf("  成功率: %.1f%%\n", successRate)
	}
	fmt.Println()

	output.Highlightln("资源使用:")
	output.Dimf("  平均内存: %.2f MB\n", avgMemory)
	output.Dimf("  最大 Goroutines: %d\n", maxGoroutines)
	fmt.Println()

	output.Highlightln("下载统计:")
	output.Dimf("  平均下载速度: %.2f MB/s\n", avgDownloadSpeed)
	fmt.Println()

	// 显示最新指标
	latest := snapshots[len(snapshots)-1].Metrics
	output.Highlightln("当前指标:")
	printMetricsDetail(latest)

	return nil
}

// perfStatusAction 查看当前状态
func perfStatusAction(c *cli.Context) error {
	m := metrics.GetMetrics()

	output.Infoln("========================================")
	output.Infoln("当前性能状态")
	output.Infoln("========================================")
	fmt.Println()

	printMetricsDetail(m)

	return nil
}

// perfJSPoolAction 查看 JS 池状态
func perfJSPoolAction(c *cli.Context) error {
	m := metrics.GetMetrics()

	output.Infoln("========================================")
	output.Infoln("JS 引擎池状态")
	output.Infoln("========================================")
	fmt.Println()

	output.Highlightln("池信息:")
	output.Dimf("  池大小: %d\n", m.JSPoolSize)
	output.Dimf("  活跃引擎: %d\n", m.JSPoolActive)
	output.Dimf("  空闲引擎: %d\n", m.JSPoolSize-m.JSPoolActive)
	output.Dimf("  利用率: %.1f%%\n", m.JSPoolUtilization)
	fmt.Println()

	output.Highlightln("缓存信息:")
	output.Dimf("  缓存大小: %d\n", m.JSCacheSize)
	output.Dimf("  缓存命中率: %.1f%%\n", m.JSCacheHitRate)
	fmt.Println()

	// 健康状态
	if m.JSPoolUtilization > 90 {
		output.Warningln("警告: JS 池利用率过高")
	} else if m.JSPoolUtilization < 10 && m.JSPoolSize > 0 {
		output.Dimln("提示: JS 池利用率较低，可考虑减小池大小")
	} else {
		output.Successln("JS 池状态良好")
	}

	return nil
}

// printMonitorHeader 打印监控表头
func printMonitorHeader() {
	fmt.Println("时间        任务    内存(MB)  Goroutines  下载速度(MB/s)  JS池利用")
	fmt.Println("----------  ------  --------  ----------  --------------  --------")
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
	output.Highlightln("任务指标:")
	output.Dimf("  进行中: %d\n", m.TaskInProgress)
	output.Dimf("  已完成: %d\n", m.CompletedTasks)
	output.Dimf("  失败: %d\n", m.FailedTasks)
	output.Dimf("  提交速率: %.2f 任务/秒\n", m.TaskSubmitRate)
	output.Dimf("  完成速率: %.2f 任务/秒\n", m.TaskCompleteRate)
	fmt.Println()

	output.Highlightln("资源使用:")
	output.Dimf("  内存: %.2f MB\n", m.MemoryUsage)
	output.Dimf("  Goroutines: %d\n", m.GoroutineCount)
	output.Dimf("  GC 次数: %d\n", m.NumGC)
	fmt.Println()

	output.Highlightln("下载指标:")
	output.Dimf("  活跃下载: %d\n", m.ActiveDownloads)
	output.Dimf("  下载速度: %.2f MB/s\n", m.DownloadSpeed)
	output.Dimf("  下载错误: %d\n", m.DownloadErrors)
	fmt.Println()

	output.Highlightln("搜索指标:")
	output.Dimf("  活跃搜索: %d\n", m.ActiveSearches)
	output.Dimf("  缓存命中率: %.1f%%\n", m.SearchCacheHitRate)
	fmt.Println()

	output.Highlightln("安装指标:")
	output.Dimf("  活跃安装: %d\n", m.ActiveInstalls)
	output.Dimf("  队列大小: %d\n", m.InstallQueueSize)
}

// clearScreen 清屏
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// moveCursor 移动光标
func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// clearBelow 清除光标下方内容
func clearBelow() {
	fmt.Print("\033[J")
}
