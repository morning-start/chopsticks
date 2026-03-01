// Package logx 提供日志功能，基于 slog + lumberjack 实现持久化。
package logx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
)

// Config 日志配置
type Config struct {
	Filename   string     // 日志文件路径
	MaxSize    int        // 单个日志文件最大大小（MB）
	MaxBackups int        // 保留的旧日志文件数量
	MaxAge     int        // 日志文件保留天数
	Compress   bool       // 是否压缩旧日志
	Level      slog.Level // 日志级别
	JSONFormat bool       // 是否使用 JSON 格式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Filename:   filepath.Join(os.TempDir(), "chopsticks", "logs", "chopsticks.log"),
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
		Level:      slog.LevelInfo,
		JSONFormat: true,
	}
}

// Logger 封装 slog.Logger
type Logger struct {
	logger *slog.Logger
	level  slog.Level
	writer io.WriteCloser
	mu     sync.RWMutex
}

// New 创建新的 Logger
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录: %w", err)
	}

	// 配置 lumberjack 日志轮转
	logWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 创建 slog handler
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	if cfg.JSONFormat {
		handler = slog.NewJSONHandler(logWriter, opts)
	} else {
		handler = slog.NewTextHandler(logWriter, opts)
	}

	return &Logger{
		logger: slog.New(handler),
		level:  cfg.Level,
		writer: logWriter,
	}, nil
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() slog.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// Sync 刷新日志缓冲区
func (l *Logger) Sync() error {
	if l.writer != nil {
		// lumberjack.Logger 没有 Sync 方法，直接返回 nil
		return nil
	}
	return nil
}

// Close 关闭日志
func (l *Logger) Close() error {
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}

// 标准日志方法

// Debug 输出调试日志
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

// Info 输出信息日志
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// Warn 输出警告日志
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

// Error 输出错误日志
func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

// 带上下文的方法

// DebugContext 输出带上下文的调试日志
func (l *Logger) DebugContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// InfoContext 输出带上下文的信息日志
func (l *Logger) InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// WarnContext 输出带上下文的警告日志
func (l *Logger) WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// ErrorContext 输出带上下文的错误日志
func (l *Logger) ErrorContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// 操作日志专用方法

// LogInstall 记录安装操作日志
func (l *Logger) LogInstall(app, version string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("operation", "install"),
		slog.String("app", app),
		slog.String("version", version),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Bool("success", err == nil),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("应用安装失败", attrs...)
	} else {
		l.Info("应用安装完成", attrs...)
	}
}

// LogUninstall 记录卸载操作日志
func (l *Logger) LogUninstall(app string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("operation", "uninstall"),
		slog.String("app", app),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Bool("success", err == nil),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("应用卸载失败", attrs...)
	} else {
		l.Info("应用卸载完成", attrs...)
	}
}

// LogUpdate 记录更新操作日志
func (l *Logger) LogUpdate(app, fromVersion, toVersion string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("operation", "update"),
		slog.String("app", app),
		slog.String("from_version", fromVersion),
		slog.String("to_version", toVersion),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Bool("success", err == nil),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("应用更新失败", attrs...)
	} else {
		l.Info("应用更新完成", attrs...)
	}
}

// LogDownload 记录下载操作日志
func (l *Logger) LogDownload(url string, size int64, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("operation", "download"),
		slog.String("url", url),
		slog.Int64("size", size),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Bool("success", err == nil),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("文件下载失败", attrs...)
	} else {
		l.Info("文件下载完成", attrs...)
	}
}

// LogBucket 记录软件源操作日志
func (l *Logger) LogBucket(operation, name, url string, err error) {
	attrs := []slog.Attr{
		slog.String("operation", "bucket_"+operation),
		slog.String("bucket_name", name),
		slog.String("bucket_url", url),
		slog.Bool("success", err == nil),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("软件源操作失败", attrs...)
	} else {
		l.Info("软件源操作完成", attrs...)
	}
}

// 全局默认 Logger

var (
	defaultLogger *Logger
	defaultMu     sync.RWMutex
)

// InitDefault 初始化全局默认 Logger
func InitDefault(cfg *Config) error {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	logger, err := New(cfg)
	if err != nil {
		return err
	}

	if defaultLogger != nil {
		defaultLogger.Close()
	}

	defaultLogger = logger
	return nil
}

// GetDefault 获取全局默认 Logger
func GetDefault() *Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultLogger
}

// 全局便捷函数

// Debug 使用全局 Logger 输出调试日志
func Debug(msg string, attrs ...slog.Attr) {
	if l := GetDefault(); l != nil {
		l.Debug(msg, attrs...)
	}
}

