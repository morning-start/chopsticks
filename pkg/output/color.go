// Package output 提供输出格式化功能，包括彩色输出。
package output

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	// 全局颜色开关
	colorEnabled = true
	colorMu      sync.RWMutex
)

// 颜色定义
var (
	// 基础颜色
	colorGreen   = lipgloss.Color("#00FF00")
	colorRed     = lipgloss.Color("#FF0000")
	colorYellow  = lipgloss.Color("#FFFF00")
	colorBlue    = lipgloss.Color("#0000FF")
	colorCyan    = lipgloss.Color("#00FFFF")
	colorHiBlack = lipgloss.Color("#666666")
	colorWhite   = lipgloss.Color("#FFFFFF")

	// 样式定义
	styleSuccess   lipgloss.Style
	styleError     lipgloss.Style
	styleWarning   lipgloss.Style
	styleInfo      lipgloss.Style
	styleHighlight lipgloss.Style
	styleDim       lipgloss.Style
	styleBold      lipgloss.Style
)

func init() {
	updateStyles()
}

func sInterfaceToString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func joinStrings(a []interface{}) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return sInterfaceToString(a[0])
	}
	var b strings.Builder
	b.WriteString(sInterfaceToString(a[0]))
	for i := 1; i < len(a); i++ {
		b.WriteString(" ")
		b.WriteString(sInterfaceToString(a[i]))
	}
	return b.String()
}

// updateStyles 根据颜色开关更新样式
func updateStyles() {
	colorMu.RLock()
	enabled := colorEnabled
	colorMu.RUnlock()

	if enabled {
		styleSuccess = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
		styleError = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
		styleWarning = lipgloss.NewStyle().Foreground(colorYellow)
		styleInfo = lipgloss.NewStyle().Foreground(colorBlue)
		styleHighlight = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
		styleDim = lipgloss.NewStyle().Foreground(colorHiBlack)
		styleBold = lipgloss.NewStyle().Bold(true)
	} else {
		// 禁用颜色时，使用无颜色样式
		styleSuccess = lipgloss.NewStyle()
		styleError = lipgloss.NewStyle()
		styleWarning = lipgloss.NewStyle()
		styleInfo = lipgloss.NewStyle()
		styleHighlight = lipgloss.NewStyle()
		styleDim = lipgloss.NewStyle()
		styleBold = lipgloss.NewStyle()
	}
}

// SetColorEnabled 启用/禁用颜色输出
func SetColorEnabled(enabled bool) {
	colorMu.Lock()
	colorEnabled = enabled
	colorMu.Unlock()

	updateStyles()
}

// IsColorEnabled 检查颜色是否启用
func IsColorEnabled() bool {
	colorMu.RLock()
	defer colorMu.RUnlock()
	return colorEnabled
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
	if len(a) == 0 {
		fmt.Print(styleSuccess.Render(format))
	} else {
		fmt.Print(styleSuccess.Render(fmt.Sprintf(format, a...)))
	}
}

// Successln 输出成功消息并换行
func Successln(a ...interface{}) {
	fmt.Println(styleSuccess.Render(joinStrings(a)))
}

// Successf 输出格式化的成功消息
func Successf(format string, a ...interface{}) {
	fmt.Println(styleSuccess.Render(fmt.Sprintf(format, a...)))
}

// Error 输出错误消息（红色）
func Error(format string, a ...interface{}) {
	fmt.Fprint(os.Stderr, styleError.Render(fmt.Sprintf(format, a...)))
}

// Errorln 输出错误消息并换行
func Errorln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, styleError.Render(joinStrings(a)))
}

// Errorf 输出格式化的错误消息
func Errorf(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, styleError.Render(fmt.Sprintf(format, a...)))
}

// Warning 输出警告消息（黄色）
func Warning(format string, a ...interface{}) {
	fmt.Print(styleWarning.Render(fmt.Sprintf(format, a...)))
}

// Warn 输出警告消息（黄色，Warning 的别名）
func Warn(format string, a ...interface{}) {
	Warning(format, a...)
}

