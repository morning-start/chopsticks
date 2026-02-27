// Package chopsticksx 提供 Chopsticks 包系统 API。
package chopsticksx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
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
	// 打开用户环境变量注册表键
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表键: %w", err)
	}
	defer key.Close()

	// 读取当前 PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("读取 PATH: %w", err)
	}

	// 检查路径是否已存在
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), path) {
			return nil // 已存在
		}
	}

	// 添加新路径
	newPath := currentPath
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += path

	// 写入注册表
	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("写入 PATH: %w", err)
	}

	// 通知系统环境变量已更改
	return notifyEnvironmentChange()
}

// RemoveFromPath 从环境变量 PATH 移除路径。
func (m *Module) RemoveFromPath(path string) error {
	// 打开用户环境变量注册表键
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表键: %w", err)
	}
	defer key.Close()

	// 读取当前 PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("读取 PATH: %w", err)
	}

	// 移除路径
	paths := strings.Split(currentPath, ";")
	var newPaths []string
	for _, p := range paths {
		if !strings.EqualFold(strings.TrimSpace(p), path) {
			newPaths = append(newPaths, p)
		}
	}

	newPath := strings.Join(newPaths, ";")

	// 写入注册表
	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("写入 PATH: %w", err)
	}

	// 通知系统环境变量已更改
	return notifyEnvironmentChange()
}

// notifyEnvironmentChange 通知系统环境变量已更改
func notifyEnvironmentChange() error {
	// 使用 rundll32 通知环境变量更改
	cmd := exec.Command("rundll32", "user32.dll,UpdatePerUserSystemParameters")
	return cmd.Run()
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
	if err := os.MkdirAll(m.shimsPath, 0755); err != nil {
		return fmt.Errorf("创建 shims 目录: %w", err)
	}

	shimPath := filepath.Join(m.shimsPath, name+".exe")

	// 如果源文件是 .exe，创建符号链接
	if strings.HasSuffix(strings.ToLower(source), ".exe") {
		// 删除已存在的 shim
		if _, err := os.Stat(shimPath); err == nil {
			if err := os.Remove(shimPath); err != nil {
				return fmt.Errorf("删除旧 shim: %w", err)
			}
		}

		// 创建符号链接
		if err := os.Symlink(source, shimPath); err != nil {
			// 如果符号链接失败，尝试复制文件
			if err := copyFile(source, shimPath); err != nil {
				return fmt.Errorf("创建 shim: %w", err)
			}
		}
	} else {
		// 创建批处理文件作为 shim
		batPath := filepath.Join(m.shimsPath, name+".bat")
		content := fmt.Sprintf(`@echo off
"%s" %%*
`, source)
		if err := os.WriteFile(batPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("创建 bat shim: %w", err)
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

// RemoveShim 移除命令链接（shim）。
func (m *Module) RemoveShim(name string) error {
	// 尝试删除 .exe shim
	shimPath := filepath.Join(m.shimsPath, name+".exe")
	if _, err := os.Stat(shimPath); err == nil {
		if err := os.Remove(shimPath); err != nil {
			return fmt.Errorf("删除 exe shim: %w", err)
		}
	}

	// 尝试删除 .bat shim
	batPath := filepath.Join(m.shimsPath, name+".bat")
	if _, err := os.Stat(batPath); err == nil {
		if err := os.Remove(batPath); err != nil {
			return fmt.Errorf("删除 bat shim: %w", err)
		}
	}

	return nil
}

// PersistData 持久化数据目录。
func (m *Module) PersistData(name string, dirs []string) error {
	if err := os.MkdirAll(m.persistPath, 0755); err != nil {
		return fmt.Errorf("创建持久化目录: %w", err)
	}

	appPersistPath := filepath.Join(m.persistPath, name)
	if err := os.MkdirAll(appPersistPath, 0755); err != nil {
		return fmt.Errorf("创建应用持久化目录: %w", err)
	}

	// 为每个需要持久化的目录创建 Junction
	for _, dir := range dirs {
		targetPath := filepath.Join(appPersistPath, filepath.Base(dir))
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("创建目标目录: %w", err)
		}

		// 如果源目录存在，先备份数据
		if _, err := os.Stat(dir); err == nil {
			// 复制现有数据到持久化目录
			if err := copyDirContents(dir, targetPath); err != nil {
				return fmt.Errorf("复制数据: %w", err)
			}
			// 删除原目录
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("删除原目录: %w", err)
			}
		}

		// 创建 Junction（Windows 符号链接）
		if err := createJunction(dir, targetPath); err != nil {
			return fmt.Errorf("创建 Junction: %w", err)
		}
	}

	return nil
}

// copyDirContents 复制目录内容
func copyDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDirContents(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// createJunction 创建 Windows Junction
func createJunction(link, target string) error {
	// 使用 mklink 命令创建 Junction
	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, target)
	return cmd.Run()
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
	// 获取开始菜单路径
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenuPath, opts.Name+".lnk")

	// 使用 PowerShell 创建快捷方式
	psScript := fmt.Sprintf(`
$WshShell = New-Object -comObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("%s")
$Shortcut.TargetPath = "%s"
$Shortcut.Description = "%s"
`, shortcutPath, opts.Source, opts.Description)

	if opts.Icon != "" {
		psScript += fmt.Sprintf(`$Shortcut.IconLocation = "%s"
`, opts.Icon)
	}

	psScript += `$Shortcut.Save()`

	cmd := exec.Command("powershell", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建快捷方式: %w", err)
	}

	return nil
}

// RemoveShortcut 移除快捷方式。
func (m *Module) RemoveShortcut(name string) error {
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenuPath, name+".lnk")

	if _, err := os.Stat(shortcutPath); err == nil {
		if err := os.Remove(shortcutPath); err != nil {
			return fmt.Errorf("删除快捷方式: %w", err)
		}
	}

	return nil
}
