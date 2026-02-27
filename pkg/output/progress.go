// Package output 提供输出格式化功能，包括进度条。
package output

import (
	"fmt"
	"io"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// ProgressManager mpb 进度管理器
type ProgressManager struct {
	progress *mpb.Progress
}

// NewProgressManager 创建新的进度管理器
func NewProgressManager() *ProgressManager {
	return &ProgressManager{
		progress: mpb.New(mpb.WithWidth(64)),
	}
}

// NewProgressManagerWithOutput 创建新的进度管理器，指定输出
func NewProgressManagerWithOutput(w io.Writer) *ProgressManager {
	return &ProgressManager{
		progress: mpb.New(mpb.WithWidth(64), mpb.WithOutput(w)),
	}
}

// AddDownloadBar 添加下载进度条
// name: 任务名称
// total: 总大小（字节）
// 返回: mpb.Bar 实例
func (pm *ProgressManager) AddDownloadBar(name string, total int64) *mpb.Bar {
	return pm.progress.AddBar(total,
		mpb.PrependDecorators(
			// 显示任务名称
			decor.Name(name+" ", decor.WCSyncWidth),
			// 显示当前/总计
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			// 显示百分比
			decor.Percentage(decor.WCSyncSpace),
			// 显示速度
			decor.EwmaSpeed(decor.SizeB1024(0), " % .2f", 60),
			// 显示剩余时间
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), " 完成"),
		),
	)
}

// AddInstallBar 添加安装进度条（多阶段）
// name: 应用名称
// stages: 阶段名称列表
// 返回: MultiStageBar 实例
func (pm *ProgressManager) AddInstallBar(name string, stages []string) *MultiStageBar {
	total := int64(len(stages) * 100) // 每个阶段100个单位

	bar := pm.progress.AddBar(total,
		mpb.PrependDecorators(
			// 显示应用名称
			decor.Name(name+" ", decor.WCSyncWidth),
			// 显示当前阶段
			decor.Any(func(s decor.Statistics) string {
				currentStage := int(s.Current / 100)
				if currentStage >= len(stages) {
					currentStage = len(stages) - 1
				}
				return fmt.Sprintf("[%s] ", stages[currentStage])
			}, decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			// 显示百分比
			decor.Percentage(decor.WCSyncSpace),
			// 显示进度
			decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
		),
	)

	return &MultiStageBar{
		bar:      bar,
		stages:   stages,
		total:    total,
		perStage: 100,
	}
}

// AddBatchBar 添加批量操作进度条
// total: 总任务数
// 返回: BatchBar 实例
func (pm *ProgressManager) AddBatchBar(total int) *BatchBar {
	return &BatchBar{
		bar:      pm.progress.AddBar(int64(total)),
		total:    total,
		current:  0,
		itemName: "",
	}
}

// Wait 等待所有进度条完成
func (pm *ProgressManager) Wait() {
	pm.progress.Wait()
}

// MultiStageBar 多阶段进度条
type MultiStageBar struct {
	bar      *mpb.Bar
	stages   []string
	total    int64
	perStage int64
	current  int
}

// SetStage 设置当前阶段
func (msb *MultiStageBar) SetStage(stage int) {
	if stage < 0 || stage >= len(msb.stages) {
		return
	}
	msb.current = stage
	target := int64(stage) * msb.perStage
	msb.bar.SetCurrent(target)
}

// CompleteStage 完成当前阶段
func (msb *MultiStageBar) CompleteStage() {
	msb.current++
	target := int64(msb.current) * msb.perStage
	if target > msb.total {
		target = msb.total
	}
	msb.bar.SetCurrent(target)
}

// SetProgress 设置当前阶段的进度 (0-100)
func (msb *MultiStageBar) SetProgress(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	current := int64(msb.current)*msb.perStage + int64(percent)*msb.perStage/100
	msb.bar.SetCurrent(current)
}

// Complete 标记完成
func (msb *MultiStageBar) Complete() {
	msb.bar.SetCurrent(msb.total)
}

// CurrentStage 返回当前阶段索引
func (msb *MultiStageBar) CurrentStage() int {
	return msb.current
}

// TotalStages 返回总阶段数
func (msb *MultiStageBar) TotalStages() int {
	return len(msb.stages)
}

// BatchBar 批量操作进度条
type BatchBar struct {
	bar      *mpb.Bar
	total    int
	current  int
	itemName string
}

// NextItem 进入下一项
func (bb *BatchBar) NextItem(name string) {
	bb.current++
	bb.itemName = name
	bb.bar.Increment()
}

// SetItemName 设置当前项名称
func (bb *BatchBar) SetItemName(name string) {
	bb.itemName = name
}

// Current 返回当前进度
func (bb *BatchBar) Current() int {
	return bb.current
}

// Total 返回总任务数
func (bb *BatchBar) Total() int {
	return bb.total
}

// Complete 标记完成
func (bb *BatchBar) Complete() {
	bb.bar.SetCurrent(int64(bb.total))
}

// 兼容旧版 ProgressBar 的 API

// ProgressBar 兼容旧版的进度条结构
type ProgressBar struct {
	bar    *mpb.Bar
	total  int64
	prefix string
}

// NewProgressBar 创建新的进度条（兼容旧版）
func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		total: total,
	}
}

