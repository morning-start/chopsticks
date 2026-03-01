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
		return "Prepare"
	case StageDownload:
		return "Download"
	case StageVerify:
		return "Verify"
	case StageExtract:
		return "Extract"
	case StageExecuteScript:
		return "Execute Script"
	case StageRegister:
		return "Register"
	case StageComplete:
		return "Complete"
	default:
		return "Unknown"
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
		return "Pending"
	case StatusRunning:
		return "Running"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusSkipped:
		return "Skipped"
	default:
		return "Unknown"
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
	// Set progress callback
	i.progressCallback = opts.ProgressCallback

	// 1. Resolve dependency graph
	graph, err := i.resolver.Resolve(ctx, spec.App)
	if err != nil {
		return fmt.Errorf("resolve dependencies: %w", err)
	}

	// 2. Topological sort into layers
	layers, err := i.toLayers(graph)
	if err != nil {
		return fmt.Errorf("layer dependencies: %w", err)
	}

	// 3. Install dependencies layer by layer in parallel
	totalApps := len(graph.Nodes)
	completedApps := 0

	for layerIdx, layer := range layers {
		if opts.SkipDeps && layerIdx < len(layers)-1 {
			// Skip dependency installation
			continue
		}

		i.reportProgress(ProgressUpdate{
			Message:       fmt.Sprintf("Installing layer %d/%d (%d apps)", layerIdx+1, len(layers), len(layer)),
			Layer:         layerIdx + 1,
			TotalLayers:   len(layers),
			CompletedApps: completedApps,
			TotalApps:     totalApps,
		})

		// Install all apps in current layer in parallel
		if err := i.installLayerParallel(ctx, layer, opts, layerIdx, len(layers), &completedApps, totalApps); err != nil {
			return fmt.Errorf("install layer %d: %w", layerIdx+1, err)
		}
	}

	// 4. Install main app (if it's the last in dependency graph)
	if !opts.SkipDeps {
		i.reportProgress(ProgressUpdate{
			AppName:       spec.App.Script.Name,
			Status:        StatusRunning,
			Message:       "Installing main app",
			Layer:         len(layers),
			TotalLayers:   len(layers),
			CompletedApps: completedApps,
			TotalApps:     totalApps,
		})

		if err := i.installApp(ctx, spec, opts); err != nil {
			return fmt.Errorf("install main app: %w", err)
		}
	}

	i.reportProgress(ProgressUpdate{
		Status:        StatusCompleted,
		Message:       fmt.Sprintf("%s installation complete", spec.App.Script.Name),
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

	// Calculate depth for each node (max distance to leaf/dependency)
	// A node with no dependencies has depth 0
	// A node that depends on others has depth = max(dependency depths) + 1
	depths := make(map[string]int)

	var calculateDepth func(node *DependencyNode, visited map[string]bool) int
	calculateDepth = func(node *DependencyNode, visited map[string]bool) int {
		if depth, ok := depths[node.App.Script.Name]; ok {
			return depth
		}

		if visited[node.App.Script.Name] {
			// Circular dependency should have been detected in Resolve
			return 0
		}

		visited[node.App.Script.Name] = true
		maxDepDepth := -1

		for _, dep := range node.Dependencies {
			depDepth := calculateDepth(dep, visited)
			if depDepth > maxDepDepth {
				maxDepDepth = depDepth
			}
		}

		delete(visited, node.App.Script.Name)

		// If no dependencies, depth is 0
		// Otherwise, depth is max dependency depth + 1
		depth := 0
		if maxDepDepth >= 0 {
			depth = maxDepDepth + 1
		}
		depths[node.App.Script.Name] = depth
		return depth
	}

	// Calculate depths for all nodes
	for _, node := range graph.Nodes {
		calculateDepth(node, make(map[string]bool))
	}

	// Find max depth
	maxDepth := 0
	for _, depth := range depths {
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Create layers - dependencies (depth 0) go in layer 0, then depth 1, etc.
	layers := make([][]*DependencyNode, maxDepth+1)
	for name, depth := range depths {
		layers[depth] = append(layers[depth], graph.Nodes[name])
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

	// Use errgroup for concurrency control
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(i.maxLayerParallel)

	// Progress synchronization
	var mu sync.Mutex

	for _, node := range layer {
		node := node // Capture loop variable
		g.Go(func() error {
			spec := LayeredInstallSpec{
				App:     node.App,
				Version: node.Version,
				Arch:    DefaultArch,
			}

			i.reportProgress(ProgressUpdate{
				AppName:       node.App.Script.Name,
				Status:        StatusRunning,
				Message:       fmt.Sprintf("Starting installation of %s", node.App.Script.Name),
				Layer:         layerIdx + 1,
				TotalLayers:   totalLayers,
				CompletedApps: *completedApps,
				TotalApps:     totalApps,
			})

			if err := i.installApp(ctx, spec, opts); err != nil {
				i.reportProgress(ProgressUpdate{
					AppName: node.App.Script.Name,
					Status:  StatusFailed,
					Message: fmt.Sprintf("Installation of %s failed: %v", node.App.Script.Name, err),
				})
				return fmt.Errorf("install %s: %w", node.App.Script.Name, err)
			}

			mu.Lock()
			*completedApps++
			currentCompleted := *completedApps
			mu.Unlock()

			i.reportProgress(ProgressUpdate{
				AppName:       node.App.Script.Name,
				Status:        StatusCompleted,
				Message:       fmt.Sprintf("%s installation complete", node.App.Script.Name),
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
			CanStart: idx == 0, // First layer can start immediately
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

	// If no dependencies, install in parallel directly
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

	// Build merged dependency graph
	mergedGraph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Order: []string{},
	}

	for _, spec := range specs {
		graph, err := i.resolver.Resolve(ctx, spec.App)
		if err != nil {
			return fmt.Errorf("resolve dependencies for %s: %w", spec.App.Script.Name, err)
		}

		// Merge nodes
		for name, node := range graph.Nodes {
			if _, exists := mergedGraph.Nodes[name]; !exists {
				mergedGraph.Nodes[name] = node
			}
		}
	}

	// Re-topological sort
	if err := i.topologicalSort(mergedGraph); err != nil {
		return err
	}

	// Layered installation
	layers, err := i.toLayers(mergedGraph)
	if err != nil {
		return err
	}

	completedApps := 0
	totalApps := len(mergedGraph.Nodes)

	for layerIdx, layer := range layers {
		i.reportProgress(ProgressUpdate{
			Message:       fmt.Sprintf("Batch install - Layer %d/%d", layerIdx+1, len(layers)),
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
		return errors.Newf(errors.KindInvalidInput, "circular dependency detected in dependency graph")
	}

	// Reverse result
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
