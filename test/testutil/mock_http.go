// Package testutil 提供集成测试的辅助工具。
package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
)

// MockDownloadServer 模拟下载服务器。
type MockDownloadServer struct {
	Server   *httptest.Server
	BaseURL  string
	Handlers map[string]http.HandlerFunc
}

// SetupMockDownloadServer 创建模拟下载服务器。
func SetupMockDownloadServer() *MockDownloadServer {
	mock := &MockDownloadServer{
		Handlers: make(map[string]http.HandlerFunc),
	}

	mux := http.NewServeMux()

	// 默认下载处理器
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		// 从路径中提取文件名
		filename := filepath.Base(r.URL.Path)

		// 检查是否有自定义处理器
		if handler, ok := mock.Handlers[filename]; ok {
			handler(w, r)
			return
		}

		// 默认返回模拟的 zip 文件内容
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Length", "100")

		// 返回模拟的 zip 文件头
		w.Write([]byte("PK\x03\x04")) // ZIP 文件魔数
		w.Write(make([]byte, 96))     // 填充内容
	})

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mock.Server = httptest.NewServer(mux)
	mock.BaseURL = mock.Server.URL

	return mock
}

// Close 关闭服务器。
func (m *MockDownloadServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// RegisterHandler 注册自定义处理器。
func (m *MockDownloadServer) RegisterHandler(filename string, handler http.HandlerFunc) {
	m.Handlers[filename] = handler
}

// GetDownloadURL 获取下载 URL。
func (m *MockDownloadServer) GetDownloadURL(filename string) string {
	return fmt.Sprintf("%s/download/%s", m.BaseURL, filename)
}

// SetupMockDownloadServerWithFile 创建带有特定文件内容的模拟下载服务器。
func SetupMockDownloadServerWithFile(filename string, content []byte) *MockDownloadServer {
	mock := SetupMockDownloadServer()

	mock.RegisterHandler(filename, func(w http.ResponseWriter, r *http.Request) {
		// 根据内容类型设置响应头
		contentType := detectContentType(filename)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

		w.Write(content)
	})

	return mock
}

// detectContentType 根据文件名检测内容类型。
func detectContentType(filename string) string {
	lowerFilename := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lowerFilename, ".zip"):
		return "application/zip"
	case strings.HasSuffix(lowerFilename, ".tar.gz") || strings.HasSuffix(lowerFilename, ".tgz"):
		return "application/gzip"
	case strings.HasSuffix(lowerFilename, ".tar"):
		return "application/x-tar"
	case strings.HasSuffix(lowerFilename, ".exe"):
		return "application/octet-stream"
	case strings.HasSuffix(lowerFilename, ".msi"):
		return "application/x-msi"
	case strings.HasSuffix(lowerFilename, ".json"):
		return "application/json"
	case strings.HasSuffix(lowerFilename, ".yaml") || strings.HasSuffix(lowerFilename, ".yml"):
		return "application/x-yaml"
	default:
		return "application/octet-stream"
	}
}
