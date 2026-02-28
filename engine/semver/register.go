package semver

import (
	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 semver 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 semver 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	semverObj := vm.NewObject()

	semverObj.Set("parse", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()
		v, err := Parse(version)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"major":  v.Major,
				"minor":  v.Minor,
				"patch":  v.Patch,
				"string": v.String(),
			},
		})
	})

	semverObj.Set("compare", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		result, err := CompareStrings(v1, v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	semverObj.Set("gt", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ver1.GT(ver2)})
	})

	semverObj.Set("lt", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ver1.LT(ver2)})
	})

	semverObj.Set("eq", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ver1.EQ(ver2)})
	})

	semverObj.Set("gte", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ver1.GTE(ver2)})
	})

	semverObj.Set("lte", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ver1.LTE(ver2)})
	})

	semverObj.Set("satisfies", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()
		constraint := call.Argument(1).String()

		ok, err := Satisfies(version, constraint)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ok})
	})

	vm.Set("semver", semverObj)
}
