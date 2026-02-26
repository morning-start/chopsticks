// Package cli 提供命令行界面功能。
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"chopsticks/core/app"
)

// CompletionCommand 处理自动补全命令。
func CompletionCommand(ctx context.Context, application app.Application, args []string) error {
	if len(args) < 1 {
		printCompletionUsage()
		return nil
	}

	shell := strings.ToLower(args[0])

	switch shell {
	case "bash":
		return generateBashCompletion()
	case "zsh":
		return generateZshCompletion()
	case "powershell", "pwsh":
		return generatePowerShellCompletion()
	case "fish":
		return generateFishCompletion()
	default:
		fmt.Fprintf(os.Stderr, "不支持的 shell: %s\n\n", shell)
		printCompletionUsage()
		return fmt.Errorf("不支持的 shell: %s", shell)
	}
}

// printCompletionUsage 打印补全命令用法。
func printCompletionUsage() {
	fmt.Println("生成 shell 自动补全脚本")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  chopsticks completion <shell>")
	fmt.Println()
	fmt.Println("支持的 shell:")
	fmt.Println("  bash       生成 Bash 补全脚本")
	fmt.Println("  zsh        生成 Zsh 补全脚本")
	fmt.Println("  powershell 生成 PowerShell 补全脚本")
	fmt.Println("  fish       生成 Fish 补全脚本")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # Bash")
	fmt.Println("  chopsticks completion bash > /etc/bash_completion.d/chopsticks")
	fmt.Println("  # 或")
	fmt.Println("  chopsticks completion bash >> ~/.bashrc")
	fmt.Println()
	fmt.Println("  # Zsh")
	fmt.Println("  chopsticks completion zsh > /usr/local/share/zsh/site-functions/_chopsticks")
	fmt.Println("  # 或")
	fmt.Println("  chopsticks completion zsh >> ~/.zshrc")
	fmt.Println()
	fmt.Println("  # PowerShell")
	fmt.Println("  chopsticks completion powershell > $PROFILE")
	fmt.Println()
	fmt.Println("  # Fish")
	fmt.Println("  chopsticks completion fish > ~/.config/fish/completions/chopsticks.fish")
}

// generateBashCompletion 生成 Bash 补全脚本。
func generateBashCompletion() error {
	script := `# Chopsticks Bash 自动补全脚本
# 将此脚本保存到 /etc/bash_completion.d/chopsticks 或添加到 ~/.bashrc

_chopsticks_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # 主命令列表
    local commands="install serve i uninstall clear remove rm update refresh upgrade up search find s list ls bucket bow b help"

    # bucket 子命令
    local bucket_commands="add a remove rm delete del list ls update up upgrade"

    # 全局选项
    local global_opts="--help -h"

    # 根据当前位置提供补全
    case "${COMP_CWORD}" in
        1)
            # 第一个参数：主命令
            COMPREPLY=( $(compgen -W "${commands} ${global_opts}" -- "${cur}") )
            ;;
        *)
            # 检查前一个参数
            case "${prev}" in
                install|serve|i|uninstall|clear|remove|rm|update|refresh|upgrade|up)
                    # 这些命令接受软件包名
                    COMPREPLY=( $(compgen -W "git nodejs python vscode --force --purge --all" -- "${cur}") )
                    ;;
                search|find|s)
                    # 搜索命令
                    COMPREPLY=( $(compgen -W "--bow" -- "${cur}") )
                    ;;
                bucket|bow|b)
                    # bucket 子命令
                    COMPREPLY=( $(compgen -W "${bucket_commands}" -- "${cur}") )
                    ;;
                add|a)
                    # add 子命令后接软件源名称
                    COMPREPLY=( $(compgen -W "main extras" -- "${cur}") )
                    ;;
                --bow)
                    # --bow 选项后接软件源名称
                    COMPREPLY=( $(compgen -W "main extras" -- "${cur}") )
                    ;;
                *)
                    # 其他情况：提供常用选项
                    COMPREPLY=( $(compgen -W "--force --purge --all --help" -- "${cur}") )
                    ;;
            esac
            ;;
    esac
}

# 注册补全函数
complete -F _chopsticks_completions chopsticks
`
	fmt.Print(script)
	return nil
}

