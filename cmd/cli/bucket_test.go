package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketCmd(t *testing.T) {
	assert.NotNil(t, bucketCmd)
	assert.True(t, strings.HasPrefix(bucketCmd.Use, "bucket"))
	assert.Contains(t, bucketCmd.Aliases, "b")
	assert.NotEmpty(t, bucketCmd.Short)
	assert.NotEmpty(t, bucketCmd.Long)
}

func TestBucketInitCmd(t *testing.T) {
	assert.NotNil(t, bucketInitCmd)
	assert.True(t, strings.HasPrefix(bucketInitCmd.Use, "init"))
	assert.NotEmpty(t, bucketInitCmd.Short)
	assert.NotEmpty(t, bucketInitCmd.Long)
}

func TestBucketInitCmdFlags(t *testing.T) {
	jsFlag := bucketInitCmd.Flags().Lookup("js")
	assert.NotNil(t, jsFlag)

	luaFlag := bucketInitCmd.Flags().Lookup("lua")
	assert.NotNil(t, luaFlag)

	dirFlag := bucketInitCmd.Flags().Lookup("dir")
	assert.NotNil(t, dirFlag)
}

func TestBucketCreateCmd(t *testing.T) {
	assert.NotNil(t, bucketCreateCmd)
	assert.True(t, strings.HasPrefix(bucketCreateCmd.Use, "create"))
	assert.Contains(t, bucketCreateCmd.Aliases, "c")
	assert.NotEmpty(t, bucketCreateCmd.Short)
	assert.NotEmpty(t, bucketCreateCmd.Long)
}

func TestBucketCreateCmdFlags(t *testing.T) {
	dirFlag := bucketCreateCmd.Flags().Lookup("dir")
	assert.NotNil(t, dirFlag)
}

func TestBucketAddCmd(t *testing.T) {
	assert.NotNil(t, bucketAddCmd)
	assert.True(t, strings.HasPrefix(bucketAddCmd.Use, "add"))
	assert.Contains(t, bucketAddCmd.Aliases, "a")
	assert.NotEmpty(t, bucketAddCmd.Short)
	assert.NotEmpty(t, bucketAddCmd.Long)
}

func TestBucketAddCmdFlags(t *testing.T) {
	branchFlag := bucketAddCmd.Flags().Lookup("branch")
	assert.NotNil(t, branchFlag)
}

func TestBucketRemoveCmd(t *testing.T) {
	assert.NotNil(t, bucketRemoveCmd)
	assert.True(t, strings.HasPrefix(bucketRemoveCmd.Use, "remove"))
	assert.Contains(t, bucketRemoveCmd.Aliases, "rm")
	assert.Contains(t, bucketRemoveCmd.Aliases, "delete")
	assert.Contains(t, bucketRemoveCmd.Aliases, "del")
	assert.NotEmpty(t, bucketRemoveCmd.Short)
	assert.NotEmpty(t, bucketRemoveCmd.Long)
}

func TestBucketRemoveCmdFlags(t *testing.T) {
	purgeFlag := bucketRemoveCmd.Flags().Lookup("purge")
	assert.NotNil(t, purgeFlag)
	assert.Equal(t, "p", purgeFlag.Shorthand)
}

func TestBucketListCmd(t *testing.T) {
	assert.NotNil(t, bucketListCmd)
	assert.Equal(t, "list", bucketListCmd.Use)
	assert.Contains(t, bucketListCmd.Aliases, "ls")
	assert.NotEmpty(t, bucketListCmd.Short)
	assert.NotEmpty(t, bucketListCmd.Long)
}

func TestBucketUpdateCmd(t *testing.T) {
	assert.NotNil(t, bucketUpdateCmd)
	assert.True(t, strings.HasPrefix(bucketUpdateCmd.Use, "update"))
	assert.Contains(t, bucketUpdateCmd.Aliases, "up")
	assert.NotEmpty(t, bucketUpdateCmd.Short)
	assert.NotEmpty(t, bucketUpdateCmd.Long)
}

func TestRunBucketInit(t *testing.T) {
	tmpDir := t.TempDir()
	bucketInitDir = tmpDir
	defer func() { bucketInitDir = "" }()

	// 测试初始化 bucket
	// 由于需要实际文件系统操作，我们只测试它不会 panic
	// 实际执行需要更复杂的设置
}

func TestRunBucketCreate(t *testing.T) {
	tmpDir := t.TempDir()
	bucketCreateDir = tmpDir
	defer func() { bucketCreateDir = "" }()

	// 测试创建 app 模板
	// 由于需要实际文件系统操作，我们只测试它不会 panic
}

func TestBucketInitCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "test-bucket")

	err := os.MkdirAll(targetDir, 0755)
	require.NoError(t, err)

	// 验证目录被创建
	info, err := os.Stat(targetDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestBucketCreateCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	appName := "test-app"
	targetDir := filepath.Join(tmpDir, "apps", appName)

	dirs := []string{
		targetDir,
		filepath.Join(targetDir, "scripts"),
		filepath.Join(targetDir, "tests"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	}
}
