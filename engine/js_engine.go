package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"chopsticks/engine/archive"
	"chopsticks/engine/checksum"
	"chopsticks/engine/chopsticksx"
	"chopsticks/engine/execx"
	"chopsticks/engine/fetch"
	"chopsticks/engine/fsutil"
	"chopsticks/engine/installerx"
	"chopsticks/engine/jsonx"
	"chopsticks/engine/logx"
	"chopsticks/engine/pathx"
	"chopsticks/engine/registry"
	"chopsticks/engine/semver"
	"chopsticks/engine/symlink"

	"github.com/dop251/goja"
)

// 文件权限常量
const (
	FilePerm0644 = 0644
	FilePerm0755 = 0755
)

// 预定义错误变量
var (
	ErrRequireModuleName   = errors.New("require() requires module name argument")
	ErrModuleLoadFailed    = errors.New("failed to load module")
	ErrModuleExecuteFailed = errors.New("failed to execute module")
)

var _ Engine = (*JSEngine)(nil)

type JSEngine struct {
	vm         *goja.Runtime
	installCtx map[string]interface{}
}

func (e *JSEngine) GetVM() *goja.Runtime {
	return e.vm
}

func NewJSEngine() *JSEngine {
	vm := goja.New()

	consoleObj := vm.NewObject()
	vm.Set("console", consoleObj)
	consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		for _, arg := range args {
			println(arg.String())
		}
		return goja.Undefined()
	})

	vm.Set("require", requireFunc(vm))
	vm.Set("module", vm.NewObject())
	vm.Set("exports", vm.NewObject())

	vm.RunString(appBaseClass)
	vm.RunString(installContextClass)

	RegisterJSAll(vm,
		&fsutil.Module{},
		&fetch.Module{},
		&symlink.Module{},
		&registry.Module{},
		&installerx.Module{},
		&execx.Module{},
		&archive.Module{},
		&checksum.Module{},
		&pathx.Module{},
		&logx.Module{},
		&jsonx.Module{},
		&semver.Module{},
	)

	// 单独注册 chopsticksx 模块（需要构造函数参数）
	chopsticksMod := chopsticksx.NewModule(
		filepath.Join(os.Getenv("LOCALAPPDATA"), "chopsticks", "apps"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "chopsticks", "shims"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "chopsticks", "persist"),
	)
	chopsticksMod.RegisterJS(vm)

	return &JSEngine{
		vm:         vm,
		installCtx: make(map[string]interface{}),
	}
}

var requireFunc = func(vm *goja.Runtime) func(call goja.FunctionCall) (goja.Value, error) {
	return func(call goja.FunctionCall) (goja.Value, error) {
		if len(call.Arguments) == 0 {
			return goja.Undefined(), ErrRequireModuleName
		}

		moduleName := call.Arguments[0].String()

		// 检查内置模块
		if builtin := getBuiltinModule(vm, moduleName); builtin != nil {
			return builtin, nil
		}

		// 尝试加载本地文件模块
		scriptPath := resolveModulePath(moduleName)

		// 读取并执行模块文件
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			return goja.Undefined(), fmt.Errorf("%w '%s': %w", ErrModuleLoadFailed, moduleName, err)
		}

		return executeModule(vm, moduleName, content)
	}
}

// getBuiltinModule 获取内置模块
func getBuiltinModule(vm *goja.Runtime, moduleName string) goja.Value {
	switch moduleName {
	case "fs":
		return vm.ToValue(map[string]interface{}{
			"readFile": func(path string) (string, error) {
				data, err := os.ReadFile(path)
				return string(data), err
			},
			"writeFile": func(path string, content string) error {
				return os.WriteFile(path, []byte(content), FilePerm0644)
			},
			"exists": func(path string) bool {
				_, err := os.Stat(path)
				return err == nil
			},
			"mkdir": func(path string) error {
				return os.MkdirAll(path, FilePerm0755)
			},
		})
	case "path":
		return vm.ToValue(map[string]interface{}{
			"join":  filepath.Join,
			"dir":   filepath.Dir,
			"base":  filepath.Base,
			"ext":   filepath.Ext,
			"clean": filepath.Clean,
		})
	case "os":
		return vm.ToValue(map[string]interface{}{
			"getenv": os.Getenv,
			"setenv": os.Setenv,
			"homedir": func() string {
				home, _ := os.UserHomeDir()
				return home
			},
		})
	}
	return nil
}

