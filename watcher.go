package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors directories for file changes, applying glob filtering.
type Watcher struct {
	logger   *Logger
	fw       *fsnotify.Watcher
	patterns []string
	ignores  []string
}

func NewWatcher(logger *Logger, patterns, ignores []string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		logger:   logger,
		fw:       fw,
		patterns: patterns,
		ignores:  ignores,
	}, nil
}

// AddRecursive registers dir and all its subdirectories with fsnotify.
func (w *Watcher) AddRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip hidden directories (e.g. .git)
			if base := filepath.Base(path); len(base) > 1 && base[0] == '.' {
				return filepath.SkipDir
			}
			if err := w.fw.Add(path); err != nil {
				w.logger.Warn(fmt.Sprintf("Cannot watch %s: %v", path, err))
			}
		}
		return nil
	})
}

// Events returns a read-only channel that emits changed file paths.
// Glob filtering is applied; new directories are automatically watched.
func (w *Watcher) Events() <-chan string {
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		for {
			select {
			case event, ok := <-w.fw.Events:
				if !ok {
					return
				}
				// Watch newly created directories
				if event.Has(fsnotify.Create) {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						_ = w.AddRecursive(event.Name)
					}
				}
				if !w.matches(event.Name) {
					continue
				}
				// Non-blocking send
				select {
				case ch <- event.Name:
				default:
				}
			case err, ok := <-w.fw.Errors:
				if !ok {
					return
				}
				w.logger.Error(fmt.Sprintf("Watcher error: %v", err))
			}
		}
	}()
	return ch
}

func (w *Watcher) Close() error {
	return w.fw.Close()
}

// matches returns true if path matches any include pattern and no ignore pattern.
func (w *Watcher) matches(path string) bool {
	rel := path
	// Normalize to relative path for consistent glob matching
	if abs, err := filepath.Abs(path); err == nil {
		if wd, err := os.Getwd(); err == nil {
			if r, err := filepath.Rel(wd, abs); err == nil {
				rel = r
			}
		}
	}
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)

	// Check ignore patterns first
	for _, pat := range w.ignores {
		if matched, _ := doublestar.Match(pat, rel); matched {
			return false
		}
		if matched, _ := doublestar.Match(pat, base); matched {
			return false
		}
	}

	// Check include patterns
	for _, pat := range w.patterns {
		if matched, _ := doublestar.Match(pat, rel); matched {
			return true
		}
		if matched, _ := doublestar.Match(pat, base); matched {
			return true
		}
	}
	return false
}
