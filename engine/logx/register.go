package logx

import (
	"fmt"

	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLua 向 Lua 状态注册 log 模块。
func (m *Module) RegisterLua(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "debug", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		m.logger.Debug(msg)
		return 0
	}))

	L.SetField(mod, "info", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		m.logger.Info(msg)
		return 0
	}))

	L.SetField(mod, "warn", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		m.logger.Warn(msg)
		return 0
	}))

	L.SetField(mod, "error", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		m.logger.Error(msg)
		return 0
	}))

	L.SetGlobal("log", mod)
}

// RegisterJS 向 JavaScript 运行时注册 log 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	logObj := vm.NewObject()

	logObj.Set("debug", func(call goja.FunctionCall) goja.Value {
		msg := call.Argument(0).String()
		m.logger.Debug(msg)
		return goja.Undefined()
	})

	logObj.Set("info", func(call goja.FunctionCall) goja.Value {
		msg := call.Argument(0).String()
		m.logger.Info(msg)
		return goja.Undefined()
	})

	logObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		msg := call.Argument(0).String()
		m.logger.Warn(msg)
		return goja.Undefined()
	})

	logObj.Set("error", func(call goja.FunctionCall) goja.Value {
		msg := call.Argument(0).String()
		m.logger.Error(msg)
		return goja.Undefined()
	})

	vm.Set("log", logObj)
}

// GlobalModule 是全局 log 模块实例。
var GlobalModule = NewModule()

// RegisterLuaGlobal 使用全局模块注册到 Lua。
func RegisterLuaGlobal(L *lua.LState) {
	GlobalModule.RegisterLua(L)
}

// RegisterJSGlobal 使用全局模块注册到 JS。
func RegisterJSGlobal(vm *goja.Runtime) {
	GlobalModule.RegisterJS(vm)
}

// printf 辅助函数用于格式化输出。
func printf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
