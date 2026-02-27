package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := Read(testFile)
	if err != nil {
		t.Errorf("Read() failed: %v", err)
	}

	if result != content {
		t.Errorf("Read() = %q, want %q", result, content)
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, err := Read("/nonexistent/file.txt")
	if err == nil {
		t.Error("Read() should return error for non-existent file")
	}
}

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")
	content := "hello world"

	err := Write(testFile, content)
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Written content = %q, want %q", string(data), content)
	}
}

func TestAppend(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// First write
	if err := Write(testFile, "hello"); err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	// Then append
	if err := Append(testFile, " world"); err != nil {
		t.Errorf("Append() failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(data) != "hello world" {
		t.Errorf("Content = %q, want %q", string(data), "hello world")
	}
}

func TestAppendCreateNew(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "new.txt")

	// Append to non-existent file should create it
	if err := Append(testFile, "content"); err != nil {
		t.Errorf("Append() failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(data) != "content" {
		t.Errorf("Content = %q, want %q", string(data), "content")
	}
}

func TestMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "subdir1", "subdir2")

	err := Mkdir(testDir)
	if err != nil {
		t.Errorf("Mkdir() failed: %v", err)
	}

	info, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("Failed to stat directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Mkdir() did not create a directory")
	}
}

func TestRmdir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a file inside
	testFile := filepath.Join(testDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rmdir(testDir)
	if err != nil {
		t.Errorf("Rmdir() failed: %v", err)
	}

	_, err = os.Stat(testDir)
	if !os.IsNotExist(err) {
		t.Error("Rmdir() did not remove directory")
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Remove(testFile)
	if err != nil {
		t.Errorf("Remove() failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Error("Remove() did not remove file")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	nonExistingFile := filepath.Join(tmpDir, "notexists.txt")

	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	exists, err := Exists(existingFile)
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("Exists() should return true for existing file")
	}

	exists, err = Exists(nonExistingFile)
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if exists {
		t.Error("Exists() should return false for non-existing file")
	}
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	testFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isDir, err := IsDir(testDir)
	if err != nil {
		t.Errorf("IsDir() failed: %v", err)
	}
	if !isDir {
		t.Error("IsDir() should return true for directory")
	}

	isDir, err = IsDir(testFile)
	if err != nil {
		t.Errorf("IsDir() failed: %v", err)
	}
	if isDir {
		t.Error("IsDir() should return false for file")
	}
}

func TestIsDirNotFound(t *testing.T) {
	_, err := IsDir("/nonexistent/path")
	if err == nil {
		t.Error("IsDir() should return error for non-existent path")
	}
}

func TestIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	testFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isFile, err := IsFile(testFile)
	if err != nil {
		t.Errorf("IsFile() failed: %v", err)
	}
	if !isFile {
		t.Error("IsFile() should return true for file")
	}

	isFile, err = IsFile(testDir)
	if err != nil {
		t.Errorf("IsFile() failed: %v", err)
	}
	if isFile {
		t.Error("IsFile() should return false for directory")
	}
}

func TestIsFileNotFound(t *testing.T) {
	_, err := IsFile("/nonexistent/path")
	if err == nil {
		t.Error("IsFile() should return error for non-existent path")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("test"), 0644)

	entries, err := List(tmpDir)
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Error("List() returned empty list")
	}

	// Check that all expected entries are present
	expected := map[string]bool{
		"file1.txt":       false,
		"file2.txt":       false,
		"subdir":          false,
	}

	for _, entry := range entries {
		if _, ok := expected[entry]; ok {
			expected[entry] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected entry %q not found in list", name)
		}
	}

	// Also check that subdir/file3.txt exists in the list
	foundSubdirFile := false
	for _, entry := range entries {
		if entry == "subdir/file3.txt" || entry == "subdir\\file3.txt" {
			foundSubdirFile = true
			break
		}
	}
	if !foundSubdirFile {
		t.Error("Expected entry \"subdir/file3.txt\" not found in list")
	}
}

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")
	content := "test content"

	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err := Copy(srcFile, dstFile)
	if err != nil {
		t.Errorf("Copy() failed: %v", err)
	}

	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Copied content = %q, want %q", string(data), content)
	}
}

func TestCopySourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dstFile := filepath.Join(tmpDir, "dst.txt")

	err := Copy("/nonexistent/source.txt", dstFile)
	if err == nil {
		t.Error("Copy() should return error for non-existent source")
	}
}

func TestRename(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "old.txt")
	dstFile := filepath.Join(tmpDir, "new.txt")
	content := "test content"

	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err := Rename(srcFile, dstFile)
	if err != nil {
		t.Errorf("Rename() failed: %v", err)
	}

	// Source should not exist
	_, err = os.Stat(srcFile)
	if !os.IsNotExist(err) {
		t.Error("Source file should not exist after rename")
	}

	// Destination should exist with correct content
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Renamed content = %q, want %q", string(data), content)
	}
}

func TestRenameSourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dstFile := filepath.Join(tmpDir, "new.txt")

	err := Rename("/nonexistent/old.txt", dstFile)
	if err == nil {
		t.Error("Rename() should return error for non-existent source")
	}
}
