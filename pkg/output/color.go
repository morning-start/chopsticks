package output

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
)

type Color int

const (
	Reset Color = iota
	Bold
	Faint
	Italic
	Underline

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

	Success = Green
	Warning = Yellow
	Error   = Red
	Info    = Cyan
)

var (
	stdout = colorable.NewColorableStdout()
	stderr = colorable.NewColorableStderr()
	colors  = map[Color]string{
		Reset:          ansi.Reset,
		Bold:           ansi.Bold,
		Faint:          ansi.Faint,
		Italic:         ansi.Italic,
		Underline:      ansi.Underline,
		Black:          ansi.ColorCode("black"),
		Red:            ansi.ColorCode("red"),
		Green:          ansi.ColorCode("green"),
		Yellow:         ansi.ColorCode("yellow"),
		Blue:           ansi.ColorCode("blue"),
		Magenta:        ansi.ColorCode("magenta"),
		Cyan:           ansi.ColorCode("cyan"),
		White:          ansi.ColorCode("white"),
		BrightBlack:    ansi.ColorCode("black+h"),
		BrightRed:      ansi.ColorCode("red+h"),
		BrightGreen:    ansi.ColorCode("green+h"),
		BrightYellow:   ansi.ColorCode("yellow+h"),
		BrightBlue:     ansi.ColorCode("blue+h"),
		BrightMagenta:  ansi.ColorCode("magenta+h"),
		BrightCyan:     ansi.ColorCode("cyan+h"),
		BrightWhite:    ansi.ColorCode("white+h"),
	}
	disableColor bool
)

func DisableColor() {
	disableColor = true
}

func EnableColor() {
	disableColor = false
}

func IsColorEnabled() bool {
	return !disableColor && isTerminal()
}

func isTerminal() bool {
	return isStdoutTerminal() || isStderrTerminal()
}

func isStdoutTerminal() bool {
	return isTty(os.Stdout.Fd())
}

func isStderrTerminal() bool {
	return isTty(os.Stderr.Fd())
}

func isTty(fd uintptr) bool {
	return false
}

func Colorize(color Color, message string) string {
	if disableColor || !isTerminal() {
		return message
	}
	c, ok := colors[color]
	if !ok {
		return message
	}
	return c + message + ansi.Reset
}

func ColorizeBg(color Color, bgColor Color, message string) string {
	if disableColor || !isTerminal() {
		return message
	}
	fg, ok := colors[color]
	if !ok {
		return message
	}
	bg, ok := colors[bgColor]
	if !ok {
		return message
	}
	return fg + bg + message + ansi.Reset
}

func Print(color Color, format string, args ...interface{}) {
	if IsColorEnabled() {
		fmt.Fprint(stdout, Colorize(color, fmt.Sprintf(format, args...)))
	} else {
		fmt.Fprintf(stdout, format, args...)
	}
}

func Println(color Color, args ...interface{}) {
	if IsColorEnabled() {
		var argsWithColor []interface{}
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				argsWithColor = append(argsWithColor, Colorize(color, s))
			} else {
				argsWithColor = append(argsWithColor, arg)
			}
		}
		fmt.Fprintln(stdout, argsWithColor...)
	} else {
		fmt.Fprintln(stdout, args...)
	}
}

func Printf(color Color, format string, args ...interface{}) {
	if IsColorEnabled() {
		message := Colorize(color, fmt.Sprintf(format, args...))
		fmt.Fprint(stdout, message)
	} else {
		fmt.Fprintf(stdout, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if IsColorEnabled() {
		message := Colorize(Error, fmt.Sprintf(format, args...))
		fmt.Fprint(stderr, message)
	} else {
		fmt.Fprintf(stderr, format, args...)
	}
}

func Errorln(args ...interface{}) {
	if IsColorEnabled() {
		var argsWithColor []interface{}
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				argsWithColor = append(argsWithColor, Colorize(Error, s))
			} else {
				argsWithColor = append(argsWithColor, arg)
			}
		}
		fmt.Fprintln(stderr, argsWithColor...)
	} else {
		fmt.Fprintln(stderr, args...)
	}
}

func Success(format string, args ...interface{}) {
	Print(Success, format, args...)
}

func Successln(args ...interface{}) {
	Println(Success, args...)
}

func Warning(format string, args ...interface{}) {
	Print(Warning, format, args...)
}

func Warningln(args ...interface{}) {
	Println(Warning, args...)
}

func Info(format string, args ...interface{}) {
	Print(Info, format, args...)
}

func Infoln(args ...interface{}) {
	Println(Info, args...)
}

func Header(format string, args ...interface{}) {
	Print(Bold+Cyan, format, args...)
}

func SubHeader(format string, args ...interface{}) {
	Print(Bold+Blue, format, args...)
}

func Item(label, value string) {
	if IsColorEnabled() {
		fmt.Fprintf(stdout, "%s %s\n", Colorize(Bold, label+":"), value)
	} else {
		fmt.Printf("%s %s\n", label+":", value)
	}
}

func Checkmark() {
	Println(Success, "✓")
}

func CheckmarkWith(msg string) {
	Println(Success, "✓ "+msg)
}

func Crossmark() {
	Println(Error, "✗")
}

func CrossmarkWith(msg string) {
	Println(Error, "✗ "+msg)
}

func Arrow() {
	Print(Info, "→ ")
}

func ArrowWith(msg string) {
	Print(Info, "→ "+msg)
}
