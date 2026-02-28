package logx

import (
	"github.com/dop251/goja"
)

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

// RegisterJSGlobal 使用全局模块注册到 JS。
func RegisterJSGlobal(vm *goja.Runtime) {
	GlobalModule.RegisterJS(vm)
}
