package fsutil

import (
	"fmt"

	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 fsutil 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 fsutil 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	fsObj := vm.NewObject()

	fsObj.Set("readFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content, err := Read(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "data": "", "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": content, "error": nil})
	})

	fsObj.Set("writeFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content := call.Argument(1).String()
		err := Write(path, content)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("append", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		content := call.Argument(1).String()
		err := Append(path, content)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("mkdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Mkdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("rmdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Rmdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("remove", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Remove(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("exists", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		exists, err := Exists(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "exists": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "exists": exists, "error": nil})
	})

	fsObj.Set("isdir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isDir, err := IsDir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "isdir": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "isdir": isDir, "error": nil})
	})

	fsObj.Set("readDir", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		entries, err := List(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "entries": []string{}, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "entries": entries, "error": nil})
	})

	fsObj.Set("copy", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dst := call.Argument(1).String()
		err := Copy(src, dst)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("removeAll", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Rmdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("mkdirAll", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := Mkdir(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	fsObj.Set("isFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		isFile, err := IsFile(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "isFile": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "isFile": isFile, "error": nil})
	})

	fsObj.Set("stat", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		info, err := Stat(path)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"size":     0,
				"isDirectory": false,
				"isFile":   false,
				"mtime":    "",
				"error":    err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success":     true,
			"size":        info.Size,
			"isDirectory": info.IsDir,
			"isFile":      info.IsFile,
			"mtime":       fmt.Sprintf("%d", info.ModTime),
			"error":       nil,
		})
	})

	vm.Set("fs", fsObj)
}
