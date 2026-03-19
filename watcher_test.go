package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMatchesIncludePatterns(t *testing.T) {
	w := &Watcher{
		logger:   NewLogger(),
		patterns: []string{"**/*.go"},
		ignores:  nil,
	}

	wd, _ := os.Getwd()
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(wd, "main.go"), true},
		{filepath.Join(wd, "sub/dir/file.go"), true},
		{filepath.Join(wd, "readme.md"), false},
		{filepath.Join(wd, "file.txt"), false},
	}

	for _, tt := range tests {
		if got := w.matches(tt.path); got != tt.want {
			t.Errorf("matches(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchesIgnorePatterns(t *testing.T) {
	w := &Watcher{
		logger:   NewLogger(),
		patterns: []string{"**/*.go"},
		ignores:  []string{"**/*_test.go"},
	}

	wd, _ := os.Getwd()
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(wd, "main.go"), true},
		{filepath.Join(wd, "main_test.go"), false},
		{filepath.Join(wd, "sub/handler_test.go"), false},
	}

	for _, tt := range tests {
		if got := w.matches(tt.path); got != tt.want {
			t.Errorf("matches(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchesWildcard(t *testing.T) {
	w := &Watcher{
		logger:   NewLogger(),
		patterns: []string{"**/*"},
		ignores:  nil,
	}

	wd, _ := os.Getwd()

	if !w.matches(filepath.Join(wd, "anything.txt")) {
		t.Error("**/* should match any file")
	}
}

func TestWatcherDetectsFileChange(t *testing.T) {
	dir := t.TempDir()

	w, err := NewWatcher(NewLogger(), []string{"**/*.txt"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.AddRecursive(dir); err != nil {
		t.Fatal(err)
	}

	events := w.Events()

	// Create a matching file
	time.Sleep(50 * time.Millisecond) // let watcher settle
	f := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(f, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case path := <-events:
		if filepath.Base(path) != "hello.txt" {
			t.Errorf("expected hello.txt, got %s", path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file change event")
	}
}

func TestWatcherIgnoresNonMatchingFile(t *testing.T) {
	dir := t.TempDir()

	w, err := NewWatcher(NewLogger(), []string{"**/*.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.AddRecursive(dir); err != nil {
		t.Fatal(err)
	}

	events := w.Events()

	// Create a non-matching file
	time.Sleep(50 * time.Millisecond)
	f := filepath.Join(dir, "readme.md")
	if err := os.WriteFile(f, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case path := <-events:
		t.Errorf("should not receive event for .md file, got %s", path)
	case <-time.After(300 * time.Millisecond):
		// Expected: no event
	}
}

func TestWatcherDetectsNewSubdirectory(t *testing.T) {
	dir := t.TempDir()

	w, err := NewWatcher(NewLogger(), []string{"**/*.txt"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.AddRecursive(dir); err != nil {
		t.Fatal(err)
	}

	events := w.Events()
	time.Sleep(50 * time.Millisecond)

	// Create subdirectory, then a file in it
	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0755)
	time.Sleep(100 * time.Millisecond) // let watcher pick up new dir

	f := filepath.Join(subdir, "new.txt")
	os.WriteFile(f, []byte("hi"), 0644)

	select {
	case path := <-events:
		if filepath.Base(path) != "new.txt" {
			t.Errorf("expected new.txt, got %s", path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event in new subdirectory")
	}
}
