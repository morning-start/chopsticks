// Package symlink 提供符号链接功能。
package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Create 创建从 oldName 到 newName 的符号链接。
func Create(oldName, newName string) error {
	if runtime.GOOS == "windows" {
		oldName = filepath.Clean(oldName)
		newName = filepath.Clean(newName)
	}

	if _, err := os.Lstat(newName); err == nil {
		if err := os.Remove(newName); err != nil {
			return fmt.Errorf("删除现有文件: %w", err)
		}
	}

	if err := os.Symlink(oldName, newName); err != nil {
		return fmt.Errorf("创建符号链接: %w", err)
	}
	return nil
}

// Is 检查 path 是否为符号链接。
func Is(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("lstat 路径: %w", err)
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

// Read 读取符号链接的目标。
func Read(path string) (string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("读取符号链接: %w", err)
	}
	return target, nil
}

// Remove 删除符号链接。
func Remove(path string) error {
	isLink, err := Is(path)
	if err != nil {
		return err
	}
	if !isLink {
		return fmt.Errorf("路径不是符号链接: %s", path)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("删除符号链接: %w", err)
	}
	return nil
}
