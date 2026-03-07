package errors

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected string
	}{
		{"ErrPermissionDenied", ErrPermissionDenied, "CHP-1001"},
		{"ErrAppNotFound", ErrAppNotFound, "CHP-4001"},
		{"ErrBucketNotFound", ErrBucketNotFound, "CHP-3001"},
		{"ErrDependencyConflict", ErrDependencyConflict, "CHP-5001"},
		{"ErrUnknown", ErrUnknown, "CHP-9001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code) != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.code, tt.expected)
			}
		})
	}
}

func TestErrorCategoryConstants(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected string
	}{
		{"CategorySystem", CategorySystem, "system"},
		{"CategoryNetwork", CategoryNetwork, "network"},
		{"CategoryBucket", CategoryBucket, "bucket"},
		{"CategoryInstall", CategoryInstall, "install"},
		{"CategoryDependency", CategoryDependency, "dependency"},
		{"CategoryConfig", CategoryConfig, "config"},
		{"CategoryOther", CategoryOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.category, tt.expected)
			}
		})
	}
}

func TestGetCategoryByCode(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected ErrorCategory
	}{
		{"system error", ErrPermissionDenied, CategorySystem},
		{"network error", ErrNetworkConnection, CategoryNetwork},
		{"bucket error", ErrBucketNotFound, CategoryBucket},
		{"install error", ErrAppNotFound, CategoryInstall},
		{"dependency error", ErrDependencyConflict, CategoryDependency},
		{"config error", ErrConfigInvalid, CategoryConfig},
		{"other error", ErrUnknown, CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategoryByCode(tt.code)
			if got != tt.expected {
				t.Errorf("GetCategoryByCode(%v) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

func TestStructuredError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StructuredError
		expected string
	}{
		{
			name: "error without cause",
			err:  NewStructured(ErrAppNotFound, "软件未找到"),
			expected: "[CHP-4001] 软件未找到",
		},
		{
			name: "error with cause",
			err:  NewStructured(ErrDownloadFailed, "下载失败").WithCause(errors.New("network timeout")),
			expected: "CHP-2003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.expected) {
				t.Errorf("Error() = %v, want to contain %v", got, tt.expected)
			}
		})
	}
}

