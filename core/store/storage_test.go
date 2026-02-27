package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chopsticks/core/manifest"
)

func setupTestStorage(t *testing.T) (Storage, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	storage, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test storage: %v", err)
	}

	return storage, dbPath
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	storage, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Error("New() returned nil storage")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("New() did not create database file")
	}
}

func TestNewCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	storage, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer storage.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "subdir")); os.IsNotExist(err) {
		t.Error("New() did not create directory")
	}
}

func TestSaveAndGetInstalledApp(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Save app
	err := storage.SaveInstalledApp(ctx, app)
	if err != nil {
		t.Fatalf("SaveInstalledApp() error = %v", err)
	}

	// Get app
	retrieved, err := storage.GetInstalledApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetInstalledApp() error = %v", err)
	}

	if retrieved.Name != app.Name {
		t.Errorf("Name = %s, want %s", retrieved.Name, app.Name)
	}
	if retrieved.Version != app.Version {
		t.Errorf("Version = %s, want %s", retrieved.Version, app.Version)
	}
	if retrieved.Bucket != app.Bucket {
		t.Errorf("Bucket = %s, want %s", retrieved.Bucket, app.Bucket)
	}
	if retrieved.InstallDir != app.InstallDir {
		t.Errorf("InstallDir = %s, want %s", retrieved.InstallDir, app.InstallDir)
	}
}

func TestGetInstalledAppNotFound(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	_, err := storage.GetInstalledApp(ctx, "non-existent-app")
	if err == nil {
		t.Error("GetInstalledApp() should return error for non-existent app")
	}
}

func TestDeleteInstalledApp(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Save and delete
	storage.SaveInstalledApp(ctx, app)
	err := storage.DeleteInstalledApp(ctx, "test-app")
	if err != nil {
		t.Fatalf("DeleteInstalledApp() error = %v", err)
	}

	// Verify deletion
	_, err = storage.GetInstalledApp(ctx, "test-app")
	if err == nil {
		t.Error("GetInstalledApp() should return error after deletion")
	}
}

func TestListInstalledApps(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// Initially empty
	apps, err := storage.ListInstalledApps(ctx)
	if err != nil {
		t.Fatalf("ListInstalledApps() error = %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("len(apps) = %d, want 0", len(apps))
	}

	// Add some apps
	appNames := []string{"app-a", "app-b", "app-c"}
	for _, name := range appNames {
		app := &manifest.InstalledApp{
			Name:        name,
			Version:     "1.0.0",
			Bucket:      "main",
			InstallDir:  "/apps/" + name,
			InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		storage.SaveInstalledApp(ctx, app)
	}

	// List apps
	apps, err = storage.ListInstalledApps(ctx)
	if err != nil {
		t.Fatalf("ListInstalledApps() error = %v", err)
	}
	if len(apps) != 3 {
		t.Errorf("len(apps) = %d, want 3", len(apps))
	}
}

func TestIsInstalled(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// Not installed initially
	installed, err := storage.IsInstalled(ctx, "test-app")
	if err != nil {
		t.Fatalf("IsInstalled() error = %v", err)
	}
	if installed {
		t.Error("IsInstalled() = true, want false")
	}

	// Install app
	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	storage.SaveInstalledApp(ctx, app)

	// Now should be installed
	installed, err = storage.IsInstalled(ctx, "test-app")
	if err != nil {
		t.Fatalf("IsInstalled() error = %v", err)
	}
	if !installed {
		t.Error("IsInstalled() = false, want true")
	}
}

func TestSaveAndGetBucket(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	bucket := &manifest.BucketConfig{
		ID:          "main",
		Name:        "Main Bucket",
		Author:      "Test Author",
		Description: "Test Description",
		Homepage:    "https://example.com",
		License:     "MIT",
		Repository: manifest.RepositoryInfo{
			Type:   "git",
			URL:    "https://github.com/example/bucket",
			Branch: "main",
		},
	}

	// Save bucket
	err := storage.SaveBucket(ctx, bucket)
	if err != nil {
		t.Fatalf("SaveBucket() error = %v", err)
	}

	// Get bucket
	retrieved, err := storage.GetBucket(ctx, "main")
	if err != nil {
		t.Fatalf("GetBucket() error = %v", err)
	}

	if retrieved.ID != bucket.ID {
		t.Errorf("ID = %s, want %s", retrieved.ID, bucket.ID)
	}
	if retrieved.Name != bucket.Name {
		t.Errorf("Name = %s, want %s", retrieved.Name, bucket.Name)
	}
	if retrieved.Author != bucket.Author {
		t.Errorf("Author = %s, want %s", retrieved.Author, bucket.Author)
	}
}

func TestGetBucketNotFound(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	_, err := storage.GetBucket(ctx, "non-existent-bucket")
	if err == nil {
		t.Error("GetBucket() should return error for non-existent bucket")
	}
}

func TestDeleteBucket(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	bucket := &manifest.BucketConfig{
		ID:   "test-bucket",
		Name: "Test Bucket",
		Repository: manifest.RepositoryInfo{
			Type:   "git",
			URL:    "https://github.com/example/bucket",
			Branch: "main",
		},
	}

	// Save and delete
	storage.SaveBucket(ctx, bucket)
	err := storage.DeleteBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("DeleteBucket() error = %v", err)
	}

	// Verify deletion
	_, err = storage.GetBucket(ctx, "test-bucket")
	if err == nil {
		t.Error("GetBucket() should return error after deletion")
	}
}

func TestListBuckets(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// Initially empty
	buckets, err := storage.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}
	if len(buckets) != 0 {
		t.Errorf("len(buckets) = %d, want 0", len(buckets))
	}

	// Add some buckets
	bucketIDs := []string{"bucket-a", "bucket-b", "bucket-c"}
	for i, id := range bucketIDs {
		bucket := &manifest.BucketConfig{
			ID:   id,
			Name: "Bucket " + string(rune('A'+i)),
			Repository: manifest.RepositoryInfo{
				Type:   "git",
				URL:    "https://github.com/example/" + id,
				Branch: "main",
			},
		}
		storage.SaveBucket(ctx, bucket)
	}

	// List buckets
	buckets, err = storage.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}
	if len(buckets) != 3 {
		t.Errorf("len(buckets) = %d, want 3", len(buckets))
	}
}

