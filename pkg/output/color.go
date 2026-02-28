// Package output 提供输出格式化功能，包括彩色输出。
package output

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
)

var (
	// 全局颜色开关
	colorEnabled = true
	colorMu      sync.RWMutex
)

// 预定义的颜色对象
var (
	// ColorSuccess 成功颜色（绿色加粗）
	ColorSuccess = color.New(color.FgGreen, color.Bold)
	// ColorError 错误颜色（红色加粗）
	ColorError = color.New(color.FgRed, color.Bold)
	// ColorWarning 警告颜色（黄色）
	ColorWarning = color.New(color.FgYellow)
	// ColorInfo 信息颜色（蓝色）
	ColorInfo = color.New(color.FgBlue)
	// ColorHighlight 高亮颜色（青色加粗）
	ColorHighlight = color.New(color.FgCyan, color.Bold)
	// ColorDim 暗淡颜色（灰色）
	ColorDim = color.New(color.FgHiBlack)
	// ColorBold 粗体
	ColorBold = color.New(color.Bold)
)

func init() {
	// 初始化时根据环境自动检测颜色支持
	updateColorState()
}

// updateColorState 根据全局开关更新颜色状态
func updateColorState() {
	colorMu.RLock()
	enabled := colorEnabled
	colorMu.RUnlock()

	color.NoColor = !enabled
}

// SetColorEnabled 启用/禁用颜色输出
func SetColorEnabled(enabled bool) {
	colorMu.Lock()
	colorEnabled = enabled
	colorMu.Unlock()

	color.NoColor = !enabled
}

// IsColorEnabled 检查颜色是否启用
func IsColorEnabled() bool {
	colorMu.RLock()
	defer colorMu.RUnlock()
	return colorEnabled && !color.NoColor
}

// DisableColor 禁用颜色输出
func DisableColor() {
	SetColorEnabled(false)
}

// EnableColor 启用颜色输出
func EnableColor() {
	SetColorEnabled(true)
}

// Success 输出成功消息（绿色）
func Success(format string, a ...interface{}) {
	ColorSuccess.Printf(format, a...)
}

// Successln 输出成功消息并换行
func Successln(a ...interface{}) {
	ColorSuccess.Println(a...)
}

// Successf 输出格式化的成功消息
func Successf(format string, a ...interface{}) {
	ColorSuccess.Printf(format, a...)
}

// Error 输出错误消息（红色）
func Error(format string, a ...interface{}) {
	ColorError.Fprintf(os.Stderr, format, a...)
}

// Errorln 输出错误消息并换行
func Errorln(a ...interface{}) {
	ColorError.Fprintln(os.Stderr, a...)
}

// Errorf 输出格式化的错误消息
func Errorf(format string, a ...interface{}) {
	ColorError.Fprintf(os.Stderr, format, a...)
}

// Warning 输出警告消息（黄色）
func Warning(format string, a ...interface{}) {
	ColorWarning.Printf(format, a...)
}

// Warn 输出警告消息（黄色，Warning 的别名）
func Warn(format string, a ...interface{}) {
	ColorWarning.Printf(format, a...)
}

// Warningln 输出警告消息并换行
func Warningln(a ...interface{}) {
	ColorWarning.Println(a...)
}

// Warningf 输出格式化的警告消息
func Warningf(format string, a ...interface{}) {
	ColorWarning.Printf(format, a...)
}

// Info 输出信息消息（蓝色）
func Info(format string, a ...interface{}) {
	ColorInfo.Printf(format, a...)
}

// Infoln 输出信息消息并换行
func Infoln(a ...interface{}) {
	ColorInfo.Println(a...)
}

// Infof 输出格式化的信息消息
func Infof(format string, a ...interface{}) {
	ColorInfo.Printf(format, a...)
}

// Highlight 输出高亮消息（青色）
func Highlight(format string, a ...interface{}) {
	ColorHighlight.Printf(format, a...)
}

// Highlightln 输出高亮消息并换行
func Highlightln(a ...interface{}) {
	ColorHighlight.Println(a...)
}

// Highlightf 输出格式化的高亮消息
func Highlightf(format string, a ...interface{}) {
	ColorHighlight.Printf(format, a...)
}

// Dim 输出暗淡消息（灰色）
func Dim(format string, a ...interface{}) {
	ColorDim.Printf(format, a...)
}

// Dimln 输出暗淡消息并换行
func Dimln(a ...interface{}) {
	ColorDim.Println(a...)
}

// Dimf 输出格式化的暗淡消息
func Dimf(format string, a ...interface{}) {
	ColorDim.Printf(format, a...)
}

// Print 使用指定颜色输出
func Print(c *color.Color, format string, a ...interface{}) {
	c.Printf(format, a...)
}

