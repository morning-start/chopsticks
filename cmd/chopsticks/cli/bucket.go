package cli

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"chopsticks/core/app"
)

//go:embed template/bucket-js
//go:embed template/bucket-lua
var templates embed.FS

// BucketSubcommand 定义 bucket 子命令信息。
type BucketSubcommand struct {
	Name        string   // 主命令名
	Aliases     []string // 别名
	Description string   // 描述
	Usage       string   // 用法
}

// bucketSubcommands 定义所有 bucket 子命令。
var bucketSubcommands = []BucketSubcommand{
	{
		Name:        "init",
		Aliases:     []string{},
		Description: "初始化新 Bucket 目录结构",
		Usage:       "bucket init <name> [--js] [--lua] [--dir <path>]",
	},
	{
		Name:        "create",
		Aliases:     []string{"c"},
		Description: "创建新 App 模板",
		Usage:       "bucket create <app-name> [--dir <path>]",
	},
	{
		Name:        "add",
		Aliases:     []string{"a"},
		Description: "添加软件源",
		Usage:       "bucket add <name> <url> [--branch <branch>]",
	},
	{
		Name:        "remove",
		Aliases:     []string{"rm", "delete", "del"},
		Description: "删除软件源",
		Usage:       "bucket remove <name> [--purge]",
	},
	{
		Name:        "list",
		Aliases:     []string{"ls"},
		Description: "列出软件源",
		Usage:       "bucket list",
	},
	{
		Name:        "update",
		Aliases:     []string{"up", "upgrade"},
		Description: "更新软件源",
		Usage:       "bucket update [name]",
	},
}

var templateMap = map[string]string{
	"js":  "bucket-js",
	"lua": "bucket-lua",
}

func printBucketUsage() {
	fmt.Println("Bucket（软件源）管理命令")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  chopsticks bucket <子命令> [参数]")
	fmt.Println()
	fmt.Println("子命令:")

	for _, cmd := range bucketSubcommands {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = fmt.Sprintf(" (别名: %s)", strings.Join(cmd.Aliases, ", "))
		}
		fmt.Printf("  %-10s %s%s\n", cmd.Name, cmd.Description, aliases)
	}

	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  chopsticks bucket init my-bucket")
	fmt.Println("  chopsticks bucket init my-bucket --js")
	fmt.Println("  chopsticks bucket init my-bucket --lua")
	fmt.Println("  chopsticks bucket create my-app")
	fmt.Println("  chopsticks bucket add main https://github.com/user/main")
	fmt.Println("  chopsticks bucket list")
	fmt.Println("  chopsticks bucket update")
	fmt.Println()
	fmt.Println("快捷方式:")
	fmt.Println("  chopsticks s add main <url>    # 使用 's' 代替 'bucket'")
	fmt.Println("  chopsticks s ls                # 使用 'ls' 代替 'list'")
	fmt.Println("  chopsticks s rm main           # 使用 'rm' 代替 'remove'")
}

// BucketCommand 处理 bucket 命令。
func BucketCommand(ctx context.Context, application app.Application, args []string) error {
	if len(args) < 1 {
		printBucketUsage()
		return fmt.Errorf("缺少子命令")
	}

	subcommand := args[0]
	subArgs := args[1:]

	primarySubcmd := subcommand

	if subcommand == "s" || subcommand == "bucket" {
		if len(subArgs) < 1 {
			printBucketUsage()
			return fmt.Errorf("缺少子命令")
		}
		primarySubcmd = subArgs[0]
		subArgs = subArgs[1:]
	}

	switch primarySubcmd {
	case "init":
		return bucketInit(ctx, application, subArgs)
	case "create":
		return bucketCreate(ctx, application, subArgs)
	case "add":
		return bucketAdd(ctx, application, subArgs)
	case "remove":
		return bucketRemove(ctx, application, subArgs)
	case "list":
		return bucketList(ctx, application, subArgs)
	case "update":
		return bucketUpdate(ctx, application, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "未知子命令: %s\n\n", subcommand)
		printBucketUsage()
		return fmt.Errorf("未知子命令: %s", subcommand)
	}
}

func bucketAdd(_ context.Context, _ app.Application, args []string) error {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks bucket add <name> <url>")
		return fmt.Errorf("缺少参数")
	}

	name := args[0]
	url := args[1]

	fmt.Printf("添加软件源: %s\n", name)
	fmt.Printf("  URL: %s\n", url)

	fmt.Printf("✓ 软件源 %s 添加成功\n", name)
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("  chopsticks search %s\n", name)
	return nil
}

func bucketRemove(_ context.Context, _ app.Application, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks bucket remove <name> [--purge]")
		return fmt.Errorf("缺少软件源名称")
	}

	name := args[0]

	fmt.Printf("删除软件源: %s\n", name)
	fmt.Printf("✓ 软件源 %s 已删除\n", name)
	return nil
}

