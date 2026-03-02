package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/cmd/cli/template"
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

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		output.ErrorCrossf("Failed to create directory: %v", err)
		return err
	}

	// 复制模板文件（从嵌入的文件系统）
	if err := template.CopyTemplateDir(templateType, targetDir); err != nil {
		output.ErrorCrossf("Failed to copy template files: %v", err)
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

// copyTemplateDir 递归复制模板目录到目标目录
func copyTemplateDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read template directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("create directory %s: %w", dstPath, err)
			}
			if err := copyTemplateDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("read file %s: %w", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("write file %s: %w", dstPath, err)
			}
		}
	}
	return nil
}

func runBucketCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	bucketDir := bucketCreateDir

	// 确定 apps 目录路径
	appsDir := "./apps"
	if bucketDir != "" {
		appsDir = filepath.Join(bucketDir, "apps")
	}

	// 确保 apps 目录存在
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		output.ErrorCrossf("Failed to create apps directory: %v", err)
		return err
	}

	output.Infof("Creating App: ")
	output.Highlightln(name)

	// 从嵌入的文件系统读取模板
	templateContent, err := template.ReadTemplateFileByName("bucket-js", "apps/_example_.js")
	if err != nil {
		output.ErrorCrossf("Failed to read template file: %v", err)
		return err
	}

	// 替换模板内容
	content := string(templateContent)
	// 替换类名 ExampleApp -> AppName (首字母大写，移除连字符)
	className := toClassName(name) + "App"
	content = strings.ReplaceAll(content, "ExampleApp", className)
	// 替换 name
	content = strings.ReplaceAll(content, `name: "example"`, fmt.Sprintf(`name: "%s"`, name))
	// 替换 description
	content = strings.ReplaceAll(content, `description: "Example Application"`, fmt.Sprintf(`description: "%s Application"`, capitalizeFirst(name)))
	// 替换 bucket 名称
	content = strings.ReplaceAll(content, `bucket: "{{.Name}}"`, `bucket: "main"`)
	// 更新注释
	content = strings.ReplaceAll(content, "Example App", capitalizeFirst(name)+" App")
	content = strings.ReplaceAll(content, "A sample app", name+" application")
	content = strings.ReplaceAll(content, "@module apps/_example_", fmt.Sprintf("@module apps/%s", name))

	// 写入文件
	appFilePath := filepath.Join(appsDir, name+".js")
	if err := os.WriteFile(appFilePath, []byte(content), 0644); err != nil {
		output.ErrorCrossf("Failed to create app file: %v", err)
		return err
	}
	output.Dimf("  Created file: %s\n", appFilePath)

	output.SuccessCheckf("App %s created successfully", name)
	output.Highlightln("\nNext steps:")
	output.Dimf("  Edit %s to customize the app\n", appFilePath)
	output.Dimln("  Implement checkVersion() and getDownloadInfo() methods")
	return nil
}

// capitalizeFirst 将字符串首字母大写
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// toClassName 将应用名转换为有效的类名（首字母大写，移除连字符，驼峰命名）
func toClassName(s string) string {
	if s == "" {
		return ""
	}
	// 分割连字符
	parts := strings.Split(s, "-")
	var result string
	for _, part := range parts {
		if part != "" {
			result += capitalizeFirst(part)
		}
	}
	return result
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
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

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
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

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
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

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
	if application == nil {
		return fmt.Errorf("应用未初始化")
	}

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
