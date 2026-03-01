package manifest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App{
		Script: &AppScript{
			Name:        "test-app",
			Description: "A test application",
			Homepage:    "https://example.com/app",
			License:     "MIT",
			Category:    "development",
			Tags:        []string{"test", "dev"},
			Maintainer:  "test@example.com",
			Bucket:      "main",
			Dependencies: []Dependency{
				{
					Name:     "dep1",
					Version:  ">=1.0.0",
					Optional: false,
				},
				{
					Name:     "dep2",
					Version:  "^2.0.0",
					Optional: true,
					Conditions: map[string]string{
						"os": "windows",
					},
				},
			},
		},
		Meta: &AppMeta{
			Version: "1.0.0",
			Versions: map[string]VersionInfo{
				"1.0.0": {
					Version:    "1.0.0",
					ReleasedAt: time.Now(),
					Downloads: map[string]DownloadInfo{
						"amd64": {
							URL:  "https://example.com/app-1.0.0-amd64.zip",
							Hash: "sha256:abc123",
							Size: 1024000,
							Type: "zip",
						},
					},
				},
			},
		},
		Ref: &AppRef{
			Name:        "test-app",
			Description: "A test application",
			Version:     "1.0.0",
			Category:    "development",
			Tags:        []string{"test", "dev"},
			ScriptPath:  "/path/to/app.lua",
			MetaPath:    "/path/to/app.meta.json",
		},
	}

	assert.Equal(t, "test-app", app.Script.Name)
	assert.Equal(t, "1.0.0", app.Meta.Version)
	assert.Equal(t, "test-app", app.Ref.Name)
	assert.Len(t, app.Script.Dependencies, 2)
}

func TestApp_Empty(t *testing.T) {
	app := App{}

	assert.Nil(t, app.Script)
	assert.Nil(t, app.Meta)
	assert.Nil(t, app.Ref)
}

func TestAppScript(t *testing.T) {
	script := AppScript{
		Name:        "test-app",
		Description: "Test description",
		Homepage:    "https://example.com",
		License:     "MIT",
		Category:    "tools",
		Tags:        []string{"cli", "tool"},
		Maintainer:  "maintainer@example.com",
		Bucket:      "main",
		Dependencies: []Dependency{
			{Name: "dep1", Version: ">=1.0.0"},
			{Name: "dep2", Version: "^2.0.0", Optional: true},
		},
	}

	assert.Equal(t, "test-app", script.Name)
	assert.Equal(t, "Test description", script.Description)
	assert.Len(t, script.Tags, 2)
	assert.Len(t, script.Dependencies, 2)
}

func TestAppScript_NoDependencies(t *testing.T) {
	script := AppScript{
		Name:         "standalone-app",
		Description:  "An app without dependencies",
		Dependencies: nil,
	}

	assert.Nil(t, script.Dependencies)
	assert.Empty(t, script.Dependencies)
}

func TestDependency(t *testing.T) {
	dep := Dependency{
		Name:     "test-dep",
		Version:  ">=1.0.0",
		Optional: true,
		Conditions: map[string]string{
			"os":   "windows",
			"arch": "amd64",
		},
	}

	assert.Equal(t, "test-dep", dep.Name)
	assert.Equal(t, ">=1.0.0", dep.Version)
	assert.True(t, dep.Optional)
	assert.Len(t, dep.Conditions, 2)
}

func TestDependency_NoConditions(t *testing.T) {
	dep := Dependency{
		Name:    "simple-dep",
		Version: "1.0.0",
	}

	assert.Nil(t, dep.Conditions)
	assert.False(t, dep.Optional)
}

func TestAppMeta(t *testing.T) {
	now := time.Now()
	meta := AppMeta{
		Version: "2.0.0",
		Versions: map[string]VersionInfo{
			"1.0.0": {
				Version:    "1.0.0",
				ReleasedAt: now.Add(-30 * 24 * time.Hour),
				Downloads:  map[string]DownloadInfo{},
			},
			"2.0.0": {
				Version:    "2.0.0",
				ReleasedAt: now,
				Downloads: map[string]DownloadInfo{
					"amd64": {
						URL:  "https://example.com/v2.0.0-amd64.zip",
						Hash: "sha256:def456",
						Size: 2048000,
						Type: "zip",
					},
					"arm64": {
						URL:  "https://example.com/v2.0.0-arm64.zip",
						Hash: "sha256:ghi789",
						Size: 1892000,
						Type: "zip",
					},
				},
			},
		},
	}

	assert.Equal(t, "2.0.0", meta.Version)
	assert.Len(t, meta.Versions, 2)
	assert.Contains(t, meta.Versions, "1.0.0")
	assert.Contains(t, meta.Versions, "2.0.0")

	v2 := meta.Versions["2.0.0"]
	assert.Len(t, v2.Downloads, 2)
	assert.Contains(t, v2.Downloads, "amd64")
	assert.Contains(t, v2.Downloads, "arm64")
}

func TestVersionInfo(t *testing.T) {
	now := time.Now()
	info := VersionInfo{
		Version:    "1.0.0",
		ReleasedAt: now,
		Downloads: map[string]DownloadInfo{
			"amd64": {
				URL:  "https://example.com/app.zip",
				Hash: "sha256:abc123",
				Size: 1024000,
				Type: "zip",
			},
		},
	}

	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, now, info.ReleasedAt)
	assert.Len(t, info.Downloads, 1)
}

func TestDownloadInfo(t *testing.T) {
	download := DownloadInfo{
		URL:  "https://example.com/app.zip",
		Hash: "sha256:abc123def456",
		Size: 1024000,
		Type: "zip",
	}

	assert.Equal(t, "https://example.com/app.zip", download.URL)
	assert.Equal(t, "sha256:abc123def456", download.Hash)
	assert.Equal(t, int64(1024000), download.Size)
	assert.Equal(t, "zip", download.Type)
}

func TestAppRef(t *testing.T) {
	ref := AppRef{
		Name:        "test-app",
		Description: "Test application",
		Version:     "1.0.0",
		Category:    "tools",
		Tags:        []string{"test", "dev"},
		ScriptPath:  "/buckets/main/app.lua",
		MetaPath:    "/buckets/main/app.meta.json",
	}

	assert.Equal(t, "test-app", ref.Name)
	assert.Equal(t, "1.0.0", ref.Version)
	assert.Equal(t, "/buckets/main/app.lua", ref.ScriptPath)
}

func TestInstalledApp(t *testing.T) {
	now := time.Now()
	installed := InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "test-app", installed.Name)
	assert.Equal(t, "1.0.0", installed.Version)
	assert.Equal(t, "main", installed.Bucket)
	assert.Equal(t, "/apps/test-app", installed.InstallDir)
	assert.Equal(t, now, installed.InstalledAt)
}

func TestAppInfo(t *testing.T) {
	info := AppInfo{
		Name:             "test-app",
		Description:      "Test application",
		Homepage:         "https://example.com",
		License:          "MIT",
		Category:         "tools",
		Tags:             []string{"test"},
		Version:          "1.0.0",
		Bucket:           "main",
		Installed:        true,
		InstalledVersion: "1.0.0",
	}

	assert.Equal(t, "test-app", info.Name)
	assert.True(t, info.Installed)
	assert.Equal(t, "1.0.0", info.InstalledVersion)
}

func TestAppInfo_NotInstalled(t *testing.T) {
	info := AppInfo{
		Name:        "test-app",
		Description: "Test application",
		Version:     "1.0.0",
		Bucket:      "main",
		Installed:   false,
	}

	assert.False(t, info.Installed)
	assert.Empty(t, info.InstalledVersion)
}