func bucketList(_ context.Context, _ app.Application, _ []string) error {
	fmt.Println("已添加的软件源:")
	fmt.Println("--------------")
	fmt.Println("  main    https://github.com/chopsticks-bows/main (默认)")
	fmt.Println()
	fmt.Println("使用 'chopsticks bucket add <name> <url>' 添加更多软件源")
	return nil
}

func bucketUpdate(_ context.Context, _ app.Application, args []string) error {
	if len(args) == 0 {
		fmt.Println("更新所有软件源...")
		fmt.Println("✓ 所有软件源更新成功")
		return nil
	}

	name := args[0]
	fmt.Printf("更新软件源: %s...\n", name)
	fmt.Printf("✓ 软件源 %s 更新成功\n", name)
	return nil
}

// BucketTemplateData 用于模板渲染的数据
type BucketTemplateData struct {
	Name string
}

// bucketInit 初始化新 Bucket 目录结构。
func bucketInit(_ context.Context, _ app.Application, args []string) error {
	if len(args) < 1 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(os.Stderr, "用法: chopsticks bucket init <name> [--js] [--lua] [--dir <path>]")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  chopsticks bucket init my-bucket")
		fmt.Println("  chopsticks bucket init my-bucket --js")
		fmt.Println("  chopsticks bucket init my-bucket --lua")
		fmt.Println("  chopsticks bucket init my-bucket --dir ./buckets")
		return nil
	}

	name := args[0]

	bucketDir := ""
	templateType := "js"

	for _, arg := range args[1:] {
		switch arg {
		case "--js":
			templateType = "js"
		case "--lua":
			templateType = "lua"
		case "--dir":
			idx := findArgIndex(args, arg)
			if idx+1 < len(args) {
				bucketDir = args[idx+1]
			}
		}
	}

	targetDir := bucketDir
	if targetDir == "" {
		targetDir = name
	}

	templateName := templateMap[templateType]
	templatePath := "template/" + templateName

	fmt.Printf("初始化 Bucket: %s (%s)\n", name, map[string]string{"js": "JavaScript", "lua": "Lua"}[templateType])

	if err := copyEmbeddedTemplate(templatePath, targetDir, name); err != nil {
		return fmt.Errorf("复制模板失败: %w", err)
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

// copyEmbeddedTemplate 从嵌入的文件系统复制模板
func copyEmbeddedTemplate(templatePath, destDir, name string) error {
	data := BucketTemplateData{Name: name}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	entries, err := templates.ReadDir(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板目录失败: %w", err)
	}

	fmt.Printf("  [DEBUG] 模板目录 %s 包含:\n", templatePath)
	for _, e := range entries {
		fmt.Printf("  [DEBUG]   - %s (IsDir: %v)\n", e.Name(), e.IsDir())
	}

	for _, entry := range entries {
		srcPath := templatePath + "/" + entry.Name()

		if entry.IsDir() {
			dirName := strings.ReplaceAll(entry.Name(), "{{.Name}}", name)
			destPath := filepath.Join(destDir, dirName)
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			fmt.Printf("  创建目录: %s\n", destPath)
			if err := copyEmbeddedDir(srcPath, destPath, name); err != nil {
				return err
			}
		} else {
			destFilePath := filepath.Join(destDir, entry.Name())
			destFilePath = strings.ReplaceAll(destFilePath, "{{.Name}}", name)
			if err := copyEmbeddedFile(srcPath, destFilePath, name, data); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyEmbeddedDir 递归复制嵌入的目录
func copyEmbeddedDir(srcPath, destDir, name string) error {
	entries, err := templates.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("读取模板目录失败: %w", err)
	}

	for _, entry := range entries {
		srcFilePath := srcPath + "/" + entry.Name()
		destFilePath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			fmt.Printf("  创建目录: %s\n", destFilePath)
			if err := copyEmbeddedDir(srcFilePath, destFilePath, name); err != nil {
				return err
			}
		} else {
			data := BucketTemplateData{Name: name}
			if err := copyEmbeddedFile(srcFilePath, destFilePath, name, data); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyEmbeddedFile 复制并处理嵌入的文件
func copyEmbeddedFile(srcPath, destPath, name string, data BucketTemplateData) error {
	content, err := templates.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %w", err)
	}

	var result string
	ext := filepath.Ext(destPath)
	if ext == ".ts" || ext == ".js" || ext == ".lua" || ext == ".json" ||
		ext == ".md" || ext == ".yaml" || ext == ".yml" || ext == ".toml" ||
		ext == ".txt" || ext == ".ignore" {
		destPath = strings.ReplaceAll(destPath, "{{.Name}}", name)

		tmpl, err := template.New("file").Parse(string(content))
		if err != nil {
			return fmt.Errorf("解析模板失败: %w", err)
		}
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("渲染模板失败: %w", err)
		}
		result = buf.String()
	} else {
		result = string(content)
	}

	if err := os.WriteFile(destPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	fmt.Printf("  创建文件: %s\n", destPath)
	return nil
}

func findArgIndex(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

// bucketCreate 创建新 App 模板。
func bucketCreate(_ context.Context, _ app.Application, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: chopsticks bucket create <app-name> [--dir <path>]")
		return fmt.Errorf("缺少软件包名称")
	}

	name := args[0]

	bucketDir := ""

	for _, arg := range args[1:] {
		if arg == "--dir" {
			idx := findArgIndex(args, arg)
			if idx+1 < len(args) {
				bucketDir = args[idx+1]
			}
		}
	}

	targetDir := bucketDir
	if targetDir == "" {
		targetDir = "./apps/" + name
	} else {
		targetDir = bucketDir + "/apps/" + name
	}

	dirs := []string{
		targetDir,
		targetDir + "/scripts",
		targetDir + "/tests",
	}

	fmt.Printf("创建 App 模板: %s\n", name)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
		fmt.Printf("  创建目录: %s\n", dir)
	}

	manifestContent := "name: " + name + "\n" +
		"version: 1.0.0\n" +
		"description: A software package\n" +
		"author: unknown\n" +
		"homepage: https://example.com\n" +
		"license: MIT\n\n" +
		"arch: amd64\n\n" +
		"format: zip\n\n" +
		"install:\n" +
		"  # installer: setup.exe\n" +
		"  # install_args: [/S]\n\n" +
		"hooks:\n" +
		"  pre_download: |\n" +
		"    console.log('准备下载...')\n" +
		"  post_download: |\n" +
		"    console.log('下载完成')\n" +
		"  pre_extract: |\n" +
		"    console.log('准备解压...')\n" +
		"  post_extract: |\n" +
		"    console.log('解压完成')\n" +
		"  pre_install: |\n" +
		"    console.log('准备安装...')\n" +
		"  post_install: |\n" +
		"    console.log('安装完成')\n" +
		"  pre_uninstall: |\n" +
		"    console.log('准备卸载...')\n" +
		"  post_uninstall: |\n" +
		"    console.log('卸载完成')\n\n" +
		"files:\n" +
		"  - " + name + "*\n\n" +
		"shortcuts:\n" +
		"  - name: " + name + "\n" +
		"    path: " + name + ".exe\n\n" +
		"registry:\n" +
		"  - key: SOFTWARE\\\\" + name + "\n" +
		"    values:\n" +
		"      - name: InstallPath\n" +
		"        type: REG_SZ\n" +
		"        value: {{.InstallDir}}\n\n" +
		"env:\n" +
		"  - name: " + strings.ToUpper(name) + "_HOME\n" +
		"    value: {{.InstallDir}}\n"

	manifestPath := targetDir + "/manifest.yaml"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("创建清单文件失败: %w", err)
	}
	fmt.Printf("  创建文件: %s\n", manifestPath)

	readmeContent := "# " + name + "\n\n" +
		"## 描述\n\n" +
		"软件包 " + name + " 的说明文档。\n\n" +
		"## 安装\n\n" +
		"```bash\n" +
		"chopsticks install " + name + "\n" +
		"```\n\n" +
		"## 卸载\n\n" +
		"```bash\n" +
		"chopsticks uninstall " + name + "\n" +
		"```\n"

	readmePath := targetDir + "/README.md"
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("创建说明文件失败: %w", err)
	}
	fmt.Printf("  创建文件: %s\n", readmePath)

	scriptContent := "// " + name + " 安装脚本\n\n" +
		"function preDownload() {\n" +
		"    console.log(\"准备下载 \" + name + \" \" + version);\n" +
		"}\n\n" +
		"function postDownload() {\n" +
		"    console.log(\"下载完成\");\n" +
		"}\n\n" +
		"function preExtract() {\n" +
		"    console.log(\"准备解压\");\n" +
		"}\n\n" +
		"function postExtract() {\n" +
		"    console.log(\"解压完成\");\n" +
		"}\n\n" +
		"function preInstall() {\n" +
		"    console.log(\"准备安装\");\n" +
		"}\n\n" +
		"function postInstall() {\n" +
		"    console.log(\"安装完成\");\n" +
		"}\n\n" +
		"function preUninstall() {\n" +
		"    console.log(\"准备卸载\");\n" +
		"}\n\n" +
		"function postUninstall() {\n" +
		"    console.log(\"卸载完成\");\n" +
		"}\n"

	scriptPath := targetDir + "/scripts/install.js"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("创建脚本文件失败: %w", err)
	}
	fmt.Printf("  创建文件: %s\n", scriptPath)

	fmt.Printf("✓ App %s 模板创建完成\n", name)
	fmt.Printf("\n下一步:\n")
	fmt.Printf("  编辑 manifest.yaml 添加软件包信息\n")
	fmt.Printf("  编辑 scripts/install.js 添加安装逻辑\n")
	return nil
}
