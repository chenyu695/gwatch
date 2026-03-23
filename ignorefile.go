package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// parseIgnoreFile reads a .gwatchignore file and returns the glob patterns.
// Empty lines and lines starting with # are skipped.
func parseIgnoreFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// loadIgnoreFiles reads .gwatchignore from each directory and returns the
// combined ignore patterns. Directories without a .gwatchignore are skipped.
func loadIgnoreFiles(dirs []string) []string {
	var patterns []string
	for _, dir := range dirs {
		p, err := parseIgnoreFile(filepath.Join(dir, ".gwatchignore"))
		if err != nil {
			continue // file doesn't exist or unreadable, skip
		}
		patterns = append(patterns, p...)
	}
	return patterns
}
