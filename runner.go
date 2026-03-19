package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// Runner executes shell commands, killing any previously running process first.
type Runner struct {
	logger *Logger
	mu     sync.Mutex
	cmd    *exec.Cmd
	done   chan struct{} // closed when current process exits
}

func NewRunner(logger *Logger) *Runner {
	return &Runner{logger: logger}
}

// Run kills the previous process group (if any) and starts cmd via sh -c.
func (r *Runner) Run(cmd string) {
	r.mu.Lock()
	if r.cmd != nil && r.cmd.Process != nil {
		r.logger.Warn("Killing previous process...")
		// Kill entire process group (negative pid)
		_ = syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
	}
	prevDone := r.done
	r.mu.Unlock()

	// Wait for previous goroutine to finish its Wait() call
	if prevDone != nil {
		<-prevDone
	}

	r.logger.Exec(fmt.Sprintf("$ %s", cmd))

	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// Create a new process group so we can kill the whole tree
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := c.Start(); err != nil {
		r.logger.Error(fmt.Sprintf("Failed to start: %v", err))
		return
	}

	done := make(chan struct{})
	r.mu.Lock()
	r.cmd = c
	r.done = done
	r.mu.Unlock()

	go func() {
		defer close(done)
		err := c.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// ExitCode -1 means killed by signal (expected on restart)
				if exitErr.ExitCode() != -1 {
					r.logger.Warn(fmt.Sprintf("Exited with code %d", exitErr.ExitCode()))
				}
			}
		} else {
			r.logger.Info("Process finished")
		}
	}()
}
