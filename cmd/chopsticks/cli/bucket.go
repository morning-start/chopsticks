package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/bucket"

	"github.com/urfave/cli/v2"
)

// bucketCommand 返回 bucket 命令定义。
func bucketCommand() *cli.Command {
	return &cli.Command{
		Name:    "bucket",
		Aliases: []string{"b"},
		Usage:   "软件源管理",
		Description: `管理软件源（Bucket）。

软件源是存储软件包配置的 Git 仓库。`,
		Subcommands: []*cli.Command{
			bucketInitCommand(),
			bucketCreateCommand(),
			bucketAddCommand(),
			bucketRemoveCommand(),
			bucketListCommand(),
			bucketUpdateCommand(),
		},
	}
}

// bucketInitCommand 返回 bucket init 子命令。
func bucketInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "初始化新 Bucket 目录结构",
		ArgsUsage: "<name>",
		Description: `初始化一个新的 Bucket 目录结构。

示例:
  chopsticks bucket init my-bucket
  chopsticks bucket init my-bucket --js
  chopsticks bucket init my-bucket --lua --dir ./buckets`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "js",
				Usage: "使用 JavaScript 模板",
			},
			&cli.BoolFlag{
				Name:  "lua",
				Usage: "使用 Lua 模板",
			},
			&cli.StringFlag{
				Name:  "dir",
				Usage: "指定目标目录",
			},
		},
		Action: bucketInitAction,
	}
}

// bucketCreateCommand 返回 bucket create 子命令。
func bucketCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Aliases:   []string{"c"},
		Usage:     "创建新 App 模板",
		ArgsUsage: "<app-name>",
		Description: `在 Bucket 中创建一个新的 App 模板。

示例:
  chopsticks bucket create my-app
  chopsticks bucket create my-app --dir ./bucket`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "指定 Bucket 目录",
			},
		},
		Action: bucketCreateAction,
	}
}

// bucketAddCommand 返回 bucket add 子命令。
func bucketAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Aliases:   []string{"a"},
		Usage:     "添加软件源",
		ArgsUsage: "<name> <url>",
		Description: `添加一个新的软件源。

示例:
  chopsticks bucket add main https://github.com/chopsticks-bucket/main
  chopsticks bucket add extras https://github.com/chopsticks-bucket/extras --branch develop`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "branch",
				Usage: "指定 Git 分支",
			},
		},
		Action: bucketAddAction,
	}
}

// bucketRemoveCommand 返回 bucket remove 子命令。
func bucketRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm", "delete", "del"},
		Usage:     "删除软件源",
		ArgsUsage: "<name>",
		Description: `删除一个软件源。

示例:
  chopsticks bucket remove main
  chopsticks bucket rm extras --purge`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "purge",
				Aliases: []string{"p"},
				Usage:   "彻底删除本地数据",
			},
		},
		Action: bucketRemoveAction,
	}
}

// bucketListCommand 返回 bucket list 子命令。
func bucketListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Usage:     "列出软件源",
		ArgsUsage: " ",
		Description: `列出所有已添加的软件源。

示例:
  chopsticks bucket list
  chopsticks bucket ls`,
		Action: bucketListAction,
	}
}

// bucketUpdateCommand 返回 bucket update 子命令。
func bucketUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Aliases:   []string{"up"},
		Usage:     "更新软件源",
		ArgsUsage: "[name]",
		Description: `更新软件源。如果不指定名称，则更新所有软件源。

示例:
  chopsticks bucket update
  chopsticks bucket update main
  chopsticks bucket up extras`,
		Action: bucketUpdateAction,
	}
}

