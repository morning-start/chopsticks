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

var (
	stdout = colorable.NewColorableStdout()
	stderr = colorable.NewColorableStderr()
	colors = map[Color]string{
		Reset:         ansi.Reset,
		Black:         ansi.ColorCode("black"),
		Red:           ansi.ColorCode("red"),
		Green:         ansi.ColorCode("green"),
		Yellow:        ansi.ColorCode("yellow"),
		Blue:          ansi.ColorCode("blue"),
		Magenta:       ansi.ColorCode("magenta"),
		Cyan:          ansi.ColorCode("cyan"),
		White:         ansi.ColorCode("white"),
		BrightBlack:   ansi.ColorCode("black+h"),
		BrightRed:     ansi.ColorCode("red+h"),
		BrightGreen:   ansi.ColorCode("green+h"),
		BrightYellow:  ansi.ColorCode("yellow+h"),
		BrightBlue:    ansi.ColorCode("blue+h"),
		BrightMagenta: ansi.ColorCode("magenta+h"),
		BrightCyan:    ansi.ColorCode("cyan+h"),
		BrightWhite:   ansi.ColorCode("white+h"),
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

func PrintError(format string, args ...interface{}) {
	if IsColorEnabled() {
		message := Colorize(ErrorColor, fmt.Sprintf(format, args...))
		fmt.Fprint(stderr, message)
	} else {
		fmt.Fprintf(stderr, format, args...)
	}
}

func PrintErrorln(args ...interface{}) {
	if IsColorEnabled() {
		var argsWithColor []interface{}
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				argsWithColor = append(argsWithColor, Colorize(ErrorColor, s))
			} else {
				argsWithColor = append(argsWithColor, arg)
			}
		}
		fmt.Fprintln(stderr, argsWithColor...)
	} else {
		fmt.Fprintln(stderr, args...)
	}
}

func PrintSuccess(format string, args ...interface{}) {
	Print(SuccessColor, format, args...)
}

func PrintSuccessln(args ...interface{}) {
	Println(SuccessColor, args...)
}

func PrintWarning(format string, args ...interface{}) {
	Print(WarningColor, format, args...)
}

func PrintWarningln(args ...interface{}) {
	Println(WarningColor, args...)
}

func PrintInfo(format string, args ...interface{}) {
	Print(InfoColor, format, args...)
}

func PrintInfoln(args ...interface{}) {
	Println(InfoColor, args...)
}

func Header(format string, args ...interface{}) {
	Print(Cyan, format, args...)
}

func SubHeader(format string, args ...interface{}) {
	Print(Blue, format, args...)
}

func Item(label, value string) {
	if IsColorEnabled() {
		fmt.Fprintf(stdout, "%s %s\n", Colorize(BrightBlack, label+":"), value)
	} else {
		fmt.Printf("%s %s\n", label+":", value)
	}
}

func Checkmark() {
	Println(SuccessColor, "✓")
}

func CheckmarkWith(msg string) {
	Println(SuccessColor, "✓ "+msg)
}

func Crossmark() {
	Println(ErrorColor, "✗")
}

func CrossmarkWith(msg string) {
	Println(ErrorColor, "✗ "+msg)
}

func Arrow() {
	Print(InfoColor, "→ ")
}

func ArrowWith(msg string) {
	Print(InfoColor, "→ "+msg)
}
