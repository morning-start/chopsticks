// Package cli 提供命令行界面功能。
package cli

// Command 定义命令信息。
type Command struct {
	Name        string   // 主命令名（用户友好）
	Aliases     []string // 别名（内部名或其他别名）
	Description string   // 描述
	Usage       string   // 用法
	Examples    []string // 示例
}

// commands 定义所有可用命令。
var commands = []Command{
	{
		Name:        "install",
		Aliases:     []string{"serve", "i"},
		Description: "安装软件包",
		Usage:       "install <package>[@version] [--force]",
		Examples: []string{
			"chopsticks install git",
			"chopsticks install nodejs@18.17.0",
			"chopsticks install git --force",
		},
	},
	{
		Name:        "uninstall",
		Aliases:     []string{"clear", "remove", "rm"},
		Description: "卸载软件包",
		Usage:       "uninstall <package> [--purge]",
		Examples: []string{
			"chopsticks uninstall git",
			"chopsticks uninstall git --purge",
		},
	},
	{
		Name:        "update",
		Aliases:     []string{"refresh", "upgrade", "up"},
		Description: "更新软件包",
		Usage:       "update [package] [--all] [--force]",
		Examples: []string{
			"chopsticks update git",
			"chopsticks update --all",
			"chopsticks update git --force",
		},
	},
	{
		Name:        "search",
		Aliases:     []string{"find", "s"},
		Description: "搜索软件包",
		Usage:       "search <query> [--bow <bow>]",
		Examples: []string{
			"chopsticks search git",
			"chopsticks search editor --bow extras",
		},
	},
	{
		Name:        "list",
		Aliases:     []string{"ls"},
		Description: "列出软件包",
		Usage:       "list [--installed]",
		Examples: []string{
			"chopsticks list",
			"chopsticks list --installed",
		},
	},
	{
		Name:        "bucket",
		Aliases:     []string{"bow", "b"},
		Description: "软件源管理",
		Usage:       "bucket <subcommand>",
		Examples: []string{
			"chopsticks bucket add main https://github.com/chopsticks-bows/main",
			"chopsticks bucket list",
			"chopsticks bucket update",
		},
	},
	{
		Name:        "completion",
		Aliases:     []string{},
		Description: "生成自动补全脚本",
		Usage:       "completion <shell>",
		Examples: []string{
			"chopsticks completion bash",
			"chopsticks completion zsh",
			"chopsticks completion powershell",
			"chopsticks completion fish",
		},
	},
	{
		Name:        "help",
		Aliases:     []string{"--help", "-h"},
		Description: "显示帮助信息",
		Usage:       "help [command]",
		Examples: []string{
			"chopsticks help",
			"chopsticks help install",
		},
	},
}

// commandMap 存储命令名到主命令的映射。
var commandMap = make(map[string]string)

func init() {
	// 构建命令映射
	for _, cmd := range commands {
		// 主命令名映射到自身
		commandMap[cmd.Name] = cmd.Name
		// 别名映射到主命令名
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd.Name
		}
	}
}

// resolveCommand 解析命令名，返回主命令名。
func resolveCommand(name string) (string, bool) {
	primary, ok := commandMap[name]
	return primary, ok
}

// getCommand 获取命令信息。
func getCommand(name string) (*Command, bool) {
	primary, ok := resolveCommand(name)
	if !ok {
		return nil, false
	}
	for _, cmd := range commands {
		if cmd.Name == primary {
			return &cmd, true
		}
	}
	return nil, false
}
