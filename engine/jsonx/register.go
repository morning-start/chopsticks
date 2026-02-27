package jsonx

import (
	"encoding/json"

	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLua 向 Lua 状态注册 json 模块。
func (m *Module) RegisterLua(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "stringify", L.NewFunction(func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		data, err := luaTableToJSON(tbl)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		jsonStr, err := json.Marshal(data)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(jsonStr))
		return 1
	}))

	L.SetField(mod, "parse", L.NewFunction(func(L *lua.LState) int {
		jsonStr := L.CheckString(1)
		var data interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := jsonToLuaTable(L, data)
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("json", mod)
}

// RegisterJS 向 JavaScript 运行时注册 json 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	jsonObj := vm.NewObject()

	jsonObj.Set("stringify", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).Export()
		jsonStr, err := json.Marshal(data)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    string(jsonStr),
		})
	})

	jsonObj.Set("parse", func(call goja.FunctionCall) goja.Value {
		jsonStr := call.Argument(0).String()
		var data interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	})

	vm.Set("json", jsonObj)
}

// luaTableToJSON 将 Lua 表转换为 Go 值。
func luaTableToJSON(tbl *lua.LTable) (interface{}, error) {
	result := make(map[string]interface{})
	tbl.ForEach(func(k lua.LValue, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			result[string(ks)] = luaValueToGo(v)
		}
	})
	return result, nil
}

// luaValueToGo 将 Lua 值转换为 Go 值。
func luaValueToGo(v lua.LValue) interface{} {
	switch v := v.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		result, _ := luaTableToJSON(v)
		return result
	default:
		return nil
	}
}

// jsonToLuaTable 将 JSON 数据转换为 Lua 表。
func jsonToLuaTable(L *lua.LState, data interface{}) *lua.LTable {
	tbl := L.NewTable()
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			L.SetField(tbl, key, goValueToLua(L, val))
		}
	case []interface{}:
		for i, val := range v {
			tbl.RawSetInt(i+1, goValueToLua(L, val))
		}
	}
	return tbl
}

// goValueToLua 将 Go 值转换为 Lua 值。
func goValueToLua(L *lua.LState, v interface{}) lua.LValue {
	switch v := v.(type) {
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	case map[string]interface{}:
		return jsonToLuaTable(L, v)
	case []interface{}:
		return jsonToLuaTable(L, v)
	default:
		return lua.LNil
	}
}
