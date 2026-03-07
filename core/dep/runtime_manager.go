// Package dep 提供依赖管理功能
package dep

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chopsticks/core/manifest"
	"chopsticks/pkg/errors"
)

// RuntimeManager 运行时库管理器接口
type RuntimeManager interface {
	// Install 安装运行时库
	Install(ctx context.Context, dep, version, appName string, size int64) error
	// Uninstall 卸载运行时库
	Uninstall(ctx context.Context, dep, appName string) error
	// GetInfo 获取运行时库信息
	GetInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error)
	// List 列出所有运行时库
	List(ctx context.Context) map[string]*manifest.RuntimeInfo
	// Cleanup 清理无用运行时库
	Cleanup(ctx context.Context) error
	// GetRefCount 获取引用计数
	GetRefCount(ctx context.Context, dep string) int
	// GetRequiredBy 获取依赖此运行时库的软件列表
	GetRequiredBy(ctx context.Context, dep string) []string
}

// runtimeManager 运行时库管理器实现
type runtimeManager struct {
	mu         sync.RWMutex
	index      *RuntimeIndex
	rootPath   string
	runtimeDir string
}

// NewRuntimeManager 创建运行时库管理器
func NewRuntimeManager(rootPath string) (RuntimeManager, error) {
	runtimeDir := filepath.Join(rootPath, "runtimes")

	// 创建运行时目录
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	return &runtimeManager{
		index:      NewRuntimeIndex(rootPath),
		rootPath:   rootPath,
		runtimeDir: runtimeDir,
	}, nil
}

// Install 安装运行时库
func (m *runtimeManager) Install(ctx context.Context, dep, version, appName string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 检查是否已存在
	info, exists := m.index.Get(dep)
	if exists {
		// 已存在，检查版本是否一致
		if info.Version != version {
			// 版本不一致，如果引用计数>0 则警告
			if info.RefCount > 0 {
				fmt.Printf("warning: runtime %s version conflict: existing=%s, required=%s\n",
					dep, info.Version, version)
			}
		}

		// 增加引用计数
		if err := m.index.Add(dep, version, size); err != nil {
			return fmt.Errorf("failed to update runtime: %w", err)
		}
	} else {
		// 不存在，创建新条目
		if err := m.index.Add(dep, version, size); err != nil {
			return fmt.Errorf("failed to add runtime: %w", err)
		}

		// 创建运行时目录
		runtimePath := filepath.Join(m.runtimeDir, dep, version)
		if err := os.MkdirAll(runtimePath, 0755); err != nil {
			return fmt.Errorf("failed to create runtime directory: %w", err)
		}
	}

	// 保存索引
	if err := m.index.Save(); err != nil {
		return fmt.Errorf("failed to save runtime index: %w", err)
	}

	return nil
}

// Uninstall 卸载运行时库
func (m *runtimeManager) Uninstall(ctx context.Context, dep, appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 检查是否存在
	info, exists := m.index.Get(dep)
	if !exists {
		return errors.Newf(errors.KindNotFound, "runtime %s not found", dep)
	}

	// 减少引用计数
	info.RefCount--

	// 从依赖者列表中移除
	for i, name := range info.RequiredBy {
		if name == appName {
			info.RequiredBy = append(info.RequiredBy[:i], info.RequiredBy[i+1:]...)
			break
		}
	}

	// 检查引用计数是否为 0
	if info.RefCount <= 0 {
		// 引用计数为 0，可以删除
		info.RefCount = 0

		// 删除运行时目录
		runtimePath := filepath.Join(m.runtimeDir, dep)
		if err := os.RemoveAll(runtimePath); err != nil {
			fmt.Printf("warning: failed to remove runtime directory %s: %v\n", runtimePath, err)
		}

		// 从索引中删除
		delete(m.index.runtimes, dep)
	}

	// 保存索引
	if err := m.index.Save(); err != nil {
		return fmt.Errorf("failed to save runtime index: %w", err)
	}

	return nil
}

