package cli

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// completionCommand 返回 completion 命令定义。
func completionCommand() *cli.Command {
	return &cli.Command{
		Name:      "completion",
		Usage:     "生成自动补全脚本",
		ArgsUsage: "<shell>",
		Description: `生成指定 shell 的自动补全脚本。

支持的 shell:
  bash       生成 Bash 补全脚本
  zsh        生成 Zsh 补全脚本
  powershell 生成 PowerShell 补全脚本
  fish       生成 Fish 补全脚本

示例:
  # Bash
  chopsticks completion bash > /etc/bash_completion.d/chopsticks
  chopsticks completion bash >> ~/.bashrc

  # Zsh
  chopsticks completion zsh > /usr/local/share/zsh/site-functions/_chopsticks
  chopsticks completion zsh >> ~/.zshrc

  # PowerShell
  chopsticks completion powershell > $PROFILE

  # Fish
  chopsticks completion fish > ~/.config/fish/completions/chopsticks.fish`,
		Action: completionAction,
	}
}

// completionAction 处理 completion 命令。
func completionAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少 shell 参数\n用法: chopsticks completion <shell>", 1)
	}

	shell := strings.ToLower(c.Args().First())

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
		return cli.Exit(fmt.Sprintf("不支持的 shell: %s\n支持的 shell: bash, zsh, powershell, fish", shell), 1)
	}
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
    local commands="install i uninstall remove rm update upgrade up search find s list ls bucket b help"

    # bucket 子命令
    local bucket_commands="init create c add a remove rm delete del list ls update up"

    # 全局选项
    local global_opts="--config -c --verbose -v --no-color --help -h --version -V"

    # 根据当前位置提供补全
    case "${COMP_CWORD}" in
        1)
            # 第一个参数：主命令
            COMPREPLY=( $(compgen -W "${commands} ${global_opts}" -- "${cur}") )
            ;;
        *)
            # 检查前一个参数
            case "${prev}" in
                install|i|uninstall|remove|rm|update|upgrade|up)
                    # 这些命令接受软件包名
                    COMPREPLY=( $(compgen -W "git nodejs python vscode --force -f --arch -a --purge -p --all" -- "${cur}") )
                    ;;
                search|find|s)
                    # 搜索命令
                    COMPREPLY=( $(compgen -W "--bucket -b" -- "${cur}") )
                    ;;
                list|ls)
                    # list 命令
                    COMPREPLY=( $(compgen -W "--installed -i" -- "${cur}") )
                    ;;
                bucket|b)
                    # bucket 子命令
                    COMPREPLY=( $(compgen -W "${bucket_commands}" -- "${cur}") )
                    ;;
                add|a)
                    # add 子命令后接软件源名称
                    COMPREPLY=( $(compgen -W "main extras" -- "${cur}") )
                    ;;
                --bucket|-b)
                    # --bucket 选项后接软件源名称
                    COMPREPLY=( $(compgen -W "main extras" -- "${cur}") )
                    ;;
                --arch|-a)
                    # --arch 选项后接架构
                    COMPREPLY=( $(compgen -W "amd64 x86 arm64" -- "${cur}") )
                    ;;
                *)
                    # 其他情况：提供常用选项
                    COMPREPLY=( $(compgen -W "--force -f --purge -p --all -a --help -h" -- "${cur}") )
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
        '(-c --config)'{-c,--config}'[指定配置文件路径]:config file:_files' \
        '(-v --verbose)'{-v,--verbose}'[启用详细输出]' \
        '--no-color[禁用彩色输出]' \
        '(-h --help)'{-h,--help}'[显示帮助信息]' \
        '(-V --version)'{-V,--version}'[显示版本信息]' \
        '1: :_chopsticks_commands' \
        '*:: :->args'

    case "$line[1]" in
        install|i)
            _arguments \
                '(-f --force)'{-f,--force}'[强制安装]' \
                '(-a --arch)'{-a,--arch}'[指定架构]:arch:(amd64 x86 arm64)' \
                '(-b --bucket)'{-b,--bucket}'[指定软件源]:bucket:_chopsticks_buckets' \
                '*:package:_chopsticks_packages'
            ;;
        uninstall|remove|rm)
            _arguments \
                '(-p --purge)'{-p,--purge}'[彻底清除]' \
                '*:package:_chopsticks_installed_packages'
            ;;
        update|upgrade|up)
            _arguments \
                '(-a --all)'{-a,--all}'[更新所有]' \
                '(-f --force)'{-f,--force}'[强制更新]' \
                '*:package:_chopsticks_installed_packages'
            ;;
        search|find|s)
            _arguments \
                '(-b --bucket)'{-b,--bucket}'[指定软件源]:bucket:_chopsticks_buckets' \
                '*:query:'
            ;;
        list|ls)
            _arguments \
                '(-i --installed)'{-i,--installed}'[显示已安装]'
            ;;
        bucket|b)
            _arguments \
                '1: :_chopsticks_bucket_commands' \
                '*:: :->bucket_args'
            case "$line[2]" in
                init)
                    _arguments \
                        '--js[使用 JavaScript 模板]' \
                        '--lua[使用 Lua 模板]' \
                        '--dir[指定目标目录]:directory:_directories' \
                        '1:name:'
                    ;;
                create|c)
                    _arguments \
                        '--dir[指定 Bucket 目录]:directory:_directories' \
                        '1:app-name:'
                    ;;
                add|a)
                    _arguments \
                        '--branch[指定分支]:branch:' \
                        '1:name:' \
                        '2:url:_urls'
                    ;;
                remove|rm|delete|del)
                    _arguments \
                        '(-p --purge)'{-p,--purge}'[删除本地数据]' \
                        '*:bucket:_chopsticks_buckets'
                    ;;
                update|up)
                    _arguments \
                        '*:bucket:_chopsticks_buckets'
                    ;;
            esac
            ;;
        completion)
            _arguments \
                '1:shell:(bash zsh powershell fish)'
            ;;
    esac
}

