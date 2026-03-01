package cli

import (
	"context"
	"errors"
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

// 常量定义
const (
	// 默认并发数
	defaultWorkers = 4
	// 进度报告间隔
	progressReportInterval = 100 * time.Millisecond
	// 进度增量
	progressIncrement = 5.0
	// 进度上限
	progressMax = 90.0
	// 进度条宽度
	progressBarWidth = 30
)

// 预定义错误
var (
	ErrMissingPackageName = errors.New("missing package name")
)

// installAsyncAction 异步安装命令
func installAsyncAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("Error: missing package name")
		output.Dimln("Usage: chopsticks install <package>[@version] ... --async")
		return cli.Exit("", 1)
	}

	force := c.Bool("force")
	arch := c.String("arch")
	bucket := c.String("bucket")
	maxWorkers := c.Int("workers")
	if maxWorkers <= 0 {
		maxWorkers = defaultWorkers
	}

	ctx, cancel := context.WithCancel(getContext(c))
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

	// 获取所有要安装的包
	packages := make([]packageSpec, c.NArg())

	for i := 0; i < c.NArg(); i++ {
		spec := c.Args().Get(i)
		name, version := parseAppSpec(spec)
		packages[i] = packageSpec{name: name, version: version, spec: spec}
	}

	total := len(packages)

	output.Infoln("========================================")
	output.Infof("Starting async installation of %d packages (max concurrency: %d)\n", total, maxWorkers)
	output.Infoln("========================================")
	fmt.Println()

	// 创建任务池
	pool := parallel.NewPool(maxWorkers)
	results := make([]installResult, total)
	var mu sync.Mutex

	for i, pkg := range packages {
		pool.Add(func(idx int, p packageSpec) func() error {
			return func() error {
				result := installPackage(ctx, application.AppManager(), p.name, p.version, bucket, arch, force)
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
	return printInstallResults(results, err)
}

// installPackage 安装单个包
func installPackage(ctx context.Context, mgr app.Manager, name, version, bucket, arch string, force bool) installResult {
	opts := app.InstallOptions{
		Arch:  arch,
		Force: force,
	}

	installSpec := app.InstallSpec{
		Bucket:  bucket,
		Name:    name,
		Version: version,
	}

	startTime := time.Now()
	err := mgr.Install(ctx, installSpec, opts)
	duration := time.Since(startTime)

	return installResult{
		name:     name,
		version:  version,
		duration: duration,
		err:      err,
	}
}

// installResult 安装结果 - 优化内存布局（按字段大小排序）
type installResult struct {
	duration time.Duration
	err      error
	name     string
	version  string
}

// printInstallResults 打印安装结果
func printInstallResults(results []installResult, poolErr error) error {
	var successCount, failCount int
	var failedApps []string
	var totalDuration time.Duration

	for _, result := range results {
		if result.err != nil {
			failCount++
			failedApps = append(failedApps, result.name)
			output.ErrorCross(fmt.Sprintf("%s failed: %v", result.name, result.err))
			continue
		}
		successCount++
		totalDuration += result.duration
		output.SuccessCheck(fmt.Sprintf("%s installed successfully (%.2fs)", result.name, result.duration.Seconds()))
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("Async installation completed")
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
		return cli.Exit("", 1)
	}
	output.SuccessCheck("All packages processed")
	return nil
}

// installWithProgress 带进度显示的安装
func installWithProgress(ctx context.Context, mgr app.Manager, name, version, bucket, arch string, force bool) error {
	opts := app.InstallOptions{
		Arch:  arch,
		Force: force,
	}

	installSpec := app.InstallSpec{
		Bucket:  bucket,
		Name:    name,
		Version: version,
	}

	// 创建进度显示
	progress := NewProgressDisplay(name)
	progress.Start()
	defer progress.Stop()

	// 在后台更新进度
	done := make(chan error, 1)
	go func() {
		done <- mgr.Install(ctx, installSpec, opts)
	}()

	// 模拟进度更新
	ticker := time.NewTicker(progressReportInterval)
	defer ticker.Stop()

	progressPercent := 0.0
	for {
		select {
		case err := <-done:
			if err != nil {
				progress.SetError(err)
				return err
			}
			progress.SetComplete()
			return nil
		case <-ticker.C:
			progressPercent += progressIncrement
			if progressPercent > progressMax {
				progressPercent = progressMax
			}
			progress.SetProgress(progressPercent)
		case <-ctx.Done():
			progress.SetError(ctx.Err())
			return ctx.Err()
		}
	}
}

// ProgressDisplay 进度显示 - 优化内存布局（按字段大小排序）
type ProgressDisplay struct {
	mu       sync.RWMutex
	stopChan chan struct{}
	err      error
	name     string
	progress float64
	stopped  bool
	complete bool
}

// NewProgressDisplay 创建进度显示
func NewProgressDisplay(name string) *ProgressDisplay {
	return &ProgressDisplay{
		name:     name,
		stopChan: make(chan struct{}),
	}
}

// Start 开始显示进度
func (p *ProgressDisplay) Start() {
	go func() {
		ticker := time.NewTicker(progressReportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-p.stopChan:
				return
			case <-ticker.C:
				p.draw()
			}
		}
	}()
}

// Stop 停止显示进度
func (p *ProgressDisplay) Stop() {
	p.mu.Lock()
	if !p.stopped {
		p.stopped = true
		close(p.stopChan)
	}
	p.mu.Unlock()
	p.draw()
	fmt.Println()
}

// SetProgress 设置进度
func (p *ProgressDisplay) SetProgress(percent float64) {
	p.mu.Lock()
	p.progress = percent
	p.mu.Unlock()
}

// SetError 设置错误
func (p *ProgressDisplay) SetError(err error) {
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
}

// SetComplete 设置完成
func (p *ProgressDisplay) SetComplete() {
	p.mu.Lock()
	p.complete = true
	p.progress = 100.0
	p.mu.Unlock()
}

// draw 绘制进度条
func (p *ProgressDisplay) draw() {
	p.mu.RLock()
	name := p.name
	progress := p.progress
	err := p.err
	complete := p.complete
	p.mu.RUnlock()

	width := progressBarWidth
	filled := int(float64(width) * progress / 100.0)
	if filled > width {
		filled = width
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	status := ""
	switch {
	case err != nil:
		status = " Error"
	case complete:
		status = " Done"
	default:
		status = " In Progress"
	}

	fmt.Printf("\r%s %s %3.0f%%%s", name, bar, progress, status)
}
