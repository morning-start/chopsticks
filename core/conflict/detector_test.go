package conflict

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/manifest"
	"chopsticks/core/store"
)

func TestDetector_DetectFileConflicts(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建存储
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Close()

	// 创建检测器
	detector := NewDetector(storage, tempDir)

	// 测试场景1: 没有冲突
	t.Run("no conflicts", func(t *testing.T) {
		app := &manifest.App{
			Script: &manifest.AppScript{
				Name:   "test-app",
				Bucket: "main",
			},
		}

		result, err := detector.Detect(context.Background(), app)
		if err != nil {
			t.Fatalf("检测失败: %v", err)
		}

		if len(result.Conflicts) != 0 {
			t.Errorf("期望没有冲突，但检测到 %d 个冲突", len(result.Conflicts))
		}
	})

	// 测试场景2: 同名应用已安装
	t.Run("app already installed", func(t *testing.T) {
		// 先安装一个应用
		installedApp := &manifest.InstalledApp{
			Name:       "existing-app",
			Version:    "1.0.0",
			Bucket:     "main",
			InstallDir: filepath.Join(tempDir, "existing-app"),
		}
		if err := storage.SaveInstalledApp(context.Background(), installedApp); err != nil {
			t.Fatalf("保存安装记录失败: %v", err)
		}

		// 尝试安装同名应用
		app := &manifest.App{
			Script: &manifest.AppScript{
				Name:   "existing-app",
				Bucket: "main",
			},
		}

		result, err := detector.Detect(context.Background(), app)
		if err != nil {
			t.Fatalf("检测失败: %v", err)
		}

		if len(result.Conflicts) == 0 {
			t.Error("期望检测到冲突，但没有检测到")
		}

		// 验证冲突类型
		foundFileConflict := false
		for _, c := range result.Conflicts {
			if c.Type == ConflictTypeFile && c.CurrentApp == "existing-app" {
				foundFileConflict = true
				if c.Severity != SeverityWarning {
					t.Errorf("期望严重程度为 warning，但得到 %s", c.Severity)
				}
			}
		}
		if !foundFileConflict {
			t.Error("期望找到文件冲突，但没有找到")
		}
	})
}

func TestDetector_DetectPortConflicts(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Close()

	detector := NewDetector(storage, tempDir)

	// 测试 MySQL 端口检测
	t.Run("mysql port detection", func(t *testing.T) {
		app := &manifest.App{
			Script: &manifest.AppScript{
				Name:   "mysql",
				Bucket: "main",
			},
		}

		conflicts, err := detector.DetectPortConflicts(context.Background(), app)
		if err != nil {
			t.Fatalf("检测失败: %v", err)
		}

		// 检查是否包含 3306 端口
		found3306 := false
		for _, c := range conflicts {
			if c.Target == "3306" {
				found3306 = true
				if c.Type != ConflictTypePort {
					t.Errorf("期望冲突类型为 port，但得到 %s", c.Type)
				}
			}
		}

		// 如果 3306 被占用，应该检测到
		// 如果没有被占用，也可能没有冲突
		t.Logf("检测到 %d 个端口冲突，包含 3306: %v", len(conflicts), found3306)
	})
}

func TestIsPortInUse(t *testing.T) {
	// 测试端口检测函数
	// 注意：这个测试可能会受到系统环境影响

	// 测试一个不太可能使用的端口（高位端口）
	unlikelyPort := 54321
	if isPortInUse(unlikelyPort) {
		t.Logf("端口 %d 被占用（可能是测试环境特殊）", unlikelyPort)
	} else {
		t.Logf("端口 %d 未被占用", unlikelyPort)
	}
}

