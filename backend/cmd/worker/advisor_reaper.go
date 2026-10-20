package main

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/repository"
)

// startAdvisorReaper runs independently from recurring/budget passes so a
// blocked business job cannot leave provider runs permanently active.
func startAdvisorReaper(ctx context.Context, store repository.AdvisorRuntimeRepository) func() {
	return startAdvisorReaperWithInterval(ctx, store, 15*time.Second)
}

func startAdvisorReaperWithInterval(ctx context.Context, store repository.AdvisorRuntimeRepository, interval time.Duration) func() {
	if ctx == nil || store == nil || interval <= 0 {
		return func() {}
	}
	loopCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		reap := func() {
			probeCtx, probeCancel := context.WithTimeout(loopCtx, 5*time.Second)
			_, _ = store.ReapExpiredRuns(probeCtx, 100)
			probeCancel()
		}
		for {
			select {
			case <-loopCtx.Done():
				return
			case <-ticker.C:
				reap()
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}
