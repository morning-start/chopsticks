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
		return vm.ToValue(result)
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
			"data":    result,
		})
	})

	pathObj.Set("base", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(Base(path))
	})

	pathObj.Set("dir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(Dir(path))
	})

	pathObj.Set("ext", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(Ext(path))
	})

	pathObj.Set("clean", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(Clean(path))
	})

	pathObj.Set("isAbs", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(IsAbs(path))
	})

	pathObj.Set("exists", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(Exists(path))
	})

	pathObj.Set("isDir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(IsDir(path))
	})

	vm.Set("path", pathObj)
}
