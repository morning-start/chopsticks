package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/engine/checksum"
	"chopsticks/pkg/async"
)

// mockAppInstaller 模拟应用安装器
type mockAppInstaller struct {
	installFunc func(ctx context.Context, app *manifest.App, opts InstallOptions) error
	callCount   int
}

func (m *mockAppInstaller) Install(ctx context.Context, app *manifest.App, opts InstallOptions) error {
	m.callCount++
	if m.installFunc != nil {
		return m.installFunc(ctx, app, opts)
	}
	return nil
}

func (m *mockAppInstaller) Download(url, dest string) error {
	return nil
}

func (m *mockAppInstaller) Verify(path, hash string, alg checksum.Algorithm) error {
	return nil
}

func (m *mockAppInstaller) Extract(src, dest string) error {
	return nil
}

// TestNewLayeredParallelInstaller 测试创建分层并行安装器
func TestNewLayeredParallelInstaller(t *testing.T) {
	dispatcher := async.NewSmartDispatcher(async.DefaultConfig())
	resolver := &DependencyResolver{}
	appInstaller := &mockAppInstaller{}

	installer := NewLayeredParallelInstaller(dispatcher, resolver, appInstaller)

	if installer == nil {
		t.Fatal("NewLayeredParallelInstaller returned nil")
	}

	if installer.maxLayerParallel != 4 {
		t.Errorf("maxLayerParallel = %d, want 4", installer.maxLayerParallel)
	}

	if installer.maxDepsParallel != 4 {
		t.Errorf("maxDepsParallel = %d, want 4", installer.maxDepsParallel)
	}
}

// TestLayeredParallelInstaller_SetMaxParallel 测试设置最大并行数
func TestLayeredParallelInstaller_SetMaxParallel(t *testing.T) {
	installer := &LayeredParallelInstaller{
		maxLayerParallel: 4,
		maxDepsParallel:  4,
	}

	installer.SetMaxParallel(8, 6)

	if installer.maxLayerParallel != 8 {
		t.Errorf("maxLayerParallel = %d, want 8", installer.maxLayerParallel)
	}

	if installer.maxDepsParallel != 6 {
		t.Errorf("maxDepsParallel = %d, want 6", installer.maxDepsParallel)
	}
}

