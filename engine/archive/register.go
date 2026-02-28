package archive

import (
	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 archive 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 archive 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	archiveObj := vm.NewObject()

	archiveObj.Set("extract", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := Extract(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	archiveObj.Set("extractZip", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractZip(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	archiveObj.Set("extract7z", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := Extract7z(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	archiveObj.Set("extractTar", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractTar(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	archiveObj.Set("extractTarGz", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractTarGz(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	archiveObj.Set("list", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()

		typ := DetectType(src)
		var files []FileInfo
		var err error

		switch typ {
		case ZIP:
			files, err = ListZip(src)
		case TAR, TARGZ:
			files, err = ListTar(src)
		default:
			return vm.ToValue(map[string]interface{}{"success": false, "error": "不支持的压缩类型"})
		}

		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}

		var result []map[string]interface{}
		for _, f := range files {
			result = append(result, map[string]interface{}{
				"name":  f.Name,
				"size":  f.Size,
				"isDir": f.IsDir,
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	archiveObj.Set("detectType", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		typ := DetectType(path)
		return vm.ToValue(map[string]interface{}{"success": true, "data": typ})
	})

	archiveObj.Set("isArchive", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(map[string]interface{}{"success": true, "data": IsArchive(path)})
	})

	vm.Set("archive", archiveObj)
}
