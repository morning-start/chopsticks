package checksum

import (
	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 checksum 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 checksum 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	checksumObj := vm.NewObject()

	checksumObj.Set("md5", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, MD5)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	checksumObj.Set("sha256", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, SHA256)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	checksumObj.Set("sha512", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		sum, err := CalculateFile(path, SHA512)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"data": sum, "error": nil})
	})

	checksumObj.Set("verify", func(call goja.FunctionCall) goja.Value {
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

	checksumObj.Set("string", func(call goja.FunctionCall) goja.Value {
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

	vm.Set("checksum", checksumObj)
}
