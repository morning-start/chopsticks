package fetch

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 fetch 注册。
type Module struct{}

// parseOptionsFromJS 从 JavaScript 对象解析 RequestOptions
func parseOptionsFromJS(options interface{}) *RequestOptions {
	opts := &RequestOptions{}
	if optionsMap, ok := options.(map[string]interface{}); ok {
		if method, ok := optionsMap["method"].(string); ok {
			opts.Method = method
		}
		if headers, ok := optionsMap["headers"].(map[string]interface{}); ok {
			opts.Headers = map[string]string{}
			for k, v := range headers {
				if vs, ok := v.(string); ok {
					opts.Headers[k] = vs
				}
			}
		}
		if body, ok := optionsMap["body"]; ok {
			opts.Body = body
		}
		if contentType, ok := optionsMap["contentType"].(string); ok {
			opts.ContentType = contentType
		}
		if timeout, ok := optionsMap["timeout"].(float64); ok && timeout > 0 {
			opts.Timeout = time.Duration(timeout) * time.Millisecond
		}
	}
	return opts
}

// respToJSWithOk 将 Response 转换为 JavaScript 对象，包含 ok 字段
func respToJSWithOk(resp *Response) map[string]interface{} {
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
		"body":    resp.Body,
		"headers": resp.Headers,
	}
}

// respToJS 将 Response 转换为 JavaScript 对象
func respToJS(resp *Response) map[string]interface{} {
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    resp.Body,
		"headers": resp.Headers,
		"rawBody": string(resp.RawBody),
	}
}

