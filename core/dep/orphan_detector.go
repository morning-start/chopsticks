// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chopsticks/core/manifest"
)

// OrphanDetector 孤儿依赖检测器接口
type OrphanDetector interface {
	// Detect 检测孤儿依赖
	Detect(ctx context.Context) (*manifest.Orphans, error)
	// DetectRuntimeOrphans 检测孤儿运行时库
	DetectRuntimeOrphans(ctx context.Context) ([]string, error)
	// DetectToolOrphans 检测孤儿工具软件
	DetectToolOrphans(ctx context.Context) ([]string, error)
	// Cleanup 清理孤儿依赖
	Cleanup(ctx context.Context, orphans *manifest.Orphans) error
	// CleanupRuntime 清理孤儿运行时库
	CleanupRuntime(ctx context.Context, orphans []string) error
	// CleanupTools 清理孤儿工具软件
	CleanupTools(ctx context.Context, orphans []string) error
	// GetOrphanInfo 获取孤儿依赖信息
	GetOrphanInfo(ctx context.Context, name string) (*OrphanInfo, error)
	// DryRun 预演清理操作
	DryRun(ctx context.Context, orphans *manifest.Orphans) error
}

// OrphanInfo 表示孤儿依赖信息
type OrphanInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "runtime" or "tool"
	Size        int64  `json:"size"`
	InstalledAt string `json:"installed_at"`
	Reason      string `json:"reason"`
}

// orphanDetector 孤儿依赖检测器实现
type orphanDetector struct {
	mu             sync.RWMutex
	runtimeManager RuntimeManager
	depsIndex      *DepsIndex
	rootPath       string
	installDir     string
	runtimeDir     string
}

// NewOrphanDetector 创建孤儿依赖检测器
func NewOrphanDetector(rootPath string, runtimeMgr RuntimeManager, depsIndex *DepsIndex) OrphanDetector {
	return &orphanDetector{
		runtimeManager: runtimeMgr,
		depsIndex:      depsIndex,
		rootPath:       rootPath,
		installDir:     filepath.Join(rootPath, "apps"),
		runtimeDir:     filepath.Join(rootPath, "runtimes"),
	}
}

// Detect 检测孤儿依赖
func (d *orphanDetector) Detect(ctx context.Context) (*manifest.Orphans, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 检测孤儿运行时库
	runtimeOrphans, err := d.DetectRuntimeOrphans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect runtime orphans: %w", err)
	}

	// 检测孤儿工具软件
	toolOrphans, err := d.DetectToolOrphans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect tool orphans: %w", err)
	}

	return &manifest.Orphans{
		Runtime: runtimeOrphans,
		Tools:   toolOrphans,
	}, nil
}

// DetectRuntimeOrphans 检测孤儿运行时库
func (d *orphanDetector) DetectRuntimeOrphans(ctx context.Context) ([]string, error) {
	// 使用运行时管理器的引用计数检测
	if d.runtimeManager != nil {
		// 获取所有运行时库
		runtimes := d.runtimeManager.List(ctx)

		var orphans []string
		for name, info := range runtimes {
			if info.RefCount <= 0 {
				orphans = append(orphans, name)
			}
		}

		return orphans, nil
	}

	// 备用方案：检查运行时目录
	return d.detectRuntimeOrphansFromFS(ctx)
}

// detectRuntimeOrphansFromFS 从文件系统检测孤儿运行时库
func (d *orphanDetector) detectRuntimeOrphansFromFS(ctx context.Context) ([]string, error) {
	// 扫描运行时目录
	entries, err := os.ReadDir(d.runtimeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read runtime directory: %w", err)
	}

	// 加载依赖索引
	if err := d.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	// 构建所有被依赖的运行时库集合
	dependedRuntimes := make(map[string]bool)
	for _, appDeps := range d.depsIndex.apps {
		for _, dep := range appDeps.Dependencies {
			// 检查是否是运行时库（在运行时目录中存在）
			runtimePath := filepath.Join(d.runtimeDir, dep)
			if _, err := os.Stat(runtimePath); err == nil {
				dependedRuntimes[dep] = true
			}
		}
	}

	// 找出没有被依赖的运行时库
	var orphans []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		runtimeName := entry.Name()
		if !dependedRuntimes[runtimeName] {
			orphans = append(orphans, runtimeName)
		}
	}

	return orphans, nil
}

// DetectToolOrphans 检测孤儿工具软件
func (d *orphanDetector) DetectToolOrphans(ctx context.Context) ([]string, error) {
	// 加载依赖索引
	if err := d.depsIndex.Load(); err != nil {
		return nil, fmt.Errorf("failed to load deps index: %w", err)
	}

	// 扫描所有已安装的软件
	entries, err := os.ReadDir(d.installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read apps directory: %w", err)
	}

	// 构建所有被依赖的软件集合
	dependedApps := make(map[string]bool)
	for _, appDeps := range d.depsIndex.apps {
		for _, dep := range appDeps.Dependencies {
			dependedApps[dep] = true
		}
	}

	// 找出没有被依赖且不是用户主动安装的软件
	var orphans []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appName := entry.Name()
		// 如果是被依赖的，则不是孤儿
		if dependedApps[appName] {
			continue
		}

		// 检查是否是工具软件（通过标记文件判断）
		toolMarker := filepath.Join(d.installDir, appName, ".tool")
		if _, err := os.Stat(toolMarker); err == nil {
			// 是工具软件且没有被依赖，是孤儿
			orphans = append(orphans, appName)
		}
	}

	return orphans, nil
}

