package main

import (
	"os"
	"testing"
	"time"
)

func TestWorkerIntervalDefaultsAndOverrides(t *testing.T) {
	t.Setenv("WORKER_INTERVAL_SECONDS", "")
	if got := workerInterval(); got != 60*time.Second {
		t.Fatalf("default interval=%s", got)
	}
	t.Setenv("WORKER_INTERVAL_SECONDS", "15")
	if got := workerInterval(); got != 15*time.Second {
		t.Fatalf("override interval=%s", got)
	}
	_ = os.Unsetenv("WORKER_INTERVAL_SECONDS")
}

func TestWorkerOnceFlag(t *testing.T) {
	t.Setenv("WORKER_ONCE", "true")
	if !workerOnce() {
		t.Fatal("expected once mode")
	}
	t.Setenv("WORKER_ONCE", "no")
	if workerOnce() {
		t.Fatal("expected loop mode")
	}
}
