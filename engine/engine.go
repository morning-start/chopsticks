// Package engine 提供 JavaScript 脚本引擎的抽象。
package engine

import (
	"github.com/dop251/goja"
)

// JSState 是 JavaScript 运行时指针的别名。
type JSState = *goja.Runtime

// Engine 定义脚本引擎实现的接口。
type Engine interface {
	// LoadFile 从给定路径加载脚本文件。
	LoadFile(path string) error
	// CallFunction 调用指定名称的函数并传入参数。
	CallFunction(name string, args ...interface{}) error
	// Close 释放脚本引擎持有的资源。
	Close()
}

// JSRegistrar 定义注册 JavaScript 函数和模块的接口。
type JSRegistrar interface {
	// RegisterJS 向给定的 JavaScript 运行时注册函数和模块。
	RegisterJS(vm JSState)
}

// RegisterJSAll 将所有给定的 JavaScript 注册器注册到 JavaScript 运行时。
func RegisterJSAll(vm JSState, registrars ...JSRegistrar) {
	for _, r := range registrars {
		r.RegisterJS(vm)
	}
}
