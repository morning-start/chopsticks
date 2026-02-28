package execx

import (
	"github.com/dop251/goja"
)

// RegisterJS 向 JavaScript 运行时注册 exec 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	execObj := vm.NewObject()

	execObj.Set("exec", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		if len(args) == 0 {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "no command specified",
				"success":  false,
			})
		}

		name := args[0].String()
		execArgs := make([]string, len(args)-1)
		for i := 1; i < len(args); i++ {
			execArgs[i-1] = args[i].String()
		}

		result, err := Exec(name, execArgs...)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   err.Error(),
				"success":  false,
			})
		}

		return vm.ToValue(map[string]interface{}{
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"success":  result.Success,
		})
	})

	execObj.Set("shell", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "no command specified",
				"success":  false,
			})
		}

		command := call.Argument(0).String()
		result, err := ExecShell(command)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   err.Error(),
				"success":  false,
			})
		}

		return vm.ToValue(map[string]interface{}{
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"success":  result.Success,
		})
	})

	execObj.Set("powershell", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "no command specified",
				"success":  false,
			})
		}

		command := call.Argument(0).String()
		result, err := ExecPowerShell(command)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"exitCode": -1,
				"stdout":   "",
				"stderr":   err.Error(),
				"success":  false,
			})
		}

		return vm.ToValue(map[string]interface{}{
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"success":  result.Success,
		})
	})

	vm.Set("exec", execObj)
}
