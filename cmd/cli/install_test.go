package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAppSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		wantName    string
		wantVersion string
	}{
		{
			name:        "只有名称",
			spec:        "git",
			wantName:    "git",
			wantVersion: "",
		},
		{
			name:        "名称和版本",
			spec:        "git@2.40.0",
			wantName:    "git",
			wantVersion: "2.40.0",
		},
		{
			name:        "复杂版本号",
			spec:        "nodejs@18.17.0",
			wantName:    "nodejs",
			wantVersion: "18.17.0",
		},
		{
			name:        "名称包含特殊字符",
			spec:        "my-app@1.0.0",
			wantName:    "my-app",
			wantVersion: "1.0.0",
		},
		{
			name:        "空字符串",
			spec:        "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "多个@符号",
			spec:        "user@repo@1.0.0",
			wantName:    "user@repo",
			wantVersion: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVersion := parseAppSpec(tt.spec)
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantVersion, gotVersion)
		})
	}
}

func TestPackageSpec(t *testing.T) {
	spec := packageSpec{
		name:    "test-app",
		version: "1.0.0",
		spec:    "test-app@1.0.0",
	}

	assert.Equal(t, "test-app", spec.name)
	assert.Equal(t, "1.0.0", spec.version)
	assert.Equal(t, "test-app@1.0.0", spec.spec)
}

func TestBatchResult(t *testing.T) {
	result := batchResult{
		name:    "test-app",
		success: true,
		err:     nil,
	}

	assert.Equal(t, "test-app", result.name)
	assert.True(t, result.success)
	assert.Nil(t, result.err)
}

func TestInstallCmd(t *testing.T) {
	assert.NotNil(t, installCmd)
	assert.True(t, strings.HasPrefix(installCmd.Use, "install"))
	assert.Contains(t, installCmd.Aliases, "i")
	assert.NotEmpty(t, installCmd.Short)
	assert.NotEmpty(t, installCmd.Long)
}

func TestInstallCmdFlags(t *testing.T) {
	// 测试 install 命令标志
	forceFlag := installCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	archFlag := installCmd.Flags().Lookup("arch")
	assert.NotNil(t, archFlag)
	assert.Equal(t, "a", archFlag.Shorthand)

	bucketFlag := installCmd.Flags().Lookup("bucket")
	assert.NotNil(t, bucketFlag)
	assert.Equal(t, "b", bucketFlag.Shorthand)

	asyncFlag := installCmd.Flags().Lookup("async")
	assert.NotNil(t, asyncFlag)

	workersFlag := installCmd.Flags().Lookup("workers")
	assert.NotNil(t, workersFlag)
	assert.Equal(t, "w", workersFlag.Shorthand)
}

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, "amd64", defaultArch)
	assert.Equal(t, "main", defaultBucket)
}
