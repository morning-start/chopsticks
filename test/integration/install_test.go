//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/app"
	"chopsticks/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullInstallWorkflow(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 执行安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证 MockInstaller 记录了安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"), "应用应该被标记为已安装")

	// 获取安装记录
	record, ok := mockInstaller.GetInstalledApp("test-app")
	require.True(t, ok, "应该能找到安装记录")
	assert.Equal(t, "test-app", record.App.Script.Name)
}

func TestInstallWithSpecificVersion(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 执行安装指定版本
	spec := app.InstallSpec{
		Bucket:  "main",
		Name:    "test-app",
		Version: "1.0.0",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证安装记录中的版本
	installed, err := components.Storage.GetInstalledApp(ctx, "test-app")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", installed.Version)
}

func TestInstallAlreadyInstalled(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 第一次安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 第二次安装（应该失败或跳过）
	err = components.AppMgr.Install(ctx, spec, opts)
	// 根据实现，可能返回错误也可能跳过
	// 这里我们只验证应用仍然安装着
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}

func TestInstallInvalidApp(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 尝试安装不存在的应用
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "non-existent-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	assert.Error(t, err)
}

func TestInstallWithForce(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 第一次安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 使用 Force 选项重新安装
	opts.Force = true
	err = components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证应用仍然安装着
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}

func TestInstallWithArchitecture(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 使用特定架构安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{
		Arch: "64bit",
	}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证应用已安装
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}

func TestInstallCreatesInstallDirectory(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 执行安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证安装目录存在
	installDir := filepath.Join(components.TmpDir, "apps", "test-app")
	_, err = os.Stat(installDir)
	assert.NoError(t, err, "安装目录应该存在")
}

func TestInstallRecordsInStorage(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 执行安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证存储中有记录
	installed, err := components.Storage.GetInstalledApp(ctx, "test-app")
	require.NoError(t, err)
	assert.NotNil(t, installed)
	assert.Equal(t, "test-app", installed.Name)
	assert.NotEmpty(t, installed.InstallDir)
	assert.NotEmpty(t, installed.InstalledAt)
}

func TestInstallMultipleApps(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个应用到测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 安装第一个应用
	spec1 := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec1, app.InstallOptions{})
	require.NoError(t, err)

	// 安装第二个应用
	spec2 := app.InstallSpec{
		Bucket: "main",
		Name:   "git",
	}
	err = components.AppMgr.Install(ctx, spec2, app.InstallOptions{})
	require.NoError(t, err)

	// 验证两个应用都已安装
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
	testutil.AssertAppInstalled(t, components.Storage, "git")

	// 验证列表中有两个应用
	apps, err := components.Storage.ListInstalledApps(ctx)
	require.NoError(t, err)
	assert.Len(t, apps, 2)
}

func TestInstallWithDefaultBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加应用到 main bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 不指定 bucket 进行安装
	spec := app.InstallSpec{
		Name: "test-app",
	}
	opts := app.InstallOptions{}

	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证应用已安装
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}
