package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("期望 GET 方法，得到 %s", r.Method)
		}

		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		case "/servererror":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		case "/headers":
			w.Header().Set("X-Custom-Header", "custom-value")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("headers test"))
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("default response"))
		}
	}))
	defer ts.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
		wantErr    bool
	}{
		{
			name:       "成功的请求",
			path:       "/success",
			wantStatus: http.StatusOK,
			wantBody:   `{"message": "success"}`,
			wantErr:    false,
		},
		{
			name:       "404 错误",
			path:       "/notfound",
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
			wantErr:    false, // Get 不会将非 200 状态码视为错误
		},
		{
			name:       "500 错误",
			path:       "/servererror",
			wantStatus: http.StatusInternalServerError,
			wantBody:   "server error",
			wantErr:    false,
		},
		{
			name:       "自定义响应头",
			path:       "/headers",
			wantStatus: http.StatusOK,
			wantBody:   "headers test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Get(nil, ts.URL+tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantBody, resp.Body)
		})
	}
}

func TestPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，得到 %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		contentType := r.Header.Get("Content-Type")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"received": "%s", "contentType": "%s"}`, string(body), contentType)
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		body        interface{}
		contentType string
		wantErr     bool
	}{
		{
			name:        "POST JSON 对象",
			body:        map[string]string{"key": "value"},
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:        "POST 字符串",
			body:        "plain text body",
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "POST 字节数组",
			body:        []byte("byte array body"),
			contentType: "application/octet-stream",
			wantErr:     false,
		},
		{
			name:        "POST nil",
			body:        nil,
			contentType: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Post(nil, ts.URL, tt.body, tt.contentType)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestDownload(t *testing.T) {
	content := []byte("This is test file content for download testing")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "downloaded.txt")

	err := Download(ts.URL, destPath)
	require.NoError(t, err)

	// 验证文件内容
	downloadedContent, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, downloadedContent)
}

func TestDownload_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "failed.txt")

	err := Download(ts.URL, destPath)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDownloadFailed)
}

func TestDownload_InvalidURL(t *testing.T) {
	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "test.txt")

	err := Download("not-a-valid-url", destPath)
	assert.Error(t, err)
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *URLInfo
		wantErr bool
	}{
		{
			name: "完整的 URL",
			url:  "https://example.com/path?key=value&foo=bar#fragment",
			want: &URLInfo{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/path",
				Query:    map[string]string{"key": "value", "foo": "bar"},
				Fragment: "fragment",
			},
			wantErr: false,
		},
		{
			name: "只有路径的 URL",
			url:  "/path/to/resource",
			want: &URLInfo{
				Scheme: "",
				Host:   "",
				Path:   "/path/to/resource",
				Query:  map[string]string{},
			},
			wantErr: false,
		},
		{
			name:    "无效的 URL",
			url:     "://invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Scheme, got.Scheme)
			assert.Equal(t, tt.want.Host, got.Host)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.Fragment, got.Fragment)
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		params  map[string]string
		want    string
		wantErr bool
	}{
		{
			name:    "添加参数到 URL",
			baseURL: "https://example.com/api",
			params:  map[string]string{"key": "value", "foo": "bar"},
			want:    "https://example.com/api?foo=bar&key=value",
			wantErr: false,
		},
		{
			name:    "覆盖现有参数",
			baseURL: "https://example.com/api?key=old",
			params:  map[string]string{"key": "new"},
			want:    "https://example.com/api?key=new",
			wantErr: false,
		},
		{
			name:    "无效的 base URL",
			baseURL: "://invalid",
			params:  map[string]string{"key": "value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildURL(tt.baseURL, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回请求信息
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"method": "%s", "path": "%s"}`, r.Method, r.URL.Path)
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		opts    *RequestOptions
		wantErr bool
	}{
		{
			name: "GET 请求",
			opts: &RequestOptions{
				Method: "GET",
			},
			wantErr: false,
		},
		{
			name: "POST 请求带 body",
			opts: &RequestOptions{
				Method:      "POST",
				Body:        map[string]string{"key": "value"},
				ContentType: "application/json",
			},
			wantErr: false,
		},
		{
			name: "带自定义 header",
			opts: &RequestOptions{
				Method:  "GET",
				Headers: map[string]string{"X-Custom": "value"},
			},
			wantErr: false,
		},
		{
			name:    "nil options (使用默认 GET)",
			opts:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Request(nil, ts.URL, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "test", "value": 123}`))
	}))
	defer ts.Close()

	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err := GetJSON(nil, ts.URL, &result)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 123, result.Value)
}

func TestPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body) // 回显请求体
	}))
	defer ts.Close()

	body := map[string]string{"message": "hello"}
	var result map[string]string

	err := PostJSON(nil, ts.URL, body, &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result["message"])
}

func TestDownloadWithContext(t *testing.T) {
	content := []byte("test content for context download")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	ctx := context.Background()
	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "context_download.txt")

	err := DownloadWithContext(ctx, ts.URL, destPath)
	require.NoError(t, err)

	downloadedContent, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, downloadedContent)
}

func TestDownloadWithContext_Cancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟慢速响应
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	}))
	defer ts.Close()

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "cancelled.txt")

	err := DownloadWithContext(ctx, ts.URL, destPath)
	assert.Error(t, err)
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	assert.Equal(t, defaultClient.Timeout, client.Timeout)
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 60 * time.Second
	client := NewClientWithTimeout(timeout)
	assert.NotNil(t, client)
	assert.Equal(t, timeout, client.Timeout)
}

func TestSetDefaultTimeout(t *testing.T) {
	originalTimeout := defaultClient.Timeout
	defer SetDefaultTimeout(originalTimeout) // 恢复原始超时

	newTimeout := 45 * time.Second
	SetDefaultTimeout(newTimeout)
	assert.Equal(t, newTimeout, defaultClient.Timeout)
}

func TestResponse(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"status": "ok"}`,
		RawBody:    []byte(`{"status": "ok"}`),
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Headers["Content-Type"])
	assert.Equal(t, `{"status": "ok"}`, resp.Body)
}

