package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd(t *testing.T) {
	assert.NotNil(t, updateCmd)
	assert.True(t, strings.HasPrefix(updateCmd.Use, "update"))
	assert.Contains(t, updateCmd.Aliases, "upgrade")
	assert.Contains(t, updateCmd.Aliases, "up")
	assert.NotEmpty(t, updateCmd.Short)
	assert.NotEmpty(t, updateCmd.Long)
}

func TestUpdateCmdFlags(t *testing.T) {
	// 测试 update 命令标志
	allFlag := updateCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
	assert.Equal(t, "a", allFlag.Shorthand)

	forceFlag := updateCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	asyncFlag := updateCmd.Flags().Lookup("async")
	assert.NotNil(t, asyncFlag)

	workersFlag := updateCmd.Flags().Lookup("workers")
	assert.NotNil(t, workersFlag)
	assert.Equal(t, "w", workersFlag.Shorthand)
}

func TestPrintUpdateResults(t *testing.T) {
	tests := []struct {
		name         string
		results      []batchResult
		wantErr      bool
		successCount int
		failCount    int
	}{
		{
			name: "全部成功",
			results: []batchResult{
				{name: "app1", success: true, err: nil},
				{name: "app2", success: true, err: nil},
			},
			wantErr:      false,
			successCount: 2,
			failCount:    0,
		},
		{
			name: "部分失败",
			results: []batchResult{
				{name: "app1", success: true, err: nil},
				{name: "app2", success: false, err: assert.AnError},
			},
			wantErr:      true,
			successCount: 1,
			failCount:    1,
		},
		{
			name:         "空结果",
			results:      []batchResult{},
			wantErr:      false,
			successCount: 0,
			failCount:    0,
		},
		{
			name: "全部失败",
			results: []batchResult{
				{name: "app1", success: false, err: assert.AnError},
				{name: "app2", success: false, err: assert.AnError},
			},
			wantErr:      true,
			successCount: 0,
			failCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := printUpdateResults(tt.results)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
