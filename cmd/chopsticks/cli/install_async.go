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

// installAsyncAction 异步安装命令
func installAsyncAction(c *cli.Context) error {
	if c.NArg() < 1 {
		output.Errorln("错误: 缺少软件包名称")
		output.Dimln("用法: chopsticks install <package>[@version] ... --async")
		return cli.Exit("", 1)
	}

	force := c.Bool("force")
	arch := c.String("arch")
	bucket := c.String("bucket")
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

	// 获取所有要安装的包
	packages := make([]struct {
		name    string
		version string
		spec    string
	}, c.NArg())

	for i := 0; i < c.NArg(); i++ {
		spec := c.Args().Get(i)
		name, version := parseAppSpec(spec)
		packages[i] = struct {
			name    string
			version string
			spec    string
		}{name: name, version: version, spec: spec}
	}

	total := len(packages)

	output.Infoln("========================================")
	output.Infof("开始异步安装 %d 个软件包 (最大并发: %d)\n", total, maxWorkers)
	output.Infoln("========================================")
	fmt.Println()

	// 创建任务池
	pool := parallel.NewPool(maxWorkers)
	results := make([]installResult, total)
	var mu sync.Mutex

	for i, pkg := range packages {
		idx := i
		pkg := pkg
		pool.Add(func() error {
			result := installPackage(ctx, application.AppManager(), pkg.name, pkg.version, bucket, arch, force)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
			return result.err
		})
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

// installResult 安装结果
type installResult struct {
	name     string
	version  string
	duration time.Duration
	err      error
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
			output.ErrorCross(fmt.Sprintf("%s 失败: %v", result.name, result.err))
		} else {
			successCount++
			totalDuration += result.duration
			output.SuccessCheck(fmt.Sprintf("%s 安装成功 (%.2fs)", result.name, result.duration.Seconds()))
		}
	}

	fmt.Println()
	output.Infoln("========================================")
	output.Infoln("异步安装完成")
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
	output.SuccessCheck("所有软件包处理完成")
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
	ticker := time.NewTicker(100 * time.Millisecond)
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
			progressPercent += 5.0
			if progressPercent > 90.0 {
				progressPercent = 90.0
			}
			progress.SetProgress(progressPercent)
		case <-ctx.Done():
			progress.SetError(ctx.Err())
			return ctx.Err()
		}
	}
}

// ProgressDisplay 进度显示
type ProgressDisplay struct {
	mu       sync.RWMutex
	stopChan chan struct{}
	name     string
	progress float64
	err      error
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
		ticker := time.NewTicker(100 * time.Millisecond)
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

	width := 30
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
	if err != nil {
		status = " 错误"
	} else if complete {
		status = " 完成"
	} else {
		status = " 进行中"
	}

	fmt.Printf("\r%s %s %3.0f%%%s", name, bar, progress, status)
}
