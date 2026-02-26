// Package installer 提供安装程序静默安装功能。
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Type string

const (
	NSIS   Type = "nsis"
	MSI    Type = "msi"
	Inno   Type = "inno"
	Unknown Type = "unknown"
)

type Options struct {
	Silent     bool
	InstallDir string
	ExtraArgs  []string
}

func DetectType(installerPath string) Type {
	ext := strings.ToLower(filepath.Ext(installerPath))
	switch ext {
	case ".exe":
		return detectExeType(installerPath)
	case ".msi":
		return MSI
	default:
		return Unknown
	}
}

func detectExeType(installerPath string) Type {
	data, err := os.ReadFile(installerPath)
	if err != nil {
		return Unknown
	}

	content := string(data)
	if strings.Contains(content, "NSIS") || strings.Contains(content, "Nullsoft") {
		return NSIS
	}
	if strings.Contains(content, "Inno Setup") || strings.Contains(content, "Inno") {
		return Inno
	}

	return Unknown
}

func Run(installerPath string, opts Options) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("仅支持 Windows 平台")
	}

	typ := DetectType(installerPath)
	if typ == Unknown {
		return fmt.Errorf("无法识别的安装程序类型: %s", installerPath)
	}

	var args []string
	switch typ {
	case NSIS:
		args = buildNSISArgs(installerPath, opts)
	case MSI:
		args = buildMSIArgs(installerPath, opts)
	case Inno:
		args = buildInnoArgs(installerPath, opts)
	}

	cmd := exec.Command(installerPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	return nil
}

func buildNSISArgs(_ string, opts Options) []string {
	args := []string{}

	if opts.Silent {
		args = append(args, "/S")
	}

	if opts.InstallDir != "" {
		args = append(args, "/D="+opts.InstallDir)
	}

	args = append(args, opts.ExtraArgs...)

	return args
}

func buildMSIArgs(installerPath string, opts Options) []string {
	args := []string{}

	args = append(args, "/i")
	args = append(args, installerPath)

	if opts.Silent {
		args = append(args, "/quiet")
		args = append(args, "/norestart")
	}

	if opts.InstallDir != "" {
		args = append(args, fmt.Sprintf("INSTALLDIR=%s", opts.InstallDir))
	}

	args = append(args, opts.ExtraArgs...)

	return args
}

func buildInnoArgs(_ string, opts Options) []string {
	args := []string{}

	if opts.Silent {
		args = append(args, "/VERYSILENT")
		args = append(args, "/SUPPRESSMSGBOXES")
		args = append(args, "/NORESTART")
	}

	if opts.InstallDir != "" {
		args = append(args, "/DIR="+opts.InstallDir)
	}

	args = append(args, opts.ExtraArgs...)

	return args
}

func IsSupported(path string) bool {
	typ := DetectType(path)
	return typ != Unknown
}

func GetTypeName(typ Type) string {
	switch typ {
	case NSIS:
		return "NSIS"
	case MSI:
		return "MSI"
	case Inno:
		return "Inno Setup"
	default:
		return "Unknown"
	}
}
