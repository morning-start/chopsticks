package engine

import (
	"os"

	"chopsticks/engine/fetch"
	"chopsticks/engine/fsutil"
	"chopsticks/engine/symlink"

	"github.com/dop251/goja"
)

var _ Engine = (*JSEngine)(nil)

type JSEngine struct {
	vm           *goja.Runtime
	installCtx   map[string]interface{}
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

	vm.RunString(dishBaseClass)
	vm.RunString(dishContextClass)

	RegisterJSAll(vm,
		&fsutil.Module{},
		&fetch.Module{},
		&symlink.Module{},
	)

	return &JSEngine{
		vm:         vm,
		installCtx: make(map[string]interface{}),
	}
}

var requireFunc = func(_ *goja.Runtime) func(call goja.FunctionCall) (goja.Value, error) {
	return func(call goja.FunctionCall) (goja.Value, error) {
		return goja.Undefined(), nil
	}
}

const dishBaseClass = `
class Dish {
    constructor(config) {
        this.name = config.name || '';
        this.description = config.description || '';
        this.homepage = config.homepage || '';
        this.license = config.license || 'MIT';
        this.version = config.version || '0.0.0';
        this.bow = config.bow || 'main';
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

const dishContextClass = `
class DishContext {
    constructor(data) {
        this.version = data.version || 'latest';
        this.arch = data.arch || 'amd64';
        this.cookDir = data.cookDir || '';
        this.dishName = data.dishName || '';
        this.bow = data.bow || 'main';
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
		return nil
	}

	jsArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		jsArgs[i] = e.vm.ToValue(arg)
	}

	_, err := fn(goja.Undefined(), jsArgs...)
	return err
}

func (e *JSEngine) GetDishInstance() (map[string]interface{}, error) {
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

func (e *JSEngine) CallDishMethod(methodName string, ctx map[string]interface{}) error {
	obj := e.vm.Get("exports")
	if obj == goja.Undefined() {
		return nil
	}

	dishObj := obj.ToObject(e.vm)
	dishVal := dishObj.Get("dish")
	if dishVal == goja.Undefined() {
		return nil
	}

	fn, ok := goja.AssertFunction(dishVal)
	if !ok {
		return nil
	}

	dishInstance := dishVal.ToObject(e.vm)
	ctxObj := e.vm.NewObject()
	for k, v := range ctx {
		ctxObj.Set(k, v)
	}

	_, err := fn(dishInstance, ctxObj)
	return err
}

func (e *JSEngine) Close() {
}
