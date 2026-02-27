package symlink

import (
	"os"
	"runtime"

	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 symlink 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 symlink 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	symlinkTable := L.NewTable()

	symlinkTable.RawSetString("create", L.NewFunction(func(L *lua.LState) int {
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

	symlinkTable.RawSetString("createDir", L.NewFunction(func(L *lua.LState) int {
		oldname := L.CheckString(1)
		newname := L.CheckString(2)
		if err := CreateDir(oldname, newname); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	symlinkTable.RawSetString("createHard", L.NewFunction(func(L *lua.LState) int {
		oldname := L.CheckString(1)
		newname := L.CheckString(2)
		if err := CreateHard(oldname, newname); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	symlinkTable.RawSetString("createJunction", L.NewFunction(func(L *lua.LState) int {
		oldname := L.CheckString(1)
		newname := L.CheckString(2)
		if err := CreateJunction(oldname, newname); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	symlinkTable.RawSetString("is", L.NewFunction(func(L *lua.LState) int {
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

	symlinkTable.RawSetString("read", L.NewFunction(func(L *lua.LState) int {
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

	symlinkTable.RawSetString("remove", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Remove(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("symlink", symlinkTable)
}

// RegisterJS 向 JavaScript 运行时注册 symlink 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	symlinkObj := vm.NewObject()

	symlinkObj.Set("create", func(call goja.FunctionCall) goja.Value {
		oldname := call.Argument(0).String()
		newname := call.Argument(1).String()
		err := Create(oldname, newname)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	symlinkObj.Set("createDir", func(call goja.FunctionCall) goja.Value {
		oldname := call.Argument(0).String()
		newname := call.Argument(1).String()
		err := CreateDir(oldname, newname)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	symlinkObj.Set("createHard", func(call goja.FunctionCall) goja.Value {
		oldname := call.Argument(0).String()
		newname := call.Argument(1).String()
		err := CreateHard(oldname, newname)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	symlinkObj.Set("createJunction", func(call goja.FunctionCall) goja.Value {
		oldname := call.Argument(0).String()
		newname := call.Argument(1).String()
		err := CreateJunction(oldname, newname)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	symlinkObj.Set("is", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isLink, _ := Is(path)
		return vm.ToValue(isLink)
	})

	symlinkObj.Set("read", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		target, err := Read(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": target, "error": nil})
	})

	symlinkObj.Set("remove", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Remove(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("symlink", symlinkObj)
}

// CreateDir 创建目录符号链接（Windows 需要管理员权限）。
func CreateDir(oldName, newName string) error {
	if runtime.GOOS == "windows" {
		return Create(oldName, newName)
	}
	return os.Symlink(oldName, newName)
}

// CreateHard 创建硬链接。
func CreateHard(oldName, newName string) error {
	if _, err := os.Lstat(newName); err == nil {
		if err := os.Remove(newName); err != nil {
			return err
		}
	}
	return os.Link(oldName, newName)
}

// CreateJunction 创建 Junction（仅 Windows）。
func CreateJunction(oldName, newName string) error {
	if runtime.GOOS != "windows" {
		return os.Symlink(oldName, newName)
	}
	// Windows 上使用 mklink /J 创建 Junction
	return Create(oldName, newName)
}