// Warningln 输出警告消息并换行
func Warningln(a ...interface{}) {
	fmt.Println(styleWarning.Render(joinStrings(a)))
}

// Warningf 输出格式化的警告消息
func Warningf(format string, a ...interface{}) {
	fmt.Println(styleWarning.Render(fmt.Sprintf(format, a...)))
}

// Info 输出信息消息（蓝色）
func Info(format string, a ...interface{}) {
	fmt.Print(styleInfo.Render(fmt.Sprintf(format, a...)))
}

// Infoln 输出信息消息并换行
func Infoln(a ...interface{}) {
	fmt.Println(styleInfo.Render(joinStrings(a)))
}

// Infof 输出格式化的信息消息
func Infof(format string, a ...interface{}) {
	fmt.Println(styleInfo.Render(fmt.Sprintf(format, a...)))
}

// Highlight 输出高亮消息（青色）
func Highlight(format string, a ...interface{}) {
	fmt.Print(styleHighlight.Render(fmt.Sprintf(format, a...)))
}

// Highlightln 输出高亮消息并换行
func Highlightln(a ...interface{}) {
	fmt.Println(styleHighlight.Render(joinStrings(a)))
}

// Highlightf 输出格式化的高亮消息
func Highlightf(format string, a ...interface{}) {
	fmt.Println(styleHighlight.Render(fmt.Sprintf(format, a...)))
}

// Dim 输出暗淡消息（灰色）
func Dim(format string, a ...interface{}) {
	fmt.Print(styleDim.Render(fmt.Sprintf(format, a...)))
}

// Dimln 输出暗淡消息并换行
func Dimln(a ...interface{}) {
	fmt.Println(styleDim.Render(joinStrings(a)))
}

// Dimf 输出格式化的暗淡消息
func Dimf(format string, a ...interface{}) {
	fmt.Println(styleDim.Render(fmt.Sprintf(format, a...)))
}

// 带图标的消息输出

// SuccessCheck 输出带勾选的成功消息
func SuccessCheck(msg string) {
	fmt.Println(styleSuccess.Render("✓ " + msg))
}

// SuccessCheckf 输出带勾选的格式化成功消息
func SuccessCheckf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(styleSuccess.Render("✓ " + msg))
}

// ErrorCross 输出带叉号的错误消息
func ErrorCross(msg string) {
	fmt.Fprintln(os.Stderr, styleError.Render("✗ "+msg))
}

// ErrorCrossf 输出带叉号的格式化错误消息
func ErrorCrossf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, styleError.Render("✗ "+msg))
}

// WarningSign 输出带警告符号的警告消息
func WarningSign(msg string) {
	fmt.Println(styleWarning.Render("⚠ " + msg))
}

// InfoSign 输出带信息符号的信息消息
func InfoSign(msg string) {
	fmt.Println(styleInfo.Render("ℹ " + msg))
}

// Arrow 输出带箭头的高亮消息
func Arrow(msg string) {
	fmt.Println(styleHighlight.Render("→ " + msg))
}

// Item 输出项目（标签: 值）
func Item(label, value string) {
	fmt.Printf("%s %s\n", styleDim.Render(label+":"), value)
}

// Style 返回指定颜色的 lipgloss.Style（用于高级用法）
func Style(color lipgloss.Color, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(color)
	if bold {
		s = s.Bold(true)
	}
	return s
}

// Render 使用指定样式渲染文本
func Render(style lipgloss.Style, format string, a ...interface{}) string {
	return style.Render(fmt.Sprintf(format, a...))
}

// PrintStyled 使用指定样式输出
func PrintStyled(style lipgloss.Style, format string, a ...interface{}) {
	fmt.Print(style.Render(fmt.Sprintf(format, a...)))
}

// PrintlnStyled 使用指定样式输出并换行
func PrintlnStyled(style lipgloss.Style, a ...interface{}) {
	fmt.Println(style.Render(joinStrings(a)))
}