// generateZshCompletion 生成 Zsh 补全脚本。
func generateZshCompletion() error {
	script := `#compdef chopsticks

# Chopsticks Zsh 自动补全脚本
# 将此脚本保存到 /usr/local/share/zsh/site-functions/_chopsticks

_chopsticks() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '1: :_chopsticks_commands' \
        '*:: :->args'

    case "$line[1]" in
        install|serve|i)
            _arguments \
                '--force[强制安装]' \
                '*:package:_chopsticks_packages'
            ;;
        uninstall|clear|remove|rm)
            _arguments \
                '--purge[彻底清除]' \
                '*:package:_chopsticks_installed_packages'
            ;;
        update|refresh|upgrade|up)
            _arguments \
                '--all[更新所有]' \
                '--force[强制更新]' \
                '*:package:_chopsticks_installed_packages'
            ;;
        search|find|s)
            _arguments \
                '--bow[指定软件源]:bucket:_chopsticks_buckets' \
                '*:query:'
            ;;
        list|ls)
            _arguments \
                '--installed[显示已安装]'
            ;;
        bucket|bow|b)
            _arguments \
                '1: :_chopsticks_bucket_commands' \
                '*:: :->bucket_args'
            case "$line[2]" in
                add|a)
                    _arguments \
                        '--branch[指定分支]:branch:' \
                        '1:name:' \
                        '2:url:_urls'
                    ;;
                remove|rm|delete|del)
                    _arguments \
                        '--purge[删除本地数据]' \
                        '*:bucket:_chopsticks_buckets'
                    ;;
                update|up|upgrade)
                    _arguments \
                        '*:bucket:_chopsticks_buckets'
                    ;;
            esac
            ;;
    esac
}

_chopsticks_commands() {
    local commands=(
        'install:安装软件包'
        'uninstall:卸载软件包'
        'update:更新软件包'
        'search:搜索软件包'
        'list:列出软件包'
        'bucket:软件源管理'
        'help:显示帮助信息'
    )
    _describe -t commands 'chopsticks command' commands "$@"
}

_chopsticks_bucket_commands() {
    local commands=(
        'add:添加软件源'
        'remove:删除软件源'
        'list:列出软件源'
        'update:更新软件源'
    )
    _describe -t commands 'bucket subcommand' commands "$@"
}

_chopsticks_packages() {
    # 这里可以从软件源获取可用软件包列表
    local packages=(git nodejs python vscode)
    _describe -t packages 'package' packages "$@"
}

_chopsticks_installed_packages() {
    # 这里可以从数据库获取已安装软件包列表
    local packages=(git nodejs)
    _describe -t packages 'installed package' packages "$@"
}

_chopsticks_buckets() {
    local buckets=(main extras)
    _describe -t buckets 'bucket' buckets "$@"
}

compdef _chopsticks chopsticks
`
	fmt.Print(script)
	return nil
}

// generatePowerShellCompletion 生成 PowerShell 补全脚本。
func generatePowerShellCompletion() error {
	script := `# Chopsticks PowerShell 自动补全脚本
# 将此脚本添加到 $PROFILE

# 注册参数补全器
Register-ArgumentCompleter -Native -CommandName chopsticks -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # 解析命令行
    $line = $commandAst.ToString()
    $tokens = $line -split '\s+'
    $command = $tokens[1]
    $subcommand = $tokens[2]

    # 主命令列表
    $commands = @(
        @{ Name = 'install'; Aliases = @('serve', 'i') }
        @{ Name = 'uninstall'; Aliases = @('clear', 'remove', 'rm') }
        @{ Name = 'update'; Aliases = @('refresh', 'upgrade', 'up') }
        @{ Name = 'search'; Aliases = @('find', 's') }
        @{ Name = 'list'; Aliases = @('ls') }
        @{ Name = 'bucket'; Aliases = @('bow', 'b') }
        @{ Name = 'help'; Aliases = @('--help', '-h') }
        @{ Name = 'completion'; Aliases = @() }
    )

    # bucket 子命令
    $bucketCommands = @(
        @{ Name = 'add'; Aliases = @('a') }
        @{ Name = 'remove'; Aliases = @('rm', 'delete', 'del') }
        @{ Name = 'list'; Aliases = @('ls') }
        @{ Name = 'update'; Aliases = @('up', 'upgrade') }
    )

    # 常用软件包
    $packages = @('git', 'nodejs', 'python', 'vscode', '7zip', 'notepadplusplus')

    # 软件源
    $buckets = @('main', 'extras')

    # 根据位置提供补全
    if ($tokens.Count -eq 1 -or ($tokens.Count -eq 2 -and $wordToComplete -ne '')) {
        # 第一个参数：主命令
        $commands | ForEach-Object {
            $cmd = $_
            [System.Management.Automation.CompletionResult]::new(
                $cmd.Name,
                $cmd.Name,
                'ParameterValue',
                "$($cmd.Name) command"
            )
            $cmd.Aliases | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    "$($cmd.Name) alias"
                )
            }
        }
    }
    else {
        switch -Wildcard ($command) {
            'install|serve|i' {
                if ($wordToComplete -match '^--') {
                    @('--force', '--arch') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
                else {
                    $packages | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                    }
                }
            }
            'uninstall|clear|remove|rm' {
                if ($wordToComplete -match '^--') {
                    @('--purge') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
                else {
                    # 已安装的软件包
                    $packages | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                    }
                }
            }
            'update|refresh|upgrade|up' {
                if ($wordToComplete -match '^--') {
                    @('--all', '--force') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
                else {
                    $packages | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                    }
                }
            }
            'search|find|s' {
                if ($wordToComplete -match '^--') {
                    @('--bow') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
            }
            'list|ls' {
                if ($wordToComplete -match '^--') {
                    @('--installed') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
            }
            'bucket|bow|b' {
                if ($tokens.Count -eq 2 -or ($tokens.Count -eq 3 -and $wordToComplete -ne '')) {
                    $bucketCommands | ForEach-Object {
                        $cmd = $_
                        [System.Management.Automation.CompletionResult]::new(
                            $cmd.Name,
                            $cmd.Name,
                            'ParameterValue',
                            "$($cmd.Name) subcommand"
                        )
                        $cmd.Aliases | ForEach-Object {
                            [System.Management.Automation.CompletionResult]::new(
                                $_,
                                $_,
                                'ParameterValue',
                                "$($cmd.Name) alias"
                            )
                        }
                    }
                }
                else {
                    # bucket 子命令的参数
                    switch -Wildcard ($subcommand) {
                        'add|a' {
                            if ($wordToComplete -match '^--') {
                                @('--branch') | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                                }
                            }
                            else {
                                $buckets | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                                }
                            }
                        }
                        'remove|rm|delete|del' {
                            if ($wordToComplete -match '^--') {
                                @('--purge') | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                                }
                            }
                            else {
                                $buckets | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                                }
                            }
                        }
                        'update|up|upgrade' {
                            $buckets | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                            }
                        }
                    }
                }
            }
            'completion' {
                @('bash', 'zsh', 'powershell', 'pwsh', 'fish') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                }
            }
        }
    }
}

Write-Host "Chopsticks PowerShell 自动补全已启用" -ForegroundColor Green
`
	fmt.Print(script)
	return nil
}