// SetPrefix 设置前缀
func (p *ProgressBar) SetPrefix(prefix string) *ProgressBar {
	p.prefix = prefix
	return p
}

// SetWidth 设置宽度（新版中忽略，由 mpb 管理）
func (p *ProgressBar) SetWidth(width int) *ProgressBar {
	return p
}

// SetWriter 设置输出（新版中忽略，由 mpb 管理）
func (p *ProgressBar) SetWriter(w io.Writer) *ProgressBar {
	return p
}

// ShowSpeed 显示速度（新版中默认显示）
func (p *ProgressBar) ShowSpeed(show bool) *ProgressBar {
	return p
}

// UseSpinner 使用旋转器（新版中忽略）
func (p *ProgressBar) UseSpinner(use bool) *ProgressBar {
	return p
}

// Add 增加进度
func (p *ProgressBar) Add(n int64) *ProgressBar {
	if p.bar != nil {
		p.bar.IncrBy(int(n))
	}
	return p
}

// Set 设置进度
func (p *ProgressBar) Set(n int64) *ProgressBar {
	if p.bar != nil {
		p.bar.SetCurrent(n)
	}
	return p
}

// Current 获取当前进度
func (p *ProgressBar) Current() int64 {
	if p.bar == nil {
		return 0
	}
	return p.bar.Current()
}

// Total 获取总进度
func (p *ProgressBar) Total() int64 {
	return p.total
}

// Done 标记完成
func (p *ProgressBar) Done() {
	if p.bar != nil {
		p.bar.SetCurrent(p.total)
	}
}

// BindToManager 将进度条绑定到管理器
func (p *ProgressBar) BindToManager(pm *ProgressManager, name string) {
	p.bar = pm.AddDownloadBar(name, p.total)
}

// DownloadProgress 下载进度（兼容旧版）
type DownloadProgress struct {
	*ProgressBar
	filename string
	url      string
}

// NewDownloadProgress 创建新的下载进度（兼容旧版）
func NewDownloadProgress(filename string, total int64) *DownloadProgress {
	return &DownloadProgress{
		ProgressBar: NewProgressBar(total),
		filename:    filename,
	}
}

// SetURL 设置 URL
func (p *DownloadProgress) SetURL(url string) *DownloadProgress {
	p.url = url
	return p
}

// Filename 获取文件名
func (p *DownloadProgress) Filename() string {
	return p.filename
}

// String 返回字符串表示
func (p *DownloadProgress) String() string {
	return fmt.Sprintf("下载: %s (%d/%d)", p.filename, p.Current(), p.Total())
}

// BindToManager 绑定到管理器
func (p *DownloadProgress) BindToManager(pm *ProgressManager) {
	p.bar = pm.AddDownloadBar(p.filename, p.total)
}

// ReaderWithProgress 带进度报告的 Reader
type ReaderWithProgress struct {
	r   io.Reader
	bar *mpb.Bar
}

// NewReaderWithProgress 创建带进度报告的 Reader
func NewReaderWithProgress(r io.Reader, bar *mpb.Bar) *ReaderWithProgress {
	return &ReaderWithProgress{
		r:   r,
		bar: bar,
	}
}

// Read 实现 io.Reader 接口
func (rp *ReaderWithProgress) Read(p []byte) (n int, err error) {
	n, err = rp.r.Read(p)
	if n > 0 && rp.bar != nil {
		rp.bar.IncrBy(n)
	}
	return n, err
}

// Close 关闭 Reader
func (rp *ReaderWithProgress) Close() error {
	if closer, ok := rp.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// ProxyReader 创建代理 Reader 用于报告进度
func (pm *ProgressManager) ProxyReader(r io.Reader, name string, total int64) io.Reader {
	bar := pm.AddDownloadBar(name, total)
	return &readerWithBar{
		Reader: r,
		bar:    bar,
	}
}

type readerWithBar struct {
	io.Reader
	bar *mpb.Bar
}

func (r *readerWithBar) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 {
		r.bar.IncrBy(n)
	}
	return
}

func (r *readerWithBar) Close() error {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