// RegisterJS 向 JavaScript 运行时注册 fetch 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	fetchObj := vm.NewObject()

	fetchObj.Set("download", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		dest := call.Argument(1).String()

		var headers map[string]string
		var timeout time.Duration

		if opts := call.Argument(2); opts != nil && !goja.IsUndefined(opts) {
			if optionsMap, ok := opts.Export().(map[string]interface{}); ok {
				if h, ok := optionsMap["headers"].(map[string]interface{}); ok {
					headers = map[string]string{}
					for k, v := range h {
						if vs, ok := v.(string); ok {
							headers[k] = vs
						}
					}
				}
				if t, ok := optionsMap["timeout"].(float64); ok && t > 0 {
					timeout = time.Duration(t) * time.Millisecond
				}
			}
		}

		var err error
		if headers != nil || timeout > 0 {
			client := &http.Client{Timeout: timeout}
			if timeout == 0 {
				client.Timeout = DownloadTimeout
			}
			err = DownloadFile(client, url, dest, headers)
		} else {
			err = Download(url, dest)
		}

		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fetchObj.Set("get", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		opts := parseOptionsFromJS(call.Argument(1).Export())

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"status":  resp.StatusCode,
			"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
			"body":    resp.Body,
			"headers": resp.Headers,
			"error":   nil,
		})
	})

	fetchObj.Set("post", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		body := call.Argument(1).Export()

		var contentType string
		if ct := call.Argument(2); ct != nil && !goja.IsUndefined(ct) {
			if ctStr, ok := ct.Export().(string); ok {
				contentType = ctStr
			}
		}
		if contentType == "" {
			contentType = "application/x-www-form-urlencoded"
		}

		opts := parseOptionsFromJS(call.Argument(3).Export())
		opts.Method = "POST"
		opts.Body = body
		opts.ContentType = contentType

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"status":  resp.StatusCode,
			"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
			"body":    resp.Body,
			"headers": resp.Headers,
			"error":   nil,
		})
	})

	fetchObj.Set("request", func(call goja.FunctionCall) goja.Value {
		method := call.Argument(0).String()
		url := call.Argument(1).String()
		opts := parseOptionsFromJS(call.Argument(2).Export())

		if opts.Method == "" {
			opts.Method = method
		}

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"status":  resp.StatusCode,
			"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
			"body":    resp.Body,
			"headers": resp.Headers,
			"error":   nil,
		})
	})

	fetchObj.Set("downloadFile", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		dest := call.Argument(1).String()

		var headers map[string]string
		var timeout time.Duration

		if opts := call.Argument(2); opts != nil && !goja.IsUndefined(opts) {
			if optionsMap, ok := opts.Export().(map[string]interface{}); ok {
				if h, ok := optionsMap["headers"].(map[string]interface{}); ok {
					headers = map[string]string{}
					for k, v := range h {
						if vs, ok := v.(string); ok {
							headers[k] = vs
						}
					}
				}
				if t, ok := optionsMap["timeout"].(float64); ok && t > 0 {
					timeout = time.Duration(t) * time.Millisecond
				}
			}
		}

		client := &http.Client{Timeout: timeout}
		if timeout == 0 {
			client.Timeout = DownloadTimeout
		}

		err := DownloadFile(client, url, dest, headers)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fetchObj.Set("parseURL", func(call goja.FunctionCall) goja.Value {
		urlStr := call.Argument(0).String()
		info, err := ParseURL(urlStr)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success":  true,
			"scheme":   info.Scheme,
			"host":     info.Host,
			"path":     info.Path,
			"query":    info.Query,
			"fragment": info.Fragment,
			"error":    nil,
		})
	})

	fetchObj.Set("buildURL", func(call goja.FunctionCall) goja.Value {
		baseURL := call.Argument(0).String()
		params := call.Argument(1).Export()

		paramsMap := map[string]string{}
		if paramsMapRaw, ok := params.(map[string]interface{}); ok {
			for k, v := range paramsMapRaw {
				if vs, ok := v.(string); ok {
					paramsMap[k] = vs
				}
			}
		}

		result, err := BuildURL(baseURL, paramsMap)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"url":     result,
			"error":   nil,
		})
	})

	fetchObj.Set("getJSON", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		opts := parseOptionsFromJS(call.Argument(1).Export())

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}

		var result interface{}
		if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": "failed to parse JSON: " + err.Error()})
		}

		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    result,
			"error":   nil,
		})
	})

	fetchObj.Set("postJSON", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		body := call.Argument(1).Export()
		opts := parseOptionsFromJS(call.Argument(2).Export())

		opts.Method = "POST"
		opts.Body = body
		if opts.ContentType == "" {
			opts.ContentType = "application/json"
		}

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}

		var result interface{}
		if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": "failed to parse JSON: " + err.Error()})
		}

		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    result,
			"error":   nil,
		})
	})

	fetchObj.Set("newClient", func(call goja.FunctionCall) goja.Value {
		var timeout time.Duration = DefaultTimeout
		var headers map[string]string

		if opts := call.Argument(0); opts != nil && !goja.IsUndefined(opts) {
			if optionsMap, ok := opts.Export().(map[string]interface{}); ok {
				if t, ok := optionsMap["timeout"].(float64); ok && t > 0 {
					timeout = time.Duration(t) * time.Millisecond
				}
				if h, ok := optionsMap["headers"].(map[string]interface{}); ok {
					headers = map[string]string{}
					for k, v := range h {
						if vs, ok := v.(string); ok {
							headers[k] = vs
						}
					}
				}
			}
		}

		client := &http.Client{Timeout: timeout}

		// 创建客户端对象并添加方法
		clientObj := vm.NewObject()

		clientObj.Set("get", func(call goja.FunctionCall) goja.Value {
			url := call.Argument(0).String()
			opts := parseOptionsFromJS(call.Argument(1).Export())

			// 合并默认 headers
			if headers != nil {
				if opts.Headers == nil {
					opts.Headers = map[string]string{}
				}
				for k, v := range headers {
					opts.Headers[k] = v
				}
			}

			resp, err := Request(client, url, opts)
			if err != nil {
				return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
			}
			return vm.ToValue(map[string]interface{}{
				"success": true,
				"status":  resp.StatusCode,
				"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
				"body":    resp.Body,
				"headers": resp.Headers,
				"error":   nil,
			})
		})

		clientObj.Set("post", func(call goja.FunctionCall) goja.Value {
			url := call.Argument(0).String()
			body := call.Argument(1).Export()

			var contentType string
			if ct := call.Argument(2); ct != nil && !goja.IsUndefined(ct) {
				if ctStr, ok := ct.Export().(string); ok {
					contentType = ctStr
				}
			}
			if contentType == "" {
				contentType = "application/x-www-form-urlencoded"
			}

			opts := parseOptionsFromJS(call.Argument(3).Export())
			opts.Method = "POST"
			opts.Body = body
			opts.ContentType = contentType

			// 合并默认 headers
			if headers != nil {
				if opts.Headers == nil {
					opts.Headers = map[string]string{}
				}
				for k, v := range headers {
					opts.Headers[k] = v
				}
			}

			resp, err := Request(client, url, opts)
			if err != nil {
				return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
			}
			return vm.ToValue(map[string]interface{}{
				"success": true,
				"status":  resp.StatusCode,
				"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
				"body":    resp.Body,
				"headers": resp.Headers,
				"error":   nil,
			})
		})

		return vm.ToValue(map[string]interface{}{
			"success": true,
			"client":  clientObj,
		})
	})

	fetchObj.Set("setDefaultTimeout", func(call goja.FunctionCall) goja.Value {
		timeout := call.Argument(0).ToFloat()
		if timeout <= 0 {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   "timeout must be greater than 0",
			})
		}

		SetDefaultTimeout(time.Duration(timeout) * time.Millisecond)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"error":   nil,
		})
	})

	vm.Set("fetch", fetchObj)
}
