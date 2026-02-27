package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
		wantErr  bool
	}{
		{"1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"v1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"V1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"1.2", &Version{Major: 1, Minor: 2, Patch: 0}, false},
		{"1", &Version{Major: 1, Minor: 0, Patch: 0}, false},
		{"1.2.3-alpha", &Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"}, false},
		{"1.2.3+build", &Version{Major: 1, Minor: 2, Patch: 3, Build: "build"}, false},
		{"1.2.3-alpha+build", &Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha", Build: "build"}, false},
		{"invalid", nil, true},
		{"", nil, true},
	}

	for _, tt := range tests {
		result, err := Parse(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if result.Major != tt.expected.Major ||
			result.Minor != tt.expected.Minor ||
			result.Patch != tt.expected.Patch ||
			result.Pre != tt.expected.Pre ||
			result.Build != tt.expected.Build {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  *Version
		expected string
	}{
		{&Version{Major: 1, Minor: 2, Patch: 3}, "1.2.3"},
		{&Version{Major: 0, Minor: 0, Patch: 1}, "0.0.1"},
		{&Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"}, "1.2.3-alpha"},
		{&Version{Major: 1, Minor: 2, Patch: 3, Build: "build123"}, "1.2.3+build123"},
		{&Version{Major: 1, Minor: 2, Patch: 3, Pre: "beta", Build: "build"}, "1.2.3-beta+build"},
	}

	for _, tt := range tests {
		result := tt.version.String()
		if result != tt.expected {
			t.Errorf("Version.String() = %q, want %q", result, tt.expected)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       *Version
		v2       *Version
		expected int
	}{
		// Major version comparison
		{&Version{Major: 2, Minor: 0, Patch: 0}, &Version{Major: 1, Minor: 0, Patch: 0}, 1},
		{&Version{Major: 1, Minor: 0, Patch: 0}, &Version{Major: 2, Minor: 0, Patch: 0}, -1},
		// Minor version comparison
		{&Version{Major: 1, Minor: 2, Patch: 0}, &Version{Major: 1, Minor: 1, Patch: 0}, 1},
		{&Version{Major: 1, Minor: 1, Patch: 0}, &Version{Major: 1, Minor: 2, Patch: 0}, -1},
		// Patch version comparison
		{&Version{Major: 1, Minor: 0, Patch: 2}, &Version{Major: 1, Minor: 0, Patch: 1}, 1},
		{&Version{Major: 1, Minor: 0, Patch: 1}, &Version{Major: 1, Minor: 0, Patch: 2}, -1},
		// Equal versions
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 1, Minor: 2, Patch: 3}, 0},
		// Pre-release comparison
		{&Version{Major: 1, Minor: 0, Patch: 0}, &Version{Major: 1, Minor: 0, Patch: 0, Pre: "alpha"}, 1},
		{&Version{Major: 1, Minor: 0, Patch: 0, Pre: "alpha"}, &Version{Major: 1, Minor: 0, Patch: 0}, -1},
	}

	for _, tt := range tests {
		result := tt.v1.Compare(tt.v2)
		if result != tt.expected {
			t.Errorf("Compare(%s, %s) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestComparisonMethods(t *testing.T) {
	v1 := &Version{Major: 1, Minor: 2, Patch: 3}
	v2 := &Version{Major: 1, Minor: 2, Patch: 4}
	v3 := &Version{Major: 1, Minor: 2, Patch: 3}

	if !v2.GT(v1) {
		t.Error("v2 should be greater than v1")
	}

	if !v1.LT(v2) {
		t.Error("v1 should be less than v2")
	}

	if !v1.EQ(v3) {
		t.Error("v1 should equal v3")
	}

	if !v2.GTE(v1) {
		t.Error("v2 should be greater than or equal to v1")
	}

	if !v1.LTE(v2) {
		t.Error("v1 should be less than or equal to v2")
	}

	if !v1.GTE(v3) {
		t.Error("v1 should be greater than or equal to v3")
	}

	if !v1.LTE(v3) {
		t.Error("v1 should be less than or equal to v3")
	}
}

func TestCompareStrings(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
		wantErr  bool
	}{
		{"1.2.3", "1.2.4", -1, false},
		{"1.2.4", "1.2.3", 1, false},
		{"1.2.3", "1.2.3", 0, false},
		{"2.0.0", "1.9.9", 1, false},
		{"invalid", "1.0.0", 0, true},
		{"1.0.0", "invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := CompareStrings(tt.v1, tt.v2)
		if tt.wantErr {
			if err == nil {
				t.Errorf("CompareStrings(%q, %q) expected error", tt.v1, tt.v2)
			}
			continue
		}
		if err != nil {
			t.Errorf("CompareStrings(%q, %q) unexpected error: %v", tt.v1, tt.v2, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("CompareStrings(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.2.3", true},
		{"v1.2.3", true},
		{"1.2", true},
		{"1", true},
		{"1.2.3-alpha", true},
		{"invalid", false},
		{"", false},
		{"1.2.3.4", false},
	}

	for _, tt := range tests {
		result := IsValid(tt.version)
		if result != tt.valid {
			t.Errorf("IsValid(%q) = %v, want %v", tt.version, result, tt.valid)
		}
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
		wantErr    bool
	}{
		// Exact version
		{"1.2.3", "1.2.3", true, false},
		{"1.2.3", "=1.2.3", true, false},
		{"1.2.3", "1.2.4", false, false},
		// Greater than
		{"1.2.3", ">1.2.2", true, false},
		{"1.2.3", ">1.2.3", false, false},
		// Greater than or equal
		{"1.2.3", ">=1.2.3", true, false},
		{"1.2.3", ">=1.2.4", false, false},
		// Less than
		{"1.2.3", "<1.2.4", true, false},
		{"1.2.3", "<1.2.3", false, false},
		// Less than or equal
		{"1.2.3", "<=1.2.3", true, false},
		{"1.2.3", "<=1.2.2", false, false},
		// Compatible version (^)
		{"1.2.3", "^1.2.0", true, false},
		{"1.2.3", "^1.0.0", true, false},
		{"2.0.0", "^1.0.0", false, false},
		// Approximate version (~)
		{"1.2.3", "~1.2.0", true, false},
		{"1.2.3", "~1.2.2", true, false},
		{"1.3.0", "~1.2.0", false, false},
		// Invalid
		{"invalid", ">=1.0.0", false, true},
		{"1.0.0", ">=invalid", false, true},
	}

	for _, tt := range tests {
		result, err := Satisfies(tt.version, tt.constraint)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Satisfies(%q, %q) expected error", tt.version, tt.constraint)
			}
			continue
		}
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", tt.version, tt.constraint, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Satisfies(%q, %q) = %v, want %v", tt.version, tt.constraint, result, tt.expected)
		}
	}
}

func TestComparePre(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},
		{"alpha", "alpha", 0},
		{"1", "2", -1},
		{"2", "1", 1},
		{"1", "1", 0},
		{"1", "alpha", -1}, // numeric < alphanumeric
		{"alpha", "1", 1},
		{"alpha.1", "alpha.2", -1},
		{"alpha.1", "alpha", 1}, // more fields = greater
	}

	for _, tt := range tests {
		result := comparePre(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("comparePre(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestVersionWithBuildMetadata(t *testing.T) {
	v, err := Parse("1.2.3+build.123")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if v.Build != "build.123" {
		t.Errorf("Build = %q, want %q", v.Build, "build.123")
	}

	// Build metadata should not affect comparison
	v1 := &Version{Major: 1, Minor: 2, Patch: 3, Build: "build1"}
	v2 := &Version{Major: 1, Minor: 2, Patch: 3, Build: "build2"}
	if v1.Compare(v2) != 0 {
		t.Error("Build metadata should not affect version comparison")
	}
}

func TestPreReleaseComparison(t *testing.T) {
	// 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0
	tests := []struct {
		v1       string
		v2       string
		expected int // -1: v1 < v2, 0: equal, 1: v1 > v2
	}{
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta.2", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-beta.11", "1.0.0-rc.1", -1},
		{"1.0.0-rc.1", "1.0.0", -1},
	}

	for _, tt := range tests {
		v1, err := Parse(tt.v1)
		if err != nil {
			t.Fatalf("Parse(%q) failed: %v", tt.v1, err)
		}
		v2, err := Parse(tt.v2)
		if err != nil {
			t.Fatalf("Parse(%q) failed: %v", tt.v2, err)
		}

		result := v1.Compare(v2)
		if result != tt.expected {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}
