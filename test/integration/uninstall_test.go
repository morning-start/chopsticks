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

func TestUninstallWorkflow(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证应用已安装
	testutil.AssertAppInstalled(t, components.Storage, "test-app")

	// 执行卸载
	opts := app.RemoveOptions{}
	err = components.AppMgr.Remove(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证应用已卸载
	testutil.AssertAppNotInstalled(t, components.Storage, "test-app")
}

func TestUninstallWithPurge(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 获取安装目录
	installDir := filepath.Join(components.TmpDir, "apps", "test-app")

	// 执行带清理的卸载
	opts := app.RemoveOptions{Purge: true}
	err = components.AppMgr.Remove(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证应用已卸载
	testutil.AssertAppNotInstalled(t, components.Storage, "test-app")

	// 验证安装目录已被删除
	_, err = os.Stat(installDir)
	assert.True(t, os.IsNotExist(err), "安装目录应该被删除")
}

func TestUninstallNotInstalled(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 尝试卸载未安装的应用
	opts := app.RemoveOptions{}
	err := components.AppMgr.Remove(ctx, "not-installed-app", opts)
	assert.Error(t, err)
}

func TestUninstallRemovesInstallDirectory(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证安装目录存在
	installDir := filepath.Join(components.TmpDir, "apps", "test-app")
	_, err = os.Stat(installDir)
	require.NoError(t, err, "安装目录应该存在")

	// 执行卸载
	opts := app.RemoveOptions{Purge: true}
	err = components.AppMgr.Remove(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证安装目录已被删除
	_, err = os.Stat(installDir)
	assert.True(t, os.IsNotExist(err), "安装目录应该被删除")
}

func TestUninstallRemovesFromStorage(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证存储中有记录
	_, err = components.Storage.GetInstalledApp(ctx, "test-app")
	require.NoError(t, err)

	// 执行卸载
	opts := app.RemoveOptions{}
	err = components.AppMgr.Remove(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证存储中无记录
	_, err = components.Storage.GetInstalledApp(ctx, "test-app")
	assert.Error(t, err, "存储中应该没有该应用的记录")
}

func TestUninstallMultipleApps(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装多个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	err := components.AppMgr.Install(ctx, app.InstallSpec{Bucket: "main", Name: "test-app"}, app.InstallOptions{})
	require.NoError(t, err)

	err = components.AppMgr.Install(ctx, app.InstallSpec{Bucket: "main", Name: "git"}, app.InstallOptions{})
	require.NoError(t, err)

	// 验证两个应用都已安装
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
	testutil.AssertAppInstalled(t, components.Storage, "git")

	// 卸载第一个应用
	err = components.AppMgr.Remove(ctx, "test-app", app.RemoveOptions{})
	require.NoError(t, err)

	// 验证第一个应用已卸载，第二个应用仍在
	testutil.AssertAppNotInstalled(t, components.Storage, "test-app")
	testutil.AssertAppInstalled(t, components.Storage, "git")
}