// Info 使用全局 Logger 输出信息日志
func Info(msg string, attrs ...slog.Attr) {
	if l := GetDefault(); l != nil {
		l.Info(msg, attrs...)
	}
}

// Warn 使用全局 Logger 输出警告日志
func Warn(msg string, attrs ...slog.Attr) {
	if l := GetDefault(); l != nil {
		l.Warn(msg, attrs...)
	}
}

// Error 使用全局 Logger 输出错误日志
func Error(msg string, attrs ...slog.Attr) {
	if l := GetDefault(); l != nil {
		l.Error(msg, attrs...)
	}
}

// LogInstall 使用全局 Logger 记录安装操作日志
func LogInstall(app, version string, duration time.Duration, err error) {
	if l := GetDefault(); l != nil {
		l.LogInstall(app, version, duration, err)
	}
}

// LogUninstall 使用全局 Logger 记录卸载操作日志
func LogUninstall(app string, duration time.Duration, err error) {
	if l := GetDefault(); l != nil {
		l.LogUninstall(app, duration, err)
	}
}

// LogUpdate 使用全局 Logger 记录更新操作日志
func LogUpdate(app, fromVersion, toVersion string, duration time.Duration, err error) {
	if l := GetDefault(); l != nil {
		l.LogUpdate(app, fromVersion, toVersion, duration, err)
	}
}

// LogDownload 使用全局 Logger 记录下载操作日志
func LogDownload(url string, size int64, duration time.Duration, err error) {
	if l := GetDefault(); l != nil {
		l.LogDownload(url, size, duration, err)
	}
}

// LogBucket 使用全局 Logger 记录软件源操作日志
func LogBucket(operation, name, url string, err error) {
	if l := GetDefault(); l != nil {
		l.LogBucket(operation, name, url, err)
	}
}

// 兼容旧版 API

// Level 定义日志级别（兼容旧版）
type Level int

const (
	// DebugLevel 调试级别
	DebugLevel Level = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
)

// LoggerInterface 定义日志记录器接口（兼容旧版）
type LoggerInterface interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// legacyLogger 适配旧版 Logger 接口
type legacyLogger struct {
	logger *Logger
}

// 编译时接口检查
var _ LoggerInterface = (*legacyLogger)(nil)

// NewLegacy 创建兼容旧版的 Logger
func NewLegacy() LoggerInterface {
	return &legacyLogger{logger: GetDefault()}
}

// NewLegacyWithLevel 创建指定级别的兼容 Logger
func NewLegacyWithLevel(level Level) LoggerInterface {
	return &legacyLogger{logger: GetDefault()}
}

// SetLevel 设置日志级别（兼容旧版）
func (l *legacyLogger) SetLevel(level Level) {
	// 忽略，使用全局 Logger 的级别
}

// Debug 输出调试日志（兼容旧版）
func (l *legacyLogger) Debug(msg string) {
	if l.logger != nil {
		l.logger.Debug(msg)
	}
}

// Info 输出信息日志（兼容旧版）
func (l *legacyLogger) Info(msg string) {
	if l.logger != nil {
		l.logger.Info(msg)
	}
}

// Warn 输出警告日志（兼容旧版）
func (l *legacyLogger) Warn(msg string) {
	if l.logger != nil {
		l.logger.Warn(msg)
	}
}

// Error 输出错误日志（兼容旧版）
func (l *legacyLogger) Error(msg string) {
	if l.logger != nil {
		l.logger.Error(msg)
	}
}

// Debugf 输出格式化调试日志（兼容旧版）
func (l *legacyLogger) Debugf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Debug(fmt.Sprintf(format, args...))
	}
}

// Infof 输出格式化信息日志（兼容旧版）
func (l *legacyLogger) Infof(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Info(fmt.Sprintf(format, args...))
	}
}

// Warnf 输出格式化警告日志（兼容旧版）
func (l *legacyLogger) Warnf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Warn(fmt.Sprintf(format, args...))
	}
}

// Errorf 输出格式化错误日志（兼容旧版）
func (l *legacyLogger) Errorf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Error(fmt.Sprintf(format, args...))
	}
}

// Module 为脚本引擎提供 log 注册（兼容旧版）
type Module struct {
	logger LoggerInterface
}

// NewModule 创建新的 log 模块（兼容旧版）
func NewModule() *Module {
	return &Module{
		logger: NewLegacy(),
	}
}

// NewModuleWithLogger 使用指定 Logger 创建模块（兼容旧版）
func NewModuleWithLogger(l LoggerInterface) *Module {
	return &Module{
		logger: l,
	}
}
