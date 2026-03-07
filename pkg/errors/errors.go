package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 1xxx - 系统错误
	ErrPermissionDenied    ErrorCode = "CHP-1001" // 权限不足
	ErrInsufficientDisk    ErrorCode = "CHP-1002" // 磁盘空间不足
	ErrFileNotFound        ErrorCode = "CHP-1003" // 文件不存在
	ErrFileAlreadyExists   ErrorCode = "CHP-1004" // 文件已存在
	ErrInvalidPath         ErrorCode = "CHP-1005" // 路径无效
	ErrDirectoryNotEmpty   ErrorCode = "CHP-1006" // 目录非空

	// 2xxx - 网络错误
	ErrNetworkConnection   ErrorCode = "CHP-2001" // 网络连接失败
	ErrDownloadTimeout     ErrorCode = "CHP-2002" // 下载超时
	ErrDownloadFailed      ErrorCode = "CHP-2003" // 下载失败
	ErrInvalidURL          ErrorCode = "CHP-2004" // URL 无效
	ErrProxyError          ErrorCode = "CHP-2005" // 代理错误
	ErrSSLVerification     ErrorCode = "CHP-2006" // SSL 验证失败

	// 3xxx - 软件源错误
	ErrBucketNotFound      ErrorCode = "CHP-3001" // 软件源不存在
	ErrBucketAlreadyExists ErrorCode = "CHP-3002" // 软件源已存在
	ErrBucketLoadFailed    ErrorCode = "CHP-3003" // 软件源加载失败
	ErrBucketUpdateFailed  ErrorCode = "CHP-3004" // 软件源更新失败
	ErrInvalidBucketURL    ErrorCode = "CHP-3005" // 软件源 URL 无效
	ErrManifestNotFound    ErrorCode = "CHP-3006" // 清单文件不存在

	// 4xxx - 安装错误
	ErrAppNotFound         ErrorCode = "CHP-4001" // 软件未找到
	ErrAppAlreadyInstalled ErrorCode = "CHP-4002" // 软件已安装
	ErrAppNotInstalled     ErrorCode = "CHP-4003" // 软件未安装
	ErrInstallFailed       ErrorCode = "CHP-4004" // 安装失败
	ErrUninstallFailed     ErrorCode = "CHP-4005" // 卸载失败
	ErrUpdateFailed        ErrorCode = "CHP-4006" // 更新失败
	ErrVersionNotFound     ErrorCode = "CHP-4007" // 版本未找到
	ErrVersionAlreadyExists ErrorCode = "CHP-4008" // 版本已存在
	ErrChecksumMismatch    ErrorCode = "CHP-4009" // 校验和不匹配
	ErrArchiveExtractFailed ErrorCode = "CHP-4010" // 解压失败
	ErrScriptFailed        ErrorCode = "CHP-4011" // 脚本执行失败
	ErrHookFailed          ErrorCode = "CHP-4012" // 钩子执行失败
	ErrInstallCancelled    ErrorCode = "CHP-4013" // 安装已取消

	// 5xxx - 依赖错误
	ErrDependencyConflict  ErrorCode = "CHP-5001" // 依赖冲突
	ErrDependencyNotFound  ErrorCode = "CHP-5002" // 依赖未找到
	ErrCircularDependency  ErrorCode = "CHP-5003" // 循环依赖
	ErrDependencyVersion   ErrorCode = "CHP-5004" // 依赖版本不匹配
	ErrOrphanDependency    ErrorCode = "CHP-5005" // 孤立依赖

	// 6xxx - 配置错误
	ErrConfigNotFound      ErrorCode = "CHP-6001" // 配置文件不存在
	ErrConfigInvalid       ErrorCode = "CHP-6002" // 配置文件无效
	ErrConfigReadFailed    ErrorCode = "CHP-6003" // 读取配置失败
	ErrConfigWriteFailed   ErrorCode = "CHP-6004" // 写入配置失败
	ErrConfigValueInvalid  ErrorCode = "CHP-6005" // 配置值无效

	// 9xxx - 其他错误
	ErrUnknown             ErrorCode = "CHP-9001" // 未知错误
	ErrInternal            ErrorCode = "CHP-9002" // 内部错误
	ErrCancelled           ErrorCode = "CHP-9003" // 操作已取消
	ErrTimeout             ErrorCode = "CHP-9004" // 操作超时
	ErrNotSupported        ErrorCode = "CHP-9005" // 不支持的操作
	ErrInvalidInput        ErrorCode = "CHP-9006" // 输入无效
)

