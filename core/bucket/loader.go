package bucket

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/manifest"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Loader 定义软件源加载器接口。
type Loader interface {
	Load(ctx context.Context, path string) (*manifest.Bucket, error)
	LoadFromGit(ctx context.Context, url, branch string) (*manifest.Bucket, error)
	ScanApps(ctx context.Context, bucketPath string) (map[string]*manifest.AppRef, error)
}

// loader 是 Loader 的实现。
type loader struct{}

// 编译时接口检查。
var _ Loader = (*loader)(nil)

// NewLoader 创建新的 Loader。
func NewLoader() Loader {
	return &loader{}
}

// Load 从本地路径加载软件源。
func (l *loader) Load(ctx context.Context, path string) (*manifest.Bucket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	configPath := filepath.Join(path, "bucket.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取软件源配置: %w", err)
	}

	var config manifest.BucketConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析软件源配置: %w", err)
	}

	apps, err := l.ScanApps(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("扫描应用: %w", err)
	}

	return &manifest.Bucket{
		Config: config,
		Path:   path,
		Apps:   apps,
	}, nil
}

// LoadFromGit 从 Git 仓库克隆并加载软件源。
func (l *loader) LoadFromGit(ctx context.Context, url, branch string) (*manifest.Bucket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 从 URL 提取 bucket 名称
	bucketName := extractBucketName(url)
	bucketsDir := filepath.Join(os.Getenv("USERPROFILE"), ".chopsticks", "buckets")
	destPath := filepath.Join(bucketsDir, bucketName)

	// 如果目录已存在，先删除
	if _, err := os.Stat(destPath); err == nil {
		if err := os.RemoveAll(destPath); err != nil {
			return nil, fmt.Errorf("删除旧目录: %w", err)
		}
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	}

	if branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		cloneOpts.SingleBranch = true
	}

	_, err := git.PlainCloneContext(ctx, destPath, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("克隆仓库: %w", err)
	}

	return l.Load(ctx, destPath)
}

// extractBucketName 从 Git URL 提取 bucket 名称
func extractBucketName(url string) string {
	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	// 从路径中提取最后一部分
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		if name != "" {
			return name
		}
	}

	// 如果无法提取，使用默认名称
	return "custom"
}

// ScanApps 扫描软件源目录中的所有应用。
func (l *loader) ScanApps(ctx context.Context, bucketPath string) (map[string]*manifest.AppRef, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	apps := make(map[string]*manifest.AppRef)

	appsPath := filepath.Join(bucketPath, "apps")
	entries, err := os.ReadDir(appsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return apps, nil
		}
		return nil, fmt.Errorf("读取应用目录: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return apps, ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".js" && ext != ".lua" {
			continue
		}

		appName := strings.TrimSuffix(name, ext)
		scriptPath := filepath.Join(appsPath, name)

		ref, err := l.loadAppRef(appName, scriptPath)
		if err != nil {
			continue
		}

		apps[appName] = ref
	}

	return apps, nil
}

// loadAppRef 加载单个应用的引用信息。
func (l *loader) loadAppRef(name, scriptPath string) (*manifest.AppRef, error) {
	ref := &manifest.AppRef{
		Name:       name,
		ScriptPath: scriptPath,
	}

	// 尝试读取对应的 .meta.json 文件获取更多信息
	dir := filepath.Dir(scriptPath)
	metaPath := filepath.Join(dir, name+".meta.json")

	if data, err := os.ReadFile(metaPath); err == nil {
		var meta struct {
			Description string   `json:"description"`
			Version     string   `json:"version"`
			Category    string   `json:"category"`
			Tags        []string `json:"tags"`
		}
		if err := json.Unmarshal(data, &meta); err == nil {
			ref.Description = meta.Description
			ref.Version = meta.Version
			ref.Category = meta.Category
			ref.Tags = meta.Tags
			ref.MetaPath = metaPath
		}
	}

	// 如果 meta.json 不存在或解析失败，尝试从脚本文件中提取基本信息
	if ref.Description == "" {
		if info := extractInfoFromScript(scriptPath); info != nil {
			ref.Description = info.Description
			ref.Version = info.Version
			ref.Category = info.Category
			ref.Tags = info.Tags
		}
	}

	return ref, nil
}

