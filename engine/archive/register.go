package archive

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 archive 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 archive 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	archiveTable := L.NewTable()

	archiveTable.RawSetString("extract", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dest := L.CheckString(2)

		if err := Extract(src, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	archiveTable.RawSetString("extractZip", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dest := L.CheckString(2)

		if err := ExtractZip(src, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	archiveTable.RawSetString("extract7z", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dest := L.CheckString(2)

		if err := Extract7z(src, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	archiveTable.RawSetString("extractTar", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dest := L.CheckString(2)

		if err := ExtractTar(src, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	archiveTable.RawSetString("extractTarGz", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)
		dest := L.CheckString(2)

		if err := ExtractTarGz(src, dest); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	archiveTable.RawSetString("list", L.NewFunction(func(L *lua.LState) int {
		src := L.CheckString(1)

		typ := DetectType(src)
		var files []FileInfo
		var err error

		switch typ {
		case ZIP:
			files, err = ListZip(src)
		case TAR, TARGZ:
			files, err = ListTar(src)
		default:
			L.Push(lua.LNil)
			L.Push(lua.LString("不支持的压缩类型"))
			return 2
		}

		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		for _, f := range files {
			fileTbl := L.NewTable()
			fileTbl.RawSetString("name", lua.LString(f.Name))
			fileTbl.RawSetString("size", lua.LNumber(f.Size))
			fileTbl.RawSetString("isDir", lua.LBool(f.IsDir))
			tbl.Append(fileTbl)
		}
		L.Push(tbl)
		return 1
	}))

	archiveTable.RawSetString("detectType", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		typ := DetectType(path)
		L.Push(lua.LString(typ))
		return 1
	}))

	archiveTable.RawSetString("isArchive", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		L.Push(lua.LBool(IsArchive(path)))
		return 1
	}))

	L.SetGlobal("archive", archiveTable)
}

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
