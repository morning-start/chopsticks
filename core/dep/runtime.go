// Package dep 提供依赖管理功能
package dep

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chopsticks/core/manifest"
	"chopsticks/pkg/errors"
)

// RuntimeIndex 表示运行时库索引
type RuntimeIndex struct {
	mu       sync.RWMutex
	runtimes map[string]*manifest.RuntimeInfo
	filePath string
}

// NewRuntimeIndex 创建运行时库索引
func NewRuntimeIndex(rootPath string) *RuntimeIndex {
	return &RuntimeIndex{
		runtimes: make(map[string]*manifest.RuntimeInfo),
		filePath: filepath.Join(rootPath, "runtime-index.json"),
	}
}

// Load 加载运行时库索引
func (r *RuntimeIndex) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，初始化为空
			r.runtimes = make(map[string]*manifest.RuntimeInfo)
			return nil
		}
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	if err := json.Unmarshal(data, &r.runtimes); err != nil {
		return fmt.Errorf("failed to unmarshal runtime index: %w", err)
	}

	return nil
}

// Save 保存运行时库索引
func (r *RuntimeIndex) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.MarshalIndent(r.runtimes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal runtime index: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save runtime index: %w", err)
	}

	return nil
}

// Get 获取运行时库信息
func (r *RuntimeIndex) Get(name string) (*manifest.RuntimeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.runtimes[name]
	return info, ok
}

// Add 添加或更新运行时库
func (r *RuntimeIndex) Add(name string, version string, size int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.runtimes[name]
	if exists {
		// 已存在，增加引用计数
		info.RefCount++
		info.RequiredBy = append(info.RequiredBy, name)
	} else {
		// 不存在，创建新条目
		r.runtimes[name] = &manifest.RuntimeInfo{
			Version:     version,
			InstalledAt: time.Now(),
			RequiredBy:  []string{name},
			RefCount:    1,
			Size:        size,
		}
	}

	return r.Save()
}

// Remove 减少运行时库引用计数
func (r *RuntimeIndex) Remove(name, appName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, ok := r.runtimes[name]
	if !ok {
		return errors.Newf(errors.KindNotFound, "runtime %s not found", name)
	}

	// 减少引用计数
	info.RefCount--

	// 从依赖者列表中移除
	for i, dep := range info.RequiredBy {
		if dep == appName {
			info.RequiredBy = append(info.RequiredBy[:i], info.RequiredBy[i+1:]...)
			break
		}
	}

	// 检查引用计数是否为 0
	if info.RefCount == 0 {
		return fmt.Errorf("runtime %s is no longer required (ref_count=0)", name)
	}

	return r.Save()
}

// List 列出所有运行时库
func (r *RuntimeIndex) List() map[string]*manifest.RuntimeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回副本
	result := make(map[string]*manifest.RuntimeInfo, len(r.runtimes))
	for k, v := range r.runtimes {
		result[k] = v
	}
	return result
}

// FindOrphans 查找孤儿运行时库（引用计数为 0）
func (r *RuntimeIndex) FindOrphans() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orphans []string
	for name, info := range r.runtimes {
		if info.RefCount == 0 {
			orphans = append(orphans, name)
		}
	}
	return orphans
}

// GetRefCount 获取引用计数
func (r *RuntimeIndex) GetRefCount(name string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.runtimes[name]; ok {
		return info.RefCount
	}
	return 0
}

// GetRequiredBy 获取依赖此运行时库的软件列表
func (r *RuntimeIndex) GetRequiredBy(name string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.runtimes[name]; ok {
		result := make([]string, len(info.RequiredBy))
		copy(result, info.RequiredBy)
		return result
	}
	return nil
}
