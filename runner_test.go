package main

import (
	"testing"
	"time"
)

func TestRunnerExecutesCommand(t *testing.T) {
	r := NewRunner(NewLogger())
	r.Run("echo runner_test_ok")

	r.mu.Lock()
	done := r.done
	r.mu.Unlock()

	select {
	case <-done:
		// Process completed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for process to finish")
	}
}

func TestRunnerKillsPreviousProcess(t *testing.T) {
	r := NewRunner(NewLogger())

	// Start a long-running process
	r.Run("sleep 60")
	time.Sleep(100 * time.Millisecond)

	r.mu.Lock()
	firstCmd := r.cmd
	r.mu.Unlock()

	if firstCmd == nil {
		t.Fatal("expected first process to be running")
	}
	firstPid := firstCmd.Process.Pid

	// Run again — should kill the first
	r.Run("echo replaced")
	time.Sleep(200 * time.Millisecond)

	// Verify the first process was killed (Wait should have returned)
	// ProcessState is set after Wait completes
	if firstCmd.ProcessState == nil {
		t.Error("expected first process to have exited")
	}
	_ = firstPid
}

func TestRunnerHandlesFailingCommand(t *testing.T) {
	r := NewRunner(NewLogger())
	r.Run("exit 42")

	// Wait for the process to finish
	time.Sleep(200 * time.Millisecond)

	// Should not panic or deadlock — just log the exit code
	r.mu.Lock()
	done := r.done
	r.mu.Unlock()

	if done != nil {
		select {
		case <-done:
			// good, process completed
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for failed process to finish")
		}
	}
}

func TestRunnerSequentialRuns(t *testing.T) {
	r := NewRunner(NewLogger())

	for i := 0; i < 5; i++ {
		r.Run("echo run")
	}

	time.Sleep(300 * time.Millisecond)

	r.mu.Lock()
	done := r.done
	r.mu.Unlock()

	if done != nil {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for last run to finish")
		}
	}
}
