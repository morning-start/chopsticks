// Package execx 提供命令执行功能。
package execx

import (
	"bytes"
	"context"
	"fmt"
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

// Exec 执行命令并返回结果。
func Exec(name string, args ...string) (*Result, error) {
	return ExecContext(context.Background(), name, args...)
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
	return Exec("cmd", "/c", command)
}

// ExecPowerShell 执行 PowerShell 命令。
func ExecPowerShell(command string) (*Result, error) {
	return Exec("powershell", "-Command", command)
}

// Module 为脚本引擎提供 exec 注册。
type Module struct{}

// NewModule 创建新的 exec 模块。
func NewModule() *Module {
	return &Module{}
}
