package checksum

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 checksum 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 checksum 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	L.SetGlobal("checksum_md5", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		sum, err := CalculateFile(path, MD5)
		if err != nil {
			L.Push(lua.LString(""))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(sum))
		return 1
	}))

	L.SetGlobal("checksum_sha256", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		sum, err := CalculateFile(path, SHA256)
		if err != nil {
			L.Push(lua.LString(""))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(sum))
		return 1
	}))

	L.SetGlobal("checksum_sha512", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		sum, err := CalculateFile(path, SHA512)
		if err != nil {
			L.Push(lua.LString(""))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(sum))
		return 1
	}))

	L.SetGlobal("checksum_verify", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		expected := L.CheckString(2)
		alg := L.OptString(3, "sha256")

		var algorithm Algorithm
		switch alg {
		case "md5":
			algorithm = MD5
		case "sha256":
			algorithm = SHA256
		case "sha512":
			algorithm = SHA512
		default:
			algorithm = SHA256
		}

		ok, err := VerifyFile(path, expected, algorithm)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(ok))
		return 1
	}))

	L.SetGlobal("checksum_string", L.NewFunction(func(L *lua.LState) int {
		data := L.CheckString(1)
		alg := L.OptString(2, "sha256")

		var algorithm Algorithm
		switch alg {
		case "md5":
			algorithm = MD5
		case "sha256":
			algorithm = SHA256
		case "sha512":
			algorithm = SHA512
		default:
			algorithm = SHA256
		}

		sum := New(algorithm).CalculateString(data)
		L.Push(lua.LString(sum))
		return 1
	}))
}

// RegisterJS 向 JavaScript 运行时注册 checksum 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	vm.Set("checksum_md5", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, MD5)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	vm.Set("checksum_sha256", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, SHA256)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	vm.Set("checksum_sha512", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, SHA512)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	vm.Set("checksum_verify", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		expected := call.Argument(1).String()
		alg := call.Argument(2).String()
		if alg == "" {
			alg = "sha256"
		}

		var algorithm Algorithm
		switch alg {
		case "md5":
			algorithm = MD5
		case "sha256":
			algorithm = SHA256
		case "sha512":
			algorithm = SHA512
		default:
			algorithm = SHA256
		}

		ok, err := VerifyFile(path, expected, algorithm)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": ok, "error": nil})
	})

	vm.Set("checksum_string", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).String()
		alg := call.Argument(1).String()
		if alg == "" {
			alg = "sha256"
		}

		var algorithm Algorithm
		switch alg {
		case "md5":
			algorithm = MD5
		case "sha256":
			algorithm = SHA256
		case "sha512":
			algorithm = SHA512
		default:
			algorithm = SHA256
		}

		sum := New(algorithm).CalculateString(data)
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})
}
