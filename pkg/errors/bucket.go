package errors

import (
	"fmt"
)

// NewBucketNotFound 创建软件源未找到错误
func NewBucketNotFound(name string) *StructuredError {
	return NewStructured(ErrBucketNotFound, fmt.Sprintf("软件源未找到：%s", name)).
		WithContext("bucket_name", name).
		WithOperation("bucket lookup").
		WithSuggestion(RecoverySuggestion{
			Title:       "查看已添加的软件源",
			Description: "检查软件源是否正确添加",
			Commands: []string{
				"chopsticks bucket list",
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "添加软件源",
			Description: "添加主软件源",
			Commands: []string{
				"chopsticks bucket add main https://github.com/chopsticks-bucket/main",
			},
			AutoFixable: true,
		})
}

// NewBucketAlreadyExists 创建软件源已存在错误
func NewBucketAlreadyExists(name string) *StructuredError {
	return NewStructured(ErrBucketAlreadyExists, fmt.Sprintf("软件源已存在：%s", name)).
		WithContext("bucket_name", name).
		WithOperation("bucket add")
}

// NewInvalidBucketURL 创建软件源 URL 无效错误
func NewInvalidBucketURL(url string) *StructuredError {
	return NewStructured(ErrInvalidBucketURL, fmt.Sprintf("软件源 URL 无效：%s", url)).
		WithContext("bucket_url", url).
		WithOperation("bucket validate").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查 URL 格式",
			Description: "确保 URL 是有效的 Git 仓库地址",
			Commands:    []string{},
			AutoFixable: false,
		})
}

// NewBucketLoadFailed 创建软件源加载失败错误
func NewBucketLoadFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrBucketLoadFailed, fmt.Sprintf("软件源加载失败：%s", name)).
		WithContext("bucket_name", name).
		WithOperation("bucket load")
}

// NewBucketUpdateFailed 创建软件源更新失败错误
func NewBucketUpdateFailed(name string, err error) *StructuredError {
	return WrapStructured(err, ErrBucketUpdateFailed, fmt.Sprintf("软件源更新失败：%s", name)).
		WithContext("bucket_name", name).
		WithOperation("bucket update").
		WithRetryable(true)
}

// NewManifestNotFound 创建清单文件未找到错误
func NewManifestNotFound(bucket, app string) *StructuredError {
	return NewStructured(ErrManifestNotFound, fmt.Sprintf("清单文件未找到：%s/%s", bucket, app)).
		WithContext("bucket_name", bucket).
		WithContext("app_name", app).
		WithOperation("manifest lookup").
		WithSuggestion(RecoverySuggestion{
			Title:       "更新软件源",
			Description: "软件源可能已过时，尝试更新软件源",
			Commands: []string{
				"chopsticks bucket update " + bucket,
			},
			AutoFixable: true,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "搜索其他软件源",
			Description: "在其他软件源中查找该软件",
			Commands: []string{
				"chopsticks search " + app,
			},
			AutoFixable: false,
		})
}

// NewNetworkConnection 创建网络连接失败错误
func NewNetworkConnection(host string, err error) *StructuredError {
	return WrapStructured(err, ErrNetworkConnection, fmt.Sprintf("网络连接失败：%s", host)).
		WithContext("host", host).
		WithOperation("network connect").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查网络连接",
			Description: "测试是否能访问目标主机",
			Commands: []string{
				"ping " + host,
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "配置代理",
			Description: "如果公司网络需要代理，请配置代理设置",
			Commands: []string{
				"chopsticks config set proxy.http http://proxy.company.com:8080",
				"chopsticks config set proxy.https https://proxy.company.com:8080",
			},
			AutoFixable: false,
		})
}

// NewDownloadTimeout 创建下载超时错误
func NewDownloadTimeout(url string, timeout int) *StructuredError {
	return NewStructured(ErrDownloadTimeout, fmt.Sprintf("下载超时：%s (超时：%d 秒)", url, timeout)).
		WithContext("url", url).
		WithContext("timeout_seconds", timeout).
		WithOperation("download").
		WithRetryable(true).
		WithSuggestion(RecoverySuggestion{
			Title:       "增加超时时间",
			Description: "增加网络超时时间设置",
			Commands: []string{
				"chopsticks config set network.timeout 300",
			},
			AutoFixable: true,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "使用镜像源",
			Description: "使用更快的镜像源下载",
			Commands: []string{
				"chopsticks bucket add main-mirror https://mirror.example.com/main",
			},
			AutoFixable: false,
		})
}

// NewInvalidURL 创建 URL 无效错误
func NewInvalidURL(url, reason string) *StructuredError {
	return NewStructured(ErrInvalidURL, fmt.Sprintf("URL 无效：%s - %s", url, reason)).
		WithContext("url", url).
		WithContext("reason", reason).
		WithOperation("url validation")
}

// NewProxyError 创建代理错误
func NewProxyError(proxyURL string, err error) *StructuredError {
	return WrapStructured(err, ErrProxyError, fmt.Sprintf("代理错误：%s", proxyURL)).
		WithContext("proxy_url", proxyURL).
		WithOperation("proxy connect").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查代理配置",
			Description: "验证代理地址和端口是否正确",
			Commands: []string{
				"chopsticks config list",
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "禁用代理",
			Description: "如果不需要代理，可以禁用代理设置",
			Commands: []string{
				"chopsticks config set proxy.http \"\"",
				"chopsticks config set proxy.https \"\"",
			},
			AutoFixable: true,
		})
}