_chopsticks_commands() {
    local commands=(
        'install:安装软件包'
        'i:安装软件包（别名）'
        'uninstall:卸载软件包'
        'remove:卸载软件包（别名）'
        'rm:卸载软件包（别名）'
        'update:更新软件包'
        'upgrade:更新软件包（别名）'
        'up:更新软件包（别名）'
        'search:搜索软件包'
        'find:搜索软件包（别名）'
        's:搜索软件包（别名）'
        'list:列出软件包'
        'ls:列出软件包（别名）'
        'bucket:软件源管理'
        'b:软件源管理（别名）'
        'completion:生成自动补全脚本'
        'help:显示帮助信息'
    )
    _describe -t commands 'chopsticks command' commands "$@"
}

_chopsticks_bucket_commands() {
    local commands=(
        'init:初始化新 Bucket'
        'create:创建新 App 模板'
        'c:创建新 App 模板（别名）'
        'add:添加软件源'
        'a:添加软件源（别名）'
        'remove:删除软件源'
        'rm:删除软件源（别名）'
        'delete:删除软件源（别名）'
        'del:删除软件源（别名）'
        'list:列出软件源'
        'ls:列出软件源（别名）'
        'update:更新软件源'
        'up:更新软件源（别名）'
    )
    _describe -t commands 'bucket subcommand' commands "$@"
}

