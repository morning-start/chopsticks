//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// E2EEnvironment E2E 测试环境
type E2EEnvironment struct {
	BinaryPath    string
	TmpDir        string
	TestBucketURL string
	EnvVars       []string
}

// SetupE2EEnvironment 设置 E2E 测试环境
func SetupE2EEnvironment(t *testing.T) *E2EEnvironment {
	t.Helper()

	tmpDir := t.TempDir()

	// 构建二进制文件
	binaryPath := buildBinary(t)

	// 设置测试 bucket 路径
	testBucketPath := filepath.Join(tmpDir, "test-bucket")
	setupTestBucket(t, testBucketPath)

	// 设置环境变量
	envVars := []string{
		"CHOPSTICKS_HOME=" + tmpDir,
		"CHOPSTICKS_CONFIG=" + filepath.Join(tmpDir, "config.yaml"),
		"PATH=" + os.Getenv("PATH"),
	}

	return &E2EEnvironment{
		BinaryPath:    binaryPath,
		TmpDir:        tmpDir,
		TestBucketURL: testBucketPath,
		EnvVars:       envVars,
	}
}

// buildBinary 构建 CLI 二进制文件
func buildBinary(t *testing.T) string {
	t.Helper()

	// 检查是否已有构建好的二进制文件
	binaryName := "chopsticks"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// 优先使用已存在的二进制文件
	if _, err := os.Stat(binaryName); err == nil {
		absPath, _ := filepath.Abs(binaryName)
		return absPath
	}

	// 构建新的二进制文件
	buildDir := t.TempDir()
	binaryPath := filepath.Join(buildDir, binaryName)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/cli")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("构建二进制文件失败: %v\n输出: %s", err, string(output))
	}

	return binaryPath
}

// setupTestBucket 设置测试 bucket
func setupTestBucket(t *testing.T, path string) {
	t.Helper()

	// 创建 bucket 目录结构
	appsDir := filepath.Join(path, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("创建 apps 目录失败: %v", err)
	}

	// 创建 bucket.json
	bucketConfig := `{
  "id": "test-bucket",
  "name": "test-bucket",
  "description": "Test bucket for E2E testing"
}
`
	if err := os.WriteFile(
		filepath.Join(path, "bucket.json"),
		[]byte(bucketConfig),
		0644,
	); err != nil {
		t.Fatalf("创建 bucket.json 失败: %v", err)
	}

	// 创建测试应用
	createE2ETestApp(t, appsDir, "test-app")
	createE2ETestApp(t, appsDir, "git")
}

// createE2ETestApp 创建 E2E 测试应用
func createE2ETestApp(t *testing.T, appsDir, name string) {
	t.Helper()

	scriptContent := fmt.Sprintf(`/**
 * @description E2E test app %s
 * @version 1.0.0
 */

const app = {
  name: "%s",
  version: "1.0.0",
  architecture: {
    "64bit": {
      url: "https://example.com/%s-1.0.0.zip",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["%s.exe"]
};

module.exports = app;
`, name, name, name, name)

	if err := os.WriteFile(
		filepath.Join(appsDir, name+".js"),
		[]byte(scriptContent),
		0644,
	); err != nil {
		t.Fatalf("创建应用脚本失败: %v", err)
	}
}

// RunCLI 运行 CLI 命令
func RunCLI(t *testing.T, env *E2EEnvironment, args ...string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, env.BinaryPath, args...)
	cmd.Env = env.EnvVars
	cmd.Dir = env.TmpDir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCLISuccess 运行 CLI 命令并期望成功
func RunCLISuccess(t *testing.T, env *E2EEnvironment, args ...string) string {
	t.Helper()

	output, err := RunCLI(t, env, args...)
	if err != nil {
		t.Fatalf("CLI 命令失败: %v\n输出: %s", err, output)
	}

	return output
}

// RunCLIFail 运行 CLI 命令并期望失败
func RunCLIFail(t *testing.T, env *E2EEnvironment, args ...string) (string, error) {
	t.Helper()

	output, err := RunCLI(t, env, args...)
	if err == nil {
		t.Fatalf("CLI 命令应该失败但没有: %s", output)
	}

	return output, err
}

// AssertOutputContains 断言输出包含特定内容
func AssertOutputContains(t *testing.T, output, expected string) {
	t.Helper()

	if !contains(output, expected) {
		t.Errorf("输出不包含期望的内容\n期望: %s\n实际: %s", expected, output)
	}
}

// AssertOutputNotContains 断言输出不包含特定内容
func AssertOutputNotContains(t *testing.T, output, unexpected string) {
	t.Helper()

	if contains(output, unexpected) {
		t.Errorf("输出不应该包含: %s", unexpected)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) > 0 && (s == substr || len(s) > len(substr) && containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
