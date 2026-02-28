package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/pkg/async"
	"chopsticks/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// LayeredParallelInstaller 分层并行安装器
type LayeredParallelInstaller struct {
	scheduler        *async.SmartDispatcher
	resolver         *DependencyResolver
	appInstaller     AppInstaller
	maxLayerParallel int
	maxDepsParallel  int
	progressCallback func(ProgressUpdate)
}

// ProgressUpdate 进度更新
type ProgressUpdate struct {
	AppName       string
	Stage         InstallStage
	Status        InstallStatus
	Progress      float64
	Message       string
	Layer         int
	TotalLayers   int
	CompletedApps int
	TotalApps     int
	Time          time.Time
}

// InstallStage 安装阶段
type InstallStage int

const (
	StagePrepare InstallStage = iota
	StageDownload
	StageVerify
	StageExtract
	StageExecuteScript
	StageRegister
	StageComplete
)

func (s InstallStage) String() string {
	switch s {
	case StagePrepare:
		return "准备"
	case StageDownload:
		return "下载"
	case StageVerify:
		return "校验"
	case StageExtract:
		return "解压"
	case StageExecuteScript:
		return "执行脚本"
	case StageRegister:
		return "注册"
	case StageComplete:
		return "完成"
	default:
		return "未知"
	}
}

// InstallStatus 安装状态
type InstallStatus int

const (
	StatusPending InstallStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusSkipped
)

func (s InstallStatus) String() string {
	switch s {
	case StatusPending:
		return "等待中"
	case StatusRunning:
		return "进行中"
	case StatusCompleted:
		return "已完成"
	case StatusFailed:
		return "失败"
	case StatusSkipped:
		return "已跳过"
	default:
		return "未知"
	}
}

// LayeredInstallSpec 分层安装规格
type LayeredInstallSpec struct {
	App        *manifest.App
	Version    string
	Arch       string
	InstallDir string
}

// LayeredInstallOptions 分层安装选项
type LayeredInstallOptions struct {
	SkipDeps         bool
	Force            bool
	MaxParallel      int
	ProgressCallback func(ProgressUpdate)
}

// NewLayeredParallelInstaller 创建分层并行安装器
func NewLayeredParallelInstaller(
	scheduler *async.SmartDispatcher,
	resolver *DependencyResolver,
	appInstaller AppInstaller,
) *LayeredParallelInstaller {
	return &LayeredParallelInstaller{
		scheduler:        scheduler,
		resolver:         resolver,
		appInstaller:     appInstaller,
		maxLayerParallel: 4,
		maxDepsParallel:  4,
	}
}

// Install 执行分层并行安装
func (i *LayeredParallelInstaller) Install(
	ctx context.Context,
	spec LayeredInstallSpec,
	opts LayeredInstallOptions,
) error {
	// 设置进度回调
	i.progressCallback = opts.ProgressCallback

	// 1. 解析依赖图
	graph, err := i.resolver.Resolve(ctx, spec.App)
	if err != nil {
		return fmt.Errorf("解析依赖失败: %w", err)
	}

	// 2. 拓扑排序分层
	layers, err := i.toLayers(graph)
	if err != nil {
		return fmt.Errorf("依赖分层失败: %w", err)
	}

	// 3. 逐层并行安装依赖
	totalApps := len(graph.Nodes)
	completedApps := 0

	for layerIdx, layer := range layers {
		if opts.SkipDeps && layerIdx < len(layers)-1 {
			// 跳过依赖安装
			continue
		}

		i.reportProgress(ProgressUpdate{
			Message:       fmt.Sprintf("正在安装第 %d/%d 层 (%d 个应用)", layerIdx+1, len(layers), len(layer)),
			Layer:         layerIdx + 1,
			TotalLayers:   len(layers),
			CompletedApps: completedApps,
			TotalApps:     totalApps,
		})

		// 并行安装当前层的所有应用
		if err := i.installLayerParallel(ctx, layer, opts, layerIdx, len(layers), &completedApps, totalApps); err != nil {
			return fmt.Errorf("安装第 %d 层失败: %w", layerIdx+1, err)
		}
	}

	// 4. 安装主应用（如果在依赖图中是最后一个）
	if !opts.SkipDeps {
		i.reportProgress(ProgressUpdate{
			AppName:       spec.App.Script.Name,
			Status:        StatusRunning,
			Message:       "安装主应用",
			Layer:         len(layers),
			TotalLayers:   len(layers),
			CompletedApps: completedApps,
			TotalApps:     totalApps,
		})

		if err := i.installApp(ctx, spec, opts); err != nil {
			return fmt.Errorf("安装主应用失败: %w", err)
		}
	}

	i.reportProgress(ProgressUpdate{
		Status:        StatusCompleted,
		Message:       fmt.Sprintf("%s 安装完成", spec.App.Script.Name),
		CompletedApps: totalApps,
		TotalApps:     totalApps,
	})

	return nil
}

