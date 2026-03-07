package chopsticksx

import (
	"strings"

	"github.com/dop251/goja"
)

// RegisterJS 向 JavaScript 运行时注册 chopsticks 模块。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	chopsticksObj := vm.NewObject()

	chopsticksObj.Set("getCookDir", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		version := call.Argument(1).String()
		path := m.GetCookDir(name, version)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    path,
		})
	})

	chopsticksObj.Set("getCurrentVersion", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		version, err := m.GetCurrentVersion(name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"version": version,
		})
	})

	chopsticksObj.Set("addToPath", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		scope := call.Argument(1).String()
		if err := m.AddToPath(path, scope); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("removeFromPath", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		scope := call.Argument(1).String()
		if err := m.RemoveFromPath(path, scope); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("setEnv", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		value := call.Argument(1).String()
		scope := call.Argument(2).String()
		if err := m.SetEnv(key, value, scope); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("getEnv", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		value := m.GetEnv(key)
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"value":   value,
		})
	})

	chopsticksObj.Set("createShim", func(call goja.FunctionCall) goja.Value {
		source := call.Argument(0).String()
		name := call.Argument(1).String()
		shimPath, err := m.CreateShim(source, name)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success":  false,
				"shimPath": "",
				"error":    err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success":  true,
			"shimPath": shimPath,
			"error":    nil,
		})
	})

	chopsticksObj.Set("removeShim", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		if err := m.RemoveShim(name); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("persistData", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		dirs := call.Argument(1).Export().([]interface{})
		strDirs := make([]string, len(dirs))
		for i, d := range dirs {
			strDirs[i] = d.(string)
		}
		if err := m.PersistData(name, strDirs); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("createShortcut", func(call goja.FunctionCall) goja.Value {
		opts := call.Argument(0).Export().(map[string]interface{})
		options := ShortcutOptions{}
		if v, ok := opts["source"]; ok {
			options.Source = v.(string)
		}
		if v, ok := opts["name"]; ok {
			options.Name = v.(string)
		}
		if v, ok := opts["description"]; ok {
			options.Description = v.(string)
		}
		if v, ok := opts["icon"]; ok {
			options.Icon = v.(string)
		}
		shortcutPath, err := m.CreateShortcut(options)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"success":      false,
				"shortcutPath": "",
				"error":        err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{
			"success":      true,
			"shortcutPath": shortcutPath,
			"error":        nil,
		})
	})

	chopsticksObj.Set("getCacheDir", func(call goja.FunctionCall) goja.Value {
		path := m.GetCacheDir()
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    path,
		})
	})

	chopsticksObj.Set("getConfigDir", func(call goja.FunctionCall) goja.Value {
		path := m.GetConfigDir()
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    path,
		})
	})

	chopsticksObj.Set("deleteEnv", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		scope := call.Argument(1).String()
		if err := m.DeleteEnv(key, scope); err != nil {
			return vm.ToValue(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true})
	})

	chopsticksObj.Set("getPath", func(call goja.FunctionCall) goja.Value {
		pathStr := m.GetPath()
		paths := strings.Split(pathStr, ";")
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"paths":   paths,
		})
	})

	chopsticksObj.Set("getShimDir", func(call goja.FunctionCall) goja.Value {
		path := m.GetShimDir()
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    path,
		})
	})

	chopsticksObj.Set("getPersistDir", func(call goja.FunctionCall) goja.Value {
		path := m.GetPersistDir()
		return vm.ToValue(map[string]interface{}{
			"success": true,
			"path":    path,
		})
	})

	vm.Set("chopsticks", chopsticksObj)
}
