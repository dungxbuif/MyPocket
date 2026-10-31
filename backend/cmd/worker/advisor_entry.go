package main

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type advisorEntryPassRunner interface {
	Run(context.Context, int) (int, error)
}

type advisorEntryWorker struct {
	Tasks   repository.AdvisorEntryTaskRepository
	Service *usecase.AdvisorEntryService
}

func (w *advisorEntryWorker) Run(ctx context.Context, limit int) (int, error) {
	if w == nil || w.Tasks == nil || w.Service == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	processed := 0
	for processed < limit {
		lease, err := w.Tasks.ClaimNextTask(ctx, uuid.NewString())
		if errors.Is(err, repository.ErrNotFound) {
			return processed, nil
		}
		if err != nil {
			return processed, err
		}
		workCtx, cancel := context.WithCancel(ctx)
		heartbeatDone := make(chan struct{})
		go w.heartbeat(workCtx, lease, cancel, heartbeatDone)
		processErr := w.Service.ProcessClaimed(workCtx, lease)
		cancel()
		<-heartbeatDone
		if processErr != nil {
			// ProcessClaimed records a fenced failure when possible. A lease lost
			// during recovery is expected after another worker/reaper wins.
			if errors.Is(processErr, repository.ErrAdvisorEntryLeaseLost) {
				continue
			}
			return processed, processErr
		}
		processed++
	}
	return processed, nil
}

func (w *advisorEntryWorker) heartbeat(ctx context.Context, lease repository.AdvisorEntryTaskLease, cancel context.CancelFunc, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			heartbeatCtx, heartbeatCancel := context.WithTimeout(ctx, 5*time.Second)
			err := w.Tasks.HeartbeatTask(heartbeatCtx, lease)
			heartbeatCancel()
			if err != nil {
				cancel()
				return
			}
		}
	}
}

func runAdvisorEntryPass(ctx context.Context, runner advisorEntryPassRunner, limit int) (int, error) {
	if runner == nil {
		return 0, nil
	}
	return runner.Run(ctx, limit)
}