// toLayers 将依赖图转换为分层结构
func (i *LayeredParallelInstaller) toLayers(graph *DependencyGraph) ([][]*DependencyNode, error) {
	if len(graph.Nodes) == 0 {
		return nil, nil
	}

	// 计算每个节点的深度（到根节点的最大距离）
	depths := make(map[string]int)
	
	var calculateDepth func(node *DependencyNode, visited map[string]bool) int
	calculateDepth = func(node *DependencyNode, visited map[string]bool) int {
		if depth, ok := depths[node.App.Script.Name]; ok {
			return depth
		}

		if visited[node.App.Script.Name] {
			// 循环依赖已在 Resolve 中检测，这里不应该发生
			return 0
		}

		visited[node.App.Script.Name] = true
		maxDepth := 0

		for _, dep := range node.Dependencies {
			depDepth := calculateDepth(dep, visited)
			if depDepth > maxDepth {
				maxDepth = depDepth
			}
		}

		delete(visited, node.App.Script.Name)
		depths[node.App.Script.Name] = maxDepth + 1
		return maxDepth + 1
	}

	// 计算所有节点的深度
	for _, node := range graph.Nodes {
		calculateDepth(node, make(map[string]bool))
	}

	// 按深度分组
	maxDepth := 0
	for _, depth := range depths {
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// 创建层（从最深到最浅，即依赖在前）
	layers := make([][]*DependencyNode, maxDepth)
	for name, depth := range depths {
		// depth 是 1-based，转换为 0-based 索引
		layerIdx := maxDepth - depth
		layers[layerIdx] = append(layers[layerIdx], graph.Nodes[name])
	}

	return layers, nil
}

// installLayerParallel 并行安装一层
func (i *LayeredParallelInstaller) installLayerParallel(
	ctx context.Context,
	layer []*DependencyNode,
	opts LayeredInstallOptions,
	layerIdx, totalLayers int,
	completedApps *int,
	totalApps int,
) error {
	if len(layer) == 0 {
		return nil
	}

	// 使用 errgroup 控制并发
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(i.maxLayerParallel)

	// 进度同步
	var mu sync.Mutex

	for _, node := range layer {
		node := node // 捕获循环变量
		g.Go(func() error {
			spec := LayeredInstallSpec{
				App:     node.App,
				Version: node.Version,
				Arch:    "amd64",
			}

			i.reportProgress(ProgressUpdate{
				AppName:       node.App.Script.Name,
				Status:        StatusRunning,
				Message:       fmt.Sprintf("开始安装 %s", node.App.Script.Name),
				Layer:         layerIdx + 1,
				TotalLayers:   totalLayers,
				CompletedApps: *completedApps,
				TotalApps:     totalApps,
			})

			if err := i.installApp(ctx, spec, opts); err != nil {
				i.reportProgress(ProgressUpdate{
					AppName: node.App.Script.Name,
					Status:  StatusFailed,
					Message: fmt.Sprintf("安装 %s 失败: %v", node.App.Script.Name, err),
				})
				return fmt.Errorf("安装 %s 失败: %w", node.App.Script.Name, err)
			}

			mu.Lock()
			*completedApps++
			currentCompleted := *completedApps
			mu.Unlock()

			i.reportProgress(ProgressUpdate{
				AppName:       node.App.Script.Name,
				Status:        StatusCompleted,
				Message:       fmt.Sprintf("%s 安装完成", node.App.Script.Name),
				Layer:         layerIdx + 1,
				TotalLayers:   totalLayers,
				CompletedApps: currentCompleted,
				TotalApps:     totalApps,
			})

			return nil
		})
	}

	return g.Wait()
}

// installApp 安装单个应用
func (i *LayeredParallelInstaller) installApp(ctx context.Context, spec LayeredInstallSpec, opts LayeredInstallOptions) error {
	appOpts := InstallOptions{
		Arch:       spec.Arch,
		Force:      opts.Force,
		InstallDir: spec.InstallDir,
	}

	return i.appInstaller.Install(ctx, spec.App, appOpts)
}

// reportProgress 报告进度
func (i *LayeredParallelInstaller) reportProgress(update ProgressUpdate) {
	update.Time = time.Now()
	if i.progressCallback != nil {
		i.progressCallback(update)
	}
}

// SetMaxParallel 设置最大并行数
func (i *LayeredParallelInstaller) SetMaxParallel(layerParallel, depsParallel int) {
	if layerParallel > 0 {
		i.maxLayerParallel = layerParallel
	}
	if depsParallel > 0 {
		i.maxDepsParallel = depsParallel
	}
}

// LayerInfo 层信息
type LayerInfo struct {
	Index    int
	Apps     []string
	Count    int
	CanStart bool
}

// GetLayerInfo 获取分层信息（用于调试和监控）
func (i *LayeredParallelInstaller) GetLayerInfo(graph *DependencyGraph) ([]LayerInfo, error) {
	layers, err := i.toLayers(graph)
	if err != nil {
		return nil, err
	}

	info := make([]LayerInfo, len(layers))
	for idx, layer := range layers {
		apps := make([]string, len(layer))
		for i, node := range layer {
			apps[i] = node.App.Script.Name
		}
		info[idx] = LayerInfo{
			Index:    idx,
			Apps:     apps,
			Count:    len(layer),
			CanStart: idx == 0, // 第一层可以立即开始
		}
	}

	return info, nil
}

// InstallBatch 批量安装多个应用
func (i *LayeredParallelInstaller) InstallBatch(
	ctx context.Context,
	specs []LayeredInstallSpec,
	opts LayeredInstallOptions,
) error {
	if len(specs) == 0 {
		return nil
	}

	// 如果没有依赖关系，直接并行安装
	if opts.SkipDeps {
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(i.maxLayerParallel)

		for _, spec := range specs {
			spec := spec
			g.Go(func() error {
				return i.installApp(ctx, spec, opts)
			})
		}

		return g.Wait()
	}

	// 构建合并的依赖图
	mergedGraph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}

	for _, spec := range specs {
		graph, err := i.resolver.Resolve(ctx, spec.App)
		if err != nil {
			return fmt.Errorf("解析 %s 的依赖失败: %w", spec.App.Script.Name, err)
		}

		// 合并节点
		for name, node := range graph.Nodes {
			if _, exists := mergedGraph.Nodes[name]; !exists {
				mergedGraph.Nodes[name] = node
			}
		}
	}

	// 重新拓扑排序
	if err := i.topologicalSort(mergedGraph); err != nil {
		return err
	}

	// 分层安装
	layers, err := i.toLayers(mergedGraph)
	if err != nil {
		return err
	}

	completedApps := 0
	totalApps := len(mergedGraph.Nodes)

	for layerIdx, layer := range layers {
		i.reportProgress(ProgressUpdate{
			Message:       fmt.Sprintf("批量安装 - 第 %d/%d 层", layerIdx+1, len(layers)),
			Layer:         layerIdx + 1,
			TotalLayers:   len(layers),
			CompletedApps: completedApps,
			TotalApps:     totalApps,
		})

		if err := i.installLayerParallel(ctx, layer, opts, layerIdx, len(layers), &completedApps, totalApps); err != nil {
			return err
		}
	}

	return nil
}

