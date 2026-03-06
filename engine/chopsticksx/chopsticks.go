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
	cachePath   string
	configPath  string
}

// NewModule 创建新的 chopsticks 模块。
func NewModule(appsPath, shimsPath, persistPath string) *Module {
	// Get cache directory
	cachePath := filepath.Join(os.Getenv("LOCALAPPDATA"), "chopsticks", "cache")
	// Get config directory
	configPath := filepath.Join(os.Getenv("APPDATA"), "chopsticks")

	return &Module{
		appsPath:    appsPath,
		shimsPath:   shimsPath,
		persistPath: persistPath,
		cachePath:   cachePath,
		configPath:  configPath,
	}
}

// GetCookDir 返回指定菜肴和版本的安装目录。
func (m *Module) GetCookDir(name, version string) string {
	return filepath.Join(m.appsPath, name, version)
}

// GetCurrentVersion 返回应用的当前版本（如果已安装）。
func (m *Module) GetCurrentVersion(name string) (string, error) {
	// Check version in install directory
	appPath := filepath.Join(m.appsPath, name)
	entries, err := os.ReadDir(appPath)
	if err != nil {
		return "", fmt.Errorf("app not installed: %s", name)
	}

	// Return first version directory found
	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no version found for: %s", name)
}

// AddToPath 将路径添加到环境变量 PATH。
// scope: "user" 或 "machine",默认为 "user"
func (m *Module) AddToPath(path string, scope string) error {
	if scope == "" {
		scope = "user"
	}

	// Determine registry key based on scope
	var regKey registry.Key
	if scope == "machine" {
		regKey = registry.LOCAL_MACHINE
	} else {
		regKey = registry.CURRENT_USER
	}

	// Open environment registry key
	key, err := registry.OpenKey(regKey, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	// Read current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("read PATH: %w", err)
	}

	// Check if path already exists
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), path) {
			return nil // Already exists
		}
	}

	// Add new path
	newPath := currentPath
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += path

	// Write to registry
	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("write PATH: %w", err)
	}

	// Notify system environment changed
	return notifyEnvironmentChange()
}

// RemoveFromPath 从环境变量 PATH 移除路径。
// scope: "user" 或 "machine",默认为 "user"
func (m *Module) RemoveFromPath(path string, scope string) error {
	if scope == "" {
		scope = "user"
	}

	// Determine registry key based on scope
	var regKey registry.Key
	if scope == "machine" {
		regKey = registry.LOCAL_MACHINE
	} else {
		regKey = registry.CURRENT_USER
	}

	// Open environment registry key
	key, err := registry.OpenKey(regKey, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	// Read current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("read PATH: %w", err)
	}

	// Remove path
	paths := strings.Split(currentPath, ";")
	var newPaths []string
	for _, p := range paths {
		if !strings.EqualFold(strings.TrimSpace(p), path) {
			newPaths = append(newPaths, p)
		}
	}

	newPath := strings.Join(newPaths, ";")

	// Write to registry
	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("write PATH: %w", err)
	}

	// Notify system environment changed
	return notifyEnvironmentChange()
}

// notifyEnvironmentChange 通知系统环境变量已更改
func notifyEnvironmentChange() error {
	// Use rundll32 to notify environment change
	cmd := exec.Command("rundll32", "user32.dll,UpdatePerUserSystemParameters")
	return cmd.Run()
}

