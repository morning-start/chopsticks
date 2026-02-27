package execx

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// RegisterLua 向 Lua 状态注册 exec 模块。
func (m *Module) RegisterLua(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "exec", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		n := L.GetTop()
		args := make([]string, 0, n-1)
		for i := 2; i <= n; i++ {
			args = append(args, L.CheckString(i))
		}

		result, err := Exec(name, args...)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		L.SetField(tbl, "exit_code", lua.LNumber(result.ExitCode))
		L.SetField(tbl, "stdout", lua.LString(result.Stdout))
		L.SetField(tbl, "stderr", lua.LString(result.Stderr))
		L.SetField(tbl, "success", lua.LBool(result.Success))
		L.Push(tbl)
		return 1
	}))

	L.SetField(mod, "shell", L.NewFunction(func(L *lua.LState) int {
		command := L.CheckString(1)
		result, err := ExecShell(command)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		L.SetField(tbl, "exit_code", lua.LNumber(result.ExitCode))
		L.SetField(tbl, "stdout", lua.LString(result.Stdout))
		L.SetField(tbl, "stderr", lua.LString(result.Stderr))
		L.SetField(tbl, "success", lua.LBool(result.Success))
		L.Push(tbl)
		return 1
	}))

	L.SetField(mod, "powershell", L.NewFunction(func(L *lua.LState) int {
		command := L.CheckString(1)
		result, err := ExecPowerShell(command)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		L.SetField(tbl, "exit_code", lua.LNumber(result.ExitCode))
		L.SetField(tbl, "stdout", lua.LString(result.Stdout))
		L.SetField(tbl, "stderr", lua.LString(result.Stderr))
		L.SetField(tbl, "success", lua.LBool(result.Success))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("exec", mod)
}

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
