package fetch

import (
	"time"

	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 fetch 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 fetch 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	fetchTable := L.NewTable()

	fetchTable.RawSetString("download", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		dest := L.CheckString(2)
		if err := Download(url, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fetchTable.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		resp, err := Get(nil, url)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(respToLua(L, resp))
		return 1
	}))

	fetchTable.RawSetString("post", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		body := L.CheckAny(2)
		contentType := L.OptString(3, "")
		resp, err := Post(nil, url, bodyToInterface(L, body), contentType)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(respToLua(L, resp))
		return 1
	}))

	fetchTable.RawSetString("request", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		options := L.CheckTable(2)

		opts := &RequestOptions{Method: "GET"}
		if method := L.GetField(options, "method"); method != lua.LNil {
			opts.Method = method.String()
		}
		if headers := L.GetField(options, "headers"); headers != lua.LNil {
			if headersTbl, ok := headers.(*lua.LTable); ok {
				opts.Headers = tableToMap(L, headersTbl)
			}
		}
		if body := L.GetField(options, "body"); body != lua.LNil {
			opts.Body = bodyToInterface(L, body)
		}

		resp, err := Request(nil, url, opts)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(respToLua(L, resp))
		return 1
	}))

	fetchTable.RawSetString("downloadFile", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		dest := L.CheckString(2)
		headers := L.OptTable(3, nil)

		var headerMap map[string]string
		if headers != nil {
			headerMap = tableToMap(L, headers)
		}

		if err := DownloadFile(nil, url, dest, headerMap); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fetchTable.RawSetString("parseURL", L.NewFunction(func(L *lua.LState) int {
		urlStr := L.CheckString(1)
		info, err := ParseURL(urlStr)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		tbl.RawSetString("scheme", lua.LString(info.Scheme))
		tbl.RawSetString("host", lua.LString(info.Host))
		tbl.RawSetString("path", lua.LString(info.Path))
		tbl.RawSetString("fragment", lua.LString(info.Fragment))

		queryTbl := L.NewTable()
		for k, v := range info.Query {
			queryTbl.RawSetString(k, lua.LString(v))
		}
		tbl.RawSetString("query", queryTbl)

		L.Push(tbl)
		return 1
	}))

	fetchTable.RawSetString("buildURL", L.NewFunction(func(L *lua.LState) int {
		baseURL := L.CheckString(1)
		params := L.CheckTable(2)

		paramsMap := tableToMap(L, params)

		result, err := BuildURL(baseURL, paramsMap)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(result))
		return 1
	}))

	fetchTable.RawSetString("getJSON", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		result := L.CheckTable(2)
		err := GetJSON(nil, url, result)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		L.Push(result)
		return 2
	}))

	fetchTable.RawSetString("postJSON", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		body := L.CheckTable(2)
		result := L.OptTable(3, nil)
		err := PostJSON(nil, url, body, result)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		if result != nil {
			L.Push(result)
			return 2
		}
		return 1
	}))

	fetchTable.RawSetString("newClient", L.NewFunction(func(L *lua.LState) int {
		timeout := L.OptNumber(1, 30)
		NewClientWithTimeout(time.Duration(float64(timeout)) * time.Second)
		L.Push(lua.LString("http_client"))
		return 1
	}))

	fetchTable.RawSetString("setDefaultTimeout", L.NewFunction(func(L *lua.LState) int {
		timeout := L.CheckNumber(1)
		SetDefaultTimeout(time.Duration(float64(timeout)) * time.Second)
		return 0
	}))

	L.SetGlobal("fetch", fetchTable)
}

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

func respToLua(L *lua.LState, resp *Response) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("status", lua.LNumber(resp.StatusCode))
	tbl.RawSetString("body", lua.LString(resp.Body))

	headersTbl := L.NewTable()
	for k, v := range resp.Headers {
		headersTbl.RawSetString(k, lua.LString(v))
	}
	tbl.RawSetString("headers", headersTbl)
	return tbl
}

func respToJS(resp *Response) map[string]interface{} {
	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    resp.Body,
		"headers": resp.Headers,
		"rawBody": string(resp.RawBody),
	}
}

func tableToMap(_ *lua.LState, tbl *lua.LTable) map[string]string {
	// 预分配容量以提高性能
	result := make(map[string]string, 16)
	tbl.ForEach(func(k lua.LValue, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			if vs, ok := v.(lua.LString); ok {
				result[string(ks)] = string(vs)
			}
		}
	})
	return result
}

func bodyToInterface(L *lua.LState, v lua.LValue) interface{} {
	switch val := v.(type) {
	case lua.LString:
		return string(val)
	case lua.LNumber:
		return float64(val)
	case lua.LBool:
		return bool(val)
	case *lua.LTable:
		return tableToInterface(L, val)
	case nil:
		return nil
	default:
		return v.String()
	}
}

func tableToInterface(L *lua.LState, tbl *lua.LTable) interface{} {
	// 预分配容量以提高性能
	result := make(map[string]interface{}, 16)
	tbl.ForEach(func(k lua.LValue, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			result[string(ks)] = bodyToInterface(L, v)
		}
	})
	return result
}
