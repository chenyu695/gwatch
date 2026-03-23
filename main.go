package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		watchDirs stringSlice
		patterns  stringSlice
		ignores   stringSlice
		delay     time.Duration
		cmd       string
		noStart   bool
	)

	flag.Var(&watchDirs, "w", "directory to watch (repeatable, default: .)")
	flag.Var(&patterns, "e", "glob pattern to match (repeatable, default: **/*)")
	flag.Var(&ignores, "i", "glob pattern to ignore (repeatable)")
	flag.DurationVar(&delay, "d", 300*time.Millisecond, "debounce delay")
	flag.StringVar(&cmd, "x", "", "command to execute on change")
	flag.BoolVar(&noStart, "n", false, "don't run command on startup")
	flag.Parse()

	// Command from remaining args (after --)
	if cmd == "" {
		if args := flag.Args(); len(args) > 0 {
			cmd = strings.Join(args, " ")
		}
	}
	if cmd == "" {
		fmt.Fprintln(os.Stderr, "Usage: gwatch [flags] -x \"cmd\"  or  gwatch [flags] -- cmd args")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(watchDirs) == 0 {
		watchDirs = stringSlice{"."}
	}
	if len(patterns) == 0 {
		patterns = stringSlice{"**/*"}
	}

	// Load .gwatchignore from each watch directory
	if fileIgnores := loadIgnoreFiles(watchDirs); len(fileIgnores) > 0 {
		ignores = append(ignores, fileIgnores...)
	}
	// Always ignore .gwatchignore itself
	ignores = append(ignores, ".gwatchignore")

	logger := NewLogger()
	logger.Info("Starting gwatch")
	logger.Info(fmt.Sprintf("Watching:  %s", strings.Join(watchDirs, ", ")))
	logger.Info(fmt.Sprintf("Patterns:  %s", strings.Join(patterns, ", ")))
	if len(ignores) > 0 {
		logger.Info(fmt.Sprintf("Ignoring:  %s", strings.Join(ignores, ", ")))
	}
	logger.Info(fmt.Sprintf("Command:   %s", cmd))
	logger.Info(fmt.Sprintf("Debounce:  %s", delay))

	watcher, err := NewWatcher(logger, patterns, ignores)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create watcher: %v", err))
		os.Exit(1)
	}
	defer watcher.Close()

	for _, dir := range watchDirs {
		if err := watcher.AddRecursive(dir); err != nil {
			logger.Error(fmt.Sprintf("Failed to watch %s: %v", dir, err))
			os.Exit(1)
		}
	}

	runner := NewRunner(logger)
	events := watcher.Events()

	// Debounced trigger
	trigger := Debounce(delay, func() {
		runner.Run(cmd)
	})

	// Run once on startup
	if !noStart {
		runner.Run(cmd)
	}

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case path, ok := <-events:
			if !ok {
				return
			}
			logger.Change(path)
			trigger()
		case <-sig:
			logger.Info("Shutting down")
			return
		}
	}
}