// TestInstallStage_String 测试安装阶段字符串
func TestInstallStage_String(t *testing.T) {
	tests := []struct {
		stage InstallStage
		want  string
	}{
		{StagePrepare, "Prepare"},
		{StageDownload, "Download"},
		{StageVerify, "Verify"},
		{StageExtract, "Extract"},
		{StageExecuteScript, "Execute Script"},
		{StageRegister, "Register"},
		{StageComplete, "Complete"},
		{InstallStage(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.stage.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestInstallStatus_String 测试安装状态字符串
func TestInstallStatus_String(t *testing.T) {
	tests := []struct {
		status InstallStatus
		want   string
	}{
		{StatusPending, "Pending"},
		{StatusRunning, "Running"},
		{StatusCompleted, "Completed"},
		{StatusFailed, "Failed"},
		{StatusSkipped, "Skipped"},
		{InstallStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLayeredParallelInstaller_toLayers 测试依赖图分层
func TestLayeredParallelInstaller_toLayers(t *testing.T) {
	installer := &LayeredParallelInstaller{}

	// 创建测试依赖图
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"app-a": {
				App: &manifest.App{
					Script: &manifest.AppScript{Name: "app-a"},
				},
				Dependencies: []*DependencyNode{
					{
						App: &manifest.App{
							Script: &manifest.AppScript{Name: "dep-1"},
						},
					},
				},
			},
			"dep-1": {
				App: &manifest.App{
					Script: &manifest.AppScript{Name: "dep-1"},
				},
			},
		},
	}

	layers, err := installer.toLayers(graph)
	if err != nil {
		t.Fatalf("toLayers failed: %v", err)
	}

	// 应该有两层：dep-1 在第一层，app-a 在第二层
	if len(layers) != 2 {
		t.Errorf("Expected 2 layers, got %d", len(layers))
	}

	// 第一层应该是 dep-1
	if len(layers) > 0 {
		found := false
		for _, node := range layers[0] {
			if node.App.Script.Name == "dep-1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected dep-1 in first layer")
		}
	}

	// 第二层应该是 app-a
	if len(layers) > 1 {
		found := false
		for _, node := range layers[1] {
			if node.App.Script.Name == "app-a" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected app-a in second layer")
		}
	}
}

// TestLayeredParallelInstaller_reportProgress 测试进度报告
func TestLayeredParallelInstaller_reportProgress(t *testing.T) {
	var receivedUpdate ProgressUpdate
	installer := &LayeredParallelInstaller{
		progressCallback: func(update ProgressUpdate) {
			receivedUpdate = update
		},
	}

	installer.reportProgress(ProgressUpdate{
		AppName: "test-app",
		Status:  StatusRunning,
		Message: "Testing",
	})

	if receivedUpdate.AppName != "test-app" {
		t.Errorf("AppName = %s, want test-app", receivedUpdate.AppName)
	}

	if receivedUpdate.Status != StatusRunning {
		t.Errorf("Status = %v, want StatusRunning", receivedUpdate.Status)
	}

	if receivedUpdate.Time.IsZero() {
		t.Error("Time should be set")
	}
}

// TestCalculateMetrics 测试计算安装指标
func TestCalculateMetrics(t *testing.T) {
	now := time.Now()
	updates := []ProgressUpdate{
		{AppName: "app1", Status: StatusRunning, Time: now},
		{AppName: "app1", Status: StatusCompleted, Time: now.Add(1 * time.Second)},
		{AppName: "app2", Status: StatusRunning, Time: now.Add(2 * time.Second)},
		{AppName: "app2", Status: StatusFailed, Time: now.Add(3 * time.Second)},
	}

	metrics := CalculateMetrics(updates)

	if metrics.TotalApps != 2 {
		t.Errorf("TotalApps = %d, want 2", metrics.TotalApps)
	}

	if metrics.CompletedApps != 1 {
		t.Errorf("CompletedApps = %d, want 1", metrics.CompletedApps)
	}

	if metrics.FailedApps != 1 {
		t.Errorf("FailedApps = %d, want 1", metrics.FailedApps)
	}

	if metrics.TotalDuration <= 0 {
		t.Error("TotalDuration should be > 0")
	}
}

// TestPipelineStage_String 测试流水线阶段字符串
func TestPipelineStage_String(t *testing.T) {
	tests := []struct {
		stage PipelineStage
		want  string
	}{
		{PipelineDownload, "Download"},
		{PipelineVerify, "Verify"},
		{PipelineExtract, "Extract"},
		{PipelineExecute, "Execute Script"},
		{PipelineRegister, "Register"},
		{PipelineComplete, "Complete"},
		{PipelineStage(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.stage.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewPipelineInstaller 测试创建流水线安装器
func TestNewPipelineInstaller(t *testing.T) {
	dispatcher := async.NewSmartDispatcher(async.DefaultConfig())
	installer := NewPipelineInstaller(dispatcher)

	if installer == nil {
		t.Fatal("NewPipelineInstaller returned nil")
	}

	if installer.bufferSize != 10 {
		t.Errorf("bufferSize = %d, want 10", installer.bufferSize)
	}

	if len(installer.stages) != 0 {
		t.Errorf("stages length = %d, want 0", len(installer.stages))
	}
}

// TestDefaultPipelineOptions 测试默认流水线选项
func TestDefaultPipelineOptions(t *testing.T) {
	opts := DefaultPipelineOptions()

	if opts.MaxConcurrency != 4 {
		t.Errorf("MaxConcurrency = %d, want 4", opts.MaxConcurrency)
	}
}

// TestPipelineBuilder 测试流水线构建器
func TestPipelineBuilder(t *testing.T) {
	dispatcher := async.NewSmartDispatcher(async.DefaultConfig())
	builder := NewPipelineBuilder(dispatcher)

	// 创建模拟的阶段处理器
	mockProcessor := &mockStageProcessor{}

	installer := builder.
		WithDownloadStage(mockProcessor).
		WithVerifyStage(mockProcessor).
		WithExtractStage(mockProcessor).
		Build()

	if installer == nil {
		t.Fatal("Build returned nil")
	}

	// 验证阶段已注册
	if _, ok := installer.stages[PipelineDownload]; !ok {
		t.Error("Download stage not registered")
	}
	if _, ok := installer.stages[PipelineVerify]; !ok {
		t.Error("Verify stage not registered")
	}
	if _, ok := installer.stages[PipelineExtract]; !ok {
		t.Error("Extract stage not registered")
	}
}

// mockStageProcessor 模拟阶段处理器
type mockStageProcessor struct {
	category async.TaskCategory
}

func (m *mockStageProcessor) Process(ctx context.Context, task *StageTask) (*StageTask, error) {
	return task, nil
}

func (m *mockStageProcessor) GetCategory() async.TaskCategory {
	return m.category
}

// TestLayerInfo 测试层信息
func TestLayerInfo(t *testing.T) {
	info := LayerInfo{
		Index:    0,
		Apps:     []string{"app1", "app2"},
		Count:    2,
		CanStart: true,
	}

	if info.Index != 0 {
		t.Errorf("Index = %d, want 0", info.Index)
	}

	if len(info.Apps) != 2 {
		t.Errorf("Apps length = %d, want 2", len(info.Apps))
	}

	if info.Count != 2 {
		t.Errorf("Count = %d, want 2", info.Count)
	}

	if !info.CanStart {
		t.Error("CanStart should be true")
	}
}

// BenchmarkLayeredInstall 基准测试分层安装
func BenchmarkLayeredInstall(b *testing.B) {
	installer := &LayeredParallelInstaller{
		maxLayerParallel: 4,
	}

	// 创建测试依赖图
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// 添加10个应用
	for i := 0; i < 10; i++ {
		graph.Nodes[fmt.Sprintf("app-%d", i)] = &DependencyNode{
			App: &manifest.App{
				Script: &manifest.AppScript{Name: fmt.Sprintf("app-%d", i)},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := installer.toLayers(graph)
		if err != nil {
			b.Fatal(err)
		}
	}
}
