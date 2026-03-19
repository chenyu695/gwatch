package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounceSingleCall(t *testing.T) {
	var count atomic.Int32
	trigger := Debounce(50*time.Millisecond, func() {
		count.Add(1)
	})

	trigger()
	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call, got %d", got)
	}
}

func TestDebounceRapidCalls(t *testing.T) {
	var count atomic.Int32
	trigger := Debounce(50*time.Millisecond, func() {
		count.Add(1)
	})

	// Fire 10 times rapidly — should only trigger once
	for i := 0; i < 10; i++ {
		trigger()
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call after rapid triggers, got %d", got)
	}
}

func TestDebounceMultipleBursts(t *testing.T) {
	var count atomic.Int32
	trigger := Debounce(50*time.Millisecond, func() {
		count.Add(1)
	})

	// Burst 1
	trigger()
	trigger()
	trigger()
	time.Sleep(100 * time.Millisecond)

	// Burst 2
	trigger()
	trigger()
	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 2 {
		t.Errorf("expected 2 calls for 2 bursts, got %d", got)
	}
}

func TestDebounceDelayRespected(t *testing.T) {
	var count atomic.Int32
	trigger := Debounce(80*time.Millisecond, func() {
		count.Add(1)
	})

	trigger()
	time.Sleep(40 * time.Millisecond)

	// Should not have fired yet
	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 calls before delay, got %d", got)
	}

	time.Sleep(80 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call after delay, got %d", got)
	}
}
