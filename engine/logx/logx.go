// Package logx 提供日志功能。
package logx

import (
	"log"
	"os"
)

// Level 定义日志级别。
type Level int

const (
	// DebugLevel 调试级别。
	DebugLevel Level = iota
	// InfoLevel 信息级别。
	InfoLevel
	// WarnLevel 警告级别。
	WarnLevel
	// ErrorLevel 错误级别。
	ErrorLevel
)

// Logger 定义日志记录器接口。
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// logger 是 Logger 的实现。
type logger struct {
	level  Level
	output *log.Logger
}

// 编译时接口检查。
var _ Logger = (*logger)(nil)

// New 创建新的 Logger。
func New() Logger {
	return &logger{
		level:  InfoLevel,
		output: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// NewWithLevel 创建指定级别的 Logger。
func NewWithLevel(level Level) Logger {
	return &logger{
		level:  level,
		output: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SetLevel 设置日志级别。
func (l *logger) SetLevel(level Level) {
	l.level = level
}

// Debug 输出调试日志。
func (l *logger) Debug(msg string) {
	if l.level <= DebugLevel {
		l.output.Printf("[DEBUG] %s", msg)
	}
}

// Info 输出信息日志。
func (l *logger) Info(msg string) {
	if l.level <= InfoLevel {
		l.output.Printf("[INFO] %s", msg)
	}
}

// Warn 输出警告日志。
func (l *logger) Warn(msg string) {
	if l.level <= WarnLevel {
		l.output.Printf("[WARN] %s", msg)
	}
}

// Error 输出错误日志。
func (l *logger) Error(msg string) {
	if l.level <= ErrorLevel {
		l.output.Printf("[ERROR] %s", msg)
	}
}

// Debugf 输出格式化调试日志。
func (l *logger) Debugf(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.output.Printf("[DEBUG] "+format, args...)
	}
}

// Infof 输出格式化信息日志。
func (l *logger) Infof(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.output.Printf("[INFO] "+format, args...)
	}
}

// Warnf 输出格式化警告日志。
func (l *logger) Warnf(format string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.output.Printf("[WARN] "+format, args...)
	}
}

// Errorf 输出格式化错误日志。
func (l *logger) Errorf(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.output.Printf("[ERROR] "+format, args...)
	}
}

// Module 为脚本引擎提供 log 注册。
type Module struct {
	logger Logger
}

// NewModule 创建新的 log 模块。
func NewModule() *Module {
	return &Module{
		logger: New(),
	}
}

// NewModuleWithLogger 使用指定 Logger 创建模块。
func NewModuleWithLogger(l Logger) *Module {
	return &Module{
		logger: l,
	}
}
