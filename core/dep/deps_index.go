// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chopsticks/core/manifest"
)

// DepsIndex 表示依赖索引（可重建缓存）
type DepsIndex struct {
	mu       sync.RWMutex
	apps     map[string]*AppDeps
	filePath string
}

// AppDeps 表示应用的依赖信息
type AppDeps struct {
	Dependencies []string // 依赖列表
	Dependents  []string // 反向依赖（谁依赖我）
}

// NewDepsIndex 创建依赖索引
func NewDepsIndex(rootPath string) *DepsIndex {
	return &DepsIndex{
		apps:     make(map[string]*AppDeps),
		filePath: filepath.Join(rootPath, "deps-index.json"),
	}
}

// Load 加载依赖索引
func (d *DepsIndex) Load() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := os.ReadFile(d.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，初始化为空
			d.apps = make(map[string]*AppDeps)
			return nil
		}
		return fmt.Errorf("failed to load deps index: %w", err)
	}

	if err := json.Unmarshal(data, &d.apps); err != nil {
		return fmt.Errorf("failed to unmarshal deps index: %w", err)
	}

	return nil
}

// Save 保存依赖索引
func (d *DepsIndex) Save() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	data, err := json.MarshalIndent(d.apps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deps index: %w", err)
	}

	if err := os.WriteFile(d.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save deps index: %w", err)
	}

	return nil
}

// Get 获取应用的依赖信息
func (d *DepsIndex) Get(name string) (*AppDeps, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	info, ok := d.apps[name]
	return info, ok
}

// GetDependents 获取反向依赖（谁依赖我）
func (d *DepsIndex) GetDependents(name string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var dependents []string
	for appName, appDeps := range d.apps {
		for _, dep := range appDeps.Dependencies {
			if dep == name {
				dependents = append(dependents, appName)
				break
			}
		}
	}
	return dependents
}

// FindOrphans 查找孤儿依赖
func (d *DepsIndex) FindOrphans() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 构建所有被依赖的软件集合
	dependedOn := make(map[string]bool)
	for _, appDeps := range d.apps {
		for _, dep := range appDeps.Dependencies {
			dependedOn[dep] = true
		}
	}

	// 找出没有被任何软件依赖的软件
	var orphans []string
	for appName := range d.apps {
		if !dependedOn[appName] {
			orphans = append(orphans, appName)
		}
	}

	return orphans
}

// Remove 移除应用依赖信息
func (d *DepsIndex) Remove(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.apps, name)
	return d.Save()
}

// Rebuild 重建依赖索引（扫描所有 manifest.json）
func (d *DepsIndex) Rebuild(ctx context.Context, rootPath string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 扫描所有已安装的软件
	appsDir := filepath.Join(rootPath, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return fmt.Errorf("failed to read apps directory: %w", err)
	}

	// 清空现有索引
	d.apps = make(map[string]*AppDeps)

	// 遍历所有软件
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appName := entry.Name()
		manifestPath := filepath.Join(appsDir, appName, "manifest.json")

		// 读取 manifest.json
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// 忽略无法读取的 manifest
			continue
		}

		var installedApp struct {
			Dependencies []manifest.Dependency `json:"dependencies"`
		}

		if err := json.Unmarshal(data, &installedApp); err != nil {
			// 忽略无法解析的 manifest
			continue
		}

		// 提取依赖列表
		var deps []string
		if len(installedApp.Dependencies) > 0 {
			deps = extractDependenciesFromFlat(installedApp.Dependencies)
		}

		// 创建应用依赖信息
		d.apps[appName] = &AppDeps{
			Dependencies: deps,
			Dependents: []string{}, // 稍后计算
		}
	}

	// 计算反向依赖
	d.calculateDependents()

	// 保存索引
	return d.Save()
}

// extractDependenciesFromFlat 从扁平的依赖数组中提取依赖名称列表
func extractDependenciesFromFlat(deps []manifest.Dependency) []string {
	var result []string

	for _, dep := range deps {
		result = append(result, dep.Name)
	}

	return result
}

// extractDependencies 从 Dependencies 结构体中提取依赖名称列表
func (d *DepsIndex) extractDependencies(deps *manifest.Dependencies) []string {
	var result []string

	if deps.Runtime != nil {
		for _, dep := range deps.Runtime {
			result = append(result, dep.Name)
		}
	}

	if deps.Tools != nil {
		for _, dep := range deps.Tools {
			result = append(result, dep.Name)
		}
	}

	if deps.Libraries != nil {
		for _, dep := range deps.Libraries {
			result = append(result, dep.Name)
		}
	}

	return result
}

// calculateDependents 计算反向依赖
func (d *DepsIndex) calculateDependents() {
	// 清空所有反向依赖
	for _, appDeps := range d.apps {
		appDeps.Dependents = []string{}
	}

	// 计算反向依赖
	for appName, appDeps := range d.apps {
		for _, dep := range appDeps.Dependencies {
			if depAppDeps, ok := d.apps[dep]; ok {
				depAppDeps.Dependents = append(depAppDeps.Dependents, appName)
			}
		}
	}
}