// topologicalSort 拓扑排序（复制自 DependencyResolver）
func (i *LayeredParallelInstaller) topologicalSort(graph *DependencyGraph) error {
	inDegree := make(map[string]int)

	for name, node := range graph.Nodes {
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
		for _, dep := range node.Dependencies {
			inDegree[dep.App.Script.Name]++
		}
	}

	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0, len(graph.Nodes))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		result = append(result, name)

		node := graph.Nodes[name]
		for _, dep := range node.Dependencies {
			depName := dep.App.Script.Name
			inDegree[depName]--
			if inDegree[depName] == 0 {
				queue = append(queue, depName)
			}
		}
	}

	if len(result) != len(graph.Nodes) {
		return errors.Newf(errors.KindInvalidInput, "依赖图中存在循环依赖")
	}

	// 反转结果
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	graph.Order = result
	return nil
}

// InstallMetrics 安装指标
type InstallMetrics struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalApps     int
	CompletedApps int
	FailedApps    int
	SkippedApps   int
	TotalDuration time.Duration
	AvgDuration   time.Duration
}

// CalculateMetrics 计算安装指标
func CalculateMetrics(updates []ProgressUpdate) InstallMetrics {
	if len(updates) == 0 {
		return InstallMetrics{}
	}

	metrics := InstallMetrics{
		StartTime: updates[0].Time,
		EndTime:   updates[len(updates)-1].Time,
	}

	seenApps := make(map[string]bool)
	for _, update := range updates {
		if update.AppName != "" && !seenApps[update.AppName] {
			seenApps[update.AppName] = true
			metrics.TotalApps++
		}

		switch update.Status {
		case StatusCompleted:
			metrics.CompletedApps++
		case StatusFailed:
			metrics.FailedApps++
		case StatusSkipped:
			metrics.SkippedApps++
		}
	}

	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
	if metrics.CompletedApps > 0 {
		metrics.AvgDuration = metrics.TotalDuration / time.Duration(metrics.CompletedApps)
	}

	return metrics
}
