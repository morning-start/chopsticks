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

		files, err := ExtractWithFiles(src, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "extractedFiles": files, "error": nil})
	})

	archiveObj.Set("extractZip", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		files, err := ExtractZipWithFiles(src, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "extractedFiles": files, "error": nil})
	})

	archiveObj.Set("extract7z", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		files, err := Extract7zWithFiles(src, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "extractedFiles": files, "error": nil})
	})

	archiveObj.Set("extractTar", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		files, err := ExtractTarWithFiles(src, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "extractedFiles": files, "error": nil})
	})

	archiveObj.Set("extractTarGz", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		files, err := ExtractTarGzWithFiles(src, dest)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "extractedFiles": files, "error": nil})
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
		return vm.ToValue(map[string]interface{}{"success": true, "files": result})
	})

	archiveObj.Set("detectType", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		typ := DetectType(path)
		return vm.ToValue(map[string]interface{}{"success": true, "type": typ})
	})

	archiveObj.Set("isArchive", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(map[string]interface{}{"success": true, "isArchive": IsArchive(path)})
	})

	vm.Set("archive", archiveObj)
}
