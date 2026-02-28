package registry

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 registry 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 registry 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	regTable := L.NewTable()

	regTable.RawSetString("setValue", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		name := L.CheckString(2)
		value := L.CheckString(3)

		key, err := CreateKey(keyPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer key.Close()

		if err := SetStringValue(key, name, value); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	regTable.RawSetString("getValue", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		name := L.CheckString(2)

		key, err := OpenKey(keyPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer key.Close()

		value, err := GetStringValue(key, name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(value))
		return 1
	}))

	regTable.RawSetString("setDword", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		name := L.CheckString(2)
		value := L.CheckInt(3)

		key, err := CreateKey(keyPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer key.Close()

		if err := SetDWordValue(key, name, uint32(value)); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	regTable.RawSetString("getDword", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		name := L.CheckString(2)

		key, err := OpenKey(keyPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer key.Close()

		value, err := GetDWordValue(key, name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(value))
		return 1
	}))

	regTable.RawSetString("deleteValue", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		name := L.CheckString(2)

		key, err := OpenKey(keyPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer key.Close()

		if err := DeleteValue(key, name); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	regTable.RawSetString("createKey", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)

		key, err := CreateKey(keyPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		key.Close()
		L.Push(lua.LBool(true))
		return 1
	}))

	regTable.RawSetString("deleteKey", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)

		if err := DeleteKey(keyPath); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	regTable.RawSetString("keyExists", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		exists, err := KeyExists(keyPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(exists))
		return 1
	}))

	regTable.RawSetString("listKeys", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		keys, err := ListKeys(keyPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		table := L.NewTable()
		for _, key := range keys {
			table.Append(lua.LString(key))
		}
		L.Push(table)
		return 1
	}))

	regTable.RawSetString("listValues", L.NewFunction(func(L *lua.LState) int {
		keyPath := L.CheckString(1)
		values, err := ListValues(keyPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		table := L.NewTable()
		for _, value := range values {
			table.Append(lua.LString(value))
		}
		L.Push(table)
		return 1
	}))

	L.SetGlobal("registry", regTable)
}

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
			return vm.ToValue(map[string]interface{}{"value": nil, "error": err.Error()})
		}
		defer key.Close()

		value, err := GetStringValue(key, name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"value": nil, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"value": value, "error": nil})
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
			return vm.ToValue(map[string]interface{}{"value": nil, "error": err.Error()})
		}
		defer key.Close()

		value, err := GetDWordValue(key, name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"value": nil, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"value": value, "error": nil})
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
		values, err := ListValues(keyPath)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"values":  values,
		})
	})

	vm.Set("registry", regObj)
}
