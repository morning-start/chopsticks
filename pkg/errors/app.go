package errors

import (
	"fmt"
)

// 应用安装错误
// NewAppNotFound 创建软件未找到错误
func NewAppNotFound(name string) *StructuredError {
	return NewStructured(ErrAppNotFound, fmt.Sprintf("软件未找到：%s", name)).
		WithContext("app_name", name).
		WithOperation("app lookup")
}

// NewAppAlreadyInstalled 创建软件已安装错误
func NewAppAlreadyInstalled(name, version string) *StructuredError {
	return NewStructured(ErrAppAlreadyInstalled, fmt.Sprintf("软件已安装：%s (版本：%s)", name, version)).
		WithContext("app_name", name).
		WithContext("version", version).
		WithOperation("install check")
}

// NewAppNotInstalled 创建软件未安装错误
func NewAppNotInstalled(name string) *StructuredError {
	return NewStructured(ErrAppNotInstalled, fmt.Sprintf("软件未安装：%s", name)).
		WithContext("app_name", name).
		WithOperation("app lookup")
}

// NewVersionNotFound 创建版本未找到错误
func NewVersionNotFound(app, version string) *StructuredError {
	return NewStructured(ErrVersionNotFound, fmt.Sprintf("版本未找到：%s@%s", app, version)).
		WithContext("app_name", app).
		WithContext("version", version).
		WithOperation("version lookup")
}

// NewVersionAlreadyExists 创建版本已存在错误
func NewVersionAlreadyExists(app, version string) *StructuredError {
	return NewStructured(ErrVersionAlreadyExists, fmt.Sprintf("版本已存在：%s@%s", app, version)).
		WithContext("app_name", app).
		WithContext("version", version).
		WithOperation("version check")
}

// NewDownloadFailed 创建下载失败错误
func NewDownloadFailed(url string, err error) *StructuredError {
	return WrapStructured(err, ErrDownloadFailed, fmt.Sprintf("下载失败：%s", url)).
		WithContext("url", url).
		WithOperation("download")
}

// NewChecksumMismatch 创建校验和不匹配错误
func NewChecksumMismatch(expected, actual string) *StructuredError {
	return NewStructured(ErrChecksumMismatch, fmt.Sprintf("校验和不匹配：期望 %s", expected)).
		WithContext("expected_hash", expected).
		WithContext("actual_hash", actual).
		WithOperation("checksum verify").
		WithSuggestion(RecoverySuggestion{
			Title:       "清理缓存重新下载",
			Description: "文件可能已损坏，清理缓存后重新下载",
			Commands: []string{
				"chopsticks cache clean --package {package}",
				"chopsticks install {package}",
			},
			AutoFixable: true,
		})
}

// NewInstallFailed 创建安装失败错误
func NewInstallFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrInstallFailed, fmt.Sprintf("安装失败：%s", name)).
		WithContext("app_name", name).
		WithOperation("install")
}

// NewUninstallFailed 创建卸载失败错误
func NewUninstallFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrUninstallFailed, fmt.Sprintf("卸载失败：%s", name)).
		WithContext("app_name", name).
		WithOperation("uninstall")
}

// NewUpdateFailed 创建更新失败错误
func NewUpdateFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrUpdateFailed, fmt.Sprintf("更新失败：%s", name)).
		WithContext("app_name", name).
		WithOperation("update")
}

// NewScriptFailed 创建脚本执行失败错误
func NewScriptFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrScriptFailed, fmt.Sprintf("脚本执行失败：%s", name)).
		WithContext("script_name", name).
		WithOperation("script execution")
}

// NewHookFailed 创建钩子执行失败错误
func NewHookFailed(hookName string, err error) *StructuredError {
	return WrapStructured(err, ErrHookFailed, fmt.Sprintf("钩子执行失败：%s", hookName)).
		WithContext("hook_name", hookName).
		WithOperation("hook execution")
}

// NewArchiveExtractFailed 创建解压失败错误
func NewArchiveExtractFailed(path string, err error) *StructuredError {
	return WrapStructured(err, ErrArchiveExtractFailed, fmt.Sprintf("解压失败：%s", path)).
		WithContext("archive_path", path).
		WithOperation("extract archive")
}