// Cleanup 清理孤儿依赖
func (d *orphanDetector) Cleanup(ctx context.Context, orphans *manifest.Orphans) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 清理孤儿运行时库
	if len(orphans.Runtime) > 0 {
		fmt.Printf("清理 %d 个孤儿运行时库：\n", len(orphans.Runtime))
		if err := d.CleanupRuntime(ctx, orphans.Runtime); err != nil {
			return fmt.Errorf("failed to cleanup runtime orphans: %w", err)
		}
	}

	// 清理孤儿工具软件
	if len(orphans.Tools) > 0 {
		fmt.Printf("清理 %d 个孤儿工具软件：\n", len(orphans.Tools))
		if err := d.CleanupTools(ctx, orphans.Tools); err != nil {
			return fmt.Errorf("failed to cleanup tool orphans: %w", err)
		}
	}

	fmt.Println("孤儿依赖清理完成")
	return nil
}

// CleanupRuntime 清理孤儿运行时库
func (d *orphanDetector) CleanupRuntime(ctx context.Context, orphans []string) error {
	for _, runtime := range orphans {
		// 获取运行时信息
		info, err := d.GetOrphanInfo(ctx, runtime)
		if err != nil {
			fmt.Printf("  ✗ 获取 %s 信息失败：%v\n", runtime, err)
			continue
		}

		// 删除运行时目录
		runtimePath := filepath.Join(d.runtimeDir, runtime)
		if err := os.RemoveAll(runtimePath); err != nil {
			fmt.Printf("  ✗ 清理 %s 失败：%v\n", runtime, err)
			continue
		}

		fmt.Printf("  ✓ 已清理：%s (大小：%s)\n", runtime, formatSize(info.Size))
	}

	return nil
}

// CleanupTools 清理孤儿工具软件
func (d *orphanDetector) CleanupTools(ctx context.Context, orphans []string) error {
	for _, tool := range orphans {
		// 获取工具信息
		info, err := d.GetOrphanInfo(ctx, tool)
		if err != nil {
			fmt.Printf("  ✗ 获取 %s 信息失败：%v\n", tool, err)
			continue
		}

		// 删除工具目录
		toolPath := filepath.Join(d.installDir, tool)
		if err := os.RemoveAll(toolPath); err != nil {
			fmt.Printf("  ✗ 清理 %s 失败：%v\n", tool, err)
			continue
		}

		fmt.Printf("  ✓ 已清理：%s (大小：%s)\n", tool, formatSize(info.Size))
	}

	return nil
}

// GetOrphanInfo 获取孤儿依赖信息
func (d *orphanDetector) GetOrphanInfo(ctx context.Context, name string) (*OrphanInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 检查是否是运行时库
	runtimePath := filepath.Join(d.runtimeDir, name)
	if _, err := os.Stat(runtimePath); err == nil {
		// 是运行时库
		size, err := d.calculateDirSize(runtimePath)
		if err != nil {
			size = 0
		}

		return &OrphanInfo{
			Name:   name,
			Type:   "runtime",
			Size:   size,
			Reason: "no dependent applications",
		}, nil
	}

	// 检查是否是工具软件
	toolPath := filepath.Join(d.installDir, name)
	if _, err := os.Stat(toolPath); err == nil {
		// 是工具软件
		size, err := d.calculateDirSize(toolPath)
		if err != nil {
			size = 0
		}

		return &OrphanInfo{
			Name:   name,
			Type:   "tool",
			Size:   size,
			Reason: "no dependent applications",
		}, nil
	}

	return nil, fmt.Errorf("orphan %s not found", name)
}

// calculateDirSize 计算目录大小
func (d *orphanDetector) calculateDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// DryRun 预演清理操作
func (d *orphanDetector) DryRun(ctx context.Context, orphans *manifest.Orphans) error {
	fmt.Println("=== 孤儿依赖清理预览 ===")
	fmt.Println()

	if len(orphans.Runtime) > 0 {
		fmt.Printf("将清理 %d 个孤儿运行时库：\n", len(orphans.Runtime))
		for _, runtime := range orphans.Runtime {
			info, err := d.GetOrphanInfo(ctx, runtime)
			if err != nil {
				fmt.Printf("  - %s (获取信息失败)\n", runtime)
				continue
			}
			fmt.Printf("  - %s (%s)\n", runtime, formatSize(info.Size))
		}
		fmt.Println()
	}

	if len(orphans.Tools) > 0 {
		fmt.Printf("将清理 %d 个孤儿工具软件：\n", len(orphans.Tools))
		for _, tool := range orphans.Tools {
			info, err := d.GetOrphanInfo(ctx, tool)
			if err != nil {
				fmt.Printf("  - %s (获取信息失败)\n", tool)
				continue
			}
			fmt.Printf("  - %s (%s)\n", tool, formatSize(info.Size))
		}
		fmt.Println()
	}

	totalSize, err := d.calculateTotalOrphanSize(ctx, orphans)
	if err != nil {
		fmt.Printf("无法计算总大小：%v\n", err)
	} else {
		fmt.Printf("预计释放空间：%s\n", formatSize(totalSize))
	}

	fmt.Println()
	return nil
}

// calculateTotalOrphanSize 计算孤儿依赖总大小
func (d *orphanDetector) calculateTotalOrphanSize(ctx context.Context, orphans *manifest.Orphans) (int64, error) {
	var totalSize int64

	for _, runtime := range orphans.Runtime {
		info, err := d.GetOrphanInfo(ctx, runtime)
		if err != nil {
			continue
		}
		totalSize += info.Size
	}

	for _, tool := range orphans.Tools {
		info, err := d.GetOrphanInfo(ctx, tool)
		if err != nil {
			continue
		}
		totalSize += info.Size
	}

	return totalSize, nil
}