// ErrorCategory 错误分类
type ErrorCategory string

const (
	CategorySystem     ErrorCategory = "system"     // 系统错误
	CategoryNetwork    ErrorCategory = "network"    // 网络错误
	CategoryBucket     ErrorCategory = "bucket"     // 软件源错误
	CategoryInstall    ErrorCategory = "install"    // 安装错误
	CategoryDependency ErrorCategory = "dependency" // 依赖错误
	CategoryConfig     ErrorCategory = "config"     // 配置错误
	CategoryOther      ErrorCategory = "other"      // 其他错误
)

// RecoverySuggestion 恢复建议
type RecoverySuggestion struct {
	Title       string   `json:"title"`        // 建议标题
	Description string   `json:"description"`  // 详细描述
	Commands    []string `json:"commands"`     // 推荐命令
	DocsURL     string   `json:"docs_url"`     // 相关文档链接
	AutoFixable bool     `json:"auto_fixable"` // 是否可自动修复
}

// ErrorDetail 错误详细信息
type ErrorDetail struct {
	Context     map[string]interface{} `json:"context,omitempty"`      // 错误上下文
	Timestamp   time.Time              `json:"timestamp"`              // 错误发生时间
	Operation   string                 `json:"operation,omitempty"`    // 执行的操作
	UserAction  string                 `json:"user_action,omitempty"`  // 用户操作
	SystemState string                 `json:"system_state,omitempty"` // 系统状态
}

// StructuredError 结构化错误
type StructuredError struct {
	Code           ErrorCode         `json:"code"`            // 错误码
	Message        string            `json:"message"`         // 错误消息
	Category       ErrorCategory     `json:"category"`        // 错误分类
	Cause          error             `json:"-"`               // 根本原因（不序列化）
	Suggestions    []RecoverySuggestion `json:"suggestions"`  // 恢复建议
	Details        ErrorDetail       `json:"details"`         // 详细信息
	StackTrace     []string          `json:"stack_trace"`     // 堆栈跟踪
	Recoverable    bool              `json:"recoverable"`     // 是否可恢复
	Retryable      bool              `json:"retryable"`       // 是否可重试
}

