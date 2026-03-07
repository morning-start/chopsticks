package cli

import (
	"fmt"
	"os"
	"strconv"

	"chopsticks/pkg/config"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	configListDefault bool
	configInitForce   bool
)

// configCmd 表示 config 命令
var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "管理 Chopsticks 配置",
	Long:    `管理 Chopsticks 配置，包括全局设置、软件源、代理和日志配置。`,
}

// configGetCmd 获取配置项
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "获取配置项的值",
	Long: `获取指定配置项的值。

支持的配置项:
  root_dir            - 根目录
  apps_dir            - 应用安装目录
  buckets_dir         - 软件源目录
  cache_dir           - 缓存目录
  persist_dir         - 持久化数据目录
  shim_dir            - 可执行文件 shim 目录
  storage_dir         - 数据存储目录
  parallel            - 并行下载数
  timeout             - 超时时间 (秒)
  retry               - 重试次数
  no_confirm          - 是否禁用确认提示
  color               - 是否启用彩色输出
  verbose             - 是否启用详细输出
  default_bucket      - 默认软件源
  auto_update         - 是否自动更新软件源
  proxy_enable        - 是否启用代理
  proxy_system        - 是否使用系统代理
  proxy_http          - HTTP 代理地址
  proxy_https         - HTTPS 代理地址
  proxy_no_proxy      - 不代理的地址列表
  log_level           - 日志级别
  log_file            - 日志文件路径
  log_max_size        - 日志文件最大大小 (MB)
  log_max_backups     - 日志文件备份数量
  log_max_age         - 日志文件保留天数
  log_compress        - 是否压缩日志`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd 设置配置项
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项的值",
	Long: `设置指定配置项的值。

示例:
  chopsticks config set parallel 5
  chopsticks config set timeout 600
  chopsticks config set default_bucket extras
  chopsticks config set proxy_enable true
  chopsticks config set proxy_http http://127.0.0.1:7890
  chopsticks config set log_level debug

布尔值使用 true/false，多个值使用逗号分隔。`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd 列出配置
var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有配置项",
	Long:    `列出所有配置项及其当前值。`,
	RunE:    runConfigList,
}

// configInitCmd 初始化配置
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long: `创建默认的配置文件。

如果配置文件已存在，默认会跳过创建。使用 --force 覆盖。`,
	RunE: runConfigInit,
}

// configPathCmd 显示配置路径
var configPathCmd = &cobra.Command{
	Use:     "path",
	Aliases: []string{"dir"},
	Short:   "显示配置文件路径",
	Long:    `显示配置文件的路径。`,
	RunE:    runConfigPath,
}

// configValueGetter 配置值获取函数类型
type configValueGetter func(*config.Config) (string, error)

// configValueSetter 配置值设置函数类型
type configValueSetter func(*config.Config, string) error

// configGetters 配置获取器映射
var configGetters = map[string]configValueGetter{
	"root_dir":        func(cfg *config.Config) (string, error) { return cfg.RootDir, nil },
	"apps_dir":        func(cfg *config.Config) (string, error) { return cfg.AppsDir, nil },
	"buckets_dir":     func(cfg *config.Config) (string, error) { return cfg.BucketsDir, nil },
	"cache_dir":       func(cfg *config.Config) (string, error) { return cfg.CacheDir, nil },
	"persist_dir":     func(cfg *config.Config) (string, error) { return cfg.PersistDir, nil },
	"shim_dir":        func(cfg *config.Config) (string, error) { return cfg.ShimDir, nil },
	"storage_dir":     func(cfg *config.Config) (string, error) { return cfg.StorageDir, nil },
	"parallel":        func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Parallel), nil },
	"timeout":         func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Timeout), nil },
	"retry":           func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.Retry), nil },
	"no_confirm":      func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.NoConfirm), nil },
	"color":           func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Color), nil },
	"verbose":         func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.Verbose), nil },
	"default_bucket":  func(cfg *config.Config) (string, error) { return cfg.DefaultBucket, nil },
	"auto_update":     func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.AutoUpdate), nil },
	"proxy_enable":    func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.ProxyEnable), nil },
	"proxy_system":    func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.ProxySystem), nil },
	"proxy_http":      func(cfg *config.Config) (string, error) { return cfg.ProxyHTTP, nil },
	"proxy_https":     func(cfg *config.Config) (string, error) { return cfg.ProxyHTTPS, nil },
	"proxy_no_proxy":  func(cfg *config.Config) (string, error) { return cfg.ProxyNoProxy, nil },
	"log_level":       func(cfg *config.Config) (string, error) { return cfg.LogLevel, nil },
	"log_file":        func(cfg *config.Config) (string, error) { return cfg.LogFile, nil },
	"log_max_size":    func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.LogMaxSize), nil },
	"log_max_backups": func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.LogMaxBackups), nil },
	"log_max_age":     func(cfg *config.Config) (string, error) { return strconv.Itoa(cfg.LogMaxAge), nil },
	"log_compress":    func(cfg *config.Config) (string, error) { return strconv.FormatBool(cfg.LogCompress), nil },
}

