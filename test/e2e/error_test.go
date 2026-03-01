//go:build e2e

package e2e

import (
	"testing"
)

// TestErrorInvalidCommand 测试无效命令
func TestErrorInvalidCommand(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 执行无效命令
	_, err := RunCLIFail(t, env, "invalid-command")
	if err == nil {
		t.Fatal("无效命令应该返回错误")
	}
}

// TestErrorInvalidApp 测试无效应用
func TestErrorInvalidApp(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// 尝试安装不存在的应用
	_, err := RunCLIFail(t, env, "install", "non-existent-app")
	if err == nil {
		t.Fatal("安装不存在的应用应该返回错误")
	}
}

// TestErrorInstallWithoutBucket 测试没有 bucket 时安装
func TestErrorInstallWithoutBucket(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 尝试在没有添加 bucket 的情况下安装应用
	_, err := RunCLIFail(t, env, "install", "test-app")
	if err == nil {
		t.Fatal("没有 bucket 时安装应该返回错误")
	}
}

// TestErrorUninstallNotInstalled 测试卸载未安装的应用
func TestErrorUninstallNotInstalled(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 尝试卸载未安装的应用
	_, err := RunCLIFail(t, env, "uninstall", "not-installed-app")
	if err == nil {
		t.Fatal("卸载未安装的应用应该返回错误")
	}
}

// TestErrorUpdateNotInstalled 测试更新未安装的应用
func TestErrorUpdateNotInstalled(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 尝试更新未安装的应用
	_, err := RunCLIFail(t, env, "update", "not-installed-app")
	if err == nil {
		t.Fatal("更新未安装的应用应该返回错误")
	}
}

// TestErrorRemoveNonExistentBucket 测试删除不存在的 bucket
func TestErrorRemoveNonExistentBucket(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 尝试删除不存在的 bucket
	_, err := RunCLIFail(t, env, "bucket", "remove", "non-existent-bucket")
	if err == nil {
		t.Fatal("删除不存在的 bucket 应该返回错误")
	}
}

// TestErrorAddDuplicateBucket 测试添加重复的 bucket
func TestErrorAddDuplicateBucket(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// 尝试再次添加相同的 bucket
	_, err := RunCLIFail(t, env, "bucket", "add", "main", env.TestBucketURL)
	if err == nil {
		t.Fatal("添加重复的 bucket 应该返回错误")
	}
}

// TestErrorInvalidBucketURL 测试无效的 bucket URL
func TestErrorInvalidBucketURL(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 尝试添加无效的 bucket URL
	_, err := RunCLIFail(t, env, "bucket", "add", "main", "invalid-url")
	if err == nil {
		t.Fatal("无效的 bucket URL 应该返回错误")
	}
}

// TestErrorMissingArguments 测试缺少参数
func TestErrorMissingArguments(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// 测试缺少参数的命令
	testCases := [][]string{
		{"install"},
		{"uninstall"},
		{"bucket", "add"},
		{"bucket", "remove"},
	}

	for _, args := range testCases {
		_, err := RunCLIFail(t, env, args...)
		if err == nil {
			t.Fatalf("缺少参数的命令 %v 应该返回错误", args)
		}
	}
}