// bucketInitAction 处理 bucket init 命令。
func bucketInitAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少 Bucket 名称\n用法: chopsticks bucket init <name>", 1)
	}

	name := c.Args().First()
	useLua := c.Bool("lua")
	targetDir := c.String("dir")

	// 默认使用 JS 模板
	templateType := "js"
	if useLua {
		templateType = "lua"
	}

	if targetDir == "" {
		targetDir = name
	}

	fmt.Printf("初始化 Bucket: %s (%s)\n", name, map[string]string{"js": "JavaScript", "lua": "Lua"}[templateType])

	// 这里应该调用实际的模板复制逻辑
	// 暂时使用简化的实现
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return cli.Exit(fmt.Sprintf("创建目录失败: %v", err), 1)
	}

	fmt.Printf("✓ Bucket %s 初始化完成\n", name)
	fmt.Printf("\n下一步:\n")
	fmt.Printf("  cd %s\n", targetDir)
	if templateType != "lua" {
		fmt.Printf("  npm install\n")
	}
	fmt.Printf("  chopsticks bucket create my-app\n")
	return nil
}

// bucketCreateAction 处理 bucket create 命令。
func bucketCreateAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少应用名称\n用法: chopsticks bucket create <app-name>", 1)
	}

	name := c.Args().First()
	bucketDir := c.String("dir")

	targetDir := bucketDir
	if targetDir == "" {
		targetDir = "./apps/" + name
	} else {
		targetDir = filepath.Join(bucketDir, "apps", name)
	}

	dirs := []string{
		targetDir,
		filepath.Join(targetDir, "scripts"),
		filepath.Join(targetDir, "tests"),
	}

	fmt.Printf("创建 App 模板: %s\n", name)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return cli.Exit(fmt.Sprintf("创建目录失败: %v", err), 1)
		}
		fmt.Printf("  创建目录: %s\n", dir)
	}

	// 创建 manifest.yaml
	manifestContent := fmt.Sprintf(`name: %s
version: 1.0.0
description: A software package
author: unknown
homepage: https://example.com
license: MIT

arch: amd64

format: zip

install:
  # installer: setup.exe
  # install_args: [/S]

hooks:
  pre_download: |
    console.log('准备下载...')
  post_download: |
    console.log('下载完成')
  pre_extract: |
    console.log('准备解压...')
  post_extract: |
    console.log('解压完成')
  pre_install: |
    console.log('准备安装...')
  post_install: |
    console.log('安装完成')
  pre_uninstall: |
    console.log('准备卸载...')
  post_uninstall: |
    console.log('卸载完成')

files:
  - %s*

shortcuts:
  - name: %s
    path: %s.exe

registry:
  - key: SOFTWARE\\%s
    values:
      - name: InstallPath
        type: REG_SZ
        value: {{.InstallDir}}

env:
  - name: %s_HOME
    value: {{.InstallDir}}
`, name, name, name, name, name, strings.ToUpper(name))

	manifestPath := filepath.Join(targetDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return cli.Exit(fmt.Sprintf("创建清单文件失败: %v", err), 1)
	}
	fmt.Printf("  创建文件: %s\n", manifestPath)

	// 创建 README.md
	readmeContent := fmt.Sprintf(`# %s

## 描述

软件包 %s 的说明文档。

## 安装

`+"```bash\n"+`chopsticks install %s
`+"```\n\n"+`## 卸载

`+"```bash\n"+`chopsticks uninstall %s
`+"```\n", name, name, name, name)

	readmePath := filepath.Join(targetDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return cli.Exit(fmt.Sprintf("创建说明文件失败: %v", err), 1)
	}
	fmt.Printf("  创建文件: %s\n", readmePath)

	// 创建安装脚本
	scriptContent := fmt.Sprintf(`// %s 安装脚本

function preDownload() {
    console.log("准备下载 " + name + " " + version);
}

function postDownload() {
    console.log("下载完成");
}

function preExtract() {
    console.log("准备解压");
}

function postExtract() {
    console.log("解压完成");
}

function preInstall() {
    console.log("准备安装");
}

function postInstall() {
    console.log("安装完成");
}

function preUninstall() {
    console.log("准备卸载");
}

function postUninstall() {
    console.log("卸载完成");
}
`, name)

	scriptPath := filepath.Join(targetDir, "scripts", "install.js")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return cli.Exit(fmt.Sprintf("创建脚本文件失败: %v", err), 1)
	}
	fmt.Printf("  创建文件: %s\n", scriptPath)

	fmt.Printf("✓ App %s 模板创建完成\n", name)
	fmt.Printf("\n下一步:\n")
	fmt.Printf("  编辑 manifest.yaml 添加软件包信息\n")
	fmt.Printf("  编辑 scripts/install.js 添加安装逻辑\n")
	return nil
}