// resolveModulePath 解析模块路径
func resolveModulePath(moduleName string) string {
	scriptPath := moduleName
	if !filepath.IsAbs(scriptPath) {
		// 相对路径，尝试在当前目录查找
		scriptPath = filepath.Join(".", moduleName)
	}

	// 添加 .js 扩展名（如果没有）
	if filepath.Ext(scriptPath) == "" {
		scriptPath += ".js"
	}

	return scriptPath
}

// executeModule 执行模块代码
func executeModule(vm *goja.Runtime, moduleName string, content []byte) (goja.Value, error) {
	// 创建模块导出对象
	moduleObj := vm.NewObject()
	exportsObj := vm.NewObject()
	moduleObj.Set("exports", exportsObj)

	// 在模块上下文中执行代码
	moduleCode := string(content)
	_, err := vm.RunScript(moduleName, moduleCode)
	if err != nil {
		return goja.Undefined(), fmt.Errorf("%w '%s': %w", ErrModuleExecuteFailed, moduleName, err)
	}

	// 返回模块的 exports
	return exportsObj, nil
}

const appBaseClass = `
class App {
    constructor(config) {
        this.name = config.name || '';
        this.description = config.description || '';
        this.homepage = config.homepage || '';
        this.license = config.license || 'MIT';
        this.version = config.version || '0.0.0';
        this.bucket = config.bucket || 'main';
        this.category = config.category || '';
        this.tags = config.tags || [];
        this.maintainer = config.maintainer || '';
    }

    async checkVersion() {
        throw new Error('checkVersion() must be implemented');
    }

    async getDownloadInfo(version, arch) {
        throw new Error('getDownloadInfo() must be implemented');
    }

    async onPreDownload(ctx) {}
    async onPostDownload(ctx) {}
    async onPreExtract(ctx) {}
    async onPostExtract(ctx) {}
    async onPreInstall(ctx) {}
    async onInstall(ctx) {}
    async onPostInstall(ctx) {}
    async onPreUninstall(ctx) {}
    async onUninstall(ctx) {}
    async onPostUninstall(ctx) {}

    getDepends() { return []; }
    getConflicts() { return []; }
    getEnvPath() { return []; }
    getBin() { return []; }
    getPersist() { return []; }
}
`

const installContextClass = `
class InstallContext {
    constructor(data) {
        this.version = data.version || 'latest';
        this.arch = data.arch || 'amd64';
        this.cookDir = data.cookDir || '';
        this.name = data.name || '';
        this.bucket = data.bucket || 'main';
        this.downloadPath = data.downloadPath || '';
        this.installDir = data.installDir || '';
    }
}
`

func (e *JSEngine) SetContext(ctx map[string]interface{}) {
	for k, v := range ctx {
		e.installCtx[k] = v
	}
}

func (e *JSEngine) LoadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = e.vm.RunString(string(content))
	return err
}

func (e *JSEngine) CallFunction(name string, args ...interface{}) error {
	fn, ok := goja.AssertFunction(e.vm.Get(name))
	if !ok {
		return fmt.Errorf("function %s not found", name)
	}

	jsArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		jsArgs[i] = e.vm.ToValue(arg)
	}

	_, err := fn(goja.Undefined(), jsArgs...)
	return err
}

func (e *JSEngine) GetAppInstance() (map[string]interface{}, error) {
	exports := e.vm.Get("exports")
	if exports == goja.Undefined() {
		return nil, nil
	}

	obj := exports.ToObject(e.vm)
	result := make(map[string]interface{})

	keys := obj.Keys()
	for _, key := range keys {
		value := obj.Get(key)
		result[key] = value.Export()
	}

	return result, nil
}

func (e *JSEngine) CallAppMethod(methodName string, ctx map[string]interface{}) error {
	obj := e.vm.Get("exports")
	if obj == goja.Undefined() {
		return nil
	}

	appObj := obj.ToObject(e.vm)
	appVal := appObj.Get("app")
	if appVal == goja.Undefined() {
		return nil
	}

	fn, ok := goja.AssertFunction(appVal)
	if !ok {
		return nil
	}

	appInstance := appVal.ToObject(e.vm)
	ctxObj := e.vm.NewObject()
	for k, v := range ctx {
		ctxObj.Set(k, v)
	}

	_, err := fn(appInstance, ctxObj)
	return err
}

func (e *JSEngine) Close() {
}
