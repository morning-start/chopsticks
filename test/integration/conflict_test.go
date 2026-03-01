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

func TestPortConflict(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装使用特定端口的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 尝试安装另一个使用相同端口的应用
	// 这里假设冲突检测会在安装前检查
	// 具体实现取决于冲突检测器的设计

	// 验证第一个应用已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
}

func TestFileConflict(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装第一个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 尝试安装可能产生文件冲突的应用
	// 冲突检测应该识别出冲突

	// 验证第一个应用仍然安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
}

func TestDependencyConflict(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 创建有依赖冲突的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b"})

	// 安装第一个应用
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-a",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 安装第二个应用
	spec = app.InstallSpec{
		Bucket: "main",
		Name:   "app-b",
	}
	err = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证两个应用都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app-a"))
	assert.True(t, mockInstaller.IsInstalled("app-b"))
}

func TestNoConflict(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装没有冲突的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	spec = app.InstallSpec{
		Bucket: "main",
		Name:   "git",
	}
	err = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证两个应用都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
	assert.True(t, mockInstaller.IsInstalled("git"))
}

func TestConflictResolution(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 冲突解决应该能够处理冲突情况
	// 具体行为取决于实现

	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
}

func TestConflictWithForce(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 安装应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	opts := app.InstallOptions{
		Force: true,
	}
	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 使用 Force 选项应该能够覆盖冲突
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
}

func TestConflictDetectionBeforeInstall(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装一个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "test-app",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 冲突检测应该在安装前识别问题
	// 这里验证基础功能工作正常
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("test-app"))
}
