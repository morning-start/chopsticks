package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConflictCmd(t *testing.T) {
	assert.NotNil(t, conflictCmd)
	assert.True(t, strings.HasPrefix(conflictCmd.Use, "conflict"))
	assert.Contains(t, conflictCmd.Aliases, "check")
	assert.NotEmpty(t, conflictCmd.Short)
	assert.NotEmpty(t, conflictCmd.Long)
}

func TestConflictCmdFlags(t *testing.T) {
	// 测试 conflict 命令标志
	bucketFlag := conflictCmd.Flags().Lookup("bucket")
	assert.NotNil(t, bucketFlag)
	assert.Equal(t, "b", bucketFlag.Shorthand)

	verboseFlag := conflictCmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
}
