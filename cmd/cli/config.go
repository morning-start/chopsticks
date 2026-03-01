// Package cli 提供命令行界面功能。
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"chopsticks/pkg/config"
	"chopsticks/pkg/output"

	"github.com/urfave/cli/v2"
)

// configCommand 返回配置管理命令
func configCommand() *cli.Command {
	return &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "管理 Chopsticks 配置",
		Subcommands: []*cli.Command{
			configGetCommand(),
			configSetCommand(),
			configListCommand(),
			configInitCommand(),
			configPathCommand(),
		},
	}
}

// configGetCommand 返回配置获取子命令
func configGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "获取配置项的值",
		ArgsUsage: "<key>",
		Description: `获取指定配置项的值。

支持的配置项:
  global.apps_path      - 应用安装路径
  global.buckets_path   - 软件源路径
  global.cache_path     - 缓存路径
  global.storage_path   - 数据库路径
  global.parallel       - 并行下载数
  global.timeout        - 超时时间(秒)
  global.retry          - 重试次数
  global.no_confirm     - 是否禁用确认提示
  global.color          - 是否启用彩色输出
  global.verbose        - 是否启用详细输出
  buckets.default       - 默认软件源
  buckets.auto_update   - 是否自动更新软件源
  proxy.enable          - 是否启用代理
  proxy.system          - 是否使用系统代理(从环境变量读取)
  proxy.http            - HTTP 代理地址
  proxy.https           - HTTPS 代理地址
  proxy.no_proxy        - 不代理的地址列表
  log.level             - 日志级别(debug/info/warn/error)
  log.file              - 日志文件路径
  log.max_size          - 日志文件最大大小(MB)
  log.max_backups       - 日志文件备份数量
  log.max_age           - 日志文件保留天数
  log.compress          - 是否压缩日志`,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("please specify config key, e.g., chopsticks config get global.parallel")
			}

			key := c.Args().First()
			cfg, err := config.LoadDefault()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			value, err := getConfigValue(cfg, key)
			if err != nil {
				return err
			}

			fmt.Println(value)
			return nil
		},
	}
}

// configValueGetter 配置值获取函数类型
type configValueGetter func(*config.Config) (string, error)

// configValueSetter 配置值设置函数类型
type configValueSetter func(*config.Config, string) error

// configGetters 配置获取器映射
var configGetters = map[string]map[string]configValueGetter{
	"global": {
		"apps_path":    func(cfg *config.Config) (string, error) { return cfg.Global.AppsPath, nil },
		"buckets_path": func(cfg *config.Config) (string, error) { return cfg.Global.BucketsPath, nil },
		"cache_path":   func(cfg *config.Config) (string, error) { return cfg.Global.CachePath, nil },
		"storage_path": func(cfg *config.Config) (string, error) { return cfg.Global.StoragePath, nil },
		"parallel":     func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Global.Parallel), nil },
		"timeout":      func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Global.Timeout), nil },
		"retry":        func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Global.Retry), nil },
		"no_confirm":   func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Global.NoConfirm), nil },
		"color":        func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Global.Color), nil },
		"verbose":      func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Global.Verbose), nil },
	},
	"buckets": {
		"default":     func(cfg *config.Config) (string, error) { return cfg.Buckets.Default, nil },
		"auto_update": func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Buckets.AutoUpdate), nil },
	},
	"proxy": {
		"enable":   func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Proxy.Enable), nil },
		"system":   func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Proxy.System), nil },
		"http":     func(cfg *config.Config) (string, error) { return cfg.Proxy.HTTP, nil },
		"https":    func(cfg *config.Config) (string, error) { return cfg.Proxy.HTTPS, nil },
		"no_proxy": func(cfg *config.Config) (string, error) { return cfg.Proxy.NoProxy, nil },
	},
	"log": {
		"level":       func(cfg *config.Config) (string, error) { return cfg.Log.Level, nil },
		"file":        func(cfg *config.Config) (string, error) { return cfg.Log.File, nil },
		"max_size":    func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Log.MaxSize), nil },
		"max_backups": func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Log.MaxBackups), nil },
		"max_age":     func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Log.MaxAge), nil },
		"compress":    func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Log.Compress), nil },
	},
}