// SetEnv 设置环境变量。
// scope: "user" 或 "machine",默认为 "user"
func (m *Module) SetEnv(key, value, scope string) error {
	if scope == "" {
		scope = "user"
	}

	// Determine registry key based on scope
	var regKey registry.Key
	if scope == "machine" {
		regKey = registry.LOCAL_MACHINE
	} else {
		regKey = registry.CURRENT_USER
	}

	// Open environment registry key
	keyReg, err := registry.OpenKey(regKey, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer keyReg.Close()

	// Set environment variable in registry
	if err := keyReg.SetStringValue(key, value); err != nil {
		return fmt.Errorf("set environment variable: %w", err)
	}

	// Also set in current process
	os.Setenv(key, value)

	// Notify system environment changed
	return notifyEnvironmentChange()
}

// GetEnv 获取环境变量。
func (m *Module) GetEnv(key string) string {
	return os.Getenv(key)
}

// CreateShim 创建命令链接（shim）。
func (m *Module) CreateShim(source, name string) (string, error) {
	if err := os.MkdirAll(m.shimsPath, 0755); err != nil {
		return "", fmt.Errorf("create shims directory: %w", err)
	}

	shimPath := filepath.Join(m.shimsPath, name+".exe")

	// If source is .exe, create symlink
	if strings.HasSuffix(strings.ToLower(source), ".exe") {
		// Remove existing shim
		if _, err := os.Stat(shimPath); err == nil {
			if err := os.Remove(shimPath); err != nil {
				return "", fmt.Errorf("remove old shim: %w", err)
			}
		}

		// Create symlink
		if err := os.Symlink(source, shimPath); err != nil {
			// If symlink fails, try copying file
			if err := copyFile(source, shimPath); err != nil {
				return "", fmt.Errorf("create shim: %w", err)
			}
		}
	} else {
		// Create batch file as shim
		batPath := filepath.Join(m.shimsPath, name+".bat")
		content := fmt.Sprintf(`@echo off
"%s" %%*
`, source)
		if err := os.WriteFile(batPath, []byte(content), 0755); err != nil {
			return "", fmt.Errorf("create bat shim: %w", err)
		}
		shimPath = batPath
	}

	return shimPath, nil
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
	// Try to delete .exe shim
	shimPath := filepath.Join(m.shimsPath, name+".exe")
	if _, err := os.Stat(shimPath); err == nil {
		if err := os.Remove(shimPath); err != nil {
			return fmt.Errorf("remove exe shim: %w", err)
		}
	}

	// Try to delete .bat shim
	batPath := filepath.Join(m.shimsPath, name+".bat")
	if _, err := os.Stat(batPath); err == nil {
		if err := os.Remove(batPath); err != nil {
			return fmt.Errorf("remove bat shim: %w", err)
		}
	}

	return nil
}

// PersistData 持久化数据目录。
func (m *Module) PersistData(name string, dirs []string) error {
	if err := os.MkdirAll(m.persistPath, 0755); err != nil {
		return fmt.Errorf("create persist directory: %w", err)
	}

	appPersistPath := filepath.Join(m.persistPath, name)
	if err := os.MkdirAll(appPersistPath, 0755); err != nil {
		return fmt.Errorf("create app persist directory: %w", err)
	}

	// Create Junction for each directory that needs persistence
	for _, dir := range dirs {
		targetPath := filepath.Join(appPersistPath, filepath.Base(dir))
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("create target directory: %w", err)
		}

		// If source directory exists, backup data first
		if _, err := os.Stat(dir); err == nil {
			// Copy existing data to persist directory
			if err := copyDirContents(dir, targetPath); err != nil {
				return fmt.Errorf("copy data: %w", err)
			}
			// Delete original directory
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("remove original directory: %w", err)
			}
		}

		// Create Junction (Windows symlink)
		if err := createJunction(dir, targetPath); err != nil {
			return fmt.Errorf("create Junction: %w", err)
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
	// Use mklink command to create Junction
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
func (m *Module) CreateShortcut(opts ShortcutOptions) (string, error) {
	// Get start menu path
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenuPath, opts.Name+".lnk")

	// Use PowerShell to create shortcut
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
		return "", fmt.Errorf("create shortcut: %w", err)
	}

	return shortcutPath, nil
}

// RemoveShortcut 移除快捷方式。
func (m *Module) RemoveShortcut(name string) error {
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenuPath, name+".lnk")

	if _, err := os.Stat(shortcutPath); err == nil {
		if err := os.Remove(shortcutPath); err != nil {
			return fmt.Errorf("remove shortcut: %w", err)
		}
	}

	return nil
}

// GetCacheDir 返回缓存目录。
func (m *Module) GetCacheDir() string {
	return m.cachePath
}

// GetConfigDir 返回配置目录。
func (m *Module) GetConfigDir() string {
	return m.configPath
}

// DeleteEnv 删除环境变量。
// scope: "user" 或 "machine",默认为 "user"
func (m *Module) DeleteEnv(key, scope string) error {
	if scope == "" {
		scope = "user"
	}

	// Determine registry key based on scope
	var regKey registry.Key
	if scope == "machine" {
		regKey = registry.LOCAL_MACHINE
	} else {
		regKey = registry.CURRENT_USER
	}

	// Open environment registry key
	keyReg, err := registry.OpenKey(regKey, `Environment`, registry.READ|registry.WRITE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer keyReg.Close()

	// Delete environment variable
	if err := keyReg.DeleteValue(key); err != nil {
		if err == registry.ErrNotExist {
			return nil // Ignore if not exists
		}
		return fmt.Errorf("delete environment variable: %w", err)
	}

	// Also unset in current process
	os.Unsetenv(key)

	// Notify system environment changed
	return notifyEnvironmentChange()
}

// GetPath 获取 PATH 环境变量。
func (m *Module) GetPath() string {
	return os.Getenv("PATH")
}

// GetShimDir 返回 shims 目录路径
func (m *Module) GetShimDir() string {
	return m.shimsPath
}

// GetPersistDir 返回 persist 目录路径
func (m *Module) GetPersistDir() string {
	return m.persistPath
}
