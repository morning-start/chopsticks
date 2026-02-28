package fsutil

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 fsutil 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 fsutil 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	fsTable := L.NewTable()

	fsTable.RawSetString("readFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		content, err := Read(path)
		if err != nil {
			L.Push(lua.LString(""))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(content))
		return 1
	}))

	fsTable.RawSetString("writeFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		content := L.CheckString(2)
		if err := Write(path, content); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("append", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		content := L.CheckString(2)
		if err := Append(path, content); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("mkdir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Mkdir(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("rmdir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Rmdir(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("remove", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Remove(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("exists", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		exists, err := Exists(path)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(exists))
		return 1
	}))

	fsTable.RawSetString("isdir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		isDir, err := IsDir(path)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(isDir))
		return 1
	}))

	fsTable.RawSetString("readDir", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		entries, err := List(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		table := L.NewTable()
		for _, entry := range entries {
			table.Append(lua.LString(entry))
		}
		L.Push(table)
		return 1
	}))

	fsTable.RawSetString("copy", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dst := L.CheckString(2)
		if err := Copy(src, dst); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("removeAll", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Rmdir(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("mkdirAll", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := Mkdir(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	fsTable.RawSetString("isFile", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		isFile, err := IsFile(path)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(isFile))
		return 1
	}))

	fsTable.RawSetString("stat", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		info, err := Stat(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// 创建结果表
		result := L.NewTable()
		L.SetField(result, "name", lua.LString(info.Name))
		L.SetField(result, "size", lua.LNumber(info.Size))
		L.SetField(result, "isDir", lua.LBool(info.IsDir))
		L.SetField(result, "isFile", lua.LBool(info.IsFile))
		L.SetField(result, "modTime", lua.LNumber(info.ModTime))
		L.SetField(result, "mode", lua.LNumber(info.Mode))

		L.Push(result)
		return 1
	}))

	L.SetGlobal("fs", fsTable)
}

// RegisterJS 向 JavaScript 运行时注册 fsutil 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	fsObj := vm.NewObject()

	fsObj.Set("readFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content, err := Read(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": content, "error": nil})
	})

	fsObj.Set("writeFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content := call.Argument(1).String()
		err := Write(path, content)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("append", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content := call.Argument(1).String()
		err := Append(path, content)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("mkdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Mkdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("rmdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Rmdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("remove", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Remove(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("exists", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		exists, _ := Exists(path)
		return vm.ToValue(exists)
	})

	fsObj.Set("isdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isDir, _ := IsDir(path)
		return vm.ToValue(isDir)
	})

	fsObj.Set("readDir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		entries, _ := List(path)
		return vm.ToValue(entries)
	})

	fsObj.Set("copy", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dst := call.Argument(1).String()
		err := Copy(src, dst)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("removeAll", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Rmdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("mkdirAll", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Mkdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("isFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isFile, _ := IsFile(path)
		return vm.ToValue(isFile)
	})

	fsObj.Set("stat", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		info, err := Stat(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    info,
		})
	})

	vm.Set("fs", fsObj)
}