// NewDependencyConflict 创建依赖冲突错误
func NewDependencyConflict(name, reason string) *StructuredError {
	return NewStructured(ErrDependencyConflict, fmt.Sprintf("依赖冲突：%s - %s", name, reason)).
		WithContext("dependency_name", name).
		WithContext("reason", reason).
		WithOperation("dependency resolution").
		WithSuggestion(RecoverySuggestion{
			Title:       "查看依赖冲突详情",
			Description: "使用详细模式查看依赖冲突的详细信息",
			Commands: []string{
				"chopsticks install {package} --verbose",
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "清理依赖缓存",
			Description: "清理依赖解析缓存后重试",
			Commands: []string{
				"chopsticks deps clean",
			},
			AutoFixable: true,
		})
}

// NewDependencyNotFound 创建依赖未找到错误
func NewDependencyNotFound(name string) *StructuredError {
	return NewStructured(ErrDependencyNotFound, fmt.Sprintf("依赖未找到：%s", name)).
		WithContext("dependency_name", name).
		WithOperation("dependency resolution")
}

// NewCircularDependency 创建循环依赖错误
func NewCircularDependency(deps []string) *StructuredError {
	return NewStructured(ErrCircularDependency, fmt.Sprintf("检测到循环依赖：%v", deps)).
		WithContext("dependencies", deps).
		WithOperation("dependency resolution").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查依赖配置",
			Description: "循环依赖通常由配置错误引起，请检查相关软件的依赖配置",
			Commands:    []string{},
			AutoFixable: false,
		})
}

// NewDependencyVersion 创建依赖版本不匹配错误
func NewDependencyVersion(name, required, available string) *StructuredError {
	return NewStructured(ErrDependencyVersion, fmt.Sprintf("依赖版本不匹配：%s (需要：%s, 可用：%s)", name, required, available)).
		WithContext("dependency_name", name).
		WithContext("required_version", required).
		WithContext("available_version", available).
		WithOperation("dependency version check")
}

// NewInstallCancelled 创建安装取消错误
func NewInstallCancelled(reason string) *StructuredError {
	return NewStructured(ErrInstallCancelled, fmt.Sprintf("安装已取消：%s", reason)).
		WithContext("reason", reason).
		WithOperation("install").
		WithRecoverable(false)
}

// NewPermissionDenied 创建权限不足错误
func NewPermissionDenied(operation string) *StructuredError {
	return NewStructured(ErrPermissionDenied, fmt.Sprintf("权限不足：%s", operation)).
		WithContext("operation", operation).
		WithOperation("permission check").
		WithSuggestion(RecoverySuggestion{
			Title:       "以管理员身份运行",
			Description: "右键点击终端，选择'以管理员身份运行'，然后重试操作",
			Commands:    []string{},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "修改安装目录到用户目录",
			Description: "将安装目录配置到用户目录下，避免需要管理员权限",
			Commands: []string{
				`chopsticks config set install_dir "%USERPROFILE%\tools"`,
			},
			AutoFixable: true,
		})
}

// NewInsufficientDisk 创建磁盘空间不足错误
func NewInsufficientDisk(required, available uint64) *StructuredError {
	return NewStructured(ErrInsufficientDisk, fmt.Sprintf("磁盘空间不足：需要 %d MB, 可用 %d MB", required/1024/1024, available/1024/1024)).
		WithContext("required_bytes", required).
		WithContext("available_bytes", available).
		WithOperation("disk space check").
		WithSuggestion(RecoverySuggestion{
			Title:       "清理缓存",
			Description: "清理下载缓存释放磁盘空间",
			Commands: []string{
				"chopsticks cache clean",
			},
			AutoFixable: true,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "检查磁盘空间",
			Description: "查看当前磁盘使用情况",
			Commands: []string{
				"chopsticks doctor",
			},
			AutoFixable: false,
		})
}

// NewFileNotFound 创建文件不存在错误
func NewFileNotFound(path string) *StructuredError {
	return NewStructured(ErrFileNotFound, fmt.Sprintf("文件不存在：%s", path)).
		WithContext("file_path", path).
		WithOperation("file lookup")
}

// NewFileAlreadyExists 创建文件已存在错误
func NewFileAlreadyExists(path string) *StructuredError {
	return NewStructured(ErrFileAlreadyExists, fmt.Sprintf("文件已存在：%s", path)).
		WithContext("file_path", path).
		WithOperation("file creation")
}

// NewInvalidPath 创建路径无效错误
func NewInvalidPath(path, reason string) *StructuredError {
	return NewStructured(ErrInvalidPath, fmt.Sprintf("路径无效：%s - %s", path, reason)).
		WithContext("path", path).
		WithContext("reason", reason).
		WithOperation("path validation")
}

// NewAppManifestNotFound 创建应用清单文件未找到错误
func NewAppManifestNotFound(bucket, app string) *StructuredError {
	return NewStructured(ErrManifestNotFound, fmt.Sprintf("应用清单文件未找到：%s/%s", bucket, app)).
		WithContext("bucket_name", bucket).
		WithContext("app_name", app).
		WithOperation("manifest lookup")
}
