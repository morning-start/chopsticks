//go:build e2e

package e2e

import (
	"testing"
)

// TestScenarioFreshInstall 测试全新安装场景
func TestScenarioFreshInstall(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	output := RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)
	AssertOutputContains(t, output, "added")

	// Step 2: 搜索应用
	output = RunCLISuccess(t, env, "search", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 3: 安装应用
	output = RunCLISuccess(t, env, "install", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 4: 验证安装
	output = RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")

	// Step 5: 查看状态
	output = RunCLISuccess(t, env, "status", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 6: 卸载应用
	output = RunCLISuccess(t, env, "uninstall", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 7: 验证卸载
	output = RunCLISuccess(t, env, "list")
	AssertOutputNotContains(t, output, "test-app")
}

// TestScenarioDeveloperWorkflow 测试开发者工作流场景
func TestScenarioDeveloperWorkflow(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// Step 2: 安装开发工具
	tools := []string{"test-app", "git"}
	for _, tool := range tools {
		output := RunCLISuccess(t, env, "install", tool)
		AssertOutputContains(t, output, tool)
	}

	// Step 3: 验证所有工具已安装
	output := RunCLISuccess(t, env, "list")
	for _, tool := range tools {
		AssertOutputContains(t, output, tool)
	}

	// Step 4: 更新所有工具
	output = RunCLISuccess(t, env, "update")
	// 应该显示更新状态

	// Step 5: 检查状态
	output = RunCLISuccess(t, env, "status")
	for _, tool := range tools {
		AssertOutputContains(t, output, tool)
	}

	// Step 6: 清理 - 卸载所有工具
	for _, tool := range tools {
		RunCLISuccess(t, env, "uninstall", tool)
	}

	// Step 7: 验证清理完成
	output = RunCLISuccess(t, env, "list")
	for _, tool := range tools {
		AssertOutputNotContains(t, output, tool)
	}
}

// TestScenarioUpgradeWorkflow 测试升级工作流场景
func TestScenarioUpgradeWorkflow(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// Step 2: 安装旧版本应用
	RunCLISuccess(t, env, "install", "test-app")

	// Step 3: 验证安装
	output := RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")

	// Step 4: 检查更新
	output = RunCLISuccess(t, env, "status")
	AssertOutputContains(t, output, "test-app")

	// Step 5: 执行更新
	output = RunCLISuccess(t, env, "update", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 6: 验证更新后状态
	output = RunCLISuccess(t, env, "status", "test-app")
	AssertOutputContains(t, output, "test-app")
}

// TestScenarioCleanupWorkflow 测试清理工作流场景
func TestScenarioCleanupWorkflow(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// Step 2: 安装多个应用
	apps := []string{"test-app", "git"}
	for _, app := range apps {
		RunCLISuccess(t, env, "install", app)
	}

	// Step 3: 验证安装
	output := RunCLISuccess(t, env, "list")
	for _, app := range apps {
		AssertOutputContains(t, output, app)
	}

	// Step 4: 删除 bucket（应该清理所有相关应用）
	output = RunCLISuccess(t, env, "bucket", "remove", "main")
	AssertOutputContains(t, output, "removed")

	// Step 5: 验证 bucket 已删除
	output = RunCLISuccess(t, env, "bucket", "list")
	AssertOutputNotContains(t, output, "main")
}

// TestScenarioMultipleBuckets 测试多 bucket 场景
func TestScenarioMultipleBuckets(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加多个 bucket
	buckets := []string{"main", "extras"}
	for _, bucket := range buckets {
		output := RunCLISuccess(t, env, "bucket", "add", bucket, env.TestBucketURL)
		AssertOutputContains(t, output, "added")
	}

	// Step 2: 验证所有 bucket 已添加
	output := RunCLISuccess(t, env, "bucket", "list")
	for _, bucket := range buckets {
		AssertOutputContains(t, output, bucket)
	}

	// Step 3: 从不同 bucket 安装应用
	RunCLISuccess(t, env, "install", "main/test-app")
	RunCLISuccess(t, env, "install", "main/git")

	// Step 4: 验证安装
	output = RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")
	AssertOutputContains(t, output, "git")

	// Step 5: 搜索所有 bucket
	output = RunCLISuccess(t, env, "search", "test")
	AssertOutputContains(t, output, "test-app")
}

// TestScenarioSearchAndInstall 测试搜索并安装场景
func TestScenarioSearchAndInstall(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// Step 2: 模糊搜索
	output := RunCLISuccess(t, env, "search", "test")
	AssertOutputContains(t, output, "test-app")

	// Step 3: 精确搜索
	output = RunCLISuccess(t, env, "search", "test-app")
	AssertOutputContains(t, output, "test-app")

	// Step 4: 安装找到的应用
	RunCLISuccess(t, env, "install", "test-app")

	// Step 5: 验证安装
	output = RunCLISuccess(t, env, "list")
	AssertOutputContains(t, output, "test-app")
}

// TestScenarioBatchOperations 测试批量操作场景
func TestScenarioBatchOperations(t *testing.T) {
	env := SetupE2EEnvironment(t)

	// Step 1: 添加 bucket
	RunCLISuccess(t, env, "bucket", "add", "main", env.TestBucketURL)

	// Step 2: 批量安装
	apps := []string{"test-app", "git"}
	for _, app := range apps {
		RunCLISuccess(t, env, "install", app)
	}

	// Step 3: 批量更新
	output := RunCLISuccess(t, env, "update")
	// 应该更新所有应用

	// Step 4: 验证所有应用仍在
	output = RunCLISuccess(t, env, "list")
	for _, app := range apps {
		AssertOutputContains(t, output, app)
	}

	// Step 5: 批量卸载
	for _, app := range apps {
		RunCLISuccess(t, env, "uninstall", app)
	}

	// Step 6: 验证所有应用已卸载
	output = RunCLISuccess(t, env, "list")
	for _, app := range apps {
		AssertOutputNotContains(t, output, app)
	}
}
