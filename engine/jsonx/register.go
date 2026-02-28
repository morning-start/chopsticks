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
		jsonStr, err := json.Marshal(data)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    string(jsonStr),
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
