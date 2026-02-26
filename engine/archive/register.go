package archive

import (
	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

// Module 为脚本引擎提供 archive 注册。
type Module struct{}

// RegisterLua 向 Lua 状态注册 archive 函数。
func (m *Module) RegisterLua(L *lua.LState) {
	L.SetGlobal("archive_extract", L.NewFunction(func(L *lua.LState) int {
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

	L.SetGlobal("archive_extract_zip", L.NewFunction(func(L *lua.LState) int {
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

	L.SetGlobal("archive_extract_tar", L.NewFunction(func(L *lua.LState) int {
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

	L.SetGlobal("archive_extract_targz", L.NewFunction(func(L *lua.LState) int {
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

	L.SetGlobal("archive_list", L.NewFunction(func(L *lua.LState) int {
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
			fileTbl.RawSetString("is_dir", lua.LBool(f.IsDir))
			tbl.Append(fileTbl)
		}
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("archive_detect_type", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		typ := DetectType(path)
		L.Push(lua.LString(typ))
		return 1
	}))

	L.SetGlobal("archive_is_archive", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		L.Push(lua.LBool(IsArchive(path)))
		return 1
	}))
}

// RegisterJS 向 JavaScript 运行时注册 archive 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	vm.Set("archive_extract", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := Extract(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("archive_extract_zip", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractZip(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("archive_extract_tar", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractTar(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("archive_extract_targz", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()

		if err := ExtractTarGz(src, dest); err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "error": nil})
	})

	vm.Set("archive_list", func(call goja.FunctionCall) goja.Value {
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
				"name":   f.Name,
				"size":   f.Size,
				"is_dir": f.IsDir,
			})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "data": result})
	})

	vm.Set("archive_detect_type", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		typ := DetectType(path)
		return vm.ToValue(map[string]interface{}{"success": true, "data": typ})
	})

	vm.Set("archive_is_archive", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		return vm.ToValue(map[string]interface{}{"success": true, "data": IsArchive(path)})
	})
}
