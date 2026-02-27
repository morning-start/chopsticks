package checksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	calc := New(SHA256)
	if calc == nil {
		t.Fatal("New() returned nil")
	}
}

func TestCalculateString(t *testing.T) {
	tests := []struct {
		alg      Algorithm
		input    string
		expected string
	}{
		{MD5, "hello", "5d41402abc4b2a76b9719d911017c592"},
		{SHA256, "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	}

	for _, tt := range tests {
		calc := New(tt.alg)
		result := calc.CalculateString(tt.input)
		if result != tt.expected {
			t.Errorf("CalculateString(%q) with %s = %q, want %q", tt.input, tt.alg, result, tt.expected)
		}
	}
}

func TestCalculate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	calc := New(SHA256)
	hash, err := calc.Calculate(testFile)
	if err != nil {
		t.Errorf("Calculate() failed: %v", err)
	}

	if hash == "" {
		t.Error("Calculate() returned empty hash")
	}

	// Verify the hash is correct
	expectedHash := calc.CalculateString(content)
	if hash != expectedHash {
		t.Errorf("Calculate() = %q, want %q", hash, expectedHash)
	}
}

func TestCalculateFileNotFound(t *testing.T) {
	calc := New(SHA256)
	_, err := calc.Calculate("/nonexistent/file.txt")
	if err == nil {
		t.Error("Calculate() should return error for non-existent file")
	}
}

func TestVerify(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	calc := New(SHA256)
	expectedHash := calc.CalculateString(content)

	ok, err := calc.Verify(testFile, expectedHash)
	if err != nil {
		t.Errorf("Verify() failed: %v", err)
	}

	if !ok {
		t.Error("Verify() should return true for correct hash")
	}
}

func TestVerifyWrongHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	calc := New(SHA256)

	ok, err := calc.Verify(testFile, "wronghash")
	if err != nil {
		t.Errorf("Verify() failed: %v", err)
	}

	if ok {
		t.Error("Verify() should return false for wrong hash")
	}
}

func TestVerifyString(t *testing.T) {
	calc := New(SHA256)
	data := "hello world"
	hash := calc.CalculateString(data)

	if !calc.VerifyString(data, hash) {
		t.Error("VerifyString() should return true for correct hash")
	}

	if calc.VerifyString(data, "wronghash") {
		t.Error("VerifyString() should return false for wrong hash")
	}
}

func TestCalculateFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := CalculateFile(testFile, SHA256)
	if err != nil {
		t.Errorf("CalculateFile() failed: %v", err)
	}

	if hash == "" {
		t.Error("CalculateFile() returned empty hash")
	}
}

func TestVerifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	calc := New(SHA256)
	hash := calc.CalculateString(content)

	ok, err := VerifyFile(testFile, hash, SHA256)
	if err != nil {
		t.Errorf("VerifyFile() failed: %v", err)
	}

	if !ok {
		t.Error("VerifyFile() should return true for correct hash")
	}
}

func TestCalculateBytes(t *testing.T) {
	data := []byte("test data")

	hash := CalculateBytes(data, SHA256)
	if hash == "" {
		t.Error("CalculateBytes() returned empty hash")
	}

	// Verify it matches CalculateString
	calc := New(SHA256)
	expected := calc.CalculateString(string(data))
	if hash != expected {
		t.Errorf("CalculateBytes() = %q, want %q", hash, expected)
	}
}

func TestAutoDetectAlgorithm(t *testing.T) {
	tests := []struct {
		checksum string
		expected Algorithm
	}{
		{"5d41402abc4b2a76b9719d911017c592", MD5},                    // 32 chars
		{"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", SHA256}, // 64 chars
		{"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", SHA512}, // 128 chars
		{"short", SHA256}, // default
	}

	for _, tt := range tests {
		result := AutoDetectAlgorithm(tt.checksum)
		if result != tt.expected {
			t.Errorf("AutoDetectAlgorithm(%q) = %s, want %s", tt.checksum, result, tt.expected)
		}
	}
}

func TestIsValidChecksum(t *testing.T) {
	tests := []struct {
		checksum string
		valid    bool
	}{
		{"5d41402abc4b2a76b9719d911017c592", true},                    // MD5
		{"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", true},  // SHA256
		{"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", true},  // SHA512
		{"", false},          // empty
		{"invalid!", false},  // invalid chars
		{"12345", false},     // wrong length
	}

	for _, tt := range tests {
		result := IsValidChecksum(tt.checksum)
		if result != tt.valid {
			t.Errorf("IsValidChecksum(%q) = %v, want %v", tt.checksum, result, tt.valid)
		}
	}
}

func TestDifferentAlgorithms(t *testing.T) {
	data := "test data"

	md5Calc := New(MD5)
	sha256Calc := New(SHA256)
	sha512Calc := New(SHA512)

	md5Hash := md5Calc.CalculateString(data)
	sha256Hash := sha256Calc.CalculateString(data)
	sha512Hash := sha512Calc.CalculateString(data)

	// Different algorithms should produce different hashes
	if md5Hash == sha256Hash {
		t.Error("MD5 and SHA256 should produce different hashes")
	}

	if sha256Hash == sha512Hash {
		t.Error("SHA256 and SHA512 should produce different hashes")
	}

	// Hash lengths should match algorithm
	if len(md5Hash) != 32 {
		t.Errorf("MD5 hash length = %d, want 32", len(md5Hash))
	}

	if len(sha256Hash) != 64 {
		t.Errorf("SHA256 hash length = %d, want 64", len(sha256Hash))
	}

	if len(sha512Hash) != 128 {
		t.Errorf("SHA512 hash length = %d, want 128", len(sha512Hash))
	}
}

func TestDefaultAlgorithm(t *testing.T) {
	// Unknown algorithm should default to SHA256
	calc := New(Algorithm("unknown"))

	data := "test"
	hash := calc.CalculateString(data)

	expectedCalc := New(SHA256)
	expectedHash := expectedCalc.CalculateString(data)

	if hash != expectedHash {
		t.Error("Unknown algorithm should default to SHA256")
	}
}