func TestFormatter_Format(t *testing.T) {
	formatter := NewFormatter(true)

	// 测试空结果
	t.Run("empty result", func(t *testing.T) {
		result := &Result{
			Conflicts: []Conflict{},
		}
		output := formatter.Format(result)
		if output != "✓ 未检测到冲突" {
			t.Errorf("期望 '✓ 未检测到冲突'，但得到 '%s'", output)
		}
	})

	// 测试有冲突的结果
	t.Run("with conflicts", func(t *testing.T) {
		result := &Result{
			Conflicts: []Conflict{
				{
					Type:        ConflictTypeFile,
					Severity:    SeverityWarning,
					Target:      "/test/path",
					CurrentApp:  "existing-app",
					Description: "测试冲突",
					Suggestion:  "测试建议",
				},
			},
			HasWarning: true,
		}
		output := formatter.Format(result)
		if output == "" || output == "✓ 未检测到冲突" {
			t.Error("期望有格式化输出，但得到空或未检测到冲突")
		}
		t.Logf("格式化输出:\n%s", output)
	})
}

func TestShouldBlockInstall(t *testing.T) {
	tests := []struct {
		name     string
		result   *Result
		force    bool
		expected bool
	}{
		{
			name:     "nil result",
			result:   nil,
			force:    false,
			expected: false,
		},
		{
			name:     "no conflicts",
			result:   &Result{Conflicts: []Conflict{}},
			force:    false,
			expected: false,
		},
		{
			name: "critical conflict without force",
			result: &Result{
				Conflicts: []Conflict{
					{Severity: SeverityCritical},
				},
				HasCritical: true,
			},
			force:    false,
			expected: true,
		},
		{
			name: "critical conflict with force",
			result: &Result{
				Conflicts: []Conflict{
					{Severity: SeverityCritical},
				},
				HasCritical: true,
			},
			force:    true,
			expected: false,
		},
		{
			name: "warning only",
			result: &Result{
				Conflicts: []Conflict{
					{Severity: SeverityWarning},
				},
				HasWarning: true,
			},
			force:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldBlockInstall(tt.result, tt.force)
			if got != tt.expected {
				t.Errorf("ShouldBlockInstall() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetAppPorts(t *testing.T) {
	detector := &detector{}

	tests := []struct {
		appName string
		want    []int
	}{
		{"mysql", []int{3306}},
		{"postgresql", []int{5432}},
		{"redis", []int{6379}},
		{"mongodb", []int{27017}},
		{"nginx", []int{80, 443}},
		{"nodejs", []int{3000, 8080}},
		{"unknown-app", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			app := &manifest.App{
				Script: &manifest.AppScript{
					Name: tt.appName,
				},
			}
			got := detector.getAppPorts(app)
			if len(got) != len(tt.want) {
				t.Errorf("getAppPorts() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getAppPorts()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// 集成测试：测试完整的冲突检测流程
func TestDetector_Detect_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 启用")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer storage.Close()

	detector := NewDetector(storage, tempDir)

	// 创建一个已安装应用
	installedApp := &manifest.InstalledApp{
		Name:       "test-mysql",
		Version:    "8.0.0",
		Bucket:     "main",
		InstallDir: filepath.Join(tempDir, "test-mysql"),
	}
	if err := storage.SaveInstalledApp(context.Background(), installedApp); err != nil {
		t.Fatalf("保存安装记录失败: %v", err)
	}

	// 检测同名应用
	app := &manifest.App{
		Script: &manifest.AppScript{
			Name:        "test-mysql",
			Description: "Test MySQL",
			Bucket:      "main",
		},
		Meta: &manifest.AppMeta{
			Version: "8.0.1",
		},
	}

	result, err := detector.Detect(context.Background(), app)
	if err != nil {
		t.Fatalf("检测失败: %v", err)
	}

	t.Logf("检测到 %d 个冲突:", len(result.Conflicts))
	for i, c := range result.Conflicts {
		t.Logf("  %d. [%s] %s: %s", i+1, c.Severity, c.Type, c.Description)
	}

	if len(result.Conflicts) == 0 {
		t.Error("期望检测到冲突，但没有检测到")
	}
}
