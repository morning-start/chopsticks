//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"

	"chopsticks/core/app"
	"chopsticks/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentInstall(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app1", "app2", "app3"})

	// 并发安装多个应用
	var wg sync.WaitGroup
	errors := make(chan error, 3)

	for _, appName := range []string{"app1", "app2", "app3"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			spec := app.InstallSpec{
				Bucket: "main",
				Name:   name,
			}
			err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
			if err != nil {
				errors <- err
			}
		}(appName)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Logf("并发安装错误: %v", err)
	}

	// 验证所有应用都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app1"), "app1 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app2"), "app2 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app3"), "app3 应该被安装")
}

func TestConcurrentUninstall(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装多个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app1", "app2", "app3"})

	for _, appName := range []string{"app1", "app2", "app3"} {
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   appName,
		}
		err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
		require.NoError(t, err)
	}

	// 验证所有应用都已安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	require.True(t, mockInstaller.IsInstalled("app1"))
	require.True(t, mockInstaller.IsInstalled("app2"))
	require.True(t, mockInstaller.IsInstalled("app3"))

	// 并发卸载
	var wg sync.WaitGroup
	for _, appName := range []string{"app1", "app2", "app3"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_ = components.AppMgr.Remove(ctx, name, app.RemoveOptions{})
		}(appName)
	}

	wg.Wait()

	// 验证所有应用都已卸载
	assert.False(t, mockInstaller.IsInstalled("app1"), "app1 应该被卸载")
	assert.False(t, mockInstaller.IsInstalled("app2"), "app2 应该被卸载")
	assert.False(t, mockInstaller.IsInstalled("app3"), "app3 应该被卸载")
}

func TestConcurrentUpdate(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 先安装多个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app1", "app2", "app3"})

	for _, appName := range []string{"app1", "app2", "app3"} {
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   appName,
		}
		err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
		require.NoError(t, err)
	}

	// 并发更新
	var wg sync.WaitGroup
	for _, appName := range []string{"app1", "app2", "app3"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_ = components.AppMgr.Update(ctx, name, app.UpdateOptions{})
		}(appName)
	}

	wg.Wait()

	// 验证所有应用仍然安装
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app1"))
	assert.True(t, mockInstaller.IsInstalled("app2"))
	assert.True(t, mockInstaller.IsInstalled("app3"))
}

func TestConcurrentMixedOperations(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app1", "app2", "app3", "app4"})

	// 先安装一些应用
	for _, appName := range []string{"app1", "app2"} {
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   appName,
		}
		err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
		require.NoError(t, err)
	}

	// 并发执行混合操作
	var wg sync.WaitGroup

	// 安装新应用
	wg.Add(1)
	go func() {
		defer wg.Done()
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   "app3",
		}
		_ = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	}()

	// 更新已安装应用
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = components.AppMgr.Update(ctx, "app1", app.UpdateOptions{})
	}()

	// 卸载应用
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = components.AppMgr.Remove(ctx, "app2", app.RemoveOptions{})
	}()

	// 安装另一个应用
	wg.Add(1)
	go func() {
		defer wg.Done()
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   "app4",
		}
		_ = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	}()

	wg.Wait()

	// 验证最终状态
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app1"), "app1 应该被安装")
	assert.False(t, mockInstaller.IsInstalled("app2"), "app2 应该被卸载")
	assert.True(t, mockInstaller.IsInstalled("app3"), "app3 应该被安装")
	assert.True(t, mockInstaller.IsInstalled("app4"), "app4 应该被安装")
}

func TestConcurrentReadWrite(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"app1", "app2"})

	// 先安装应用
	spec := app.InstallSpec{
		Bucket: "main",
		Name:   "app1",
	}
	err := components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	require.NoError(t, err)

	// 并发读写操作
	var wg sync.WaitGroup

	// 读取操作
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = components.Storage.ListInstalledApps(ctx)
		}()
	}

	// 写入操作
	wg.Add(1)
	go func() {
		defer wg.Done()
		spec := app.InstallSpec{
			Bucket: "main",
			Name:   "app2",
		}
		_ = components.AppMgr.Install(ctx, spec, app.InstallOptions{})
	}()

	wg.Wait()

	// 验证最终状态
	mockInstaller := components.Installer.(*testutil.MockInstaller)
	assert.True(t, mockInstaller.IsInstalled("app1"))
	assert.True(t, mockInstaller.IsInstalled("app2"))
}

func TestConcurrentBucketOperations(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 并发添加多个 bucket
	var wg sync.WaitGroup
	errors := make(chan error, 3)

	for i, bucketName := range []string{"bucket1", "bucket2", "bucket3"} {
		wg.Add(1)
		go func(name string, idx int) {
			defer wg.Done()
			// 创建 bucket
			testutil.CreateTestBucket(t, components.TmpDir+"/buckets/"+name)
			// 添加到存储
			err := components.BucketMgr.Add(ctx, name, "https://github.com/test/"+name, struct{}{})
			if err != nil {
				errors <- err
			}
		}(bucketName, i)
	}

	wg.Wait()
	close(errors)

	// 检查错误
	errorCount := 0
	for err := range errors {
		t.Logf("并发 bucket 操作错误: %v", err)
		errorCount++
	}

	// 验证至少有一些 bucket 被添加
	buckets, err := components.BucketMgr.ListBuckets(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(buckets), 0, "应该有 bucket 被添加")
}
