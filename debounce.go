package main

import (
	"sync"
	"time"
)

// Debounce returns a function that delays invoking fn until after delay has
// elapsed since the last call. Each call resets the timer.
func Debounce(delay time.Duration, fn func()) func() {
	var timer *time.Timer
	var mu sync.Mutex
	return func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, fn)
	}
}