// bucketAddAction 处理 bucket add 命令。
func bucketAddAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit("错误: 缺少参数\n用法: chopsticks bucket add <name> <url>", 1)
	}

	name := c.Args().Get(0)
	url := c.Args().Get(1)
	branch := c.String("branch")

	opts := bucket.AddOptions{
		Branch: branch,
	}

	fmt.Printf("添加软件源: %s\n", name)
	fmt.Printf("  URL: %s\n", url)
	if opts.Branch != "" {
		fmt.Printf("  分支: %s\n", opts.Branch)
	}
	fmt.Println()

	ctx := getContext(c)
	application := getApp()

	if err := application.BucketManager().Add(ctx, name, url, opts); err != nil {
		return cli.Exit(fmt.Sprintf("添加软件源失败: %v", err), 1)
	}

	fmt.Printf("✓ 软件源 %s 添加成功\n", name)
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("  chopsticks search <应用名>\n")
	return nil
}

// bucketRemoveAction 处理 bucket remove 命令。
func bucketRemoveAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("错误: 缺少软件源名称\n用法: chopsticks bucket remove <name>", 1)
	}

	name := c.Args().First()
	purge := c.Bool("purge")

	fmt.Printf("删除软件源: %s\n", name)
	if purge {
		fmt.Println("  模式: 完全删除（包括本地文件）")
	}
	fmt.Println()

	ctx := getContext(c)
	application := getApp()

	if err := application.BucketManager().Remove(ctx, name, purge); err != nil {
		return cli.Exit(fmt.Sprintf("删除软件源失败: %v", err), 1)
	}

	fmt.Printf("✓ 软件源 %s 已删除\n", name)
	return nil
}

// bucketListAction 处理 bucket list 命令。
func bucketListAction(c *cli.Context) error {
	ctx := getContext(c)
	application := getApp()

	// 获取所有软件源
	buckets, err := application.BucketManager().ListBuckets(ctx)
	if err != nil {
		return cli.Exit(fmt.Sprintf("获取软件源列表失败: %v", err), 1)
	}

	fmt.Println("已添加的软件源:")
	fmt.Println("--------------")

	if len(buckets) == 0 {
		fmt.Println("  (暂无软件源)")
	} else {
		for _, name := range buckets {
			// 获取详细信息
			b, err := application.BucketManager().GetBucket(ctx, name)
			if err != nil {
				fmt.Printf("  %-10s (无法获取详细信息)\n", name)
				continue
			}
			url := b.Repository.URL
			if url == "" {
				url = "本地"
			}
			fmt.Printf("  %-10s %s\n", name, url)
		}
	}

	fmt.Println()
	fmt.Println("使用 'chopsticks bucket add <name> <url>' 添加更多软件源")
	return nil
}

// bucketUpdateAction 处理 bucket update 命令。
func bucketUpdateAction(c *cli.Context) error {
	ctx := getContext(c)
	application := getApp()

	if c.NArg() == 0 {
		// 更新所有软件源
		fmt.Println("更新所有软件源...")
		fmt.Println()

		if err := application.BucketManager().UpdateAll(ctx); err != nil {
			return cli.Exit(fmt.Sprintf("更新软件源失败: %v", err), 1)
		}

		fmt.Println("✓ 所有软件源更新成功")
		return nil
	}

	// 更新指定软件源
	name := c.Args().First()
	fmt.Printf("更新软件源: %s...\n", name)
	fmt.Println()

	if err := application.BucketManager().Update(ctx, name); err != nil {
		return cli.Exit(fmt.Sprintf("更新软件源失败: %v", err), 1)
	}

	fmt.Printf("✓ 软件源 %s 更新成功\n", name)
	return nil
}