// configSetCommand 返回配置设置子命令
func configSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "设置配置项的值",
		ArgsUsage: "<key> <value>",
		Description: `设置指定配置项的值。

示例:
  chopsticks config set global.parallel 5
  chopsticks config set global.timeout 600
  chopsticks config set buckets.default extras
  chopsticks config set proxy.enable true
  chopsticks config set proxy.http http://127.0.0.1:7890
  chopsticks config set log.level debug

布尔值使用 true/false，多个值使用逗号分隔。`,
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				return fmt.Errorf("please specify config key and value, e.g., chopsticks config set global.parallel 5")
			}

			key := c.Args().Get(0)
			value := c.Args().Get(1)

			cfg, err := config.LoadDefault()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := setConfigValue(cfg, key, value); err != nil {
				return err
			}

			if err := config.SaveDefault(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			output.Success("Config updated: %s = %s", key, value)
			return nil
		},
	}
}

// configSetters 配置设置器映射
var configSetters = map[string]map[string]configValueSetter{
	"global": {
		"apps_path":    func(cfg *config.Config, v string) error { cfg.Global.AppsPath = v; return nil },
		"buckets_path": func(cfg *config.Config, v string) error { cfg.Global.BucketsPath = v; return nil },
		"cache_path":   func(cfg *config.Config, v string) error { cfg.Global.CachePath = v; return nil },
		"storage_path": func(cfg *config.Config, v string) error { cfg.Global.StoragePath = v; return nil },
		"parallel": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("parallel must be an integer")
			}
			cfg.Global.Parallel = i
			return nil
		},
		"timeout": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("timeout must be an integer")
			}
			cfg.Global.Timeout = i
			return nil
		},
		"retry": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("retry must be an integer")
			}
			cfg.Global.Retry = i
			return nil
		},
		"no_confirm": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("no_confirm must be true or false")
			}
			cfg.Global.NoConfirm = b
			return nil
		},
		"color": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("color must be true or false")
			}
			cfg.Global.Color = b
			return nil
		},
		"verbose": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("verbose must be true or false")
			}
			cfg.Global.Verbose = b
			return nil
		},
	},
	"buckets": {
		"default": func(cfg *config.Config, v string) error { cfg.Buckets.Default = v; return nil },
		"auto_update": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("auto_update must be true or false")
			}
			cfg.Buckets.AutoUpdate = b
			return nil
		},
	},
	"proxy": {
		"enable": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("enable must be true or false")
			}
			cfg.Proxy.Enable = b
			return nil
		},
		"system": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("system must be true or false")
			}
			cfg.Proxy.System = b
			return nil
		},
		"http":     func(cfg *config.Config, v string) error { cfg.Proxy.HTTP = v; return nil },
		"https":    func(cfg *config.Config, v string) error { cfg.Proxy.HTTPS = v; return nil },
		"no_proxy": func(cfg *config.Config, v string) error { cfg.Proxy.NoProxy = v; return nil },
	},
	"log": {
		"level": func(cfg *config.Config, v string) error { cfg.Log.Level = v; return nil },
		"file":  func(cfg *config.Config, v string) error { cfg.Log.File = v; return nil },
		"max_size": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("max_size must be an integer")
			}
			cfg.Log.MaxSize = i
			return nil
		},
		"max_backups": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("max_backups must be an integer")
			}
			cfg.Log.MaxBackups = i
			return nil
		},
		"max_age": func(cfg *config.Config, v string) error {
			i, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("max_age must be an integer")
			}
			cfg.Log.MaxAge = i
			return nil
		},
		"compress": func(cfg *config.Config, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("compress must be true or false")
			}
			cfg.Log.Compress = b
			return nil
		},
	},
}

