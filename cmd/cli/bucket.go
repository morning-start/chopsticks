package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/bucket"
	"chopsticks/pkg/output"

	"github.com/spf13/cobra"
)

var (
	bucketInitJS     bool
	bucketInitLua    bool
	bucketInitDir    string
	bucketCreateDir  string
	bucketAddBranch  string
	bucketRemovePurge bool
)

// bucketCmd 表示 bucket 命令
var bucketCmd = &cobra.Command{
	Use:     "bucket",
	Aliases: []string{"b"},
	Short:   "软件源管理",
	Long: `管理软件源（Bucket）。

软件源是存储软件包配置的 Git 仓库。`,
}

// bucketInitCmd 初始化 Bucket
var bucketInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "初始化新 Bucket 目录结构",
	Long: `初始化一个新的 Bucket 目录结构。

示例:
  chopsticks bucket init my-bucket
  chopsticks bucket init my-bucket --js
  chopsticks bucket init my-bucket --lua --dir ./buckets`,
	Args: cobra.ExactArgs(1),
	RunE: runBucketInit,
}

// bucketCreateCmd 创建 App 模板
var bucketCreateCmd = &cobra.Command{
	Use:     "create <app-name>",
	Aliases: []string{"c"},
	Short:   "创建新 App 模板",
	Long: `在 Bucket 中创建一个新的 App 模板。

示例:
  chopsticks bucket create my-app
  chopsticks bucket create my-app --dir ./bucket`,
	Args: cobra.ExactArgs(1),
	RunE: runBucketCreate,
}

// bucketAddCmd 添加软件源
var bucketAddCmd = &cobra.Command{
	Use:     "add <name> <url>",
	Aliases: []string{"a"},
	Short:   "添加软件源",
	Long: `添加一个新的软件源。

示例:
  chopsticks bucket add main https://github.com/chopsticks-bucket/main
  chopsticks bucket add extras https://github.com/chopsticks-bucket/extras --branch develop`,
	Args: cobra.ExactArgs(2),
	RunE: runBucketAdd,
}

// bucketRemoveCmd 删除软件源
var bucketRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "删除软件源",
	Long: `删除一个软件源。

示例:
  chopsticks bucket remove main
  chopsticks bucket rm extras --purge`,
	Args: cobra.ExactArgs(1),
	RunE: runBucketRemove,
}

// bucketListCmd 列出软件源
var bucketListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出软件源",
	Long: `列出所有已添加的软件源。

示例:
  chopsticks bucket list
  chopsticks bucket ls`,
	RunE: runBucketList,
}

// bucketUpdateCmd 更新软件源
var bucketUpdateCmd = &cobra.Command{
	Use:     "update [name]",
	Aliases: []string{"up"},
	Short:   "更新软件源",
	Long: `更新软件源。如果不指定名称，则更新所有软件源。

示例:
  chopsticks bucket update
  chopsticks bucket update main
  chopsticks bucket up extras`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBucketUpdate,
}

func runBucketInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	useLua := bucketInitLua
	targetDir := bucketInitDir

	// 默认使用 JS 模板
	templateType := "js"
	if useLua {
		templateType = "lua"
	}

	if targetDir == "" {
		targetDir = name
	}

	output.Infof("Initializing Bucket: ")
	output.Highlight("%s", name)
	output.Dimf(" (%s)\n", map[string]string{"js": "JavaScript", "lua": "Lua"}[templateType])

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		output.ErrorCrossf("Failed to create directory: %v", err)
		return err
	}

	output.SuccessCheckf("Bucket %s initialized", name)
	output.Highlightln("\nNext steps:")
	output.Dimf("  cd %s\n", targetDir)
	if templateType != "lua" {
		output.Dimln("  npm install")
	}
	output.Dimln("  chopsticks bucket create my-app")
	return nil
}

func runBucketCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	bucketDir := bucketCreateDir

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

	output.Infof("Creating App template: ")
	output.Highlightln(name)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			output.ErrorCrossf("Failed to create directory: %v", err)
			return err
		}
		output.Dimf("  Created directory: %s\n", dir)
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
		output.ErrorCrossf("Failed to create manifest file: %v", err)
		return err
	}
	output.Dimf("  Created file: %s\n", manifestPath)

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
		output.ErrorCrossf("Failed to create README file: %v", err)
		return err
	}
	output.Dimf("  Created file: %s\n", readmePath)

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
		output.ErrorCrossf("Failed to create script file: %v", err)
		return err
	}
	output.Dimf("  Created file: %s\n", scriptPath)

	output.SuccessCheckf("App %s template created", name)
	output.Highlightln("\nNext steps:")
	output.Dimln("  Edit manifest.yaml to add package info")
	output.Dimln("  Edit scripts/install.js to add installation logic")
	return nil
}

func runBucketAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	url := args[1]
	branch := bucketAddBranch

	opts := bucket.AddOptions{
		Branch: branch,
	}

	output.Infof("Adding bucket: ")
	output.Highlightln(name)
	output.Dimf("  URL: %s\n", url)
	if opts.Branch != "" {
		output.Dimf("  Branch: %s\n", opts.Branch)
	}
	fmt.Println()

	ctx := cmd.Context()
	application := getApp()

	if err := application.BucketManager().Add(ctx, name, url, opts); err != nil {
		output.ErrorCrossf("Failed to add bucket: %v", err)
		return err
	}

	output.SuccessCheckf("Bucket %s added successfully", name)
	output.Highlightln("\nNext steps:")
	output.Dimln("  chopsticks search <app-name>")
	return nil
}

func runBucketRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	purge := bucketRemovePurge

	output.Infof("Removing bucket: ")
	output.Highlightln(name)
	if purge {
		output.Warningln("  Mode: purge (including local files)")
	}
	fmt.Println()

	ctx := cmd.Context()
	application := getApp()

	if err := application.BucketManager().Remove(ctx, name, purge); err != nil {
		output.ErrorCrossf("Failed to remove bucket: %v", err)
		return err
	}

	output.SuccessCheckf("Bucket %s removed", name)
	return nil
}

func runBucketList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	application := getApp()

	buckets, err := application.BucketManager().ListBuckets(ctx)
	if err != nil {
		output.ErrorCrossf("Failed to list buckets: %v", err)
		return err
	}

	output.Highlightln("Added buckets:")
	output.Dimln("--------------")

	if len(buckets) == 0 {
		output.Dimln("  (no buckets)")
	} else {
		for _, name := range buckets {
			b, err := application.BucketManager().GetBucket(ctx, name)
			if err != nil {
				output.Warningf("  %-10s (unable to get details)\n", name)
				continue
			}
			url := b.Repository.URL
			if url == "" {
				url = "local"
			}
			output.Successf("  %-10s ", name)
			output.Dimln(url)
		}
	}

	output.Dimln("\nUse 'chopsticks bucket add <name> <url>' to add more buckets")
	return nil
}

func runBucketUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	application := getApp()

	if len(args) == 0 {
		output.Infoln("Updating all buckets...")
		fmt.Println()

		if err := application.BucketManager().UpdateAll(ctx); err != nil {
			output.ErrorCrossf("Failed to update buckets: %v", err)
			return err
		}

		output.SuccessCheck("All buckets updated successfully")
		return nil
	}

	name := args[0]
	output.Infof("Updating bucket: %s...\n", name)
	fmt.Println()

	if err := application.BucketManager().Update(ctx, name); err != nil {
		output.ErrorCrossf("Failed to update bucket: %v", err)
		return err
	}

	output.SuccessCheckf("Bucket %s updated successfully", name)
	return nil
}

func init() {
	bucketInitCmd.Flags().BoolVar(&bucketInitJS, "js", false, "Use JavaScript template")
	bucketInitCmd.Flags().BoolVar(&bucketInitLua, "lua", false, "Use Lua template")
	bucketInitCmd.Flags().StringVar(&bucketInitDir, "dir", "", "指定目标目录")

	bucketCreateCmd.Flags().StringVar(&bucketCreateDir, "dir", "", "指定 Bucket 目录")

	bucketAddCmd.Flags().StringVar(&bucketAddBranch, "branch", "", "指定 Git 分支")

	bucketRemoveCmd.Flags().BoolVarP(&bucketRemovePurge, "purge", "p", false, "彻底删除本地数据")

	bucketCmd.AddCommand(bucketInitCmd)
	bucketCmd.AddCommand(bucketCreateCmd)
	bucketCmd.AddCommand(bucketAddCmd)
	bucketCmd.AddCommand(bucketRemoveCmd)
	bucketCmd.AddCommand(bucketListCmd)
	bucketCmd.AddCommand(bucketUpdateCmd)

	rootCmd.AddCommand(bucketCmd)
}
