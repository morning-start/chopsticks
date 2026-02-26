package symlink

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 symlink 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 symlink 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	L.SetGlobal("symlink_create", L.NewFunction(func(L *lua.LState) int {
		oldname := L.CheckString(1)
		newname := L.CheckString(2)
		if err := Create(oldname, newname); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("symlink_is", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		isLink, err := Is(path)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(isLink))
		return 1
	}))

	L.SetGlobal("symlink_read", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		target, err := Read(path)
		if err != nil {
			L.Push(lua.LString(""))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(target))
		return 1
	}))

	L.SetGlobal("symlink_remove", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Remove(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))
}

// RegisterJS 向 JavaScript 运行时注册 symlink 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	vm.Set("symlink_create", func(call goja.FunctionCall) goja.Value {
		oldname := call.Argument(0).String()
		newname := call.Argument(1).String()
		err := Create(oldname, newname)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("symlink_is", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isLink, _ := Is(path)
		return vm.ToValue(isLink)
	})

	vm.Set("symlink_read", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		target, err := Read(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": target, "error": nil})
	})

	vm.Set("symlink_remove", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Remove(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})
}
