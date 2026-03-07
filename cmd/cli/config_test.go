package cli

import (
	"strings"
	"testing"

	"chopsticks/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigCmd(t *testing.T) {
	assert.NotNil(t, configCmd)
	assert.Equal(t, "config", configCmd.Use)
	assert.Contains(t, configCmd.Aliases, "cfg")
	assert.NotEmpty(t, configCmd.Short)
	assert.NotEmpty(t, configCmd.Long)
}

func TestConfigGetCmd(t *testing.T) {
	assert.NotNil(t, configGetCmd)
	assert.True(t, strings.HasPrefix(configGetCmd.Use, "get"))
	assert.NotEmpty(t, configGetCmd.Short)
	assert.NotEmpty(t, configGetCmd.Long)
}

func TestConfigSetCmd(t *testing.T) {
	assert.NotNil(t, configSetCmd)
	assert.True(t, strings.HasPrefix(configSetCmd.Use, "set"))
	assert.NotEmpty(t, configSetCmd.Short)
	assert.NotEmpty(t, configSetCmd.Long)
}

func TestConfigListCmd(t *testing.T) {
	assert.NotNil(t, configListCmd)
	assert.Equal(t, "list", configListCmd.Use)
	assert.Contains(t, configListCmd.Aliases, "ls")
	assert.NotEmpty(t, configListCmd.Short)
	assert.NotEmpty(t, configListCmd.Long)
}

func TestConfigInitCmd(t *testing.T) {
	assert.NotNil(t, configInitCmd)
	assert.Equal(t, "init", configInitCmd.Use)
	assert.NotEmpty(t, configInitCmd.Short)
	assert.NotEmpty(t, configInitCmd.Long)
}

func TestConfigInitCmdFlags(t *testing.T) {
	forceFlag := configInitCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
}

func TestConfigPathCmd(t *testing.T) {
	assert.NotNil(t, configPathCmd)
	assert.Equal(t, "path", configPathCmd.Use)
	assert.Contains(t, configPathCmd.Aliases, "dir")
	assert.NotEmpty(t, configPathCmd.Short)
	assert.NotEmpty(t, configPathCmd.Long)
}

func TestConfigListCmdFlags(t *testing.T) {
	defaultFlag := configListCmd.Flags().Lookup("default")
	assert.NotNil(t, defaultFlag)
	assert.Equal(t, "d", defaultFlag.Shorthand)
}

func TestConfigGetters(t *testing.T) {
	// 测试配置获取器
	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		key      string
		expected string
		wantErr  bool
	}{
		{"apps_dir", "apps_dir", cfg.AppsDir, false},
		{"parallel", "parallel", "3", false},
		{"timeout", "timeout", "300", false},
		{"default_bucket", "default_bucket", "main", false},
		{"proxy_enable", "proxy_enable", "true", false},
		{"log_level", "log_level", "info", false},
		{"invalid", "invalid.key", "", true},
		{"invalid section", "invalidsection.key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := getConfigValue(cfg, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}

func TestConfigSetters(t *testing.T) {
	// 测试配置设置器
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"set parallel", "parallel", "5", false},
		{"set timeout", "timeout", "600", false},
		{"set invalid bool", "no_confirm", "invalid", true},
		{"set invalid int", "parallel", "not-a-number", true},
		{"set invalid key", "invalid.key", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setConfigValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigSettersBoolValues(t *testing.T) {
	cfg := config.DefaultConfig()

	// 测试布尔值设置
	err := setConfigValue(cfg, "no_confirm", "true")
	assert.NoError(t, err)
	assert.True(t, cfg.NoConfirm)

	err = setConfigValue(cfg, "no_confirm", "false")
	assert.NoError(t, err)
	assert.False(t, cfg.NoConfirm)

	// 测试 1/0
	err = setConfigValue(cfg, "no_confirm", "1")
	assert.NoError(t, err)
	assert.True(t, cfg.NoConfirm)

	err = setConfigValue(cfg, "no_confirm", "0")
	assert.NoError(t, err)
	assert.False(t, cfg.NoConfirm)
}

func TestConfigSettersIntValues(t *testing.T) {
	cfg := config.DefaultConfig()

	// 测试整数值设置
	err := setConfigValue(cfg, "parallel", "10")
	assert.NoError(t, err)
	assert.Equal(t, 10, cfg.Parallel)

	err = setConfigValue(cfg, "timeout", "600")
	assert.NoError(t, err)
	assert.Equal(t, 600, cfg.Timeout)

	err = setConfigValue(cfg, "retry", "5")
	assert.NoError(t, err)
	assert.Equal(t, 5, cfg.Retry)
}

func TestConfigSettersStringValues(t *testing.T) {
	cfg := config.DefaultConfig()

	// 测试字符串值设置
	err := setConfigValue(cfg, "apps_dir", "/custom/apps")
	assert.NoError(t, err)
	assert.Equal(t, "/custom/apps", cfg.AppsDir)

	err = setConfigValue(cfg, "default_bucket", "extras")
	assert.NoError(t, err)
	assert.Equal(t, "extras", cfg.DefaultBucket)

	err = setConfigValue(cfg, "proxy_http", "http://127.0.0.1:7890")
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:7890", cfg.ProxyHTTP)
}