_chopsticks_packages() {
    # 这里可以从软件源获取可用软件包列表
    local packages=(git nodejs python vscode 7zip notepadplusplus)
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
        @{ Name = 'install'; Aliases = @('i') }
        @{ Name = 'uninstall'; Aliases = @('remove', 'rm') }
        @{ Name = 'update'; Aliases = @('upgrade', 'up') }
        @{ Name = 'search'; Aliases = @('find', 's') }
        @{ Name = 'list'; Aliases = @('ls') }
        @{ Name = 'bucket'; Aliases = @('b') }
        @{ Name = 'completion'; Aliases = @() }
        @{ Name = 'help'; Aliases = @('--help', '-h') }
    )

    # bucket 子命令
    $bucketCommands = @(
        @{ Name = 'init'; Aliases = @() }
        @{ Name = 'create'; Aliases = @('c') }
        @{ Name = 'add'; Aliases = @('a') }
        @{ Name = 'remove'; Aliases = @('rm', 'delete', 'del') }
        @{ Name = 'list'; Aliases = @('ls') }
        @{ Name = 'update'; Aliases = @('up') }
    )

    # 常用软件包
    $packages = @('git', 'nodejs', 'python', 'vscode', '7zip', 'notepadplusplus')

    # 软件源
    $buckets = @('main', 'extras')

    # 全局选项
    $globalOpts = @('--config', '-c', '--verbose', '-v', '--no-color', '--help', '-h', '--version', '-V')

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
        # 全局选项
        $globalOpts | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
    }
    else {
        switch -Wildcard ($command) {
            'install|i' {
                if ($wordToComplete -match '^--') {
                    @('--force', '-f', '--arch', '-a', '--bucket', '-b') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
                else {
                    $packages | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                    }
                }
            }
            'uninstall|remove|rm' {
                if ($wordToComplete -match '^--') {
                    @('--purge', '-p') | ForEach-Object {
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
            'update|upgrade|up' {
                if ($wordToComplete -match '^--') {
                    @('--all', '-a', '--force', '-f') | ForEach-Object {
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
                    @('--bucket', '-b') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
            }
            'list|ls' {
                if ($wordToComplete -match '^--') {
                    @('--installed', '-i') | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                    }
                }
            }
            'bucket|b' {
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
                                @('--purge', '-p') | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
                                }
                            }
                            else {
                                $buckets | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                                }
                            }
                        }
                        'update|up' {
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

# 禁用文件补全
complete -c chopsticks -f

# 全局选项
complete -c chopsticks -n '__fish_is_first_arg' -l config -s c -d '指定配置文件路径' -r
complete -c chopsticks -n '__fish_is_first_arg' -l verbose -s v -d '启用详细输出'
complete -c chopsticks -n '__fish_is_first_arg' -l no-color -d '禁用彩色输出'
complete -c chopsticks -n '__fish_is_first_arg' -l help -s h -d '显示帮助信息'
complete -c chopsticks -n '__fish_is_first_arg' -l version -s V -d '显示版本信息'

# 主命令
complete -c chopsticks -n '__fish_use_subcommand' -a 'install' -d '安装软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'i' -d '安装软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'uninstall' -d '卸载软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'remove' -d '卸载软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'rm' -d '卸载软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'update' -d '更新软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'upgrade' -d '更新软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'up' -d '更新软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'search' -d '搜索软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'find' -d '搜索软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 's' -d '搜索软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'list' -d '列出软件包'
complete -c chopsticks -n '__fish_use_subcommand' -a 'ls' -d '列出软件包（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'bucket' -d '软件源管理'
complete -c chopsticks -n '__fish_use_subcommand' -a 'b' -d '软件源管理（别名）'
complete -c chopsticks -n '__fish_use_subcommand' -a 'completion' -d '生成自动补全脚本'
complete -c chopsticks -n '__fish_use_subcommand' -a 'help' -d '显示帮助信息'

# install 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from install i' -l force -s f -d '强制安装'
complete -c chopsticks -n '__fish_seen_subcommand_from install i' -l arch -s a -d '指定架构' -a 'amd64 x86 arm64'
complete -c chopsticks -n '__fish_seen_subcommand_from install i' -l bucket -s b -d '指定软件源' -a 'main extras'

# uninstall 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from uninstall remove rm' -l purge -s p -d '彻底清除'

# update 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from update upgrade up' -l all -s a -d '更新所有'
complete -c chopsticks -n '__fish_seen_subcommand_from update upgrade up' -l force -s f -d '强制更新'

# search 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from search find s' -l bucket -s b -d '指定软件源' -a 'main extras'

# list 命令选项
complete -c chopsticks -n '__fish_seen_subcommand_from list ls' -l installed -s i -d '显示已安装'

# bucket 子命令
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'init' -d '初始化新 Bucket'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'create' -d '创建新 App 模板'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'c' -d '创建新 App 模板（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'add' -d '添加软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'a' -d '添加软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'remove' -d '删除软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'rm' -d '删除软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'delete' -d '删除软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'del' -d '删除软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'list' -d '列出软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'ls' -d '列出软件源（别名）'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'update' -d '更新软件源'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b' -a 'up' -d '更新软件源（别名）'

# bucket init 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from init' -l js -d '使用 JavaScript 模板'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from init' -l lua -d '使用 Lua 模板'
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from init' -l dir -d '指定目标目录' -r

# bucket create 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from create c' -l dir -d '指定 Bucket 目录' -r

# bucket add 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from add a' -l branch -d '指定分支'

# bucket remove 选项
complete -c chopsticks -n '__fish_seen_subcommand_from bucket b; and __fish_seen_subcommand_from remove rm delete del' -l purge -s p -d '删除本地数据'

# completion 命令
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'bash' -d 'Bash 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'zsh' -d 'Zsh 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'powershell' -d 'PowerShell 补全脚本'
complete -c chopsticks -n '__fish_seen_subcommand_from completion' -a 'fish' -d 'Fish 补全脚本'
`
	fmt.Print(script)
	return nil
}