func TestStructuredError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewStructured(ErrDownloadFailed, "下载失败").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestStructuredError_Is(t *testing.T) {
	err1 := NewStructured(ErrAppNotFound, "软件未找到")
	err2 := NewStructured(ErrAppNotFound, "另一个软件未找到")
	err3 := NewStructured(ErrBucketNotFound, "软件源未找到")

	tests := []struct {
		name     string
		err      *StructuredError
		target   error
		expected bool
	}{
		{"same code", err1, err2, true},
		{"different code", err1, err3, false},
		{"nil target", err1, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Is(tt.target)
			if got != tt.expected {
				t.Errorf("Is() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStructuredError_MarshalJSON(t *testing.T) {
	testErr := NewStructured(ErrAppNotFound, "软件未找到").
		WithContext("app_name", "test-app").
		WithOperation("install check")

	data, marshalErr := json.Marshal(testErr)
	if marshalErr != nil {
		t.Fatalf("Marshal failed: %v", marshalErr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["code"] != "CHP-4001" {
		t.Errorf("code = %v, want CHP-4001", result["code"])
	}
	if result["category"] != "install" {
		t.Errorf("category = %v, want install", result["category"])
	}
}

func TestStructuredError_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"code": "CHP-4001",
		"message": "软件未找到",
		"category": "install",
		"recoverable": true,
		"retryable": false,
		"suggestions": [
			{
				"title": "搜索软件",
				"description": "使用关键词搜索",
				"commands": ["chopsticks search test"],
				"auto_fixable": false
			}
		],
		"details": {
			"timestamp": "2026-03-06T00:00:00Z",
			"operation": "install check",
			"context": {"app_name": "test-app"}
		}
	}`

	parsedErr, unmarshalErr := FromJSON([]byte(jsonData))
	if unmarshalErr != nil {
		t.Fatalf("Unmarshal failed: %v", unmarshalErr)
	}

	if parsedErr.Code != ErrAppNotFound {
		t.Errorf("Code = %v, want CHP-4001", parsedErr.Code)
	}
	if parsedErr.Category != CategoryInstall {
		t.Errorf("Category = %v, want install", parsedErr.Category)
	}
	if !parsedErr.Recoverable {
		t.Errorf("Recoverable = false, want true")
	}
}

func TestStructuredError_WithContext(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithContext("app_name", "test-app").
		WithContext("version", "1.0.0")

	if err.Details.Context["app_name"] != "test-app" {
		t.Errorf("Context app_name = %v, want test-app", err.Details.Context["app_name"])
	}
	if err.Details.Context["version"] != "1.0.0" {
		t.Errorf("Context version = %v, want 1.0.0", err.Details.Context["version"])
	}
}

func TestStructuredError_WithOperation(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithOperation("install check")

	if err.Details.Operation != "install check" {
		t.Errorf("Operation = %v, want install check", err.Details.Operation)
	}
}

func TestStructuredError_WithSuggestion(t *testing.T) {
	suggestion := RecoverySuggestion{
		Title:       "测试建议",
		Description: "这是一个测试建议",
		Commands:    []string{"test command"},
		AutoFixable: true,
	}

	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithSuggestion(suggestion)

	if len(err.Suggestions) != 3 { // 默认 2 个 + 自定义 1 个
		t.Errorf("Suggestions count = %v, want 3", len(err.Suggestions))
	}
}

func TestStructuredError_WithRetryable(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithRetryable(true)

	if !err.Retryable {
		t.Errorf("Retryable = false, want true")
	}
}

func TestStructuredError_WithRecoverable(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithRecoverable(false)

	if err.Recoverable {
		t.Errorf("Recoverable = true, want false")
	}
}

func TestWrapStructured(t *testing.T) {
	cause := errors.New("network timeout")
	err := WrapStructured(cause, ErrDownloadFailed, "download package")

	if err.Code != ErrDownloadFailed {
		t.Errorf("Code = %v, want CHP-2003", err.Code)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want network timeout", err.Cause)
	}
	if err.Details.Operation != "download package" {
		t.Errorf("Operation = %v, want download package", err.Details.Operation)
	}
}

func TestWrapStructuredf(t *testing.T) {
	cause := errors.New("network timeout")
	err := WrapStructuredf(cause, ErrDownloadFailed, "download %s from %s", "package", "https://example.com")

	if err.Code != ErrDownloadFailed {
		t.Errorf("Code = %v, want CHP-2003", err.Code)
	}
	if !strings.Contains(err.Message, "download package from https://example.com") {
		t.Errorf("Message = %v, want to contain download package from https://example.com", err.Message)
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{"structured error", NewStructured(ErrAppNotFound, "test"), ErrAppNotFound},
		{"wrapped error", WrapStructured(errors.New("cause"), ErrDownloadFailed, "op"), ErrDownloadFailed},
		{"nil error", nil, ""},
		{"standard error", errors.New("standard error"), ErrUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorCode(tt.err)
			if got != tt.expected {
				t.Errorf("GetErrorCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetErrorCategory(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{"install error", NewStructured(ErrAppNotFound, "test"), CategoryInstall},
		{"network error", NewStructured(ErrNetworkConnection, "test"), CategoryNetwork},
		{"nil error", nil, CategoryOther},
		{"standard error", errors.New("standard"), CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorCategory(tt.err)
			if got != tt.expected {
				t.Errorf("GetErrorCategory() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetRecoverySuggestions(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到")
	suggestions := GetRecoverySuggestions(err)

	if len(suggestions) < 1 {
		t.Errorf("GetRecoverySuggestions() returned %d suggestions, want >= 1", len(suggestions))
	}
}

func TestFormatError(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "软件未找到").
		WithContext("app_name", "test-app").
		WithOperation("install check")

	formatted := FormatError(err, true)

	if !strings.Contains(formatted, "CHP-4001") {
		t.Errorf("FormatError() = %v, want to contain CHP-4001", formatted)
	}
	if !strings.Contains(formatted, "install") {
		t.Errorf("FormatError() = %v, want to contain install", formatted)
	}
	if !strings.Contains(formatted, "恢复建议") {
		t.Errorf("FormatError() = %v, want to contain 恢复建议", formatted)
	}
}

func TestToJSON(t *testing.T) {
	testErr := NewStructured(ErrAppNotFound, "软件未找到")

	jsonStr, jsonErr := ToJSON(testErr)
	if jsonErr != nil {
		t.Fatalf("ToJSON failed: %v", jsonErr)
	}

	if !strings.Contains(jsonStr, "CHP-4001") {
		t.Errorf("ToJSON() = %v, want to contain CHP-4001", jsonStr)
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := `{"code":"CHP-4001","message":"test","category":"install"}`

	parsedErr, parseErr := FromJSON([]byte(jsonData))
	if parseErr != nil {
		t.Fatalf("FromJSON failed: %v", parseErr)
	}

	if parsedErr.Code != ErrAppNotFound {
		t.Errorf("Code = %v, want CHP-4001", parsedErr.Code)
	}
}

func TestAnalyzeErrors(t *testing.T) {
	errs := []error{
		NewStructured(ErrAppNotFound, "test1"),
		NewStructured(ErrAppNotFound, "test2"),
		NewStructured(ErrNetworkConnection, "test"),
		NewStructured(ErrBucketNotFound, "test"),
		errors.New("standard error"),
	}

	summary := AnalyzeErrors(errs)

	if summary.TotalErrors != 5 {
		t.Errorf("TotalErrors = %v, want 5", summary.TotalErrors)
	}

	if summary.ByCategory[CategoryInstall] != 2 {
		t.Errorf("Install errors = %v, want 2", summary.ByCategory[CategoryInstall])
	}

	if len(summary.MostFrequent) < 1 {
		t.Errorf("MostFrequent = %v, want >= 1", summary.MostFrequent)
	}
}

func TestAppErrorConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  *StructuredError
		code ErrorCode
	}{
		{"NewAppNotFound", NewAppNotFound("test-app"), ErrAppNotFound},
		{"NewAppAlreadyInstalled", NewAppAlreadyInstalled("test-app", "1.0.0"), ErrAppAlreadyInstalled},
		{"NewAppNotInstalled", NewAppNotInstalled("test-app"), ErrAppNotInstalled},
		{"NewVersionNotFound", NewVersionNotFound("test-app", "1.0.0"), ErrVersionNotFound},
		{"NewDownloadFailed", NewDownloadFailed("https://example.com", errors.New("timeout")), ErrDownloadFailed},
		{"NewInstallFailed", NewInstallFailed("test-app", errors.New("failed")), ErrInstallFailed},
		{"NewUninstallFailed", NewUninstallFailed("test-app", errors.New("failed")), ErrUninstallFailed},
		{"NewUpdateFailed", NewUpdateFailed("test-app", errors.New("failed")), ErrUpdateFailed},
		{"NewDependencyConflict", NewDependencyConflict("dep", "reason"), ErrDependencyConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("%s Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
		})
	}
}

func TestBucketErrorConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  *StructuredError
		code ErrorCode
	}{
		{"NewBucketNotFound", NewBucketNotFound("test-bucket"), ErrBucketNotFound},
		{"NewBucketAlreadyExists", NewBucketAlreadyExists("test-bucket"), ErrBucketAlreadyExists},
		{"NewInvalidBucketURL", NewInvalidBucketURL("invalid-url"), ErrInvalidBucketURL},
		{"NewManifestNotFound", NewManifestNotFound("main", "test-app"), ErrManifestNotFound},
		{"NewNetworkConnection", NewNetworkConnection("github.com", errors.New("timeout")), ErrNetworkConnection},
		{"NewConfigInvalid", NewConfigInvalid("config.yaml", "invalid yaml"), ErrConfigInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("%s Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
		})
	}
}

func TestStructuredError_Timestamp(t *testing.T) {
	before := time.Now()
	err := NewStructured(ErrAppNotFound, "test")
	after := time.Now()

	if err.Details.Timestamp.Before(before) || err.Details.Timestamp.After(after) {
		t.Errorf("Timestamp = %v, want between %v and %v", err.Details.Timestamp, before, after)
	}
}

func TestStructuredError_Category(t *testing.T) {
	err := NewStructured(ErrAppNotFound, "test")

	if err.Category != CategoryInstall {
		t.Errorf("Category = %v, want install", err.Category)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected bool
	}{
		{"network connection", ErrNetworkConnection, true},
		{"download timeout", ErrDownloadTimeout, true},
		{"download failed", ErrDownloadFailed, true},
		{"app not found", ErrAppNotFound, false},
		{"permission denied", ErrPermissionDenied, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStructured(tt.code, "test")
			if err.Retryable != tt.expected {
				t.Errorf("Retryable = %v, want %v", err.Retryable, tt.expected)
			}
		})
	}
}

func TestErrorContextPreservation(t *testing.T) {
	err := NewStructured(ErrDownloadFailed, "下载失败").
		WithContext("url", "https://example.com/file.zip").
		WithContext("size", 1024*1024*100).
		WithContext("retry_count", 3).
		WithOperation("download package")

	if err.Details.Context["url"] != "https://example.com/file.zip" {
		t.Errorf("url context = %v, want https://example.com/file.zip", err.Details.Context["url"])
	}
	if err.Details.Context["size"] != 1024*1024*100 {
		t.Errorf("size context = %v, want 104857600", err.Details.Context["size"])
	}
	if err.Details.Context["retry_count"] != 3 {
		t.Errorf("retry_count context = %v, want 3", err.Details.Context["retry_count"])
	}
}

func TestErrorChaining(t *testing.T) {
	cause1 := errors.New("root cause")
	err1 := WrapStructured(cause1, ErrNetworkConnection, "connect to server")
	err2 := WrapStructured(err1, ErrDownloadFailed, "download package")

	// Test error chain
	if !errors.Is(err2, err1) {
		t.Error("err2 should wrap err1")
	}

	// Verify the chain can be unwrapped
	unwrapped := errors.Unwrap(err2)
	if unwrapped != err1 {
		t.Errorf("Unwrap(err2) = %v, want err1", unwrapped)
	}
}

func TestGetDefaultSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		minCount int
	}{
		{"permission denied", ErrPermissionDenied, 2},
		{"network connection", ErrNetworkConnection, 2},
		{"app not found", ErrAppNotFound, 2},
		{"dependency conflict", ErrDependencyConflict, 2},
		{"unknown error", ErrUnknown, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := GetDefaultSuggestions(tt.code)
			if len(suggestions) < tt.minCount {
				t.Errorf("GetDefaultSuggestions(%v) returned %d suggestions, want >= %d", tt.code, len(suggestions), tt.minCount)
			}
		})
	}
}
