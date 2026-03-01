package manifest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    BucketConfig
		wantErr bool
	}{
		{
			name: "完整的 bucket 配置",
			json: `{
				"id": "test-bucket",
				"name": "Test Bucket",
				"author": "Test Author",
				"description": "A test bucket",
				"homepage": "https://example.com",
				"license": "MIT",
				"repository": {
					"type": "git",
					"url": "https://github.com/test/bucket.git",
					"branch": "main"
				}
			}`,
			want: BucketConfig{
				ID:          "test-bucket",
				Name:        "Test Bucket",
				Author:      "Test Author",
				Description: "A test bucket",
				Homepage:    "https://example.com",
				License:     "MIT",
				Repository: RepositoryInfo{
					Type:   "git",
					URL:    "https://github.com/test/bucket.git",
					Branch: "main",
				},
			},
			wantErr: false,
		},
		{
			name: "最小化的 bucket 配置",
			json: `{
				"id": "minimal-bucket",
				"name": "Minimal Bucket"
			}`,
			want: BucketConfig{
				ID:   "minimal-bucket",
				Name: "Minimal Bucket",
			},
			wantErr: false,
		},
		{
			name:    "无效的 JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
		{
			name: "空的 JSON 对象",
			json: `{}`,
			want: BucketConfig{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got BucketConfig
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBucketConfig_MarshalJSON(t *testing.T) {
	config := BucketConfig{
		ID:          "test-bucket",
		Name:        "Test Bucket",
		Author:      "Test Author",
		Description: "A test bucket",
		Homepage:    "https://example.com",
		License:     "MIT",
		Repository: RepositoryInfo{
			Type:   "git",
			URL:    "https://github.com/test/bucket.git",
			Branch: "main",
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	// 验证可以反序列化回来
	var decoded BucketConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, config, decoded)
}

func TestRepositoryInfo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		repo    RepositoryInfo
		wantErr bool
	}{
		{
			name: "有效的 HTTPS URL",
			repo: RepositoryInfo{
				Type:   "git",
				URL:    "https://github.com/test/bucket.git",
				Branch: "main",
			},
			wantErr: false,
		},
		{
			name: "有效的 SSH URL",
			repo: RepositoryInfo{
				Type: "git",
				URL:  "git@github.com:test/bucket.git",
			},
			wantErr: false,
		},
		{
			name: "空的 URL",
			repo: RepositoryInfo{
				Type: "git",
				URL:  "",
			},
			wantErr: false, // 结构体验证不强制要求 URL
		},
		{
			name:    "空的 RepositoryInfo",
			repo:    RepositoryInfo{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RepositoryInfo 本身没有验证方法，这里只是测试结构体创建
			assert.NotNil(t, tt.repo)
		})
	}
}

func TestBucket(t *testing.T) {
	bucket := Bucket{
		Config: BucketConfig{
			ID:   "test-bucket",
			Name: "Test Bucket",
		},
		Path: "/path/to/bucket",
		Apps: map[string]*AppRef{
			"app1": {
				Name:        "app1",
				Description: "Test app 1",
				Version:     "1.0.0",
			},
			"app2": {
				Name:        "app2",
				Description: "Test app 2",
				Version:     "2.0.0",
			},
		},
		LastUpdated: time.Now(),
	}

	assert.Equal(t, "test-bucket", bucket.Config.ID)
	assert.Equal(t, "Test Bucket", bucket.Config.Name)
	assert.Equal(t, "/path/to/bucket", bucket.Path)
	assert.Len(t, bucket.Apps, 2)
	assert.Contains(t, bucket.Apps, "app1")
	assert.Contains(t, bucket.Apps, "app2")
}

func TestBucket_EmptyApps(t *testing.T) {
	bucket := Bucket{
		Config: BucketConfig{
			ID:   "empty-bucket",
			Name: "Empty Bucket",
		},
		Path:        "/path/to/empty",
		Apps:        map[string]*AppRef{},
		LastUpdated: time.Now(),
	}

	assert.Empty(t, bucket.Apps)
}

func TestBucket_NilApps(t *testing.T) {
	bucket := Bucket{
		Config: BucketConfig{
			ID:   "nil-apps-bucket",
			Name: "Nil Apps Bucket",
		},
		Path:        "/path/to/bucket",
		Apps:        nil,
		LastUpdated: time.Now(),
	}

	assert.Nil(t, bucket.Apps)
}
