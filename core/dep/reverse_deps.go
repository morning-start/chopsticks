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

// ReverseDepsCalculator 反向依赖计算器接口
type ReverseDepsCalculator interface {
	// Calculate 计算反向依赖
	Calculate(ctx context.Context) error
	// GetDependents 获取反向依赖（谁依赖我）
	GetDependents(appName string) []string
	// GetDependencies 获取依赖列表
	GetDependencies(appName string) []string
	// IsDependent 检查是否是依赖者
	IsDependent(appName, dependency string) bool
	// GetAllDependents 递归获取所有反向依赖
	GetAllDependents(appName string) []string
	// GetDependentsTree 获取反向依赖树
	GetDependentsTree(appName string) *DependentTree
}

// reverseDepsCalculator 反向依赖计算器实现
type reverseDepsCalculator struct {
	mu         sync.RWMutex
	apps       map[string]*AppDeps
	filePath   string
	installDir string
}

// DependentTree 表示反向依赖树
type DependentTree struct {
	Name       string           `json:"name"`
	Dependents []*DependentTree `json:"dependents,omitempty"`
	Depth      int              `json:"depth"`
}

// NewReverseDepsCalculator 创建反向依赖计算器
func NewReverseDepsCalculator(rootPath string) ReverseDepsCalculator {
	return &reverseDepsCalculator{
		apps:       make(map[string]*AppDeps),
		filePath:   filepath.Join(rootPath, "deps-index.json"),
		installDir: filepath.Join(rootPath, "apps"),
	}
}

// Calculate 计算反向依赖
func (c *reverseDepsCalculator) Calculate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 扫描所有已安装的软件
	entries, err := os.ReadDir(c.installDir)
	if err != nil {
		return fmt.Errorf("failed to read apps directory: %w", err)
	}

	// 清空现有索引
	c.apps = make(map[string]*AppDeps)

	// 遍历所有软件
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appName := entry.Name()
		manifestPath := filepath.Join(c.installDir, appName, "manifest.json")

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
			for _, dep := range installedApp.Dependencies {
				deps = append(deps, dep.Name)
			}
		}

		// 创建应用依赖信息
		c.apps[appName] = &AppDeps{
			Dependencies: deps,
			Dependents:   []string{}, // 稍后计算
		}
	}

	// 计算反向依赖
	c.calculateDependents()

	// 保存索引
	return c.save()
}

// GetDependents 获取反向依赖（谁依赖我）
func (c *reverseDepsCalculator) GetDependents(appName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var dependents []string
	for name, appDeps := range c.apps {
		for _, dep := range appDeps.Dependencies {
			if dep == appName {
				dependents = append(dependents, name)
				break
			}
		}
	}
	return dependents
}

// GetDependencies 获取依赖列表
func (c *reverseDepsCalculator) GetDependencies(appName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if appDeps, ok := c.apps[appName]; ok {
		result := make([]string, len(appDeps.Dependencies))
		copy(result, appDeps.Dependencies)
		return result
	}
	return nil
}

// IsDependent 检查是否是依赖者
func (c *reverseDepsCalculator) IsDependent(appName, dependency string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if appDeps, ok := c.apps[appName]; ok {
		for _, dep := range appDeps.Dependencies {
			if dep == dependency {
				return true
			}
		}
	}
	return false
}

// GetAllDependents 递归获取所有反向依赖（包括间接依赖）
func (c *reverseDepsCalculator) GetAllDependents(appName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	allDependents := make(map[string]bool)
	visited := make(map[string]bool)

	var collectDependents func(name string)
	collectDependents = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		for appName, appDeps := range c.apps {
			for _, dep := range appDeps.Dependencies {
				if dep == name {
					if !allDependents[appName] {
						allDependents[appName] = true
						collectDependents(appName)
					}
					break
				}
			}
		}
	}

	collectDependents(appName)

	result := make([]string, 0, len(allDependents))
	for name := range allDependents {
		result = append(result, name)
	}
	return result
}

// GetDependentsTree 获取反向依赖树
func (c *reverseDepsCalculator) GetDependentsTree(appName string) *DependentTree {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.buildDependentsTree(appName, 0, make(map[string]bool))
}

// buildDependentsTree 构建反向依赖树
func (c *reverseDepsCalculator) buildDependentsTree(appName string, depth int, visited map[string]bool) *DependentTree {
	if visited[appName] {
		// 避免循环引用
		return &DependentTree{
			Name:  appName,
			Depth: depth,
		}
	}
	visited[appName] = true

	tree := &DependentTree{
		Name:       appName,
		Dependents: []*DependentTree{},
		Depth:      depth,
	}

	// 查找直接依赖此应用的软件
	for name, appDeps := range c.apps {
		for _, dep := range appDeps.Dependencies {
			if dep == appName {
				childTree := c.buildDependentsTree(name, depth+1, visited)
				tree.Dependents = append(tree.Dependents, childTree)
				break
			}
		}
	}

	return tree
}

// calculateDependents 计算反向依赖
func (c *reverseDepsCalculator) calculateDependents() {
	// 清空所有反向依赖
	for _, appDeps := range c.apps {
		appDeps.Dependents = []string{}
	}

	// 计算反向依赖
	for appName, appDeps := range c.apps {
		for _, dep := range appDeps.Dependencies {
			if depAppDeps, ok := c.apps[dep]; ok {
				depAppDeps.Dependents = append(depAppDeps.Dependents, appName)
			}
		}
	}
}

// save 保存依赖索引
func (c *reverseDepsCalculator) save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.apps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deps index: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save deps index: %w", err)
	}

	return nil
}

// load 加载依赖索引
func (c *reverseDepsCalculator) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，初始化为空
			c.apps = make(map[string]*AppDeps)
			return nil
		}
		return fmt.Errorf("failed to load deps index: %w", err)
	}

	if err := json.Unmarshal(data, &c.apps); err != nil {
		return fmt.Errorf("failed to unmarshal deps index: %w", err)
	}

	return nil
}

// GetAppDeps 获取应用依赖信息
func (c *reverseDepsCalculator) GetAppDeps(appName string) (*AppDeps, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, ok := c.apps[appName]
	return info, ok
}

// ListAllApps 列出所有应用
func (c *reverseDepsCalculator) ListAllApps() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	apps := make([]string, 0, len(c.apps))
	for appName := range c.apps {
		apps = append(apps, appName)
	}
	return apps
}

// HasDependencies 检查应用是否有依赖
func (c *reverseDepsCalculator) HasDependencies(appName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if appDeps, ok := c.apps[appName]; ok {
		return len(appDeps.Dependencies) > 0
	}
	return false
}

// HasDependents 检查应用是否有反向依赖
func (c *reverseDepsCalculator) HasDependents(appName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if appDeps, ok := c.apps[appName]; ok {
		return len(appDeps.Dependents) > 0
	}
	return false
}
