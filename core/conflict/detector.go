// Package conflict 提供应用安装前的冲突检测功能。
package conflict

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
	"chopsticks/pkg/errors"
)

// ConflictType 表示冲突类型。
type ConflictType string

const (
	ConflictTypeFile       ConflictType = "file"       // 文件冲突
	ConflictTypePort       ConflictType = "port"       // 端口冲突
	ConflictTypeEnvVar     ConflictType = "env_var"    // 环境变量冲突
	ConflictTypeRegistry   ConflictType = "registry"   // 注册表冲突
	ConflictTypeDependency ConflictType = "dependency" // 依赖冲突
)

// Conflict 表示一个冲突。
type Conflict struct {
	Type        ConflictType // 冲突类型
	Severity    Severity     // 严重程度
	Target      string       // 冲突目标（文件路径、端口号、环境变量名等）
	CurrentApp  string       // 当前占用该资源的应用
	Description string       // 冲突描述
	Suggestion  string       // 解决建议
}

// Severity 表示冲突严重程度。
type Severity string

const (
	SeverityCritical Severity = "critical" // 严重，必须解决
	SeverityWarning  Severity = "warning"  // 警告，建议解决
	SeverityInfo     Severity = "info"     // 信息，可选解决
)

// Result 表示冲突检测结果。
type Result struct {
	Conflicts   []Conflict // 检测到的冲突列表
	HasCritical bool       // 是否有严重冲突
	HasWarning  bool       // 是否有警告级别冲突
}

// Detector 冲突检测器接口。
type Detector interface {
	Detect(ctx context.Context, app *manifest.App) (*Result, error)
	DetectFileConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error)
	DetectPortConflicts(ctx context.Context, app *manifest.App) ([]Conflict, error)
	DetectEnvVarConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error)
	DetectRegistryConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error)
}

// detector 是 Detector 的实现。
type detector struct {
	storage     store.Storage
	installBase string
}

// NewDetector 创建新的冲突检测器。
func NewDetector(storage store.Storage, installBase string) Detector {
	return &detector{
		storage:     storage,
		installBase: installBase,
	}
}

// Detect 执行完整的冲突检测。
func (d *detector) Detect(ctx context.Context, app *manifest.App) (*Result, error) {
	if app == nil || app.Script == nil {
		return nil, errors.Newf(errors.KindInvalidInput, "invalid app info")
	}

	// 获取所有已安装应用
	installed, err := d.storage.ListInstalledApps(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list installed apps")
	}

	result := &Result{
		Conflicts: make([]Conflict, 0),
	}

	// 检测文件冲突
	fileConflicts, err := d.DetectFileConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrap(err, "detect file conflicts")
	}
	result.Conflicts = append(result.Conflicts, fileConflicts...)

	// 检测端口冲突
	portConflicts, err := d.DetectPortConflicts(ctx, app)
	if err != nil {
		return nil, errors.Wrap(err, "detect port conflicts")
	}
	result.Conflicts = append(result.Conflicts, portConflicts...)

	// 检测环境变量冲突
	envConflicts, err := d.DetectEnvVarConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrap(err, "detect env var conflicts")
	}
	result.Conflicts = append(result.Conflicts, envConflicts...)

	// 检测注册表冲突
	regConflicts, err := d.DetectRegistryConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrap(err, "detect registry conflicts")
	}
	result.Conflicts = append(result.Conflicts, regConflicts...)

	// 分析结果
	for _, c := range result.Conflicts {
		if c.Severity == SeverityCritical {
			result.HasCritical = true
		}
		if c.Severity == SeverityWarning {
			result.HasWarning = true
		}
	}

	return result, nil
}

