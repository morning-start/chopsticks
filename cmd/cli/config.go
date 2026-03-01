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
				return fmt.Errorf("请指定配置项名称，例如: chopsticks config get global.parallel")
			}

			key := c.Args().First()
			cfg, err := config.LoadDefault()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
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
				return fmt.Errorf("请指定配置项名称和值，例如: chopsticks config set global.parallel 5")
			}

			key := c.Args().Get(0)
			value := c.Args().Get(1)

			cfg, err := config.LoadDefault()
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}

			if err := setConfigValue(cfg, key, value); err != nil {
				return err
			}

			if err := config.SaveDefault(cfg); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			output.Success("配置已更新: %s = %s", key, value)
			return nil
		},
	}
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
					return fmt.Errorf("加载配置失败: %w", err)
				}
			}

			fmt.Println("全局配置:")
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

			fmt.Println("\n软件源配置:")
			fmt.Printf("  default:        %s\n", cfg.Buckets.Default)
			fmt.Printf("  auto_update:    %t\n", cfg.Buckets.AutoUpdate)
			if len(cfg.Buckets.Mirrors) > 0 {
				fmt.Println("  mirrors:")
				for name, url := range cfg.Buckets.Mirrors {
					fmt.Printf("    %s: %s\n", name, url)
				}
			}

			fmt.Println("\n代理配置:")
			fmt.Printf("  enable:         %t\n", cfg.Proxy.Enable)
			fmt.Printf("  http:           %s\n", cfg.Proxy.HTTP)
			fmt.Printf("  https:          %s\n", cfg.Proxy.HTTPS)
			fmt.Printf("  no_proxy:       %s\n", cfg.Proxy.NoProxy)

			fmt.Println("\n日志配置:")
			fmt.Printf("  level:          %s\n", cfg.Log.Level)
			fmt.Printf("  file:           %s\n", cfg.Log.File)
			fmt.Printf("  max_size:       %d MB\n", cfg.Log.MaxSize)
			fmt.Printf("  max_backups:    %d\n", cfg.Log.MaxBackups)
			fmt.Printf("  max_age:        %d 天\n", cfg.Log.MaxAge)
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
				return fmt.Errorf("配置文件已存在: %s\n使用 --force 覆盖", configPath)
			}

			// 确保配置目录存在
			if err := config.EnsureConfigDir(); err != nil {
				return fmt.Errorf("创建配置目录失败: %w", err)
			}

			// 创建默认配置
			cfg := config.DefaultConfig()
			if err := config.SaveDefault(cfg); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			output.Success("配置文件已创建: %s", configPath)
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
		return "", fmt.Errorf("配置项格式错误，应为: section.key")
	}

	section := parts[0]
	name := parts[1]

	switch section {
	case "global":
		switch name {
		case "apps_path":
			return cfg.Global.AppsPath, nil
		case "buckets_path":
			return cfg.Global.BucketsPath, nil
		case "cache_path":
			return cfg.Global.CachePath, nil
		case "storage_path":
			return cfg.Global.StoragePath, nil
		case "parallel":
			return strconv.Itoa(cfg.Global.Parallel), nil
		case "timeout":
			return strconv.Itoa(cfg.Global.Timeout), nil
		case "retry":
			return strconv.Itoa(cfg.Global.Retry), nil
		case "no_confirm":
			return strconv.FormatBool(cfg.Global.NoConfirm), nil
		case "color":
			return strconv.FormatBool(cfg.Global.Color), nil
		case "verbose":
			return strconv.FormatBool(cfg.Global.Verbose), nil
		default:
			return "", fmt.Errorf("未知的配置项: global.%s", name)
		}
	case "buckets":
		switch name {
		case "default":
			return cfg.Buckets.Default, nil
		case "auto_update":
			return strconv.FormatBool(cfg.Buckets.AutoUpdate), nil
		default:
			return "", fmt.Errorf("未知的配置项: buckets.%s", name)
		}
	case "proxy":
		switch name {
		case "enable":
			return strconv.FormatBool(cfg.Proxy.Enable), nil
		case "http":
			return cfg.Proxy.HTTP, nil
		case "https":
			return cfg.Proxy.HTTPS, nil
		case "no_proxy":
			return cfg.Proxy.NoProxy, nil
		default:
			return "", fmt.Errorf("未知的配置项: proxy.%s", name)
		}
	case "log":
		switch name {
		case "level":
			return cfg.Log.Level, nil
		case "file":
			return cfg.Log.File, nil
		case "max_size":
			return strconv.Itoa(cfg.Log.MaxSize), nil
		case "max_backups":
			return strconv.Itoa(cfg.Log.MaxBackups), nil
		case "max_age":
			return strconv.Itoa(cfg.Log.MaxAge), nil
		case "compress":
			return strconv.FormatBool(cfg.Log.Compress), nil
		default:
			return "", fmt.Errorf("未知的配置项: log.%s", name)
		}
	default:
		return "", fmt.Errorf("未知的配置段: %s", section)
	}
}

// setConfigValue 设置配置项的值
func setConfigValue(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return fmt.Errorf("配置项格式错误，应为: section.key")
	}

	section := parts[0]
	name := parts[1]

	switch section {
	case "global":
		switch name {
		case "apps_path":
			cfg.Global.AppsPath = value
		case "buckets_path":
			cfg.Global.BucketsPath = value
		case "cache_path":
			cfg.Global.CachePath = value
		case "storage_path":
			cfg.Global.StoragePath = value
		case "parallel":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("parallel 必须是整数")
			}
			cfg.Global.Parallel = v
		case "timeout":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("timeout 必须是整数")
			}
			cfg.Global.Timeout = v
		case "retry":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("retry 必须是整数")
			}
			cfg.Global.Retry = v
		case "no_confirm":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("no_confirm 必须是 true 或 false")
			}
			cfg.Global.NoConfirm = v
		case "color":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("color 必须是 true 或 false")
			}
			cfg.Global.Color = v
		case "verbose":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("verbose 必须是 true 或 false")
			}
			cfg.Global.Verbose = v
		default:
			return fmt.Errorf("未知的配置项: global.%s", name)
		}
	case "buckets":
		switch name {
		case "default":
			cfg.Buckets.Default = value
		case "auto_update":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("auto_update 必须是 true 或 false")
			}
			cfg.Buckets.AutoUpdate = v
		default:
			return fmt.Errorf("未知的配置项: buckets.%s", name)
		}
	case "proxy":
		switch name {
		case "enable":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("enable 必须是 true 或 false")
			}
			cfg.Proxy.Enable = v
		case "http":
			cfg.Proxy.HTTP = value
		case "https":
			cfg.Proxy.HTTPS = value
		case "no_proxy":
			cfg.Proxy.NoProxy = value
		default:
			return fmt.Errorf("未知的配置项: proxy.%s", name)
		}
	case "log":
		switch name {
		case "level":
			cfg.Log.Level = value
		case "file":
			cfg.Log.File = value
		case "max_size":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("max_size 必须是整数")
			}
			cfg.Log.MaxSize = v
		case "max_backups":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("max_backups 必须是整数")
			}
			cfg.Log.MaxBackups = v
		case "max_age":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("max_age 必须是整数")
			}
			cfg.Log.MaxAge = v
		case "compress":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("compress 必须是 true 或 false")
			}
			cfg.Log.Compress = v
		default:
			return fmt.Errorf("未知的配置项: log.%s", name)
		}
	default:
		return fmt.Errorf("未知的配置段: %s", section)
	}

	return nil
}
