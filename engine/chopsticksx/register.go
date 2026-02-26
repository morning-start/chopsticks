package chopsticksx

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLua 向 Lua 状态注册 chopsticks 模块。
func (m *Module) RegisterLua(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "get_cook_dir", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		version := L.CheckString(2)
		result := m.GetCookDir(name, version)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "get_current_version", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		version, err := m.GetCurrentVersion(name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(version))
		return 1
	}))

	L.SetField(mod, "add_to_path", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := m.AddToPath(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "remove_from_path", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := m.RemoveFromPath(path); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "set_env", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		if err := m.SetEnv(key, value); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "get_env", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		result := m.GetEnv(key)
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(mod, "create_shim", L.NewFunction(func(L *lua.LState) int {
		source := L.CheckString(1)
		name := L.CheckString(2)
		if err := m.CreateShim(source, name); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "remove_shim", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		if err := m.RemoveShim(name); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "persist_data", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		tbl := L.CheckTable(2)
		var dirs []string
		tbl.ForEach(func(k lua.LValue, v lua.LValue) {
			if vs, ok := v.(lua.LString); ok {
				dirs = append(dirs, string(vs))
			}
		})
		if err := m.PersistData(name, dirs); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "create_shortcut", L.NewFunction(func(L *lua.LState) int {
		opts := L.CheckTable(1)
		options := ShortcutOptions{}
		if v := L.GetField(opts, "source"); v.Type() == lua.LTString {
			options.Source = string(v.(lua.LString))
		}
		if v := L.GetField(opts, "name"); v.Type() == lua.LTString {
			options.Name = string(v.(lua.LString))
		}
		if v := L.GetField(opts, "description"); v.Type() == lua.LTString {
			options.Description = string(v.(lua.LString))
		}
		if v := L.GetField(opts, "icon"); v.Type() == lua.LTString {
			options.Icon = string(v.(lua.LString))
		}
		if err := m.CreateShortcut(options); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("chopsticks", mod)
}

// RegisterJS 向 JavaScript 运行时注册 chopsticks 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	chopsticksObj := vm.NewObject()

	chopsticksObj.Set("getCookDir", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		version := call.Argument(1).String()
		result := m.GetCookDir(name, version)
		return vm.ToValue(result)
	})

	chopsticksObj.Set("getCurrentVersion", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		version, err := m.GetCurrentVersion(name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    version,
		})
	})

	chopsticksObj.Set("addToPath", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		if err := m.AddToPath(path); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("removeFromPath", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		if err := m.RemoveFromPath(path); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("setEnv", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		value := call.Argument(1).String()
		if err := m.SetEnv(key, value); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("getEnv", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		result := m.GetEnv(key)
		return vm.ToValue(result)
	})

	chopsticksObj.Set("createShim", func(call goja.FunctionCall) goja.Value {
		source := call.Argument(0).String()
		name := call.Argument(1).String()
		if err := m.CreateShim(source, name); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("removeShim", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		if err := m.RemoveShim(name); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("persistData", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		dirs := call.Argument(1).Export().([]interface{})
		strDirs := make([]string, len(dirs))
		for i, d := range dirs {
			strDirs[i] = d.(string)
		}
		if err := m.PersistData(name, strDirs); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("createShortcut", func(call goja.FunctionCall) goja.Value {
		opts := call.Argument(0).Export().(map[string]interface{})
		options := ShortcutOptions{}
		if v, ok := opts["source"]; ok {
			options.Source = v.(string)
		}
		if v, ok := opts["name"]; ok {
			options.Name = v.(string)
		}
		if v, ok := opts["description"]; ok {
			options.Description = v.(string)
		}
		if v, ok := opts["icon"]; ok {
			options.Icon = v.(string)
		}
		if err := m.CreateShortcut(options); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	vm.Set("chopsticks", chopsticksObj)
}
