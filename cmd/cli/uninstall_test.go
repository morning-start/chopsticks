package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUninstallCmd(t *testing.T) {
	assert.NotNil(t, uninstallCmd)
	assert.True(t, strings.HasPrefix(uninstallCmd.Use, "uninstall"))
	assert.Contains(t, uninstallCmd.Aliases, "remove")
	assert.Contains(t, uninstallCmd.Aliases, "rm")
	assert.NotEmpty(t, uninstallCmd.Short)
	assert.NotEmpty(t, uninstallCmd.Long)
}

func TestUninstallCmdFlags(t *testing.T) {
	// 测试 uninstall 命令标志
	purgeFlag := uninstallCmd.Flags().Lookup("purge")
	assert.NotNil(t, purgeFlag)
	assert.Equal(t, "p", purgeFlag.Shorthand)
}

func TestPrintUninstallResults(t *testing.T) {
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
			err := printUninstallResults(tt.results)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
