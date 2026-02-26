// Package engine 提供应用脚本执行功能。
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/core/manifest"

	lua "github.com/yuin/gopher-lua"
)

// ScriptExecutor 定义脚本执行器接口。
type ScriptExecutor interface {
	LoadScript(path string) error
	CheckVersion() (string, error)
	GetDownloadInfo(version, arch string) (*manifest.DownloadInfo, error)
	PreInstall(ctx *InstallContext) error
	PostInstall(ctx *InstallContext) error
	PreUninstall(ctx *InstallContext) error
	PostUninstall(ctx *InstallContext) error
	GetEnvPath() ([]string, error)
	GetBin() ([]string, error)
	GetPersist() ([]string, error)
	GetDepends() ([]Dependency, error)
	GetConflicts() ([]string, error)
}

// InstallContext 包含安装上下文。
type InstallContext struct {
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	InstallDir string `json:"install_dir"`
	AppName    string `json:"app_name"`
	Bucket     string `json:"bucket"`
}

// Dependency 定义依赖。
type Dependency struct {
	Name     string `json:"name"`
	Bucket   string `json:"bucket,omitempty"`
	Version  string `json:"version,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// luaScriptExecutor 是 Lua 脚本执行器的实现。
type luaScriptExecutor struct {
	eng  *LuaEngine
	path string
}

// 编译时接口检查。
var _ ScriptExecutor = (*luaScriptExecutor)(nil)

// NewLuaScriptExecutor 创建新的 Lua 脚本执行器。
func NewLuaScriptExecutor(eng *LuaEngine, scriptPath string) ScriptExecutor {
	return &luaScriptExecutor{
		eng:  eng,
		path: scriptPath,
	}
}

// LoadScript 加载脚本文件。
func (e *luaScriptExecutor) LoadScript(path string) error {
	return e.eng.LoadFile(path)
}

// CheckVersion 调用 check_version 函数获取最新版本。
func (e *luaScriptExecutor) CheckVersion() (string, error) {
	L := e.eng.L

	fn := L.GetGlobal("check_version")
	if fn.Type() == lua.LTNil {
		return "", fmt.Errorf("check_version 函数未定义")
	}

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return "", fmt.Errorf("调用 check_version 失败: %w", err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	if ret.Type() != lua.LTString {
		return "", fmt.Errorf("check_version 必须返回字符串")
	}

	return string(ret.(lua.LString)), nil
}

// GetDownloadInfo 调用 get_download_info 函数获取下载信息。
func (e *luaScriptExecutor) GetDownloadInfo(version, arch string) (*manifest.DownloadInfo, error) {
	L := e.eng.L

	fn := L.GetGlobal("get_download_info")
	if fn.Type() == lua.LTNil {
		return nil, fmt.Errorf("get_download_info 函数未定义")
	}

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, lua.LString(version), lua.LString(arch))
	if err != nil {
		return nil, fmt.Errorf("调用 get_download_info 失败: %w", err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	if ret.Type() != lua.LTTable {
		return nil, fmt.Errorf("get_download_info 必须返回表")
	}

	// 将 Lua 表转换为 JSON
	jsonStr, err := luaTableToJSON(L, ret.(*lua.LTable))
	if err != nil {
		return nil, fmt.Errorf("转换下载信息失败: %w", err)
	}

	var info manifest.DownloadInfo
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		return nil, fmt.Errorf("解析下载信息失败: %w", err)
	}

	return &info, nil
}

// PreInstall 调用 pre_install 钩子。
func (e *luaScriptExecutor) PreInstall(ctx *InstallContext) error {
	return e.callInstallHook("pre_install", ctx)
}

// PostInstall 调用 post_install 钩子。
func (e *luaScriptExecutor) PostInstall(ctx *InstallContext) error {
	return e.callInstallHook("post_install", ctx)
}

// PreUninstall 调用 pre_uninstall 钩子。
func (e *luaScriptExecutor) PreUninstall(ctx *InstallContext) error {
	return e.callInstallHook("pre_uninstall", ctx)
}

// PostUninstall 调用 post_uninstall 钩子。
func (e *luaScriptExecutor) PostUninstall(ctx *InstallContext) error {
	return e.callInstallHook("post_uninstall", ctx)
}

// callInstallHook 调用安装/卸载钩子。
func (e *luaScriptExecutor) callInstallHook(hookName string, ctx *InstallContext) error {
	L := e.eng.L

	fn := L.GetGlobal(hookName)
	if fn.Type() == lua.LTNil {
		// 钩子未定义，直接返回成功
		return nil
	}

	// 创建上下文表
	ctxTable := L.NewTable()
	L.SetField(ctxTable, "version", lua.LString(ctx.Version))
	L.SetField(ctxTable, "arch", lua.LString(ctx.Arch))
	L.SetField(ctxTable, "install_dir", lua.LString(ctx.InstallDir))
	L.SetField(ctxTable, "app_name", lua.LString(ctx.AppName))
	L.SetField(ctxTable, "bucket", lua.LString(ctx.Bucket))

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, ctxTable)
	if err != nil {
		return fmt.Errorf("调用 %s 失败: %w", hookName, err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	// 检查返回值
	if ret.Type() == lua.LTBool && !bool(ret.(lua.LBool)) {
		return fmt.Errorf("%s 返回 false，取消操作", hookName)
	}

	return nil
}

// GetEnvPath 获取环境变量 PATH。
func (e *luaScriptExecutor) GetEnvPath() ([]string, error) {
	return e.getStringArray("env_path")
}

// GetBin 获取可执行文件列表。
func (e *luaScriptExecutor) GetBin() ([]string, error) {
	return e.getStringArray("bin")
}

// GetPersist 获取持久化目录列表。
func (e *luaScriptExecutor) GetPersist() ([]string, error) {
	return e.getStringArray("persist")
}

// GetDepends 获取依赖列表。
func (e *luaScriptExecutor) GetDepends() ([]Dependency, error) {
	L := e.eng.L

	fn := L.GetGlobal("depends")
	if fn.Type() == lua.LTNil {
		return nil, nil
	}

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 depends 失败: %w", err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	if ret.Type() != lua.LTTable {
		return nil, nil
	}

	// 转换依赖列表
	var deps []Dependency
	table := ret.(*lua.LTable)
	table.ForEach(func(_, v lua.LValue) {
		if v.Type() == lua.LTTable {
			depTable := v.(*lua.LTable)
			dep := Dependency{
				Name:     getStringField(L, depTable, "name"),
				Bucket:   getStringField(L, depTable, "bucket"),
				Version:  getStringField(L, depTable, "version"),
				Optional: getBoolField(L, depTable, "optional"),
			}
			deps = append(deps, dep)
		}
	})

	return deps, nil
}

// GetConflicts 获取冲突列表。
func (e *luaScriptExecutor) GetConflicts() ([]string, error) {
	return e.getStringArray("conflicts")
}

// getStringArray 获取字符串数组。
func (e *luaScriptExecutor) getStringArray(funcName string) ([]string, error) {
	L := e.eng.L

	fn := L.GetGlobal(funcName)
	if fn.Type() == lua.LTNil {
		return nil, nil
	}

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 %s 失败: %w", funcName, err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	if ret.Type() != lua.LTTable {
		return nil, nil
	}

	var result []string
	table := ret.(*lua.LTable)
	table.ForEach(func(_, v lua.LValue) {
		if v.Type() == lua.LTString {
			result = append(result, string(v.(lua.LString)))
		}
	})

	return result, nil
}

// luaTableToJSON 将 Lua 表转换为 JSON 字符串。
func luaTableToJSON(L *lua.LState, table *lua.LTable) (string, error) {
	result := make(map[string]interface{})

	table.ForEach(func(key, value lua.LValue) {
		if key.Type() == lua.LTString {
			k := string(key.(lua.LString))
			switch v := value.(type) {
			case lua.LString:
				result[k] = string(v)
			case lua.LNumber:
				result[k] = float64(v)
			case lua.LBool:
				result[k] = bool(v)
			}
		}
	})

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// getStringField 获取字符串字段。
func getStringField(L *lua.LState, table *lua.LTable, field string) string {
	v := L.GetField(table, field)
	if v.Type() == lua.LTString {
		return string(v.(lua.LString))
	}
	return ""
}

// getBoolField 获取布尔字段。
func getBoolField(L *lua.LState, table *lua.LTable, field string) bool {
	v := L.GetField(table, field)
	if v.Type() == lua.LTBool {
		return bool(v.(lua.LBool))
	}
	return false
}

// LoadAppFromPath 从路径加载应用。
func LoadAppFromPath(scriptPath, metaPath string, eng *LuaEngine) (*manifest.App, error) {
	// 加载脚本
	executor := NewLuaScriptExecutor(eng, scriptPath)
	if err := executor.LoadScript(scriptPath); err != nil {
		return nil, fmt.Errorf("加载脚本失败: %w", err)
	}

	// 读取元数据
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("读取元数据失败: %w", err)
	}

	var meta manifest.AppMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("解析元数据失败: %w", err)
	}

	// 获取脚本信息（从 app 表）
	script, err := extractScriptInfo(eng)
	if err != nil {
		return nil, fmt.Errorf("提取脚本信息失败: %w", err)
	}

	return &manifest.App{
		Script: script,
		Meta:   &meta,
		Ref: &manifest.AppRef{
			Name:       script.Name,
			ScriptPath: scriptPath,
			MetaPath:   metaPath,
		},
	}, nil
}

// extractScriptInfo 从 Lua 引擎中提取脚本信息。
func extractScriptInfo(eng *LuaEngine) (*manifest.AppScript, error) {
	L := eng.L

	appTable := L.GetGlobal("app")
	if appTable.Type() != lua.LTTable {
		return nil, fmt.Errorf("app 表未定义")
	}

	table := appTable.(*lua.LTable)

	return &manifest.AppScript{
		Name:        getStringField(L, table, "name"),
		Description: getStringField(L, table, "description"),
		Homepage:    getStringField(L, table, "homepage"),
		License:     getStringField(L, table, "license"),
		Category:    getStringField(L, table, "category"),
		Bucket:      getStringField(L, table, "bucket"),
	}, nil
}

// FindAppFiles 查找应用文件。
func FindAppFiles(dir, name string) (scriptPath, metaPath string, err error) {
	// 查找脚本文件
	luaPath := filepath.Join(dir, name, "app.lua")
	if _, err := os.Stat(luaPath); err == nil {
		scriptPath = luaPath
	}

	// 查找元数据文件
	metaPath = filepath.Join(dir, name, "app.meta.json")
	if _, err := os.Stat(metaPath); err != nil {
		return "", "", fmt.Errorf("元数据文件不存在: %s", metaPath)
	}

	if scriptPath == "" {
		return "", "", fmt.Errorf("脚本文件不存在: %s", luaPath)
	}

	return scriptPath, metaPath, nil
}