// DetectFileConflicts 检测文件冲突。
func (d *detector) DetectFileConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error) {
	var conflicts []Conflict

	// 检查目标安装目录是否已被占用
	appName := app.Script.Name
	targetDir := filepath.Join(d.installBase, appName)

	for _, inst := range installed {
		if inst.Name == appName {
			// 同名应用已安装
			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeFile,
				Severity:    SeverityWarning,
				Target:      targetDir,
				CurrentApp:  inst.Name,
				Description: fmt.Sprintf("应用 '%s' 已安装于 %s", appName, inst.InstallDir),
				Suggestion:  "使用 --force 选项强制重新安装，或先卸载现有版本",
			})
			continue
		}

		// 检查安装目录是否冲突
		if inst.InstallDir == targetDir {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeFile,
				Severity:    SeverityCritical,
				Target:      targetDir,
				CurrentApp:  inst.Name,
				Description: fmt.Sprintf("安装目录 '%s' 已被应用 '%s' 占用", targetDir, inst.Name),
				Suggestion:  "选择其他安装目录或卸载冲突应用",
			})
		}
	}

	// 检查 shim 目录冲突
	shimDir := filepath.Join(d.installBase, "shims")
	exeName := appName + ".exe"
	shimPath := filepath.Join(shimDir, exeName)
	if _, err := os.Stat(shimPath); err == nil {
		conflicts = append(conflicts, Conflict{
			Type:        ConflictTypeFile,
			Severity:    SeverityWarning,
			Target:      shimPath,
			CurrentApp:  "unknown",
			Description: fmt.Sprintf("shim 文件 '%s' 已存在", exeName),
			Suggestion:  "安装时将被覆盖",
		})
	}

	return conflicts, nil
}

// DetectPortConflicts 检测端口冲突。
func (d *detector) DetectPortConflicts(ctx context.Context, app *manifest.App) ([]Conflict, error) {
	var conflicts []Conflict

	// 从应用元数据中获取端口信息
	if app.Meta == nil {
		return conflicts, nil
	}

	// 检查常用端口是否被占用
	commonPorts := d.getAppPorts(app)
	for _, port := range commonPorts {
		if isPortInUse(port) {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypePort,
				Severity:    SeverityWarning,
				Target:      fmt.Sprintf("%d", port),
				CurrentApp:  "unknown",
				Description: fmt.Sprintf("端口 %d 已被占用", port),
				Suggestion:  "应用启动时可能需要指定其他端口",
			})
		}
	}

	return conflicts, nil
}

// DetectEnvVarConflicts 检测环境变量冲突。
func (d *detector) DetectEnvVarConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error) {
	var conflicts []Conflict

	// 获取应用可能需要的环境变量
	envVars := d.getAppEnvVars(app)

	for _, envVar := range envVars {
		currentValue := os.Getenv(envVar.Name)
		if currentValue == "" {
			continue
		}

		// 检查是否有其他应用设置了此环境变量
		for _, inst := range installed {
			if inst.Name == app.Script.Name {
				continue
			}

			// 检查该应用是否可能使用相同的环境变量
			if d.isLikelyEnvVarConflict(envVar.Name, inst.Name) {
				conflicts = append(conflicts, Conflict{
					Type:        ConflictTypeEnvVar,
					Severity:    SeverityWarning,
					Target:      envVar.Name,
					CurrentApp:  inst.Name,
					Description: fmt.Sprintf("环境变量 '%s' 已被设置为 '%s'", envVar.Name, currentValue),
					Suggestion:  "可能需要更新环境变量值",
				})
			}
		}
	}

	return conflicts, nil
}

// DetectRegistryConflicts 检测注册表冲突。
func (d *detector) DetectRegistryConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error) {
	var conflicts []Conflict

	// 获取应用可能的注册表项
	regKeys := d.getAppRegistryKeys(app)

	for _, regKey := range regKeys {
		// 检查注册表项是否已存在
		exists, currentValue, err := d.checkRegistryKeyExists(regKey)
		if err != nil {
			// 注册表检查失败，记录为警告
			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeRegistry,
				Severity:    SeverityInfo,
				Target:      regKey,
				CurrentApp:  "unknown",
				Description: fmt.Sprintf("无法检查注册表项 '%s': %v", regKey, err),
				Suggestion:  "请手动检查",
			})
			continue
		}

		if exists {
			// 查找占用此注册表项的应用
			ownerApp := d.findRegistryKeyOwner(ctx, regKey, installed)

			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeRegistry,
				Severity:    SeverityWarning,
				Target:      regKey,
				CurrentApp:  ownerApp,
				Description: fmt.Sprintf("注册表项 '%s' 已存在，值为 '%s'", regKey, currentValue),
				Suggestion:  "安装时可能被覆盖",
			})
		}
	}

	return conflicts, nil
}

