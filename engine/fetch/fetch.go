// Package fetch 提供 HTTP 请求功能。
package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultClient = &http.Client{
	Timeout: 30 * time.Second,
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
		return nil, fmt.Errorf("创建请求: %w", err)
	}
	return doRequest(client, req, nil)
}

// Post 执行 HTTP POST 请求。
func Post(client *http.Client, requestURL string, body interface{}, contentType string) (*Response, error) {
	if client == nil {
		client = defaultClient
	}
	var bodyReader io.Reader
	var err error

	switch v := body.(type) {
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = bytes.NewReader(v)
	case io.Reader:
		bodyReader = v
	default:
		jsonBody, jsonErr := json.Marshal(body)
		if jsonErr != nil {
			return nil, fmt.Errorf("序列化 body: %w", jsonErr)
		}
		bodyReader = bytes.NewReader(jsonBody)
		if contentType == "" {
			contentType = "application/json"
		}
	}

	req, err := http.NewRequest("POST", requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return doRequest(client, req, nil)
}

// Download 从 url 下载文件到 destPath。
func Download(url, destPath string) error {
	return DownloadWithContext(context.Background(), url, destPath)
}

// DownloadWithContext 使用 context 下载文件。
func DownloadWithContext(ctx context.Context, url, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("获取 url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建文件: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("复制响应体: %w", err)
	}
	return nil
}

// DownloadFile 使用自定义请求头下载文件。
func DownloadFile(client *http.Client, url, destPath string, headers map[string]string) error {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建文件: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("复制响应体: %w", err)
	}
	return nil
}

// DownloadRange 下载文件的指定字节范围 (用于并行下载/断点续传)。
func DownloadRange(url, destPath string, startBytes, endBytes int64) error {
	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求: %w", err)
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startBytes, endBytes))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建文件: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("复制响应体: %w", err)
	}
	return nil
}

// ParseURL 解析 URL 字符串为组件。
func ParseURL(rawURL string) (*URLInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("解析 url: %w", err)
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
		return "", fmt.Errorf("解析基础 url: %w", err)
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

	var bodyReader io.Reader
	if opts.Body != nil {
		switch v := opts.Body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			jsonBody, jsonErr := json.Marshal(opts.Body)
			if jsonErr != nil {
				return nil, fmt.Errorf("序列化 body: %w", jsonErr)
			}
			bodyReader = bytes.NewReader(jsonBody)
			if opts.ContentType == "" {
				opts.ContentType = "application/json"
			}
		}
	}

	req, err := http.NewRequest(opts.Method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求: %w", err)
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
		return nil, fmt.Errorf("执行请求: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体: %w", err)
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
		return fmt.Errorf("反序列化 json: %w", err)
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
			return fmt.Errorf("反序列化 json: %w", err)
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
		return fmt.Errorf("反序列化 xml: %w", err)
	}
	return nil
}

// PostXML 执行带 XML body 的 POST 请求并反序列化响应。
func PostXML(client *http.Client, requestURL string, body interface{}, result interface{}) error {
	xmlBody, err := xml.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化 xml: %w", err)
	}

	resp, err := Post(client, requestURL, xmlBody, "application/xml")
	if err != nil {
		return err
	}
	if result != nil {
		if err := xml.Unmarshal([]byte(resp.Body), result); err != nil {
			return fmt.Errorf("反序列化 xml: %w", err)
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
