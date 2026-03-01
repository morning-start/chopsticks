//go:build integration

package integration

import (
	"context"
	"testing"

	"chopsticks/core/app"
	"chopsticks/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateWorkflow(t *testing.T) {
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

	// 执行更新
	opts := app.UpdateOptions{}
	err = components.AppMgr.Update(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证应用仍然安装着
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}

func TestUpdateAll(t *testing.T) {
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

	// 执行批量更新
	opts := app.UpdateOptions{}
	err = components.AppMgr.UpdateAll(ctx, opts)
	require.NoError(t, err)

	// 验证两个应用仍然安装着
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
	testutil.AssertAppInstalled(t, components.Storage, "git")
}

func TestUpdateNotInstalled(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 尝试更新未安装的应用
	opts := app.UpdateOptions{}
	err := components.AppMgr.Update(ctx, "not-installed-app", opts)
	assert.Error(t, err)
}

func TestUpdateWithForce(t *testing.T) {
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

	// 使用 Force 选项执行更新
	opts := app.UpdateOptions{Force: true}
	err = components.AppMgr.Update(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证应用仍然安装着
	testutil.AssertAppInstalled(t, components.Storage, "test-app")
}

func TestUpdatePreservesInstallDir(t *testing.T) {
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
	installed, err := components.Storage.GetInstalledApp(ctx, "test-app")
	require.NoError(t, err)
	originalInstallDir := installed.InstallDir

	// 执行更新
	opts := app.UpdateOptions{}
	err = components.AppMgr.Update(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证安装目录未改变
	installed, err = components.Storage.GetInstalledApp(ctx, "test-app")
	require.NoError(t, err)
	assert.Equal(t, originalInstallDir, installed.InstallDir)
}

func TestUpdateAllWithNoApps(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 没有安装任何应用时执行批量更新
	opts := app.UpdateOptions{}
	err := components.AppMgr.UpdateAll(ctx, opts)
	require.NoError(t, err)
}