// GetInfo 获取运行时库信息
func (m *runtimeManager) GetInfo(ctx context.Context, dep string) (*manifest.RuntimeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return nil, fmt.Errorf("failed to load runtime index: %w", err)
	}

	info, ok := m.index.Get(dep)
	if !ok {
		return nil, errors.Newf(errors.KindNotFound, "runtime %s not found", dep)
	}

	// 返回副本
	return &manifest.RuntimeInfo{
		Version:     info.Version,
		InstalledAt: info.InstalledAt,
		RequiredBy:  append([]string{}, info.RequiredBy...),
		RefCount:    info.RefCount,
		Size:        info.Size,
	}, nil
}

// List 列出所有运行时库
func (m *runtimeManager) List(ctx context.Context) map[string]*manifest.RuntimeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return nil
	}

	// 返回副本
	source := m.index.List()
	result := make(map[string]*manifest.RuntimeInfo, len(source))
	for k, v := range source {
		result[k] = &manifest.RuntimeInfo{
			Version:     v.Version,
			InstalledAt: v.InstalledAt,
			RequiredBy:  append([]string{}, v.RequiredBy...),
			RefCount:    v.RefCount,
			Size:        v.Size,
		}
	}
	return result
}

// Cleanup 清理无用运行时库
func (m *runtimeManager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	// 查找孤儿运行时库
	orphans := m.index.FindOrphans()

	if len(orphans) == 0 {
		fmt.Println("没有需要清理的运行时库")
		return nil
	}

	fmt.Printf("找到 %d 个孤儿运行时库：\n", len(orphans))
	for _, orphan := range orphans {
		fmt.Printf("  - %s\n", orphan)
	}
	fmt.Println()

	// 清理孤儿运行时库
	for _, orphan := range orphans {
		// 删除运行时目录
		runtimePath := filepath.Join(m.runtimeDir, orphan)
		if err := os.RemoveAll(runtimePath); err != nil {
			fmt.Printf("  ✗ 清理 %s 失败：%v\n", orphan, err)
			continue
		}

		// 从索引中删除
		delete(m.index.runtimes, orphan)

		fmt.Printf("  ✓ 已清理：%s\n", orphan)
	}

	// 保存索引
	if err := m.index.Save(); err != nil {
		return fmt.Errorf("failed to save runtime index: %w", err)
	}

	fmt.Println("\n运行时库清理完成")
	return nil
}

// GetRefCount 获取引用计数
func (m *runtimeManager) GetRefCount(ctx context.Context, dep string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return 0
	}

	return m.index.GetRefCount(dep)
}

// GetRequiredBy 获取依赖此运行时库的软件列表
func (m *runtimeManager) GetRequiredBy(ctx context.Context, dep string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return nil
	}

	return m.index.GetRequiredBy(dep)
}

// InstallWithPath 安装运行时库并返回路径
func (m *runtimeManager) InstallWithPath(ctx context.Context, dep, version, appName string, size int64) (string, error) {
	if err := m.Install(ctx, dep, version, appName, size); err != nil {
		return "", err
	}

	runtimePath := filepath.Join(m.runtimeDir, dep, version)
	return runtimePath, nil
}

// GetRuntimePath 获取运行时库路径
func (m *runtimeManager) GetRuntimePath(dep, version string) string {
	return filepath.Join(m.runtimeDir, dep, version)
}

// IsInstalled 检查运行时库是否已安装
func (m *runtimeManager) IsInstalled(ctx context.Context, dep string) bool {
	info, err := m.GetInfo(ctx, dep)
	return err == nil && info != nil && info.RefCount > 0
}

// UpdateSize 更新运行时库大小
func (m *runtimeManager) UpdateSize(ctx context.Context, dep string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载运行时索引
	if err := m.index.Load(); err != nil {
		return fmt.Errorf("failed to load runtime index: %w", err)
	}

	info, ok := m.index.Get(dep)
	if !ok {
		return errors.Newf(errors.KindNotFound, "runtime %s not found", dep)
	}

	info.Size = size

	// 保存索引
	if err := m.index.Save(); err != nil {
		return fmt.Errorf("failed to save runtime index: %w", err)
	}

	return nil
}
