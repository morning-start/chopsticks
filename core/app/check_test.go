package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"chopsticks/core/manifest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck_NotInstalled(t *testing.T) {
	storage, _, mgr, _ := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	result, err := mgr.Check(ctx, "nonexistent", CheckOptions{})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheck_Passed(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "path", Path: appDir},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-app", result.Name)
	assert.Equal(t, CheckStatusPassed, result.Status)
}

func TestCheck_FailedPath(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "path", Path: filepath.Join(appDir, "nonexistent-bin")},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusFailed, result.Status)
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, IssueTypePath, result.Issues[0].Type)
}

func TestCheck_FailedEnv(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "env", Key: "NONEXISTENT_ENV_VAR", Value: "test-value"},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusFailed, result.Status)
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, IssueTypeEnv, result.Issues[0].Type)
}

func TestCheck_RegistrySkipped(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "registry", Key: "HKCU\\Software\\TestApp"},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusPassed, result.Status)
	assert.Len(t, result.Issues, 0)
}

func TestCheck_AllOperations(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	binDir := filepath.Join(appDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	tmpFile := filepath.Join(appDir, "test.txt")
	err = os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "path", Path: binDir},
			{Type: "file", Path: tmpFile},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusPassed, result.Status)
	assert.Len(t, result.Issues, 0)
}

func TestCheck_CheckOptions_FilterByType(t *testing.T) {
	storage, adapter, mgr, tmpDir := createTestManager(t)
	defer storage.Close()

	ctx := context.Background()

	appDir := filepath.Join(tmpDir, "test-app")
	nonexistentBin := filepath.Join(appDir, "nonexistent-bin")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	opsFile := OperationsFile{
		Version: "1.0.0",
		Operations: []Operation{
			{Type: "path", Path: nonexistentBin},
			{Type: "env", Key: "NONEXISTENT_ENV_VAR", Value: "test-value"},
		},
	}
	data, err := json.Marshal(opsFile)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(appDir, "operations.json"), data, 0644)
	require.NoError(t, err)

	app := &manifest.InstalledApp{
		Name:       "test-app",
		Version:    "1.0.0",
		Bucket:     "main",
		InstallDir: appDir,
	}
	err = adapter.SaveInstalledApp(ctx, app)
	require.NoError(t, err)

	result, err := mgr.Check(ctx, "test-app", CheckOptions{CheckPaths: true, CheckEnv: false})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusFailed, result.Status)
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, IssueTypePath, result.Issues[0].Type)

	result, err = mgr.Check(ctx, "test-app", CheckOptions{CheckPaths: false, CheckEnv: true})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusFailed, result.Status)
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, IssueTypeEnv, result.Issues[0].Type)

	result, err = mgr.Check(ctx, "test-app", CheckOptions{CheckPaths: true, CheckEnv: true})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckStatusFailed, result.Status)
	assert.Len(t, result.Issues, 2)
}