// NewSSLVerificationError 创建 SSL 验证失败错误
func NewSSLVerificationError(url string, err error) *StructuredError {
	return WrapStructured(err, ErrSSLVerification, fmt.Sprintf("SSL 验证失败：%s", url)).
		WithContext("url", url).
		WithOperation("ssl verify").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查系统时间",
			Description: "系统时间不正确可能导致 SSL 验证失败",
			Commands:    []string{},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "检查证书",
			Description: "目标网站的 SSL 证书可能已过期或无效",
			Commands:    []string{},
			AutoFixable: false,
		})
}

// NewConfigNotFound 创建配置文件不存在错误
func NewConfigNotFound(path string) *StructuredError {
	return NewStructured(ErrConfigNotFound, fmt.Sprintf("配置文件不存在：%s", path)).
		WithContext("config_path", path).
		WithOperation("config load").
		WithSuggestion(RecoverySuggestion{
			Title:       "初始化配置",
			Description: "运行初始化命令创建默认配置文件",
			Commands: []string{
				"chopsticks config init",
			},
			AutoFixable: true,
		})
}

// NewConfigInvalid 创建配置文件无效错误
func NewConfigInvalid(path, reason string) *StructuredError {
	return NewStructured(ErrConfigInvalid, fmt.Sprintf("配置文件无效：%s - %s", path, reason)).
		WithContext("config_path", path).
		WithContext("reason", reason).
		WithOperation("config validate").
		WithSuggestion(RecoverySuggestion{
			Title:       "重置配置",
			Description: "重置为默认配置",
			Commands: []string{
				"chopsticks config reset",
			},
			AutoFixable: true,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "手动编辑配置文件",
			Description: "手动修复配置文件格式",
			Commands: []string{
				`notepad ` + path,
			},
			AutoFixable: false,
		})
}

// NewConfigReadFailed 创建读取配置失败错误
func NewConfigReadFailed(path string, err error) *StructuredError {
	return WrapStructured(err, ErrConfigReadFailed, fmt.Sprintf("读取配置失败：%s", path)).
		WithContext("config_path", path).
		WithOperation("config read")
}

// NewConfigWriteFailed 创建写入配置失败错误
func NewConfigWriteFailed(path string, err error) *StructuredError {
	return WrapStructured(err, ErrConfigWriteFailed, fmt.Sprintf("写入配置失败：%s", path)).
		WithContext("config_path", path).
		WithOperation("config write").
		WithSuggestion(RecoverySuggestion{
			Title:       "检查文件权限",
			Description: "确保对配置文件有写入权限",
			Commands:    []string{},
			AutoFixable: false,
		})
}

// NewConfigValueInvalid 创建配置值无效错误
func NewConfigValueInvalid(key, value, reason string) *StructuredError {
	return NewStructured(ErrConfigValueInvalid, fmt.Sprintf("配置值无效：%s = %s - %s", key, value, reason)).
		WithContext("config_key", key).
		WithContext("config_value", value).
		WithContext("reason", reason).
		WithOperation("config validate")
}

// NewInternalError 创建内部错误
func NewInternalError(operation string, err error) *StructuredError {
	return WrapStructured(err, ErrInternal, fmt.Sprintf("内部错误：%s", operation)).
		WithContext("operation", operation).
		WithOperation("internal").
		WithRecoverable(false).
		WithSuggestion(RecoverySuggestion{
			Title:       "查看详细日志",
			Description: "使用详细模式获取更多信息以便排查问题",
			Commands: []string{
				"chopsticks " + operation + " --verbose",
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "报告问题",
			Description: "如果问题持续，请在 GitHub 上报告此问题",
			Commands:    []string{},
			AutoFixable: false,
		})
}

// NewUnknownError 创建未知错误
func NewUnknownError(err error) *StructuredError {
	return WrapStructured(err, ErrUnknown, "未知错误").
		WithOperation("unknown").
		WithSuggestion(RecoverySuggestion{
			Title:       "查看详细日志",
			Description: "使用详细模式获取更多信息",
			Commands: []string{
				"chopsticks {command} --verbose",
			},
			AutoFixable: false,
		}).
		WithSuggestion(RecoverySuggestion{
			Title:       "运行诊断",
			Description: "运行系统诊断检查",
			Commands: []string{
				"chopsticks doctor",
			},
			AutoFixable: false,
		})
}

// NewCancelled 创建操作取消错误
func NewCancelled(operation, reason string) *StructuredError {
	return NewStructured(ErrCancelled, fmt.Sprintf("操作已取消：%s - %s", operation, reason)).
		WithContext("operation", operation).
		WithContext("reason", reason).
		WithOperation(operation).
		WithRecoverable(false)
}

// NewTimeout 创建操作超时错误
func NewTimeout(operation string, timeout int) *StructuredError {
	return NewStructured(ErrTimeout, fmt.Sprintf("操作超时：%s (超时：%d 秒)", operation, timeout)).
		WithContext("operation", operation).
		WithContext("timeout_seconds", timeout).
		WithOperation(operation).
		WithRetryable(true).
		WithSuggestion(RecoverySuggestion{
			Title:       "增加超时时间",
			Description: "增加操作的超时时间设置",
			Commands:    []string{},
			AutoFixable: true,
		})
}

// NewNotSupported 创建不支持的操作错误
func NewNotSupported(operation string) *StructuredError {
	return NewStructured(ErrNotSupported, fmt.Sprintf("不支持的操作：%s", operation)).
		WithContext("operation", operation).
		WithOperation(operation).
		WithRecoverable(false)
}

// NewInvalidInput 创建输入无效错误
func NewInvalidInput(message string) *StructuredError {
	return NewStructured(ErrInvalidInput, fmt.Sprintf("输入无效：%s", message)).
		WithOperation("input validation")
}
