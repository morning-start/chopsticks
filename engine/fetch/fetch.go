// Package fetch 提供 HTTP 请求功能。
package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chopsticks/pkg/output"
)

// HTTP 超时时间常量
const (
	DefaultTimeout       = 30 * time.Second
	DownloadTimeout      = 5 * time.Minute
	StatusOK             = http.StatusOK
	StatusPartialContent = http.StatusPartialContent
)

// 预定义错误变量
var (
	ErrCreateRequest   = errors.New("failed to create request")
	ErrExecuteRequest  = errors.New("failed to execute request")
	ErrDownloadFailed  = errors.New("download failed")
	ErrCreateDirectory = errors.New("failed to create directory")
	ErrCreateFile      = errors.New("failed to create file")
	ErrCopyResponse    = errors.New("failed to copy response body")
	ErrParseURL        = errors.New("failed to parse URL")
	ErrSerializeBody   = errors.New("failed to serialize body")
	ErrDeserializeJSON = errors.New("failed to deserialize JSON")
	ErrDeserializeXML  = errors.New("failed to deserialize XML")
	ErrSerializeXML    = errors.New("failed to serialize XML")
)

var defaultClient = &http.Client{
	Timeout: DefaultTimeout,
}

// RequestOptions 包含 HTTP 请求的选项。
type RequestOptions struct {
	Method      string
	Headers     map[string]string
	Body        interface{}
	ContentType string
	Timeout     time.Duration
}

// Response 表示 HTTP 响应。
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	RawBody    []byte
}

// URLInfo 包含解析后的 URL 组件。
type URLInfo struct {
	Scheme   string
	Host     string
	Path     string
	Query    map[string]string
	Fragment string
}

// Get 执行 HTTP GET 请求。
func Get(client *http.Client, requestURL string) (*Response, error) {
	if client == nil {
		client = defaultClient
	}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}
	return doRequest(client, req, nil)
}

// Post 执行 HTTP POST 请求。
func Post(client *http.Client, requestURL string, body interface{}, contentType string) (*Response, error) {
	if client == nil {
		client = defaultClient
	}
	bodyReader, err := prepareBody(body, &contentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return doRequest(client, req, nil)
}

// prepareBody 准备请求体
func prepareBody(body interface{}, contentType *string) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case string:
		return strings.NewReader(v), nil
	case []byte:
		return bytes.NewReader(v), nil
	case io.Reader:
		return v, nil
	default:
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSerializeBody, err)
		}
		if *contentType == "" {
			*contentType = "application/json"
		}
		return bytes.NewReader(jsonBody), nil
	}
}

// Download 从 url 下载文件到 destPath。
func Download(url, destPath string) error {
	return DownloadWithContext(context.Background(), url, destPath)
}

// DownloadWithContext 使用 context 下载文件。
func DownloadWithContext(ctx context.Context, url, destPath string) error {
	return DownloadWithProgress(ctx, url, destPath, nil)
}

// DownloadWithProgress 使用进度条下载文件。
func DownloadWithProgress(ctx context.Context, url, destPath string, pm *output.ProgressManager) error {
	client := &http.Client{Timeout: DownloadTimeout}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecuteRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != StatusOK {
		return fmt.Errorf("%w: %s", ErrDownloadFailed, resp.Status)
	}

	return saveResponseBody(resp, destPath, pm)
}

// DownloadFile 使用自定义请求头下载文件。
func DownloadFile(client *http.Client, url, destPath string, headers map[string]string) error {
	return DownloadFileWithProgress(client, url, destPath, headers, nil)
}

// DownloadFileWithProgress 使用进度条下载文件。
func DownloadFileWithProgress(client *http.Client, url, destPath string, headers map[string]string, pm *output.ProgressManager) error {
	if client == nil {
		client = &http.Client{Timeout: DownloadTimeout}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecuteRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != StatusOK {
		return fmt.Errorf("%w: %s", ErrDownloadFailed, resp.Status)
	}

	return saveResponseBody(resp, destPath, pm)
}

// saveResponseBody 保存响应体到文件
func saveResponseBody(resp *http.Response, destPath string, pm *output.ProgressManager) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrCreateDirectory, err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFile, err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if pm != nil && resp.ContentLength > 0 {
		filename := filepath.Base(destPath)
		reader = pm.ProxyReader(resp.Body, filename, resp.ContentLength)
	}

	if _, err = io.Copy(out, reader); err != nil {
		return fmt.Errorf("%w: %w", ErrCopyResponse, err)
	}
	return nil
}

