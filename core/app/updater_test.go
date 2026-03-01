package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	// 创建源目录结构
	subDir := filepath.Join(srcDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// 创建源文件
	srcFile := filepath.Join(srcDir, "file1.txt")
	err = os.WriteFile(srcFile, []byte("content1"), 0644)
	require.NoError(t, err)

	subFile := filepath.Join(subDir, "file2.txt")
	err = os.WriteFile(subFile, []byte("content2"), 0644)
	require.NoError(t, err)

	// 测试复制目录
	err = copyDir(srcDir, dstDir)
	require.NoError(t, err)

	// 验证文件被复制
	dstFile := filepath.Join(dstDir, "file1.txt")
	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content))

	dstSubFile := filepath.Join(dstDir, "subdir", "file2.txt")
	content, err = os.ReadFile(dstSubFile)
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content))
}

func TestCopyDir_SourceNotExist(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "non-existent")
	dstDir := filepath.Join(t.TempDir(), "dst")

	err := copyDir(srcDir, dstDir)
	assert.Error(t, err)
}

func TestCopyDir_SourceIsFile(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "source.txt")
	err := os.WriteFile(srcFile, []byte("content"), 0644)
	require.NoError(t, err)

	dstDir := filepath.Join(t.TempDir(), "dst")

	err = copyDir(srcFile, dstDir)
	assert.Error(t, err)
}

func TestCopyDir_EmptySource(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	err := copyDir(srcDir, dstDir)
	require.NoError(t, err)

	// 验证目标目录存在
	info, err := os.Stat(dstDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestConstants(t *testing.T) {
	// 验证常量存在且类型正确
	_ = DefaultDirPerm
	_ = DefaultFilePerm
	assert.NotZero(t, DefaultDirPerm)
	assert.NotZero(t, DefaultFilePerm)
}