// configListCommand 返回配置列子命令
func configListCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "列出所有配置项",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "default",
				Aliases: []string{"d"},
				Usage:   "显示默认值而不是当前值",
			},
		},
		Action: func(c *cli.Context) error {
			var cfg *config.Config
			var err error

			if c.Bool("default") {
				cfg = config.DefaultConfig()
			} else {
				cfg, err = config.LoadDefault()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
			}

			fmt.Println("Global Config:")
			fmt.Printf("  apps_path:      %s\n", cfg.Global.AppsPath)
			fmt.Printf("  buckets_path:   %s\n", cfg.Global.BucketsPath)
			fmt.Printf("  cache_path:     %s\n", cfg.Global.CachePath)
			fmt.Printf("  storage_path:   %s\n", cfg.Global.StoragePath)
			fmt.Printf("  parallel:       %d\n", cfg.Global.Parallel)
			fmt.Printf("  timeout:        %d\n", cfg.Global.Timeout)
			fmt.Printf("  retry:          %d\n", cfg.Global.Retry)
			fmt.Printf("  no_confirm:     %t\n", cfg.Global.NoConfirm)
			fmt.Printf("  color:          %t\n", cfg.Global.Color)
			fmt.Printf("  verbose:        %t\n", cfg.Global.Verbose)

			fmt.Println("\nBucket Config:")
			fmt.Printf("  default:        %s\n", cfg.Buckets.Default)
			fmt.Printf("  auto_update:    %t\n", cfg.Buckets.AutoUpdate)
			if len(cfg.Buckets.Mirrors) > 0 {
				fmt.Println("  mirrors:")
				for name, url := range cfg.Buckets.Mirrors {
					fmt.Printf("    %s: %s\n", name, url)
				}
			}

			fmt.Println("\nProxy Config:")
			fmt.Printf("  enable:         %t\n", cfg.Proxy.Enable)
			fmt.Printf("  system:         %t\n", cfg.Proxy.System)
			fmt.Printf("  http:           %s\n", cfg.Proxy.HTTP)
			fmt.Printf("  https:          %s\n", cfg.Proxy.HTTPS)
			fmt.Printf("  no_proxy:       %s\n", cfg.Proxy.NoProxy)
			if cfg.Proxy.Enable && cfg.Proxy.System {
				httpProxy, httpsProxy, noProxy := cfg.Proxy.GetEffectiveProxy()
				fmt.Println("  effective:")
				fmt.Printf("    http:         %s\n", httpProxy)
				fmt.Printf("    https:        %s\n", httpsProxy)
				fmt.Printf("    no_proxy:     %s\n", noProxy)
			}

			fmt.Println("\nLog Config:")
			fmt.Printf("  level:          %s\n", cfg.Log.Level)
			fmt.Printf("  file:           %s\n", cfg.Log.File)
			fmt.Printf("  max_size:       %d MB\n", cfg.Log.MaxSize)
			fmt.Printf("  max_backups:    %d\n", cfg.Log.MaxBackups)
			fmt.Printf("  max_age:        %d days\n", cfg.Log.MaxAge)
			fmt.Printf("  compress:       %t\n", cfg.Log.Compress)

			return nil
		},
	}
}

// configInitCommand 返回配置初始化子命令
func configInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "初始化配置文件",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "强制覆盖已存在的配置文件",
			},
		},
		Description: `创建默认的配置文件。

如果配置文件已存在，默认会跳过创建。使用 --force 覆盖。`,
		Action: func(c *cli.Context) error {
			configPath := config.GetConfigPath()

			// 检查文件是否已存在
			if _, err := os.Stat(configPath); err == nil && !c.Bool("force") {
				return fmt.Errorf("config file already exists: %s\nuse --force to overwrite", configPath)
			}

			// 确保配置目录存在
			if err := config.EnsureConfigDir(); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// 创建默认配置
			cfg := config.DefaultConfig()
			if err := config.SaveDefault(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			output.Success("Config file created: %s", configPath)
			return nil
		},
	}
}

// configPathCommand 返回配置路径子命令
func configPathCommand() *cli.Command {
	return &cli.Command{
		Name:    "path",
		Aliases: []string{"dir"},
		Usage:   "显示配置文件路径",
		Action: func(c *cli.Context) error {
			configPath := config.GetConfigPath()
			fmt.Println(configPath)
			return nil
		},
	}
}

// getConfigValue 获取配置项的值
func getConfigValue(cfg *config.Config, key string) (string, error) {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid config key format, expected: section.key")
	}

	section, name := parts[0], parts[1]

	sectionMap, ok := configGetters[section]
	if !ok {
		return "", fmt.Errorf("unknown config section: %s", section)
	}

	getter, ok := sectionMap[name]
	if !ok {
		return "", fmt.Errorf("unknown config key: %s.%s", section, name)
	}

	return getter(cfg)
}

// setConfigValue 设置配置项的值
func setConfigValue(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid config key format, expected: section.key")
	}

	section, name := parts[0], parts[1]

	sectionMap, ok := configSetters[section]
	if !ok {
		return fmt.Errorf("unknown config section: %s", section)
	}

	setter, ok := sectionMap[name]
	if !ok {
		return fmt.Errorf("unknown config key: %s.%s", section, name)
	}

	return setter(cfg, value)
}
