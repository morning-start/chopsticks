package bucket

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/manifest"
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
	return nil, fmt.Errorf("Git 克隆暂未实现")
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
	return &manifest.AppRef{
		Name:       name,
		ScriptPath: scriptPath,
	}, nil
}
