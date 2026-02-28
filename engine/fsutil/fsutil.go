// Package fsutil 提供文件系统工具函数。
package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Read 读取 path 处的整个文件并返回其内容为字符串。
func Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// Write 将 content 写入 path 处的文件，根据需要创建目录。
func Write(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// Append 将 content 追加到 path 处的文件，如果文件不存在则创建它。
func Append(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// Mkdir 创建目录及所有必要的父目录。
func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

// Rmdir 删除目录及其所有内容。
func Rmdir(path string) error {
	return os.RemoveAll(path)
}

// Remove 删除文件。
func Remove(path string) error {
	return os.Remove(path)
}

// Exists 检查路径是否存在。
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsDir 检查路径是否为目录。
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// IsFile 检查路径是否为普通文件。
func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// List 返回 path 处目录中的所有条目列表。
func List(path string) ([]string, error) {
	var entries []string
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(path, p)
		if rel != "." {
			entries = append(entries, rel)
		}
		return nil
	})
	return entries, err
}

// Copy 将文件从 src 复制到 dst。
func Copy(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("读取源文件: %w", err)
	}
	if err := Write(dst, string(data)); err != nil {
		return fmt.Errorf("写入目标文件: %w", err)
	}
	return nil
}

// Rename 重命名（移动）文件从 oldPath 到 newPath。
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Name    string `json:"name"`     // 文件名
	Size    int64  `json:"size"`     // 文件大小（字节）
	IsDir   bool   `json:"isDir"`    // 是否为目录
	IsFile  bool   `json:"isFile"`   // 是否为文件
	ModTime int64  `json:"modTime"`  // 修改时间（Unix时间戳）
	Mode    uint32 `json:"mode"`     // 权限模式
}

// Stat 返回文件的详细信息
func Stat(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		IsFile:  !info.IsDir(),
		ModTime: info.ModTime().Unix(),
		Mode:    uint32(info.Mode()),
	}, nil
}