// generateFishCompletion 生成 Fish 补全脚本。
func generateFishCompletion() error {
	script := `# Chopsticks Fish 自动补全脚本
# 将此脚本保存到 ~/.config/fish/completions/chopsticks.fish

# 主命令补全
complete -c chopsticks -f

# 主命令
complete -c chopsticks -n '__fish_use_subcommand' -a 'install' -d '安装软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'serve' -d '安装软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'i' -d '安装软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'uninstall' -d '卸载软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'clear' -d '卸载软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'remove' -d '卸载软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'rm' -d '卸载软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'update' -d '更新软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'refresh' -d '更新软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'upgrade' -d '更新软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'up' -d '更新软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'search' -d '搜索软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'find' -d '搜索软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 's' -d '搜索软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'list' -d '列出软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'ls' -d '列出软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'bucket' -d '软件源管理'
complete -c chopsticks -n '__fish_use_subcommand' -a 'bow' -d '软件源管理（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'b' -d '软件源管理（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'help' -d '显示帮助信息'
complete -c chopsticks -n '__fish_use_subcommand' -a 'completion' -d '生成自动补全脚本'

# install 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from install serve i' -l force -d '强制安装'
complete -c chopsticks -n '__fish_seen_subcommand_from install serve i' -l arch -d '指定架构' -a 'amd64 x86 arm64'

# uninstall 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from uninstall clear remove rm' -l purge -d '彻底清除'

# update 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from update refresh upgrade up' -l all -d '更新所有'
complete -c chopsticks -n '__fish_seen_subcommand_from update refresh upgrade up' -l force -d '强制更新'

# search 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from search find s' -l bow -d '指定软件源' -a 'main extras'

# list 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from list ls' -l installed -d '显示已安装'

# bucket 子命令
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'add' -d '添加软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'a' -d '添加软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'remove' -d '删除软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'rm' -d '删除软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'list' -d '列出软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'ls' -d '列出软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'update' -d '更新软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b' -a 'up' -d '更新软件源（别名）'

# bucket add 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b; and __fish_seen_subcommand_from add a' -l branch -d '指定分支'

# bucket remove 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket bow b; and __fish_seen_subcommand_from remove rm delete del' -l purge -d '删除本地数据'

# completion 命令
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'bash' -d 'Bash 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'zsh' -d 'Zsh 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'powershell' -d 'PowerShell 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'fish' -d 'Fish 补全脚本'
`
	fmt.Print(script)
	return nil
}
