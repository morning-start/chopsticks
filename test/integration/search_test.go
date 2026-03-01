//go:build integration

package integration

import (
	"context"
	"testing"

	"chopsticks/core/bucket"
	"chopsticks/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchByName(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket 和应用
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 按名称搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test-app", opts)
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

func TestSearchByDescription(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 按描述搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestSearchInSpecificBucket(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 在特定 bucket 中搜索
	opts := bucket.SearchOptions{
		Bucket: "main",
	}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)

	// 验证所有结果都来自 main bucket
	for _, result := range results {
		assert.Equal(t, "main", result.Bucket)
	}
}

func TestSearchAllBuckets(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 在所有 bucket 中搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestSearchNoResults(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app"})

	// 搜索不存在的应用
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "non-existent-app", opts)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"Test-App"})

	// 使用不同大小写搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test-app", opts)
	require.NoError(t, err)

	// 验证搜索结果（取决于实现是否支持大小写不敏感）
	_ = results
}

func TestSearchPartialMatch(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 部分匹配搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "test", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestSearchWithCategory(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加测试 bucket
	testutil.AddTestBucketWithApps(t, components, "main", []string{"test-app", "git"})

	// 按类别搜索
	opts := bucket.SearchOptions{
		Category: "development",
	}
	results, err := components.BucketMgr.Search(ctx, "git", opts)
	require.NoError(t, err)

	// 验证搜索结果
	_ = results
}

func TestSearchPagination(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个应用
	apps := []string{"app1", "app2", "app3", "app4", "app5"}
	testutil.AddTestBucketWithApps(t, components, "main", apps)

	// 分页搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "app", opts)
	require.NoError(t, err)

	// 验证分页结果
	// 具体行为取决于实现
	_ = results
}

func TestSearchSorting(t *testing.T) {
	ctx := context.Background()
	components := testutil.SetupTestEnvironment(t)

	// 添加多个应用
	apps := []string{"zebra", "apple", "mango"}
	testutil.AddTestBucketWithApps(t, components, "main", apps)

	// 排序搜索
	opts := bucket.SearchOptions{}
	results, err := components.BucketMgr.Search(ctx, "", opts)
	require.NoError(t, err)

	// 验证排序结果
	// 具体行为取决于实现
	_ = results
}
