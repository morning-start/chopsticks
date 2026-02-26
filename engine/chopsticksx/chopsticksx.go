// Package chopsticksx 提供 Chopsticks 包系统 API。
package chopsticksx

import (
	"fmt"
	"os"
	"path/filepath"
)

// Module 为脚本引擎提供 chopsticks 注册。
type Module struct {
	appsPath    string
	shimsPath   string
	persistPath string
}

// NewModule 创建新的 chopsticks 模块。
func NewModule(appsPath, shimsPath, persistPath string) *Module {
	return &Module{
		appsPath:    appsPath,
		shimsPath:   shimsPath,
		persistPath: persistPath,
	}
}

// GetCookDir 返回指定菜肴和版本的安装目录。
func (m *Module) GetCookDir(name, version string) string {
	return filepath.Join(m.appsPath, name, version)
}

// GetCurrentVersion 返回菜肴的当前版本（如果已安装）。
func (m *Module) GetCurrentVersion(name string) (string, error) {
	// 检查安装目录中的版本
	dishPath := filepath.Join(m.appsPath, name)
	entries, err := os.ReadDir(dishPath)
	if err != nil {
		return "", fmt.Errorf("dish not installed: %s", name)
	}

	// 返回第一个找到的版本目录
	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no version found for: %s", name)
}

// AddToPath 将路径添加到环境变量 PATH。
func (m *Module) AddToPath(path string) error {
	// TODO: 实现添加到 PATH
	return nil
}

// RemoveFromPath 从环境变量 PATH 移除路径。
func (m *Module) RemoveFromPath(path string) error {
	// TODO: 实现从 PATH 移除
	return nil
}

// SetEnv 设置环境变量。
func (m *Module) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// GetEnv 获取环境变量。
func (m *Module) GetEnv(key string) string {
	return os.Getenv(key)
}

// CreateShim 创建命令链接（shim）。
func (m *Module) CreateShim(source, name string) error {
	// TODO: 实现 shim 创建
	return nil
}

// RemoveShim 移除命令链接（shim）。
func (m *Module) RemoveShim(name string) error {
	// TODO: 实现 shim 移除
	return nil
}

// PersistData 持久化数据目录。
func (m *Module) PersistData(name string, dirs []string) error {
	// TODO: 实现数据持久化
	return nil
}

// ShortcutOptions 快捷方式选项。
type ShortcutOptions struct {
	Source      string
	Name        string
	Description string
	Icon        string
}

// CreateShortcut 创建快捷方式。
func (m *Module) CreateShortcut(opts ShortcutOptions) error {
	// TODO: 实现快捷方式创建
	return nil
}

// RemoveShortcut 移除快捷方式。
func (m *Module) RemoveShortcut(name string) error {
	// TODO: 实现快捷方式移除
	return nil
}
