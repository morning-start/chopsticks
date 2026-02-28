package installerx

import (
	"chopsticks/infra/installer"

	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 installer 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 installer 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	instObj := vm.NewObject()

	instObj.Set("run", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()

		opts := installer.Options{
			Silent: true,
		}

		if len(call.Arguments) >= 2 {
			if options, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if dir, ok := options["installDir"].(string); ok {
					opts.InstallDir = dir
				}
				if silent, ok := options["silent"].(bool); ok {
					opts.Silent = silent
				}
			}
		}

		if err := installer.Run(path, opts); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	instObj.Set("runNSIS", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()

		opts := installer.Options{
			Silent: true,
		}

		if len(call.Arguments) >= 2 {
			if options, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if dir, ok := options["installDir"].(string); ok {
					opts.InstallDir = dir
				}
			}
		}

		typ := installer.DetectType(path)
		if typ != installer.NSIS {
			return vm.ToValue(map[string]interface{}{"success": false, "error": "不是有效的 NSIS 安装程序"})
		}

		if err := installer.Run(path, opts); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	instObj.Set("runMSI", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()

		opts := installer.Options{
			Silent: true,
		}

		if len(call.Arguments) >= 2 {
			if options, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if dir, ok := options["installDir"].(string); ok {
					opts.InstallDir = dir
				}
			}
		}

		typ := installer.DetectType(path)
		if typ != installer.MSI {
			return vm.ToValue(map[string]interface{}{"success": false, "error": "不是有效的 MSI 安装程序"})
		}

		if err := installer.Run(path, opts); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	instObj.Set("runInno", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()

		opts := installer.Options{
			Silent: true,
		}

		if len(call.Arguments) >= 2 {
			if options, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if dir, ok := options["installDir"].(string); ok {
					opts.InstallDir = dir
				}
			}
		}

		typ := installer.DetectType(path)
		if typ != installer.Inno {
			return vm.ToValue(map[string]interface{}{"success": false, "error": "不是有效的 Inno Setup 安装程序"})
		}

		if err := installer.Run(path, opts); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	instObj.Set("detectType", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		typ := installer.DetectType(path)
		return vm.ToValue(string(typ))
	})

	vm.Set("installer", instObj)
}
