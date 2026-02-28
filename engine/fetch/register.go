package fetch

import (
	"time"

	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 fetch 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 fetch 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	fetchObj := vm.NewObject()

	fetchObj.Set("download", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		dest := call.Argument(1).String()
		err := Download(url, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fetchObj.Set("get", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		resp, err := Get(nil, url)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": respToJS(resp)})
	})

	fetchObj.Set("post", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		body := call.Argument(1).Export()
		contentType := call.Argument(2).String()
		resp, err := Post(nil, url, body, contentType)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": respToJS(resp)})
	})

	fetchObj.Set("request", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		options := call.Argument(1).Export()

		opts := &RequestOptions{Method: "GET"}
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
		}

		resp, err := Request(nil, url, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": respToJS(resp)})
	})

	fetchObj.Set("downloadFile", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		dest := call.Argument(1).String()
		var headers map[string]string
		if h := call.Argument(2); h != nil && !goja.IsUndefined(h) {
			if hm, ok := h.Export().(map[string]interface{}); ok {
				headers = map[string]string{}
				for k, v := range hm {
					if vs, ok := v.(string); ok {
						headers[k] = vs
					}
				}
			}
		}
		err := DownloadFile(nil, url, dest, headers)
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
			"success": true,
			"data": map[string]interface{}{
				"scheme":   info.Scheme,
				"host":     info.Host,
				"path":     info.Path,
				"query":    info.Query,
				"fragment": info.Fragment,
			},
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
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	fetchObj.Set("getJSON", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		result := call.Argument(1)
		err := GetJSON(nil, url, result)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	fetchObj.Set("postJSON", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		body := call.Argument(1).Export()
		result := call.Argument(2)
		err := PostJSON(nil, url, body, result)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	fetchObj.Set("newClient", func(call goja.FunctionCall) goja.Value {
		timeout := float64(30)
		if t := call.Argument(0).ToFloat(); t > 0 {
			timeout = t
		}
		client := NewClientWithTimeout(time.Duration(timeout) * time.Second)
		return vm.ToValue(client)
	})

	fetchObj.Set("setDefaultTimeout", func(call goja.FunctionCall) goja.Value {
		timeout := call.Argument(0).ToFloat()
		SetDefaultTimeout(time.Duration(timeout) * time.Second)
		return nil
	})

	vm.Set("fetch", fetchObj)
}

func respToJS(resp *Response) map[string]interface{} {
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    resp.Body,
		"headers": resp.Headers,
		"rawBody": string(resp.RawBody),
	}
}