// Println 使用指定颜色输出并换行
func Println(c *color.Color, a ...interface{}) {
	c.Println(a...)
}

// Printf 使用指定颜色格式化输出
func Printf(c *color.Color, format string, a ...interface{}) {
	c.Printf(format, a...)
}

// 带图标的消息输出

// SuccessCheck 输出带勾选的成功消息
func SuccessCheck(msg string) {
	ColorSuccess.Println("✓", msg)
}

// SuccessCheckf 输出带勾选的格式化成功消息
func SuccessCheckf(format string, a ...interface{}) {
	ColorSuccess.Printf("✓ "+format+"\n", a...)
}

// ErrorCross 输出带叉号的错误消息
func ErrorCross(msg string) {
	ColorError.Fprintln(os.Stderr, "✗", msg)
}

// ErrorCrossf 输出带叉号的格式化错误消息
func ErrorCrossf(format string, a ...interface{}) {
	ColorError.Fprintf(os.Stderr, "✗ "+format+"\n", a...)
}

// WarningSign 输出带警告符号的警告消息
func WarningSign(msg string) {
	ColorWarning.Println("⚠", msg)
}

// InfoSign 输出带信息符号的信息消息
func InfoSign(msg string) {
	ColorInfo.Println("ℹ", msg)
}

// Arrow 输出带箭头的高亮消息
func Arrow(msg string) {
	ColorHighlight.Println("→", msg)
}

// 兼容旧版 API

// Color 颜色类型（兼容旧版）
type Color int

const (
	Reset Color = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White

	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite

	SuccessColor = Green
	WarningColor = Yellow
	ErrorColor   = Red
	InfoColor    = Cyan
)

// Colorize 给消息添加颜色（兼容旧版）
func Colorize(c Color, message string) string {
	switch c {
	case Red:
		return ColorError.Sprint(message)
	case Green:
		return ColorSuccess.Sprint(message)
	case Yellow:
		return ColorWarning.Sprint(message)
	case Blue:
		return ColorInfo.Sprint(message)
	case Cyan:
		return ColorHighlight.Sprint(message)
	case Magenta:
		return color.MagentaString(message)
	case White:
		return color.WhiteString(message)
	case BrightBlack:
		return ColorDim.Sprint(message)
	default:
		return message
	}
}

// ColorizeBg 给消息添加前景色和背景色（兼容旧版，简化实现）
func ColorizeBg(fg Color, bg Color, message string) string {
	// 简化实现，只使用前景色
	return Colorize(fg, message)
}

// PrintError 输出错误（兼容旧版）
func PrintError(format string, args ...interface{}) {
	Errorf(format, args...)
}

// PrintErrorln 输出错误并换行（兼容旧版）
func PrintErrorln(args ...interface{}) {
	Errorln(args...)
}

// PrintSuccess 输出成功（兼容旧版）
func PrintSuccess(format string, args ...interface{}) {
	Successf(format, args...)
}

// PrintSuccessln 输出成功并换行（兼容旧版）
func PrintSuccessln(args ...interface{}) {
	Successln(args...)
}

// PrintWarning 输出警告（兼容旧版）
func PrintWarning(format string, args ...interface{}) {
	Warningf(format, args...)
}

// PrintWarningln 输出警告并换行（兼容旧版）
func PrintWarningln(args ...interface{}) {
	Warningln(args...)
}

// PrintInfo 输出信息（兼容旧版）
func PrintInfo(format string, args ...interface{}) {
	Infof(format, args...)
}

// PrintInfoln 输出信息并换行（兼容旧版）
func PrintInfoln(args ...interface{}) {
	Infoln(args...)
}

// Header 输出标题（兼容旧版）
func Header(format string, args ...interface{}) {
	Highlightf(format, args...)
}

// SubHeader 输出副标题（兼容旧版）
func SubHeader(format string, args ...interface{}) {
	Infof(format, args...)
}

// Item 输出项目（兼容旧版）
func Item(label, value string) {
	fmt.Printf("%s %s\n", ColorDim.Sprint(label+":"), value)
}

// Checkmark 输出勾选（兼容旧版）
func Checkmark() {
	Successln("✓")
}

// CheckmarkWith 输出带勾选的文本（兼容旧版）
func CheckmarkWith(msg string) {
	SuccessCheck(msg)
}

// Crossmark 输出叉号（兼容旧版）
func Crossmark() {
	Errorln("✗")
}

// CrossmarkWith 输出带叉号的文本（兼容旧版）
func CrossmarkWith(msg string) {
	ErrorCross(msg)
}

// ArrowWith 输出带箭头的文本（兼容旧版）
func ArrowWith(msg string) {
	Arrow(msg)
}