// DownloadRange 下载文件的指定字节范围 (用于并行下载/断点续传)。
func DownloadRange(url, destPath string, startBytes, endBytes int64) error {
	client := &http.Client{Timeout: DownloadTimeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startBytes, endBytes))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecuteRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != StatusOK && resp.StatusCode != StatusPartialContent {
		return fmt.Errorf("%w: %s", ErrDownloadFailed, resp.Status)
	}

	return saveResponseBody(resp, destPath, nil)
}

// ParseURL 解析 URL 字符串为组件。
func ParseURL(rawURL string) (*URLInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseURL, err)
	}

	queryParams := make(map[string]string, len(parsed.Query()))
	for k, v := range parsed.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	return &URLInfo{
		Scheme:   parsed.Scheme,
		Host:     parsed.Host,
		Path:     parsed.Path,
		Query:    queryParams,
		Fragment: parsed.Fragment,
	}, nil
}

// BuildURL 使用查询参数构建 URL。
func BuildURL(baseURL string, params map[string]string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseURL, err)
	}

	query := parsed.Query()
	for k, v := range params {
		query.Set(k, v)
	}

	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

// Request 使用选项执行 HTTP 请求。
func Request(client *http.Client, requestURL string, opts *RequestOptions) (*Response, error) {
	if client == nil {
		client = defaultClient
	}

	if opts == nil {
		opts = &RequestOptions{Method: "GET"}
	}

	bodyReader, err := prepareBody(opts.Body, &opts.ContentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(opts.Method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	if opts.ContentType != "" {
		req.Header.Set("Content-Type", opts.ContentType)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	if opts.Timeout > 0 {
		client.Timeout = opts.Timeout
	}

	return doRequest(client, req, opts)
}

func doRequest(client *http.Client, req *http.Request, _ *RequestOptions) (*Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExecuteRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCopyResponse, err)
	}

	headers := make(map[string]string, len(resp.Header))
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		RawBody:    body,
	}, nil
}

// GetJSON 执行 GET 请求并反序列化 JSON 响应。
func GetJSON(client *http.Client, requestURL string, result interface{}) error {
	resp, err := Get(client, requestURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp.Body), result); err != nil {
		return fmt.Errorf("%w: %w", ErrDeserializeJSON, err)
	}
	return nil
}

// PostJSON 执行带 JSON body 的 POST 请求并反序列化响应。
func PostJSON(client *http.Client, requestURL string, body interface{}, result interface{}) error {
	resp, err := Post(client, requestURL, body, "application/json")
	if err != nil {
		return err
	}
	if result != nil {
		if err := json.Unmarshal([]byte(resp.Body), result); err != nil {
			return fmt.Errorf("%w: %w", ErrDeserializeJSON, err)
		}
	}
	return nil
}

// GetXML 执行 GET 请求并反序列化 XML 响应。
func GetXML(client *http.Client, requestURL string, result interface{}) error {
	resp, err := Get(client, requestURL)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal([]byte(resp.Body), result); err != nil {
		return fmt.Errorf("%w: %w", ErrDeserializeXML, err)
	}
	return nil
}

// PostXML 执行带 XML body 的 POST 请求并反序列化响应。
func PostXML(client *http.Client, requestURL string, body interface{}, result interface{}) error {
	xmlBody, err := xml.Marshal(body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSerializeXML, err)
	}

	resp, err := Post(client, requestURL, xmlBody, "application/xml")
	if err != nil {
		return err
	}
	if result != nil {
		if err := xml.Unmarshal([]byte(resp.Body), result); err != nil {
			return fmt.Errorf("%w: %w", ErrDeserializeXML, err)
		}
	}
	return nil
}

// SetDefaultTimeout 设置默认客户端的超时时间。
func SetDefaultTimeout(timeout time.Duration) {
	defaultClient.Timeout = timeout
}

// NewClient 创建具有默认超时的新 HTTP 客户端。
func NewClient() *http.Client {
	return &http.Client{Timeout: defaultClient.Timeout}
}

// NewClientWithTimeout 创建具有指定超时的新 HTTP 客户端。
func NewClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
