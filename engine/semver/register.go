package semver

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 semver 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 semver 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	L.SetGlobal("semver_parse", L.NewFunction(func(L *lua.LState) int {
		version := L.CheckString(1)
		v, err := Parse(version)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		tbl.RawSetString("major", lua.LNumber(v.Major))
		tbl.RawSetString("minor", lua.LNumber(v.Minor))
		tbl.RawSetString("patch", lua.LNumber(v.Patch))
		tbl.RawSetString("string", lua.LString(v.String()))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("semver_compare", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		result, err := CompareStrings(v1, v2)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(result))
		return 1
	}))

	L.SetGlobal("semver_gt", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		ver1, err := Parse(v1)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ver2, err := Parse(v2)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(ver1.GT(ver2)))
		return 1
	}))

	L.SetGlobal("semver_lt", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		ver1, err := Parse(v1)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ver2, err := Parse(v2)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(ver1.LT(ver2)))
		return 1
	}))

	L.SetGlobal("semver_eq", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		ver1, err := Parse(v1)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ver2, err := Parse(v2)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(ver1.EQ(ver2)))
		return 1
	}))

	L.SetGlobal("semver_gte", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		ver1, err := Parse(v1)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ver2, err := Parse(v2)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(ver1.GTE(ver2)))
		return 1
	}))

	L.SetGlobal("semver_lte", L.NewFunction(func(L *lua.LState) int {
		v1 := L.CheckString(1)
		v2 := L.CheckString(2)

		ver1, err := Parse(v1)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ver2, err := Parse(v2)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(ver1.LTE(ver2)))
		return 1
	}))

	L.SetGlobal("semver_satisfies", L.NewFunction(func(L *lua.LState) int {
		version := L.CheckString(1)
		constraint := L.CheckString(2)

		ok, err := Satisfies(version, constraint)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(ok))
		return 1
	}))
}

// RegisterJS 向 JavaScript 运行时注册 semver 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	vm.Set("semver_parse", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_compare", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		result, err := CompareStrings(v1, v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	vm.Set("semver_gt", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_lt", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_eq", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_gte", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_lte", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("semver_satisfies", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()
		constraint := call.Argument(1).String()

		ok, err := Satisfies(version, constraint)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": ok})
	})
}
