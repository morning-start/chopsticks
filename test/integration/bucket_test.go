//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/bucket"
	"chopsticks/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketLifecycle(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 1. 添加 bucket
	addOpts := bucket.AddOptions{}
	err := components.BucketMgr.Add(ctx, "test-bucket", "https://github.com/test/test-bucket", addOpts)
	require.NoError(t, err)

	// 2. 列出 buckets
	buckets, err := components.BucketMgr.ListBuckets(ctx)
	require.NoError(t, err)
	assert.Contains(t, buckets, "test-bucket")

	// 3. 获取 bucket
	b, err := components.BucketMgr.GetBucket(ctx, "test-bucket")
	require.NoError(t, err)
	assert.Equal(t, "test-bucket", b.ID)

	// 4. 更新 bucket
	err = components.BucketMgr.Update(ctx, "test-bucket")
	require.NoError(t, err)

	// 5. 删除 bucket
	err = components.BucketMgr.Remove(ctx, "test-bucket", false)
	require.NoError(t, err)

	// 6. 验证已删除
	buckets, _ = components.BucketMgr.ListBuckets(ctx)
	assert.NotContains(t, buckets, "test-bucket")
}

func TestAddDuplicateBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 第一次添加
	addOpts := bucket.AddOptions{}
	err := components.BucketMgr.Add(ctx, "test-bucket", "https://github.com/test/test-bucket", addOpts)
	require.NoError(t, err)

	// 第二次添加（应该失败）
	err = components.BucketMgr.Add(ctx, "test-bucket", "https://github.com/test/test-bucket", addOpts)
	assert.Error(t, err)
}

func TestRemoveBucketWithPurge(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加 bucket
	addOpts := bucket.AddOptions{}
	err := components.BucketMgr.Add(ctx, "test-bucket", "https://github.com/test/test-bucket", addOpts)
	require.NoError(t, err)

	// 获取 bucket 路径
	bucketPath := filepath.Join(components.TmpDir, "buckets", "test-bucket")

	// 带清理删除
	err = components.BucketMgr.Remove(ctx, "test-bucket", true)
	require.NoError(t, err)

	// 验证目录已删除
	_, err = os.Stat(bucketPath)
	assert.True(t, os.IsNotExist(err), "bucket 目录应该被删除")
}

func TestUpdateAllBuckets(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个 buckets
	addOpts := bucket.AddOptions{}
	err := components.BucketMgr.Add(ctx, "bucket1", "https://github.com/test/bucket1", addOpts)
	require.NoError(t, err)

	err = components.BucketMgr.Add(ctx, "bucket2", "https://github.com/test/bucket2", addOpts)
	require.NoError(t, err)

	// 批量更新
	err = components.BucketMgr.UpdateAll(ctx)
	require.NoError(t, err)
}

func TestBucketSearch(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加 bucket 和应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 搜索应用
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// 验证搜索结果包含 test-app
	found := false
	for _, result := range results {
		if result.App.Name == "test-app" {
			found = true
			break
		}
	}
	assert.True(t, found, "搜索结果应该包含 test-app")
}

func TestBucketSearchInSpecificBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加 bucket 和应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 在特定 bucket 中搜索
	opts := bucket.SearchOptions{
		Bucket: "main",
	}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// 验证所有结果都来自 main bucket
	for _, result := range results {
		assert.Equal(t, "main", result.Bucket)
	}
}

func TestGetBucketNotFound(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 获取不存在的 bucket
	_, err := components.BucketMgr.GetBucket(ctx, "non-existent-bucket")
	assert.Error(t, err)
}

func TestRemoveBucketNotFound(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 删除不存在的 bucket
	err := components.BucketMgr.Remove(ctx, "non-existent-bucket", false)
	assert.Error(t, err)
}

func TestUpdateBucketNotFound(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 更新不存在的 bucket
	err := components.BucketMgr.Update(ctx, "non-existent-bucket")
	assert.Error(t, err)
}

func TestListAppsInBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加 bucket 和多个应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 列出 bucket 中的应用
	apps, err := components.BucketMgr.ListApps(ctx, "main")
	require.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Contains(t, apps, "test-app")
	assert.Contains(t, apps, "git")
}

func TestGetAppFromBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加 bucket 和应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 获取应用
	app, err := components.BucketMgr.GetApp(ctx, "main", "test-app")
	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "test-app", app.Script.Name)
}