// Error 实现 error 接口
func (e *StructuredError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// Is 实现 errors.Is 接口
func (e *StructuredError) Is(target error) bool {
	if te, ok := target.(*StructuredError); ok {
		return e.Code == te.Code
	}
	return false
}

// MarshalJSON 实现 json.Marshaler 接口
func (e *StructuredError) MarshalJSON() ([]byte, error) {
	type Alias StructuredError
	cause := ""
	if e.Cause != nil {
		cause = e.Cause.Error()
	}
	return json.Marshal(&struct {
		*Alias
		Cause string `json:"cause,omitempty"`
	}{
		Alias: (*Alias)(e),
		Cause: cause,
	})
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (e *StructuredError) UnmarshalJSON(data []byte) error {
	type Alias StructuredError
	aux := &struct {
		*Alias
		Cause string `json:"cause,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Cause != "" {
		e.Cause = errors.New(aux.Cause)
	}
	return nil
}

// WithCause 设置根本原因
func (e *StructuredError) WithCause(cause error) *StructuredError {
	e.Cause = cause
	return e
}

// WithContext 添加上下文信息
func (e *StructuredError) WithContext(key string, value interface{}) *StructuredError {
	if e.Details.Context == nil {
		e.Details.Context = make(map[string]interface{})
	}
	e.Details.Context[key] = value
	return e
}

// WithOperation 设置操作信息
func (e *StructuredError) WithOperation(op string) *StructuredError {
	e.Details.Operation = op
	return e
}

// WithSuggestion 添加恢复建议
func (e *StructuredError) WithSuggestion(suggestion RecoverySuggestion) *StructuredError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithRetryable 设置是否可重试
func (e *StructuredError) WithRetryable(retryable bool) *StructuredError {
	e.Retryable = retryable
	return e
}

// WithRecoverable 设置是否可恢复
func (e *StructuredError) WithRecoverable(recoverable bool) *StructuredError {
	e.Recoverable = recoverable
	return e
}

// GetCategoryByCode 根据错误码获取错误分类
func GetCategoryByCode(code ErrorCode) ErrorCategory {
	codeStr := string(code)
	switch {
	case strings.HasPrefix(codeStr, "CHP-1"):
		return CategorySystem
	case strings.HasPrefix(codeStr, "CHP-2"):
		return CategoryNetwork
	case strings.HasPrefix(codeStr, "CHP-3"):
		return CategoryBucket
	case strings.HasPrefix(codeStr, "CHP-4"):
		return CategoryInstall
	case strings.HasPrefix(codeStr, "CHP-5"):
		return CategoryDependency
	case strings.HasPrefix(codeStr, "CHP-6"):
		return CategoryConfig
	default:
		return CategoryOther
	}
}

// GetDefaultSuggestions 根据错误码获取默认恢复建议
func GetDefaultSuggestions(code ErrorCode) []RecoverySuggestion {
	switch code {
	case ErrPermissionDenied:
		return []RecoverySuggestion{
			{
				Title:       "以管理员身份运行",
				Description: "右键点击终端，选择'以管理员身份运行'，然后重试操作",
				Commands:    []string{},
				AutoFixable: false,
			},
			{
				Title:       "修改安装目录到用户目录",
				Description: "将安装目录配置到用户目录下，避免需要管理员权限",
				Commands:    []string{`chopsticks config set install_dir "%USERPROFILE%\tools"`},
				AutoFixable: true,
			},
		}

	case ErrInsufficientDisk:
		return []RecoverySuggestion{
			{
				Title:       "清理缓存",
				Description: "清理下载缓存释放磁盘空间",
				Commands:    []string{"chopsticks cache clean"},
				AutoFixable: true,
			},
			{
				Title:       "检查磁盘空间",
				Description: "查看当前磁盘使用情况",
				Commands:    []string{"chopsticks doctor"},
				AutoFixable: false,
			},
		}

	case ErrNetworkConnection:
		return []RecoverySuggestion{
			{
				Title:       "检查网络连接",
				Description: "测试是否能访问 GitHub",
				Commands:    []string{"ping github.com"},
				AutoFixable: false,
			},
			{
				Title:       "配置代理",
				Description: "如果公司网络需要代理，请配置代理设置",
				Commands:    []string{
					"chopsticks config set proxy.http http://proxy.company.com:8080",
					"chopsticks config set proxy.https https://proxy.company.com:8080",
				},
				AutoFixable: false,
			},
		}

	case ErrDownloadTimeout:
		return []RecoverySuggestion{
			{
				Title:       "增加超时时间",
				Description: "增加网络超时时间设置",
				Commands:    []string{"chopsticks config set network.timeout 300"},
				AutoFixable: true,
			},
			{
				Title:       "使用镜像源",
				Description: "使用更快的镜像源下载",
				Commands:    []string{"chopsticks bucket add main-mirror https://mirror.example.com/main"},
				AutoFixable: false,
			},
		}

	case ErrBucketNotFound:
		return []RecoverySuggestion{
			{
				Title:       "查看已添加的软件源",
				Description: "检查软件源是否正确添加",
				Commands:    []string{"chopsticks bucket list"},
				AutoFixable: false,
			},
			{
				Title:       "添加软件源",
				Description: "添加主软件源",
				Commands:    []string{"chopsticks bucket add main https://github.com/chopsticks-bucket/main"},
				AutoFixable: true,
			},
		}

	case ErrAppNotFound:
		return []RecoverySuggestion{
			{
				Title:       "搜索软件",
				Description: "使用关键词搜索可用的软件",
				Commands:    []string{"chopsticks search {keyword}"},
				AutoFixable: false,
			},
			{
				Title:       "查看所有可用软件",
				Description: "列出所有可安装的软件",
				Commands:    []string{"chopsticks list --all"},
				AutoFixable: false,
			},
		}

	case ErrAppAlreadyInstalled:
		return []RecoverySuggestion{
			{
				Title:       "查看已安装版本",
				Description: "查看当前安装的版本信息",
				Commands:    []string{"chopsticks list"},
				AutoFixable: false,
			},
			{
				Title:       "强制重新安装",
				Description: "使用 --force 参数强制重新安装",
				Commands:    []string{"chopsticks install {package} --force"},
				AutoFixable: false,
			},
		}

	case ErrDependencyConflict:
		return []RecoverySuggestion{
			{
				Title:       "查看依赖冲突详情",
				Description: "使用详细模式查看依赖冲突信息",
				Commands:    []string{"chopsticks install {package} --verbose"},
				AutoFixable: false,
			},
			{
				Title:       "清理依赖缓存",
				Description: "清理依赖解析缓存后重试",
				Commands:    []string{"chopsticks deps clean"},
				AutoFixable: true,
			},
		}

	case ErrChecksumMismatch:
		return []RecoverySuggestion{
			{
				Title:       "清理缓存重新下载",
				Description: "清理下载缓存并重新下载安装",
				Commands:    []string{"chopsticks cache clean --package {package}", "chopsticks install {package}"},
				AutoFixable: true,
			},
		}

	case ErrConfigInvalid:
		return []RecoverySuggestion{
			{
				Title:       "重置配置",
				Description: "重置为默认配置",
				Commands:    []string{"chopsticks config reset"},
				AutoFixable: true,
			},
			{
				Title:       "手动编辑配置文件",
				Description: "手动修复配置文件格式",
				Commands:    []string{`notepad %USERPROFILE%\.chopsticks\config.yaml`},
				AutoFixable: false,
			},
		}

	default:
		return []RecoverySuggestion{
			{
				Title:       "查看详细日志",
				Description: "使用详细模式获取更多信息",
				Commands:    []string{"chopsticks {command} --verbose"},
				AutoFixable: false,
			},
			{
				Title:       "运行诊断",
				Description: "运行系统诊断检查",
				Commands:    []string{"chopsticks doctor"},
				AutoFixable: false,
			},
		}
	}
}

// NewStructured 创建结构化错误
func NewStructured(code ErrorCode, message string) *StructuredError {
	return &StructuredError{
		Code:        code,
		Message:     message,
		Category:    GetCategoryByCode(code),
		Suggestions: GetDefaultSuggestions(code),
		Details: ErrorDetail{
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
		},
		Recoverable: true,
		Retryable:   isRetryable(code),
	}
}

// isRetryable 判断错误码是否可重试
func isRetryable(code ErrorCode) bool {
	retryableCodes := map[ErrorCode]bool{
		ErrNetworkConnection: true,
		ErrDownloadTimeout:   true,
		ErrDownloadFailed:    true,
		ErrProxyError:        true,
		ErrBucketUpdateFailed: true,
	}
	return retryableCodes[code]
}

// WrapStructured 包装现有错误为结构化错误
func WrapStructured(err error, code ErrorCode, operation string) *StructuredError {
	if err == nil {
		return nil
	}

	return &StructuredError{
		Code:        code,
		Message:     fmt.Sprintf("%s: %v", operation, err),
		Category:    GetCategoryByCode(code),
		Cause:       err,
		Suggestions: GetDefaultSuggestions(code),
		Details: ErrorDetail{
			Timestamp: time.Now(),
			Operation: operation,
			Context:   make(map[string]interface{}),
		},
		Recoverable: true,
		Retryable:   isRetryable(code),
	}
}

// WrapStructuredf 包装现有错误为结构化错误（支持格式化）
func WrapStructuredf(err error, code ErrorCode, format string, args ...interface{}) *StructuredError {
	if err == nil {
		return nil
	}

	return &StructuredError{
		Code:        code,
		Message:     fmt.Sprintf(format+": %v", append(args, err)...),
		Category:    GetCategoryByCode(code),
		Cause:       err,
		Suggestions: GetDefaultSuggestions(code),
		Details: ErrorDetail{
			Timestamp: time.Now(),
			Operation: fmt.Sprintf(format, args...),
			Context:   make(map[string]interface{}),
		},
		Recoverable: true,
		Retryable:   isRetryable(code),
	}
}

// GetErrorCode 从错误中获取错误码
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return ""
	}

	var se *StructuredError
	if errors.As(err, &se) {
		return se.Code
	}

	return ErrUnknown
}

// GetErrorCategory 从错误中获取错误分类
func GetErrorCategory(err error) ErrorCategory {
	if err == nil {
		return CategoryOther
	}

	var se *StructuredError
	if errors.As(err, &se) {
		return se.Category
	}

	return CategoryOther
}

// GetRecoverySuggestions 从错误中获取恢复建议
func GetRecoverySuggestions(err error) []RecoverySuggestion {
	if err == nil {
		return nil
	}

	var se *StructuredError
	if errors.As(err, &se) {
		return se.Suggestions
	}

	code := GetErrorCode(err)
	if code != "" {
		return GetDefaultSuggestions(code)
	}

	return nil
}

// FormatError 格式化错误输出
func FormatError(err error, verbose bool) string {
	if err == nil {
		return ""
	}

	var sb strings.Builder

	var se *StructuredError
	if errors.As(err, &se) {
		sb.WriteString(fmt.Sprintf("错误代码：%s\n", se.Code))
		sb.WriteString(fmt.Sprintf("错误分类：%s\n", se.Category))
		sb.WriteString(fmt.Sprintf("错误消息：%s\n", se.Message))

		if verbose {
			if se.Cause != nil {
				sb.WriteString(fmt.Sprintf("根本原因：%v\n", se.Cause))
			}
			if se.Details.Operation != "" {
				sb.WriteString(fmt.Sprintf("执行操作：%s\n", se.Details.Operation))
			}
			if len(se.Details.Context) > 0 {
				sb.WriteString("上下文信息:\n")
				keys := make([]string, 0, len(se.Details.Context))
				for k := range se.Details.Context {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					sb.WriteString(fmt.Sprintf("  - %s: %v\n", k, se.Details.Context[k]))
				}
			}
		}

		if len(se.Suggestions) > 0 {
			sb.WriteString("\n恢复建议:\n")
			for i, sug := range se.Suggestions {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, sug.Title))
				if sug.Description != "" {
					sb.WriteString(fmt.Sprintf("     %s\n", sug.Description))
				}
				if len(sug.Commands) > 0 {
					sb.WriteString("     推荐命令:\n")
					for _, cmd := range sug.Commands {
						sb.WriteString(fmt.Sprintf("       > %s\n", cmd))
					}
				}
				if sug.AutoFixable {
					sb.WriteString("     [可自动修复]\n")
				}
			}
		}

		if se.Retryable {
			sb.WriteString("\n[此错误支持重试]\n")
		}
	} else {
		sb.WriteString(fmt.Sprintf("错误：%v\n", err))
	}

	return sb.String()
}

// ToJSON 将错误转换为 JSON 格式
func ToJSON(err error) (string, error) {
	if err == nil {
		return "", nil
	}

	data, err := json.MarshalIndent(err, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FromJSON 从 JSON 格式解析错误
func FromJSON(data []byte) (*StructuredError, error) {
	var se StructuredError
	if err := json.Unmarshal(data, &se); err != nil {
		return nil, err
	}
	return &se, nil
}

// ErrorSummary 错误摘要信息
type ErrorSummary struct {
	TotalErrors   int                     `json:"total_errors"`
	ByCategory    map[ErrorCategory]int   `json:"by_category"`
	ByCode        map[ErrorCode]int       `json:"by_code"`
	MostFrequent  []ErrorCode             `json:"most_frequent"`
	Recoverable   int                     `json:"recoverable"`
	Unrecoverable int                     `json:"unrecoverable"`
}

// AnalyzeErrors 分析错误列表
func AnalyzeErrors(errs []error) *ErrorSummary {
	summary := &ErrorSummary{
		ByCategory: make(map[ErrorCategory]int),
		ByCode:     make(map[ErrorCode]int),
	}

	codeCount := make(map[ErrorCode]int)

	for _, err := range errs {
		if err == nil {
			continue
		}

		summary.TotalErrors++

		var se *StructuredError
		if errors.As(err, &se) {
			summary.ByCategory[se.Category]++
			summary.ByCode[se.Code]++
			codeCount[se.Code]++

			if se.Recoverable {
				summary.Recoverable++
			} else {
				summary.Unrecoverable++
			}
		} else {
			summary.ByCategory[CategoryOther]++
			summary.ByCode[ErrUnknown]++
			codeCount[ErrUnknown]++
			summary.Unrecoverable++
		}
	}

	// 计算最常见的错误
	for code, count := range codeCount {
		if count >= 2 {
			summary.MostFrequent = append(summary.MostFrequent, code)
		}
	}

	sort.Slice(summary.MostFrequent, func(i, j int) bool {
		return codeCount[summary.MostFrequent[i]] > codeCount[summary.MostFrequent[j]]
	})

	return summary
}

// ============================================================================
// 向后兼容的旧 API - 用于保持与现有代码的兼容性
// ============================================================================

// 保留旧的 ErrorKind 类型和常量
type ErrorKind int

const (
	KindUnknown ErrorKind = iota
	KindNotFound
	KindAlreadyExists
	KindInvalidInput
	KindPermission
	KindNetwork
	KindIO
	KindExec
	KindCancelled
	KindTimeout
	KindConflict
)

// 保留旧的错误常量（使用不同名称避免冲突）
var (
	ErrLegacyNotFound         = fmt.Errorf("not found")
	ErrLegacyAlreadyExists    = fmt.Errorf("already exists")
	ErrLegacyInvalidInput     = fmt.Errorf("invalid input")
	ErrLegacyPermissionDenied = fmt.Errorf("permission denied")
	ErrLegacyCancelled        = fmt.Errorf("operation cancelled")
	ErrLegacyTimeout          = fmt.Errorf("operation timeout")
	ErrLegacyNotSupported     = fmt.Errorf("not supported")
	ErrLegacyInternal         = fmt.Errorf("internal error")
)

// LegacyError 保留旧的错误结构
type LegacyError struct {
	Op   string
	Err  error
	Kind ErrorKind
	Key  string
}

func (e *LegacyError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Err.Error()
}

func (e *LegacyError) Unwrap() error {
	return e.Err
}

// New 创建旧式错误
func New(kind ErrorKind, msg string) error {
	return &LegacyError{
		Err:  errors.New(msg),
		Kind: kind,
	}
}

// Newf 创建旧式格式化错误
func Newf(kind ErrorKind, format string, args ...interface{}) error {
	return &LegacyError{
		Err:  fmt.Errorf(format, args...),
		Kind: kind,
	}
}

// Wrap 包装旧式错误
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &LegacyError{
		Op:  op,
		Err: err,
	}
}

// Wrapf 包装旧式格式化错误
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &LegacyError{
		Op:  fmt.Sprintf(format, args...),
		Err: err,
	}
}

// WrapWithKind 包装旧式错误并设置类型
func WrapWithKind(err error, op string, kind ErrorKind) error {
	if err == nil {
		return nil
	}
	return &LegacyError{
		Op:   op,
		Err:  err,
		Kind: kind,
	}
}

// Is 检查错误是否匹配
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 转换错误类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap 解包错误
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// GetKind 获取错误类型
func GetKind(err error) ErrorKind {
	if err == nil {
		return KindUnknown
	}
	var e *LegacyError
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindUnknown
}

// IsKind 检查错误类型
func IsKind(err error, kind ErrorKind) bool {
	return GetKind(err) == kind
}
