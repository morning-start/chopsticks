package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchCmd(t *testing.T) {
	assert.NotNil(t, searchCmd)
	assert.True(t, strings.HasPrefix(searchCmd.Use, "search"))
	assert.Contains(t, searchCmd.Aliases, "find")
	assert.Contains(t, searchCmd.Aliases, "s")
	assert.NotEmpty(t, searchCmd.Short)
	assert.NotEmpty(t, searchCmd.Long)
}

func TestSearchCmdFlags(t *testing.T) {
	// 测试 search 命令标志
	bucketFlag := searchCmd.Flags().Lookup("bucket")
	assert.NotNil(t, bucketFlag)
	assert.Equal(t, "b", bucketFlag.Shorthand)

	asyncFlag := searchCmd.Flags().Lookup("async")
	assert.NotNil(t, asyncFlag)

	workersFlag := searchCmd.Flags().Lookup("workers")
	assert.NotNil(t, workersFlag)
	assert.Equal(t, "w", workersFlag.Shorthand)
}
