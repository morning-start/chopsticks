//go:build e2e

package e2e

import (
	"testing"
)

func TestCLIInstall(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	output := RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)
	AssertOutputContains(t, output, "added") // 或其他成功标识

	// 安装应用
	output = RunCLISuccess(t, env, "install", "test-app")
	AssertOutputContains(t, output, "test-app")

	// 验证应用已安装
	output = RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")
}

func TestCLIUninstall(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)

	// 安装应用
	RunCLISuccess(t, env, "install", "test-app")

	// 验证应用已安装
	output := RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")

	// 卸载应用
	output = RunCLISuccess(t, env, "uninstall", "test-app")
	AssertOutputContains(t, output, "test-app")

	// 验证应用已卸载
	output = RunCLISuccess(t, env, "list")
	AssertOutputNotContains(t, output, "test-app")
}

func TestCLIUpdate(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)

	// 安装应用
	RunCLISuccess(t, env, "install", "test-app")

	// 更新应用
	output := RunCLISuccess(t, env, "update", "test-app")
	AssertOutputContains(t, output, "test-app")
}

func TestCLISearch(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)

	// 搜索应用
	output := RunCLISuccess(t, env, "search", "test-app")
	AssertOutputContains(t, output, "test-app")
}

func TestCLIBucketAdd(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	output := RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)
	AssertOutputContains(t, output, "added")

	// 验证 bucket 已添加
	output = RunCLISuccess(t, env, "bucket", "list")
	AssertOutputContains(t, output, "test-bucket")
}

func TestCLIBucketRemove(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)

	// 验证 bucket 已添加
	output := RunCLISuccess(t, env, "bucket", "list")
	AssertOutputContains(t, output, "test-bucket")

	// 删除 bucket
	output = RunCLISuccess(t, env, "bucket", "remove", "test-bucket")
	AssertOutputContains(t, output, "removed")

	// 验证 bucket 已删除
	output = RunCLISuccess(t, env, "bucket", "list")
	AssertOutputNotContains(t, output, "test-bucket")
}

func TestCLIBucketList(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 初始状态应该没有 bucket
	output := RunCLISuccess(t, env, "bucket", "list")
	// 输出应该包含表头或空列表提示

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)

	// 验证 bucket 在列表中
	output = RunCLISuccess(t, env, "bucket", "list")
	AssertOutputContains(t, output, "test-bucket")
}

func TestCLIList(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 初始状态应该没有应用
	output := RunCLISuccess(t, env, "list")
	// 输出应该包含空列表提示

	// 添加 bucket 并安装应用
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)
	RunCLISuccess(t, env, "install", "test-app")

	// 验证应用在列表中
	output = RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")
}

func TestCLIStatus(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket 并安装应用
	RunCLISuccess(t, env, "bucket", "add", "test-bucket", env.TestBucketURL)
	RunCLISuccess(t, env, "install", "test-app")

	// 查看状态
	output := RunCLISuccess(t, env, "status")
	AssertOutputContains(t, output, "test-app")
}

func TestCLIHelp(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 测试帮助命令
	output := RunCLISuccess(t, env, "--help")
	AssertOutputContains(t, output, "Usage")
	AssertOutputContains(t, output, "Commands")

	// 测试子命令帮助
	output = RunCLISuccess(t, env, "install", "--help")
	AssertOutputContains(t, output, "install")

	output = RunCLISuccess(t, env, "bucket", "--help")
	AssertOutputContains(t, output, "bucket")
}

func TestCLIVersion(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 测试版本命令
	output := RunCLISuccess(t, env, "--version")
	// 输出应该包含版本信息
	AssertOutputContains(t, output, "version")
}
