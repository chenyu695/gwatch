package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout captures stdout output from fn.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestLoggerInfo(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Info("hello")
	})
	if !strings.Contains(output, "[gwatch]") {
		t.Errorf("Info output missing [gwatch] tag: %q", output)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("Info output missing message: %q", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Warn("caution")
	})
	if !strings.Contains(output, "[warn]") {
		t.Errorf("Warn output missing [warn] tag: %q", output)
	}
}

func TestLoggerError(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Error("oops")
	})
	if !strings.Contains(output, "[error]") {
		t.Errorf("Error output missing [error] tag: %q", output)
	}
}

func TestLoggerChange(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Change("file.go")
	})
	if !strings.Contains(output, "[change]") {
		t.Errorf("Change output missing [change] tag: %q", output)
	}
}

func TestLoggerExec(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Exec("$ go build")
	})
	if !strings.Contains(output, "[exec]") {
		t.Errorf("Exec output missing [exec] tag: %q", output)
	}
	if !strings.Contains(output, "go build") {
		t.Errorf("Exec output missing command: %q", output)
	}
}

func TestLoggerTimestamp(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Info("ts")
	})
	// Should contain a time pattern like [HH:MM:SS]
	if !strings.Contains(output, ":") {
		t.Errorf("output missing timestamp: %q", output)
	}
}

func TestLoggerColors(t *testing.T) {
	output := captureStdout(func() {
		NewLogger().Info("color test")
	})
	// Should contain ANSI escape codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("output missing ANSI color codes: %q", output)
	}
	_ = fmt.Sprintf // suppress unused import if needed
}
