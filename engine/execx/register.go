package execx

import (
	"time"

	"github.com/dop251/goja"
)

// RegisterJS 向 JavaScript 运行时注册 exec 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	execObj := vm.NewObject()

	execObj.Set("exec", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		if len(args) == 0 {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    "no command specified",
			})
		}

		name := args[0].String()

		// 解析 args 参数（支持 array 或 string）
		var execArgs []string
		var opts *Options

		// 查找 options 参数（最后一个参数如果是 object）
		if len(args) > 1 {
			lastArg := args[len(args)-1]
			if lastArgExport := lastArg.Export(); lastArgExport != nil {
				if _, isMap := lastArgExport.(map[string]interface{}); isMap {
					// 最后一个参数是 object，作为 options
					opts = parseOptions(lastArgExport)
					// args 参数在 options 之前
					if len(args) > 1 {
						execArgs = parseArgs(args[1 : len(args)-1])
					}
				} else {
					// 最后一个参数不是 object，全部作为 args
					execArgs = parseArgs(args[1:])
				}
			} else {
				// 最后一个参数为 null 或 undefined，全部作为 args
				execArgs = parseArgs(args[1:])
			}
		}

		result, err := ExecWithOptions(name, execArgs, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    err.Error(),
			})
		}

		return vm.ToValue(map[string]interface{}{
			"success":  result.Success,
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"error":    "",
		})
	})

	execObj.Set("shell", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    "no command specified",
			})
		}

		command := call.Argument(0).String()

		// 解析 options 参数
		var opts *Options
		if len(call.Arguments) > 1 {
			if argExport := call.Argument(1).Export(); argExport != nil {
				opts = parseOptions(argExport)
			}
		}

		result, err := ExecShellWithOptions(command, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    err.Error(),
			})
		}

		return vm.ToValue(map[string]interface{}{
			"success":  result.Success,
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"error":    "",
		})
	})

	execObj.Set("powershell", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    "no command specified",
			})
		}

		command := call.Argument(0).String()

		// 解析 options 参数
		var opts *Options
		if len(call.Arguments) > 1 {
			if argExport := call.Argument(1).Export(); argExport != nil {
				opts = parseOptions(argExport)
			}
		}

		result, err := ExecPowerShellWithOptions(command, opts)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"exitCode": -1,
				"stdout":   "",
				"stderr":   "",
				"error":    err.Error(),
			})
		}

		return vm.ToValue(map[string]interface{}{
			"success":  result.Success,
			"exitCode": result.ExitCode,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
			"error":    "",
		})
	})

	vm.Set("exec", execObj)
}

// parseArgs 解析参数，支持 array 或 string 类型
func parseArgs(args []goja.Value) []string {
	if len(args) == 0 {
		return nil
	}

	// 如果只有一个参数
	if len(args) == 1 {
		arg := args[0]
		if argExport := arg.Export(); argExport != nil {
			// 检查是否是数组
			if arr, isArray := argExport.([]interface{}); isArray {
				result := make([]string, len(arr))
				for i, v := range arr {
					if str, ok := v.(string); ok {
						result[i] = str
					} else {
						result[i] = ""
					}
				}
				return result
			}
			// 如果是字符串，直接返回
			if str, isString := argExport.(string); isString {
				return []string{str}
			}
		}
	}

	// 多个参数，全部转换为 string
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = arg.String()
	}
	return result
}

// parseOptions 解析 options 对象
func parseOptions(argExport interface{}) *Options {
	if argExport == nil {
		return nil
	}

	optionsMap, ok := argExport.(map[string]interface{})
	if !ok {
		return nil
	}

	opts := &Options{}

	// 解析 cwd
	if cwd, ok := optionsMap["cwd"].(string); ok {
		opts.CWD = cwd
	}

	// 解析 env
	if env, ok := optionsMap["env"].(map[string]interface{}); ok {
		opts.Env = make(map[string]string)
		for k, v := range env {
			if val, ok := v.(string); ok {
				opts.Env[k] = val
			}
		}
	}

	// 解析 timeout（毫秒转换为 time.Duration）
	if timeout, ok := optionsMap["timeout"].(float64); ok && timeout > 0 {
		opts.Timeout = time.Duration(timeout) * time.Millisecond
	}

	return opts
}
