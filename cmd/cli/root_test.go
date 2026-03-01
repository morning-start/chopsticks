package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "chopsticks", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
	assert.Equal(t, "0.5.0-alpha", rootCmd.Version)
}

func TestRootCmdFlags(t *testing.T) {
	// 测试全局标志
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)

	noColorFlag := rootCmd.PersistentFlags().Lookup("no-color")
	assert.NotNil(t, noColorFlag)
}

func TestRootCmdSubcommands(t *testing.T) {
	// 测试子命令是否存在
	subcommands := rootCmd.Commands()
	assert.NotEmpty(t, subcommands)

	// 验证主要子命令存在
	cmdNames := make(map[string]bool)
	for _, cmd := range subcommands {
		cmdNames[cmd.Name()] = true
	}

	assert.True(t, cmdNames["install"], "install command should exist")
	assert.True(t, cmdNames["uninstall"], "uninstall command should exist")
	assert.True(t, cmdNames["update"], "update command should exist")
	assert.True(t, cmdNames["list"], "list command should exist")
	assert.True(t, cmdNames["search"], "search command should exist")
	assert.True(t, cmdNames["bucket"], "bucket command should exist")
	assert.True(t, cmdNames["config"], "config command should exist")
	assert.True(t, cmdNames["conflict"], "conflict command should exist")
}
