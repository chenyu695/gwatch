package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gwatchignore")

	content := `# build artifacts
*.exe
*.dll

# test files
**/*_test.go

# blank lines above are intentional
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	patterns, err := parseIgnoreFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"*.exe", "*.dll", "**/*_test.go"}
	if len(patterns) != len(expected) {
		t.Fatalf("got %d patterns, want %d", len(patterns), len(expected))
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestParseIgnoreFileMissing(t *testing.T) {
	_, err := parseIgnoreFile("/nonexistent/.gwatchignore")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadIgnoreFiles(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// dir1 has an ignore file, dir2 does not
	if err := os.WriteFile(filepath.Join(dir1, ".gwatchignore"), []byte("*.log\nvendor/**\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patterns := loadIgnoreFiles([]string{dir1, dir2})
	expected := []string{"*.log", "vendor/**"}
	if len(patterns) != len(expected) {
		t.Fatalf("got %d patterns, want %d", len(patterns), len(expected))
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
		}
	}
}
