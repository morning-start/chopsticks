// Package pathx 提供路径处理功能。
package pathx

import (
	"os"
	"path/filepath"
)

// Join 连接多个路径片段。
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Abs 返回绝对路径。
func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Base 返回路径的最后一个元素。
func Base(path string) string {
	return filepath.Base(path)
}

// Dir 返回路径的目录部分。
func Dir(path string) string {
	return filepath.Dir(path)
}

// Ext 返回文件扩展名。
func Ext(path string) string {
	return filepath.Ext(path)
}

// Clean 返回最短路径名。
func Clean(path string) string {
	return filepath.Clean(path)
}

// IsAbs 报告路径是否为绝对路径。
func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Exists 检查路径是否存在。
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 检查路径是否为目录。
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Module 为脚本引擎提供 path 注册。
type Module struct{}

// NewModule 创建新的 path 模块。
func NewModule() *Module {
	return &Module{}
}
