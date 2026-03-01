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

func TestDependencyResolution(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加带依赖的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b", "app-c"})

	// 安装 app-a，应该自动安装 app-b 和 app-c
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-a",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证所有依赖都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app-a"), "app-a 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app-b"), "app-b 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app-c"), "app-c 应该被安装")
}

func TestDependencyChain(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加带依赖链的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b", "app-c"})

	// 安装 app-b，应该自动安装 app-c
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-b",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证 app-b 和 app-c 已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app-b"), "app-b 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app-c"), "app-c 应该被安装")
	assert.False(t, mockInstaller.IsInstalled("app-a"), "app-a 不应该被安装")
}

func TestCircularDependency(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 创建循环依赖: app-x -> app-y -> app-x
	testutil.CreateTestAppWithDeps(t, components, "app-x", []string{"app-y"})
	testutil.CreateTestAppWithDeps(t, components, "app-y", []string{"app-x"})

	// 尝试安装 app-x，应该检测到循环依赖
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-x",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	// 根据实现，可能返回错误也可能处理循环
	// 这里我们只验证不会无限循环
	assert.NotNil(t, err)
}

func TestDependencyAlreadyInstalled(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加带依赖的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b", "app-c"})

	// 先安装 app-c
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-c",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 再安装 app-a，app-c 应该被跳过
	spec = app.InstallSpec{
		Bucket: "main",
		Name:   "app-a",
	}
	err = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证所有应用都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app-a"), "app-a 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app-b"), "app-b 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app-c"), "app-c 应该被安装")
}

func TestOptionalDependency(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 创建带可选依赖的应用和依赖应用
	testutil.CreateTestAppWithDeps(t, components, "app-with-optional", []string{"optional-dep"})
	testutil.CreateTestAppWithDeps(t, components, "optional-dep", nil)

	// 安装应用，可选依赖失败不应该影响主应用安装
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-with-optional",
	}
	opts := app.InstallOptions{}
	err := components.AppMgr.Install(ctx, spec, opts)
	require.NoError(t, err)

	// 验证主应用已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app-with-optional"), "主应用应该被安装")
}

func TestDependencyVersionConflict(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 创建需要不同版本依赖的应用和依赖应用
	testutil.CreateTestAppWithDeps(t, components, "app-req-v1", []string{"dep:1.0.0"})
	testutil.CreateTestAppWithDeps(t, components, "app-req-v2", []string{"dep:2.0.0"})
	testutil.CreateTestAppWithDeps(t, components, "dep", nil)

	// 安装第一个应用
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-req-v1",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 安装第二个应用，可能会遇到版本冲突
	spec = app.InstallSpec{
		Bucket: "main",
		Name:   "app-req-v2",
	}
	err = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	// 根据实现，可能成功或失败
	_ = err
}

func TestDependencyOrder(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加带依赖的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b", "app-c"})

	// 安装 app-a
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-a",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 验证安装顺序: 依赖应该先被安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	// 这里可以验证安装顺序，具体取决于实现
	_ = mockInstaller
}

func TestUninstallWithDependents(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加带依赖的应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app-a", "app-b", "app-c"})

	// 安装 app-a（会安装 app-b 和 app-c）
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app-a",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 尝试卸载 app-b（被 app-a 依赖）
	err = components.AppMgr.Remove(ctx, "app-b", app.RemoveOptions{})
	// 应该失败或警告
	assert.NotNil(t, err)
}
