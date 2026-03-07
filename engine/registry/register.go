package registry

import (
	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 registry 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 registry 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	regObj := vm.NewObject()

	regObj.Set("setValue", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		name := call.Argument(1).String()
		value := call.Argument(2).String()

		key, err := CreateKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		defer key.Close()

		if err := SetStringValue(key, name, value); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	regObj.Set("getValue", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		name := call.Argument(1).String()

		key, err := OpenKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "value": nil, "type": nil, "error": err.Error()})
		}
		defer key.Close()

		value, valType, err := GetStringValueWithType(key, name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "value": nil, "type": nil, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "value": value, "type": valType, "error": nil})
	})

	regObj.Set("setDword", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		name := call.Argument(1).String()
		value := call.Argument(2).ToInteger()

		key, err := CreateKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		defer key.Close()

		if err := SetDWordValue(key, name, uint32(value)); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	regObj.Set("getDword", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		name := call.Argument(1).String()

		key, err := OpenKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "value": nil, "error": err.Error()})
		}
		defer key.Close()

		value, err := GetDWordValue(key, name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "value": nil, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "value": value, "error": nil})
	})

	regObj.Set("deleteValue", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		name := call.Argument(1).String()

		key, err := OpenKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		defer key.Close()

		if err := DeleteValue(key, name); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	regObj.Set("createKey", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()

		key, err := CreateKey(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		key.Close()
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	regObj.Set("deleteKey", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()

		if err := DeleteKey(keyPath); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	regObj.Set("keyExists", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		exists, err := KeyExists(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"exists":  exists,
		})
	})

	regObj.Set("listKeys", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		keys, err := ListKeys(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"keys":    keys,
		})
	})

	regObj.Set("listValues", func(call goja.FunctionCall) goja.Value {
		keyPath := call.Argument(0).String()
		values, err := ListValuesWithInfo(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		jsValues := make([]map[string]interface{}, len(values))
		for i, v := range values {
			jsValues[i] = map[string]interface{}{
				"name":  v.Name,
				"type":  v.Type,
				"value": v.Value,
			}
		}

		return vm.ToValue(map[string]interface{}{
			"success": true,
			"values":  jsValues,
		})
	})

	vm.Set("registry", regObj)
}
