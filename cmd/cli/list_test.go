package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCmd(t *testing.T) {
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Contains(t, listCmd.Aliases, "ls")
	assert.NotEmpty(t, listCmd.Short)
	assert.NotEmpty(t, listCmd.Long)
}

func TestListCmdFlags(t *testing.T) {
	// 测试 list 命令标志
	installedFlag := listCmd.Flags().Lookup("installed")
	assert.NotNil(t, installedFlag)
	assert.Equal(t, "i", installedFlag.Shorthand)
}
