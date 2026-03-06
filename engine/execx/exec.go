// Package execx 提供命令执行功能。
package execx

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Result 存储命令执行结果。
type Result struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Success  bool   `json:"success"`
}

// Options 存储命令执行选项。
type Options struct {
	CWD     string            // 工作目录
	Env     map[string]string // 环境变量
	Timeout time.Duration     // 超时时间
}

// Exec 执行命令并返回结果。
func Exec(name string, args ...string) (*Result, error) {
	return ExecWithOptions(name, args, nil)
}

// ExecWithOptions 使用选项执行命令。
func ExecWithOptions(name string, args []string, opts *Options) (*Result, error) {
	var ctx context.Context
	if opts != nil && opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, name, args...)

	// 设置工作目录
	if opts != nil && opts.CWD != "" {
		cmd.Dir = opts.CWD
	}

	// 设置环境变量
	if opts != nil && len(opts.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("exec command: %w", err)
		}
	}

	return result, nil
}

// ExecContext 使用 context 执行命令。
func ExecContext(ctx context.Context, name string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("exec command: %w", err)
		}
	}

	return result, nil
}

// ExecWithTimeout 使用超时执行命令。
func ExecWithTimeout(timeout time.Duration, name string, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ExecContext(ctx, name, args...)
}

// ExecShell 执行 shell 命令。
func ExecShell(command string) (*Result, error) {
	return ExecWithOptions("cmd", []string{"/c", command}, nil)
}

// ExecShellWithOptions 使用选项执行 shell 命令。
func ExecShellWithOptions(command string, opts *Options) (*Result, error) {
	return ExecWithOptions("cmd", []string{"/c", command}, opts)
}

// ExecPowerShell 执行 PowerShell 命令。
func ExecPowerShell(command string) (*Result, error) {
	return ExecWithOptions("powershell", []string{"-Command", command}, nil)
}

// ExecPowerShellWithOptions 使用选项执行 PowerShell 命令。
func ExecPowerShellWithOptions(command string, opts *Options) (*Result, error) {
	return ExecWithOptions("powershell", []string{"-Command", command}, opts)
}

// Module 为脚本引擎提供 exec 注册。
type Module struct{}

// NewModule 创建新的 exec 模块。
func NewModule() *Module {
	return &Module{}
}
