package pathx

import (
	"github.com/dop251/goja"
)

// RegisterJS 向 JavaScript 运行时注册 path 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	pathObj := vm.NewObject()

	pathObj.Set("join", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		elems := make([]string, len(args))
		for i, arg := range args {
			elems[i] = arg.String()
		}
		result := Join(elems...)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    result,
		})
	})

	pathObj.Set("abs", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result, err := Abs(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    result,
		})
	})

	pathObj.Set("base", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := Base(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"name":    result,
		})
	})

	pathObj.Set("dir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := Dir(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"dir":     result,
		})
	})

	pathObj.Set("ext", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := Ext(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"ext":     result,
		})
	})

	pathObj.Set("clean", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := Clean(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    result,
		})
	})

	pathObj.Set("isAbs", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := IsAbs(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"isAbs":   result,
		})
	})

	pathObj.Set("exists", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := Exists(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"exists":  result,
		})
	})

	pathObj.Set("isDir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		result := IsDir(path)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"isDir":   result,
		})
	})

	vm.Set("path", pathObj)
}
