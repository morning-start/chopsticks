package engine

import (
	"testing"

	"github.com/dop251/goja"
	lua "github.com/yuin/gopher-lua"
)

func TestNewJSEngine(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	if engine.vm == nil {
		t.Error("JSEngine.vm should not be nil")
	}
}

func TestJSEngineGetVM(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	vm := engine.GetVM()
	if vm == nil {
		t.Error("GetVM() should not return nil")
	}
}

func TestJSEngineSetContext(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	ctx := map[string]interface{}{
		"name":    "test-app",
		"version": "1.0.0",
	}

	engine.SetContext(ctx)

	if engine.installCtx == nil {
		t.Error("installCtx should not be nil after SetContext")
	}

	if engine.installCtx["name"] != "test-app" {
		t.Error("name not set correctly in context")
	}
}

func TestJSEngineCallFunctionNotExists(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	// Calling a non-existent function should not error
	err := engine.CallFunction("nonExistentFunction")
	if err != nil {
		t.Errorf("CallFunction() for non-existent function failed: %v", err)
	}
}

func TestNewLuaEngine(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	if engine.L == nil {
		t.Error("LuaEngine.L should not be nil")
	}
}

func TestLuaEngineCallFunctionNotExists(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	// Calling a non-existent function should not error
	err := engine.CallFunction("nonExistentFunction")
	if err != nil {
		t.Errorf("CallFunction() for non-existent function failed: %v", err)
	}
}

func TestRegisterLuaAll(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test with no registrars
	RegisterLuaAll(L)

	// Test with multiple registrars
	RegisterLuaAll(L,
		&testLuaRegistrar{name: "test1"},
		&testLuaRegistrar{name: "test2"},
	)
}

func TestRegisterJSAll(t *testing.T) {
	vm := goja.New()

	// Test with no registrars
	RegisterJSAll(vm)

	// Test with multiple registrars
	RegisterJSAll(vm,
		&testJSRegistrar{name: "test1"},
		&testJSRegistrar{name: "test2"},
	)
}

// Test registrars
type testLuaRegistrar struct {
	name string
}

func (r *testLuaRegistrar) RegisterLua(L LuaState) {
	L.SetGlobal(r.name, lua.LString("test"))
}

type testJSRegistrar struct {
	name string
}

func (r *testJSRegistrar) RegisterJS(vm JSState) {
	vm.Set(r.name, "test")
}

func TestInstallContext(t *testing.T) {
	ctx := &InstallContext{
		Version:      "1.0.0",
		Arch:         "amd64",
		InstallDir:   "/apps/test",
		CookDir:      "/cook/test",
		Name:         "test-app",
		Bucket:       "main",
		DownloadPath: "/downloads/test.zip",
	}

	if ctx.Version != "1.0.0" {
		t.Error("Version mismatch")
	}

	if ctx.Arch != "amd64" {
		t.Error("Arch mismatch")
	}

	if ctx.InstallDir != "/apps/test" {
		t.Error("InstallDir mismatch")
	}

	if ctx.CookDir != "/cook/test" {
		t.Error("CookDir mismatch")
	}

	if ctx.Name != "test-app" {
		t.Error("Name mismatch")
	}

	if ctx.Bucket != "main" {
		t.Error("Bucket mismatch")
	}

	if ctx.DownloadPath != "/downloads/test.zip" {
		t.Error("DownloadPath mismatch")
	}
}

func TestDependency(t *testing.T) {
	dep := Dependency{
		Name:     "test-dep",
		Bucket:   "main",
		Version:  "1.0.0",
		Optional: true,
	}

	if dep.Name != "test-dep" {
		t.Error("Name mismatch")
	}

	if dep.Bucket != "main" {
		t.Error("Bucket mismatch")
	}

	if dep.Version != "1.0.0" {
		t.Error("Version mismatch")
	}

	if !dep.Optional {
		t.Error("Optional should be true")
	}
}

func TestJSEngineLoadInvalidFile(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	err := engine.LoadFile("/nonexistent/file.js")
	if err == nil {
		t.Error("LoadFile() should return error for non-existent file")
	}
}

func TestLuaEngineLoadInvalidFile(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	err := engine.LoadFile("/nonexistent/file.lua")
	if err == nil {
		t.Error("LoadFile() should return error for non-existent file")
	}
}

func TestJSEngineGetAppInstanceEmpty(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	// Before loading any script, exports should be empty/undefined
	instance, err := engine.GetAppInstance()
	if err != nil {
		t.Errorf("GetAppInstance() failed: %v", err)
	}

	// Should return nil or empty map
	if instance != nil && len(instance) > 0 {
		t.Error("GetAppInstance() should return empty result for fresh engine")
	}
}

func TestJSEngineCallAppMethodEmpty(t *testing.T) {
	engine := NewJSEngine()
	if engine == nil {
		t.Fatal("NewJSEngine() returned nil")
	}
	defer engine.Close()

	// Before loading any script, calling app method should not error
	err := engine.CallAppMethod("test", map[string]interface{}{})
	if err != nil {
		t.Errorf("CallAppMethod() failed: %v", err)
	}
}

func TestLuaEngineCallFunctionWithArgs(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	// Define a simple function
	engine.L.DoString(`
		function testFunc(a, b)
			return a + b
		end
	`)

	// Call with arguments
	err := engine.CallFunction("testFunc", 1, 2)
	if err != nil {
		t.Errorf("CallFunction() failed: %v", err)
	}
}

func TestLuaEngineCallFunctionWithStringArgs(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	// Define a simple function
	engine.L.DoString(`
		function greet(name)
			return "Hello, " .. name
		end
	`)

	// Call with string argument
	err := engine.CallFunction("greet", "World")
	if err != nil {
		t.Errorf("CallFunction() failed: %v", err)
	}
}

func TestLuaEngineCallFunctionWithFloatArgs(t *testing.T) {
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("NewLuaEngine() returned nil")
	}
	defer engine.Close()

	// Define a simple function
	engine.L.DoString(`
		function calc(x)
			return x * 2
		end
	`)

	// Call with float argument
	err := engine.CallFunction("calc", 3.14)
	if err != nil {
		t.Errorf("CallFunction() failed: %v", err)
	}
}