// configSetters 配置设置器映射
var configSetters = map[string]configValueSetter{
	"root_dir":    func(cfg *config.Config, v string) error { cfg.RootDir = v; return nil },
	"apps_dir":    func(cfg *config.Config, v string) error { cfg.AppsDir = v; return nil },
	"buckets_dir": func(cfg *config.Config, v string) error { cfg.BucketsDir = v; return nil },
	"cache_dir":   func(cfg *config.Config, v string) error { cfg.CacheDir = v; return nil },
	"persist_dir": func(cfg *config.Config, v string) error { cfg.PersistDir = v; return nil },
	"shim_dir":    func(cfg *config.Config, v string) error { cfg.ShimDir = v; return nil },
	"storage_dir": func(cfg *config.Config, v string) error { cfg.StorageDir = v; return nil },
	"parallel": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parallel must be an integer")
		}
		cfg.Parallel = i
		return nil
	},
	"timeout": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("timeout must be an integer")
		}
		cfg.Timeout = i
		return nil
	},
	"retry": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("retry must be an integer")
		}
		cfg.Retry = i
		return nil
	},
	"no_confirm": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("no_confirm must be true or false")
		}
		cfg.NoConfirm = b
		return nil
	},
	"color": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("color must be true or false")
		}
		cfg.Color = b
		return nil
	},
	"verbose": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("verbose must be true or false")
		}
		cfg.Verbose = b
		return nil
	},
	"default_bucket": func(cfg *config.Config, v string) error { cfg.DefaultBucket = v; return nil },
	"auto_update": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("auto_update must be true or false")
		}
		cfg.AutoUpdate = b
		return nil
	},
	"proxy_enable": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("proxy_enable must be true or false")
		}
		cfg.ProxyEnable = b
		return nil
	},
	"proxy_system": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("proxy_system must be true or false")
		}
		cfg.ProxySystem = b
		return nil
	},
	"proxy_http":     func(cfg *config.Config, v string) error { cfg.ProxyHTTP = v; return nil },
	"proxy_https":    func(cfg *config.Config, v string) error { cfg.ProxyHTTPS = v; return nil },
	"proxy_no_proxy": func(cfg *config.Config, v string) error { cfg.ProxyNoProxy = v; return nil },
	"log_level":      func(cfg *config.Config, v string) error { cfg.LogLevel = v; return nil },
	"log_file":       func(cfg *config.Config, v string) error { cfg.LogFile = v; return nil },
	"log_max_size": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("log_max_size must be an integer")
		}
		cfg.LogMaxSize = i
		return nil
	},
	"log_max_backups": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("log_max_backups must be an integer")
		}
		cfg.LogMaxBackups = i
		return nil
	},
	"log_max_age": func(cfg *config.Config, v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("log_max_age must be an integer")
		}
		cfg.LogMaxAge = i
		return nil
	},
	"log_compress": func(cfg *config.Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("log_compress must be true or false")
		}
		cfg.LogCompress = b
		return nil
	},
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	value, err := getConfigValue(cfg, key)
	if err != nil {
		return fmt.Errorf("获取配置值 [%s] 失败：%w", key, err)
	}

	fmt.Println(value)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := setConfigValue(cfg, key, value); err != nil {
		return fmt.Errorf("设置配置值 [%s] 失败：%w", key, err)
	}

	if err := config.SaveDefault(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.Success("Config updated: %s = %s", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	var cfg *config.Config
	var err error

	if configListDefault {
		cfg = config.DefaultConfig()
	} else {
		cfg, err = config.LoadDefault()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	fmt.Println("Global Config:")
	fmt.Printf("  root_dir:         %s\n", cfg.RootDir)
	fmt.Printf("  apps_dir:         %s\n", cfg.AppsDir)
	fmt.Printf("  buckets_dir:      %s\n", cfg.BucketsDir)
	fmt.Printf("  cache_dir:        %s\n", cfg.CacheDir)
	fmt.Printf("  storage_dir:      %s\n", cfg.StorageDir)
	fmt.Printf("  parallel:         %d\n", cfg.Parallel)
	fmt.Printf("  timeout:          %d\n", cfg.Timeout)
	fmt.Printf("  retry:            %d\n", cfg.Retry)
	fmt.Printf("  no_confirm:       %t\n", cfg.NoConfirm)
	fmt.Printf("  color:            %t\n", cfg.Color)
	fmt.Printf("  verbose:          %t\n", cfg.Verbose)

	fmt.Println("\nBucket Config:")
	fmt.Printf("  default_bucket:   %s\n", cfg.DefaultBucket)
	fmt.Printf("  auto_update:      %t\n", cfg.AutoUpdate)
	if len(cfg.BucketMirrors) > 0 {
		fmt.Println("  bucket_mirrors:")
		for name, url := range cfg.BucketMirrors {
			fmt.Printf("    %s: %s\n", name, url)
		}
	}

	fmt.Println("\nProxy Config:")
	fmt.Printf("  proxy_enable:     %t\n", cfg.ProxyEnable)
	fmt.Printf("  proxy_system:     %t\n", cfg.ProxySystem)
	fmt.Printf("  proxy_http:       %s\n", cfg.ProxyHTTP)
	fmt.Printf("  proxy_https:      %s\n", cfg.ProxyHTTPS)
	fmt.Printf("  proxy_no_proxy:   %s\n", cfg.ProxyNoProxy)
	if cfg.ProxyEnable && cfg.ProxySystem {
		httpProxy, httpsProxy, noProxy := cfg.GetEffectiveProxy()
		fmt.Println("  effective:")
		fmt.Printf("    http:         %s\n", httpProxy)
		fmt.Printf("    https:        %s\n", httpsProxy)
		fmt.Printf("    no_proxy:     %s\n", noProxy)
	}

	fmt.Println("\nLog Config:")
	fmt.Printf("  log_level:        %s\n", cfg.LogLevel)
	fmt.Printf("  log_file:         %s\n", cfg.LogFile)
	fmt.Printf("  log_max_size:     %d MB\n", cfg.LogMaxSize)
	fmt.Printf("  log_max_backups:  %d\n", cfg.LogMaxBackups)
	fmt.Printf("  log_max_age:      %d days\n", cfg.LogMaxAge)
	fmt.Printf("  log_compress:     %t\n", cfg.LogCompress)

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := config.GetConfigPath()

	// 检查文件是否已存在
	if _, err := os.Stat(configPath); err == nil && !configInitForce {
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
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	configPath := config.GetConfigPath()
	fmt.Println(configPath)
	return nil
}

// getConfigValue 获取配置项的值
func getConfigValue(cfg *config.Config, key string) (string, error) {
	getter, ok := configGetters[key]
	if !ok {
		return "", fmt.Errorf("unknown config key: %s", key)
	}

	return getter(cfg)
}

// setConfigValue 设置配置项的值
func setConfigValue(cfg *config.Config, key, value string) error {
	setter, ok := configSetters[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}

	return setter(cfg, value)
}

func init() {
	configListCmd.Flags().BoolVarP(&configListDefault, "default", "d", false, "显示默认值而不是当前值")
	configInitCmd.Flags().BoolVarP(&configInitForce, "force", "f", false, "强制覆盖已存在的配置文件")

	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)

	rootCmd.AddCommand(configCmd)
}
