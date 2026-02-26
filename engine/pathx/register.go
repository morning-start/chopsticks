package pathx

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLua 向 Lua 状态注册 path 模块。
func (m *Module) RegisterLua(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "join", L.NewFunction(func(L *lua.LState) int {
		n := L.GetTop()
		elems := make([]string, n)
		for i := 1; i <= n; i++ {
			elems[i-1] = L.CheckString(i)
		}
		result := Join(elems...)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "abs", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result, err := Abs(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "base", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := Base(path)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "dir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := Dir(path)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "ext", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := Ext(path)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "clean", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := Clean(path)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "is_abs", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := IsAbs(path)
		L.Push(lua.LBool(result))
		return 1
	}))

	L.SetField(mod, "exists", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := Exists(path)
		L.Push(lua.LBool(result))
		return 1
	}))

	L.SetField(mod, "is_dir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		result := IsDir(path)
		L.Push(lua.LBool(result))
		return 1
	}))

	L.SetGlobal("path", mod)
}

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