func TestSaveAndGetInstallOperation(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// First create an installed app
	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	storage.SaveInstalledApp(ctx, app)

	op := &InstallOperation{
		InstalledID: "test-app",
		Operation:   "create_symlink",
		TargetPath:  "/bin/test-app",
		TargetValue: "/apps/test-app/bin/test-app",
	}

	// Save operation
	err := storage.SaveInstallOperation(ctx, op)
	if err != nil {
		t.Fatalf("SaveInstallOperation() error = %v", err)
	}

	// Get operations
	ops, err := storage.GetInstallOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetInstallOperations() error = %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("len(ops) = %d, want 1", len(ops))
	}

	if ops[0].Operation != op.Operation {
		t.Errorf("Operation = %s, want %s", ops[0].Operation, op.Operation)
	}
}

func TestDeleteInstallOperations(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// Create app and operation
	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	storage.SaveInstalledApp(ctx, app)

	op := &InstallOperation{
		InstalledID: "test-app",
		Operation:   "create_symlink",
		TargetPath:  "/bin/test-app",
	}
	storage.SaveInstallOperation(ctx, op)

	// Delete operations
	err := storage.DeleteInstallOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("DeleteInstallOperations() error = %v", err)
	}

	// Verify deletion
	ops, err := storage.GetInstallOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetInstallOperations() error = %v", err)
	}
	if len(ops) != 0 {
		t.Errorf("len(ops) = %d, want 0", len(ops))
	}
}

func TestSaveAndGetSystemOperation(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// First create an installed app
	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	storage.SaveInstalledApp(ctx, app)

	op := &SystemOperation{
		InstalledID:   "test-app",
		Operation:     "set_env",
		TargetType:    "env",
		TargetKey:     "TEST_APP_HOME",
		TargetValue:   "/apps/test-app",
		OriginalValue: "",
	}

	// Save operation
	err := storage.SaveSystemOperation(ctx, op)
	if err != nil {
		t.Fatalf("SaveSystemOperation() error = %v", err)
	}

	// Get operations
	ops, err := storage.GetSystemOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetSystemOperations() error = %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("len(ops) = %d, want 1", len(ops))
	}

	if ops[0].Operation != op.Operation {
		t.Errorf("Operation = %s, want %s", ops[0].Operation, op.Operation)
	}
	if ops[0].TargetKey != op.TargetKey {
		t.Errorf("TargetKey = %s, want %s", ops[0].TargetKey, op.TargetKey)
	}
}

func TestDeleteSystemOperations(t *testing.T) {
	storage, _ := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	// Create app and operation
	app := &manifest.InstalledApp{
		Name:        "test-app",
		Version:     "1.0.0",
		Bucket:      "main",
		InstallDir:  "/apps/test-app",
		InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	storage.SaveInstalledApp(ctx, app)

	op := &SystemOperation{
		InstalledID: "test-app",
		Operation:   "set_env",
		TargetType:  "env",
		TargetKey:   "TEST_APP_HOME",
	}
	storage.SaveSystemOperation(ctx, op)

	// Delete operations
	err := storage.DeleteSystemOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("DeleteSystemOperations() error = %v", err)
	}

	// Verify deletion
	ops, err := storage.GetSystemOperations(ctx, "test-app")
	if err != nil {
		t.Fatalf("GetSystemOperations() error = %v", err)
	}
	if len(ops) != 0 {
		t.Errorf("len(ops) = %d, want 0", len(ops))
	}
}

func TestClose(t *testing.T) {
	storage, _ := setupTestStorage(t)

	err := storage.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