func TestRequestOptions(t *testing.T) {
	opts := &RequestOptions{
		Method:      "POST",
		Headers:     map[string]string{"Authorization": "Bearer token"},
		Body:        map[string]string{"key": "value"},
		ContentType: "application/json",
		Timeout:     30 * time.Second,
	}

	assert.Equal(t, "POST", opts.Method)
	assert.Equal(t, "Bearer token", opts.Headers["Authorization"])
	assert.Equal(t, "application/json", opts.ContentType)
}

func TestErrors(t *testing.T) {
	// 测试错误变量
	assert.NotNil(t, ErrCreateRequest)
	assert.NotNil(t, ErrExecuteRequest)
	assert.NotNil(t, ErrDownloadFailed)
	assert.NotNil(t, ErrCreateDirectory)
	assert.NotNil(t, ErrCreateFile)
	assert.NotNil(t, ErrCopyResponse)
	assert.NotNil(t, ErrParseURL)
	assert.NotNil(t, ErrSerializeBody)
	assert.NotNil(t, ErrDeserializeJSON)
	assert.NotNil(t, ErrDeserializeXML)
	assert.NotNil(t, ErrSerializeXML)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 5*time.Minute, DownloadTimeout)
	assert.Equal(t, http.StatusOK, StatusOK)
	assert.Equal(t, http.StatusPartialContent, StatusPartialContent)
}

func TestGetXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0"?><root><name>test</name><value>123</value></root>`))
	}))
	defer ts.Close()

	var result struct {
		Name  string `xml:"name"`
		Value int    `xml:"value"`
	}

	err := GetXML(nil, ts.URL, &result)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 123, result.Value)
}

func TestPostXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write(body) // 回显请求体
	}))
	defer ts.Close()

	type XMLBody struct {
		Message string `xml:"message"`
	}

	body := XMLBody{Message: "hello"}
	var result XMLBody

	err := PostXML(nil, ts.URL, body, &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Message)
}

func TestDownloadFile(t *testing.T) {
	content := []byte("test content for DownloadFile")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer token123" {
			t.Errorf("期望 Authorization header，得到 %s", authHeader)
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "downloadfile.txt")

	headers := map[string]string{
		"Authorization": "Bearer token123",
	}

	err := DownloadFile(nil, ts.URL, destPath, headers)
	require.NoError(t, err)

	downloadedContent, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, downloadedContent)
}

func TestDownloadFile_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "failed.txt")

	err := DownloadFile(nil, ts.URL, destPath, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDownloadFailed)
}

func TestDownloadRange(t *testing.T) {
	content := []byte("0123456789")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查 Range 头
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "bytes=0-4" {
			t.Errorf("期望 Range: bytes=0-4，得到 %s", rangeHeader)
		}

		w.Header().Set("Content-Length", "5")
		w.WriteHeader(http.StatusPartialContent)
		w.Write(content[0:5])
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "range_download.txt")

	err := DownloadRange(ts.URL, destPath, 0, 4)
	require.NoError(t, err)

	downloadedContent, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, "01234", string(downloadedContent))
}

func TestDownloadRange_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "failed.txt")

	err := DownloadRange(ts.URL, destPath, 0, 100)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDownloadFailed)
}

func TestURLInfo(t *testing.T) {
	info := URLInfo{
		Scheme:   "https",
		Host:     "example.com",
		Path:     "/api/v1",
		Query:    map[string]string{"key": "value"},
		Fragment: "section1",
	}

	assert.Equal(t, "https", info.Scheme)
	assert.Equal(t, "example.com", info.Host)
	assert.Equal(t, "/api/v1", info.Path)
	assert.Equal(t, "value", info.Query["key"])
	assert.Equal(t, "section1", info.Fragment)
}
