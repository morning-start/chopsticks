// Package cli 提供命令行界面功能。
package cli

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"chopsticks/core/app"
)

// Execute 使用给定的应用运行 CLI。
func Execute(ctx context.Context, application app.Application) error {
	// 获取命令行参数
	args := os.Args[1:]

	if len(args) < 1 {
		printUsage()
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	// 解析命令（支持别名）
	primaryCmd, ok := resolveCommand(command)
	if !ok {
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", command)
		printUsage()
		return fmt.Errorf("未知命令: %s", command)
	}

	// 如果使用的是别名，显示提示
	if command != primaryCmd {
		cmd, _ := getCommand(primaryCmd)
		if cmd != nil && isInternalAlias(command, cmd) {
			// 内部别名不显示提示
		}
	}

	switch primaryCmd {
	case "install":
		return ServeCommand(ctx, application, commandArgs)
	case "uninstall":
		return ClearCommand(ctx, application, commandArgs)
	case "update":
		return RefreshCommand(ctx, application, commandArgs)
	case "search":
		return SearchCommand(ctx, application, commandArgs)
	case "list":
		return ListCommand(ctx, application, commandArgs)
	case "bucket":
		return BucketCommand(ctx, application, commandArgs)
	case "completion":
		return CompletionCommand(ctx, application, commandArgs)
	case "help":
		if len(commandArgs) > 0 {
			printCommandHelp(commandArgs[0])
		} else {
			printUsage()
		}
		return nil
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", command)
		printUsage()
		return fmt.Errorf("未知命令: %s", command)
	}
}

// isInternalAlias 检查是否为内部别名（如 serve/clear/refresh/bow）。
func isInternalAlias(name string, _ *Command) bool {
	internalNames := []string{"serve", "clear", "refresh", "bucket"}
	return slices.Contains(internalNames, name)
}

// printUsage 打印使用帮助。
func printUsage() {
	fmt.Println("Chopsticks 包管理器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  chopsticks <command> [arguments]")
	fmt.Println()
	fmt.Println("命令:")

	// 按名称排序显示命令
	for _, cmd := range commands {
		if cmd.Name == "help" {
			continue // help 命令最后显示
		}
		aliases := ""
		if len(cmd.Aliases) > 0 {
			// 只显示非内部别名
			var userAliases []string
			for _, alias := range cmd.Aliases {
				if !isInternalAlias(alias, &cmd) {
					userAliases = append(userAliases, alias)
				}
			}
			if len(userAliases) > 0 {
				aliases = fmt.Sprintf(" (别名: %s)", strings.Join(userAliases, ", "))
			}
		}
		fmt.Printf("  %-12s %s%s\n", cmd.Name, cmd.Description, aliases)
	}

	fmt.Println("  help         显示帮助信息")
	fmt.Println()
	fmt.Println("使用 'chopsticks help <command>' 查看具体命令的帮助")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  chopsticks install git")
	fmt.Println("  chopsticks install nodejs@18.17.0")
	fmt.Println("  chopsticks uninstall git")
	fmt.Println("  chopsticks update --all")
	fmt.Println("  chopsticks search editor")
	fmt.Println("  chopsticks bucket add main https://github.com/chopsticks-bows/main")
}

// printCommandHelp 打印特定命令的帮助。
func printCommandHelp(name string) {
	cmd, ok := getCommand(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", name)
		return
	}

	fmt.Printf("命令: %s\n", cmd.Name)
	fmt.Printf("描述: %s\n", cmd.Description)
	if len(cmd.Aliases) > 0 {
		fmt.Printf("别名: %s\n", strings.Join(cmd.Aliases, ", "))
	}
	fmt.Println()
	fmt.Printf("用法:\n  chopsticks %s\n", cmd.Usage)
	fmt.Println()
	fmt.Println("示例:")
	for _, example := range cmd.Examples {
		fmt.Printf("  %s\n", example)
	}
}