// scriptInfo 保存从脚本中提取的信息
type scriptInfo struct {
	Description string
	Version     string
	Category    string
	Tags        []string
}

// extractInfoFromScript 从脚本文件中提取基本信息
func extractInfoFromScript(scriptPath string) *scriptInfo {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil
	}

	info := &scriptInfo{}
	ext := filepath.Ext(scriptPath)

	switch ext {
	case ".js":
		// 从 JS 文件中提取注释中的信息
		info = extractFromJS(string(content))
	case ".lua":
		// 从 Lua 文件中提取注释中的信息
		info = extractFromLua(string(content))
	}

	return info
}

// extractFromJS 从 JavaScript 文件中提取信息
func extractFromJS(content string) *scriptInfo {
	info := &scriptInfo{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 提取 @description
		if strings.Contains(line, "@description") {
			parts := strings.SplitN(line, "@description", 2)
			if len(parts) == 2 {
				info.Description = strings.TrimSpace(parts[1])
			}
		}

		// 提取 @version
		if strings.Contains(line, "@version") {
			parts := strings.SplitN(line, "@version", 2)
			if len(parts) == 2 {
				info.Version = strings.TrimSpace(parts[1])
			}
		}

		// 提取 category
		if strings.Contains(line, "category:") || strings.Contains(line, "category =") {
			if idx := strings.Index(line, `"`); idx != -1 {
				endIdx := strings.LastIndex(line, `"`)
				if endIdx > idx {
					info.Category = line[idx+1 : endIdx]
				}
			}
		}
	}

	return info
}

// extractDependenciesFromScript 从脚本文件中提取依赖信息
func extractDependenciesFromScript(scriptPath string) []manifest.Dependency {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil
	}

	ext := filepath.Ext(scriptPath)
	switch ext {
	case ".js":
		return extractDependenciesFromJS(string(content))
	default:
		return nil
	}
}

// extractDependenciesFromJS 从 JavaScript 文件中提取依赖
func extractDependenciesFromJS(content string) []manifest.Dependency {
	var deps []manifest.Dependency

	// 查找 depends: [...] 或 depends = [...] 模式
	lines := strings.Split(content, "\n")
	inDependsArray := false
	var dependsContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检测 depends 字段开始
		if !inDependsArray {
			if strings.Contains(line, "depends:") || strings.Contains(line, "depends =") {
				// 提取 depends: 后面的内容
				startIdx := strings.Index(line, "[")
				if startIdx != -1 {
					inDependsArray = true
					dependsContent.WriteString(line[startIdx:])
					if strings.Contains(line, "]") {
						inDependsArray = false
					}
				}
			}
		} else {
			// 继续在多行数组中
			dependsContent.WriteString(" ")
			dependsContent.WriteString(line)
			if strings.Contains(line, "]") {
				inDependsArray = false
			}
		}
	}

	// 解析依赖数组内容
	contentStr := dependsContent.String()
	if contentStr == "" {
		return deps
	}

	// 提取方括号内的内容
	startIdx := strings.Index(contentStr, "[")
	endIdx := strings.Index(contentStr, "]")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return deps
	}

	innerContent := contentStr[startIdx+1 : endIdx]

	// 分割依赖项
	items := strings.Split(innerContent, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		// 移除引号
		item = strings.Trim(item, `"'`)
		if item == "" {
			continue
		}

		// 解析依赖名称和版本约束
		dep := manifest.Dependency{
			Name:    item,
			Version: "",
		}

		// 检查是否有版本约束 (格式: "name:version")
		if idx := strings.Index(item, ":"); idx != -1 {
			dep.Name = item[:idx]
			dep.Version = item[idx+1:]
		}

		deps = append(deps, dep)
	}

	return deps
}

// extractFromLua 从 Lua 文件中提取信息
func extractFromLua(content string) *scriptInfo {
	info := &scriptInfo{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 提取 -- description:
		if strings.HasPrefix(line, "--") {
			line = strings.TrimPrefix(line, "--")
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "description:") {
				info.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
			if strings.HasPrefix(line, "version:") {
				info.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			}
			if strings.HasPrefix(line, "category:") {
				info.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
			}
		}
	}

	return info
}
