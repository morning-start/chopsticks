package jsonx

import (
	"encoding/json"

	"github.com/dop251/goja"
)

// RegisterJS 向 JavaScript 运行时注册 json 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	jsonObj := vm.NewObject()

	jsonObj.Set("stringify", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).Export()
		spaceArg := call.Argument(1)

		var jsonStr []byte
		var err error

		// 检查是否提供了 space 参数
		if !goja.IsUndefined(spaceArg) && !goja.IsNull(spaceArg) {
			// 使用格式化输出
			space := spaceArg.Export()
			var prefix string
			var indent string

			switch v := space.(type) {
			case int, int64, float64:
				// 数字类型：转换为空格字符串
				indent = ""
				for i := 0; i < int(v.(float64)); i++ {
					indent += " "
				}
			case string:
				// 字符串类型：直接使用
				indent = v
			default:
				// 其他类型：使用默认缩进
				indent = "  "
			}

			jsonStr, err = json.MarshalIndent(data, prefix, indent)
		} else {
			// 紧凑格式
			jsonStr, err = json.Marshal(data)
		}

		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"json":    string(jsonStr),
		})
	})

	jsonObj.Set("parse", func(call goja.FunctionCall) goja.Value {
		jsonStr := call.Argument(0).String()
		var data interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	})

	vm.Set("json", jsonObj)
}
