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
	storage     store.LegacyStorage
	installBase string
}

// NewDetector 创建新的冲突检测器。
func NewDetector(storage store.LegacyStorage, installBase string) Detector {
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
		return nil, errors.Wrap(err, "list all installed apps for conflict detection")
	}

	result := &Result{
		Conflicts: make([]Conflict, 0),
	}

	// 检测文件冲突
	fileConflicts, err := d.DetectFileConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrapf(err, "detect file conflicts for %s", app.Script.Name)
	}
	result.Conflicts = append(result.Conflicts, fileConflicts...)

	// 检测端口冲突
	portConflicts, err := d.DetectPortConflicts(ctx, app)
	if err != nil {
		return nil, errors.Wrapf(err, "detect port conflicts for %s", app.Script.Name)
	}
	result.Conflicts = append(result.Conflicts, portConflicts...)

	// 检测环境变量冲突
	envConflicts, err := d.DetectEnvVarConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrapf(err, "detect env var conflicts for %s", app.Script.Name)
	}
	result.Conflicts = append(result.Conflicts, envConflicts...)

	// 检测注册表冲突
	regConflicts, err := d.DetectRegistryConflicts(ctx, app, installed)
	if err != nil {
		return nil, errors.Wrapf(err, "detect registry conflicts for %s", app.Script.Name)
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

	// 检查 Resources 是否为 nil
	if app == nil || app.Script == nil || app.Script.Resources == nil {
		return conflicts, nil
	}

	// 从 app.Resources.Ports 读取端口声明
	for _, portDecl := range app.Script.Resources.Ports {
		if isPortInUse(portDecl.Port) {
			// 根据 Required 字段判断冲突严重程度
			severity := SeverityWarning
			if portDecl.Required {
				severity = SeverityCritical
			}

			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypePort,
				Severity:    severity,
				Target:      fmt.Sprintf("%d", portDecl.Port),
				CurrentApp:  "unknown",
				Description: fmt.Sprintf("端口 %d 已被占用 (%s)", portDecl.Port, portDecl.Description),
				Suggestion:  "应用启动时可能需要指定其他端口",
			})
		}
	}

	return conflicts, nil
}

// DetectEnvVarConflicts 检测环境变量冲突。
func (d *detector) DetectEnvVarConflicts(ctx context.Context, app *manifest.App, installed []*manifest.InstalledApp) ([]Conflict, error) {
	var conflicts []Conflict

	// 检查 Resources 是否为 nil
	if app == nil || app.Script == nil || app.Script.Resources == nil {
		return conflicts, nil
	}

	// 从 app.Resources.EnvVars 读取环境变量声明
	for _, envDecl := range app.Script.Resources.EnvVars {
		currentValue := os.Getenv(envDecl.Name)
		if currentValue == "" {
			// 环境变量未设置，如果声明为必需则添加警告
			if envDecl.Required {
				conflicts = append(conflicts, Conflict{
					Type:        ConflictTypeEnvVar,
					Severity:    SeverityWarning,
					Target:      envDecl.Name,
					CurrentApp:  "system",
					Description: fmt.Sprintf("必需的环境变量 '%s' 未设置", envDecl.Name),
					Suggestion:  fmt.Sprintf("安装时将设置为 '%s'", envDecl.Value),
				})
			}
			continue
		}

		// 环境变量已存在，检查是否有其他应用使用了此变量
		for _, inst := range installed {
			if inst.Name == app.Script.Name {
				continue
			}

			// 检查该应用是否可能使用相同的环境变量
			if d.isLikelyEnvVarConflict(envDecl.Name, inst.Name) {
				// 根据 Required 字段判断严重程度
				severity := SeverityInfo
				if envDecl.Required {
					severity = SeverityWarning
				}

				conflicts = append(conflicts, Conflict{
					Type:        ConflictTypeEnvVar,
					Severity:    severity,
					Target:      envDecl.Name,
					CurrentApp:  inst.Name,
					Description: fmt.Sprintf("环境变量 '%s' 已被设置为 '%s'", envDecl.Name, currentValue),
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

	// 检查 Resources 是否为 nil
	if app == nil || app.Script == nil || app.Script.Resources == nil {
		return conflicts, nil
	}

	// 从 app.Resources.Registry 读取注册表声明
	for _, regDecl := range app.Script.Resources.Registry {
		// 构建完整的注册表键路径
		regKey := fmt.Sprintf(`%s\%s`, regDecl.Hive, regDecl.Key)
		if regDecl.ValueName != "" {
			regKey += fmt.Sprintf(`\%s`, regDecl.ValueName)
		}

		// 检查注册表项是否已存在
		exists, currentValue, err := d.checkRegistryKeyExists(regKey)
		if err != nil {
			// 注册表检查失败，根据 Required 字段判断严重程度
			severity := SeverityInfo
			if regDecl.Required {
				severity = SeverityWarning
			}

			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeRegistry,
				Severity:    severity,
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

			// 根据 Required 字段判断严重程度
			severity := SeverityInfo
			if regDecl.Required {
				severity = SeverityWarning
			}

			conflicts = append(conflicts, Conflict{
				Type:        ConflictTypeRegistry,
				Severity:    severity,
				Target:      regKey,
				CurrentApp:  ownerApp,
				Description: fmt.Sprintf("注册表项 '%s' 已存在，值为 '%s' (%s)", regKey, currentValue, regDecl.Description),
				Suggestion:  "安装时可能被覆盖",
			})
		}
	}

	return conflicts, nil
}

// isPortInUse 检查端口是否被占用。
func isPortInUse(port int) bool {
	// 尝试在 127.0.0.1 上监听端口
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// 监听失败说明端口已被占用
		return true
	}
	// 监听成功说明端口可用，关闭监听器
	listener.Close()
	return false
}

// isLikelyEnvVarConflict 判断是否可能是环境变量冲突。
func (d *detector) isLikelyEnvVarConflict(envVarName, appName string) bool {
	// 简单的启发式判断
	appNameUpper := strings.ToUpper(appName)
	envVarUpper := strings.ToUpper(envVarName)

	// 如果环境变量名包含应用名，可能是该应用设置的
	return strings.Contains(envVarUpper, appNameUpper)
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
