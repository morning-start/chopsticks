// Package testutil 提供集成测试的辅助工具。
package testutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"chopsticks/infra/git"
)

// 编译时检查 MockGit 是否实现了 git.Git 接口
var _ git.Git = (*MockGit)(nil)

// MockGit 模拟 Git 操作。
type MockGit struct {
	ClonedRepos map[string]string // url -> dest
	PulledRepos []string          // list of pulled repos
	FetchedRepos []string         // list of fetched repos
	
	// FixturePath 用于指定测试数据目录
	FixturePath string
}

// NewMockGit 创建新的 MockGit。
func NewMockGit() *MockGit {
	return &MockGit{
		ClonedRepos:  make(map[string]string),
		PulledRepos:  []string{},
		FetchedRepos: []string{},
	}
}

// Clone 模拟克隆操作。
func (m *MockGit) Clone(ctx context.Context, url, dest string) error {
	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	// 如果指定了 fixture 路径，复制测试数据
	if m.FixturePath != "" {
		if err := copyDir(m.FixturePath, dest); err != nil {
			return fmt.Errorf("复制 fixture 数据: %w", err)
		}
	}

	m.ClonedRepos[url] = dest
	return nil
}

// Pull 模拟拉取操作。
func (m *MockGit) Pull(ctx context.Context, dir string) error {
	m.PulledRepos = append(m.PulledRepos, dir)
	return nil
}

// Fetch 模拟获取操作。
func (m *MockGit) Fetch(ctx context.Context, dir string) error {
	m.FetchedRepos = append(m.FetchedRepos, dir)
	return nil
}

// GetLatestTag 获取最新标签。
func (m *MockGit) GetLatestTag(ctx context.Context, dir string) (string, error) {
	return "v1.0.0", nil
}

// GetCommitHash 获取提交哈希。
func (m *MockGit) GetCommitHash(ctx context.Context, dir string) (string, error) {
	return "abc123def456", nil
}

// GetCurrentBranch 获取当前分支。
func (m *MockGit) GetCurrentBranch(ctx context.Context, dir string) (string, error) {
	return "main", nil
}

// copyDir 复制目录内容。
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile 复制文件。
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
