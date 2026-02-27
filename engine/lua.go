package engine

import (
	"chopsticks/engine/fetch"
	"chopsticks/engine/fsutil"
	"chopsticks/engine/registry"
	"chopsticks/engine/symlink"

	lua "github.com/yuin/gopher-lua"
)

// 编译时接口检查。
var _ Engine = (*LuaEngine)(nil)

// LuaEngine 实现 Lua 的 Engine 接口。
type LuaEngine struct {
	L *lua.LState
}

// NewLuaEngine 创建新的 LuaEngine。
func NewLuaEngine() *LuaEngine {
	L := lua.NewState()
	RegisterLuaAll(L,
		&fsutil.Module{},
		&fetch.Module{},
		&symlink.Module{},
		&registry.Module{},
	)
	return &LuaEngine{
		L: L,
	}
}

// LoadFile 加载 Lua 脚本文件。
func (e *LuaEngine) LoadFile(path string) error {
	return e.L.DoFile(path)
}

// CallFunction 调用 Lua 函数。
func (e *LuaEngine) CallFunction(name string, args ...interface{}) error {
	fn := e.L.GetGlobal(name)
	if fn.Type() == lua.LTNil {
		return nil
	}

	luaArgs := make([]lua.LValue, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			luaArgs[i] = lua.LString(v)
		case int:
			luaArgs[i] = lua.LNumber(v)
		case float64:
			luaArgs[i] = lua.LNumber(v)
		default:
			luaArgs[i] = lua.LString("")
		}
	}

	return e.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, luaArgs...)
}

// Close 关闭 Lua 引擎。
func (e *LuaEngine) Close() {
	e.L.Close()
}