// isPortInUse 检查端口是否被占用。
func isPortInUse(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return true
	}
	listener.Close()
	return false
}

// getAppPorts 获取应用可能使用的端口列表。
func (d *detector) getAppPorts(app *manifest.App) []int {
	// 这里可以根据应用类型或元数据返回常用端口
	// 实际实现中可以从应用元数据读取
	ports := []int{}

	// 根据应用名称推断常用端口
	appName := strings.ToLower(app.Script.Name)
	switch {
	case strings.Contains(appName, "mysql"):
		ports = append(ports, 3306)
	case strings.Contains(appName, "postgres"):
		ports = append(ports, 5432)
	case strings.Contains(appName, "redis"):
		ports = append(ports, 6379)
	case strings.Contains(appName, "mongo"):
		ports = append(ports, 27017)
	case strings.Contains(appName, "nginx"):
		ports = append(ports, 80, 443)
	case strings.Contains(appName, "apache"):
		ports = append(ports, 80, 443)
	case strings.Contains(appName, "node") || strings.Contains(appName, "npm"):
		ports = append(ports, 3000, 8080)
	}

	return ports
}

// EnvVar 表示环境变量信息。
type EnvVar struct {
	Name        string
	Description string
}

// getAppEnvVars 获取应用可能使用的环境变量。
func (d *detector) getAppEnvVars(app *manifest.App) []EnvVar {
	vars := []EnvVar{}

	// 根据应用名称推断常用环境变量
	appName := strings.ToUpper(app.Script.Name)
	vars = append(vars, EnvVar{
		Name:        appName + "_HOME",
		Description: app.Script.Name + " 安装目录",
	})

	return vars
}

// isLikelyEnvVarConflict 判断是否可能是环境变量冲突。
func (d *detector) isLikelyEnvVarConflict(envVarName, appName string) bool {
	// 简单的启发式判断
	appNameUpper := strings.ToUpper(appName)
	envVarUpper := strings.ToUpper(envVarName)

	// 如果环境变量名包含应用名，可能是该应用设置的
	return strings.Contains(envVarUpper, appNameUpper)
}

// getAppRegistryKeys 获取应用可能的注册表项。
func (d *detector) getAppRegistryKeys(app *manifest.App) []string {
	keys := []string{}

	// 根据应用名称推断可能的注册表项
	appName := app.Script.Name

	// 常见的注册表路径
	commonPaths := []string{
		`HKLM\SOFTWARE\` + appName,
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\` + appName,
		`HKCU\SOFTWARE\` + appName,
	}

	keys = append(keys, commonPaths...)
	return keys
}

// checkRegistryKeyExists 检查注册表项是否存在。
func (d *detector) checkRegistryKeyExists(key string) (bool, string, error) {
	// 实际实现中需要使用 Windows API 或调用 reg 命令
	// 这里简化处理，返回 false 表示不冲突
	// TODO: 实现实际的注册表检查
	return false, "", nil
}

// findRegistryKeyOwner 查找注册表项的拥有者。
func (d *detector) findRegistryKeyOwner(ctx context.Context, key string, installed []*manifest.InstalledApp) string {
	// 简化实现，返回第一个可能相关的应用
	for _, inst := range installed {
		if strings.Contains(strings.ToLower(key), strings.ToLower(inst.Name)) {
			return inst.Name
		}
	}
	return "unknown"
}
