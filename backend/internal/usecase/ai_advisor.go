package usecase

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type AdvisorService struct {
	Store         repository.AdvisorRepository
	Orchestrator  *AdvisorOrchestrator
	mu            sync.Mutex
	cancellations map[string]context.CancelFunc
}

func NewAdvisorService(store repository.AdvisorRepository, orchestrator *AdvisorOrchestrator) *AdvisorService {
	return &AdvisorService{Store: store, Orchestrator: orchestrator, cancellations: make(map[string]context.CancelFunc)}
}

// Enabled reports whether the service is wired and, when the provider exposes
// an availability probe, configured for use. It intentionally does not make a
// network call: capability discovery must be cheap and side-effect free.
func (s *AdvisorService) Enabled() bool {
	if s == nil || s.Store == nil || s.Orchestrator == nil || s.Orchestrator.Provider == nil || s.Orchestrator.Tools == nil {
		return false
	}
	if probe, ok := s.Orchestrator.Provider.(interface{ Available() bool }); ok {
		return probe.Available()
	}
	return true
}

func (s *AdvisorService) Submit(ctx context.Context, principal Principal, input repository.AdvisorRunInput) (AdvisorAnswer, error) {
	if s == nil || s.Store == nil || s.Orchestrator == nil || strings.TrimSpace(principal.OwnerID) == "" || !s.Enabled() {
		return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
	}
	terminalCtx := context.WithoutCancel(ctx)
	if s.Orchestrator.Validator != nil {
		if err := s.Orchestrator.Validator.Validate(ctx, principal); err != nil {
			return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
		}
	}
	input.OwnerID = principal.OwnerID
	run, userMessage, replay, err := s.Store.StartRun(ctx, input)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	if replay {
		if run.Status == entity.AdvisorStatusCompleted {
			return AdvisorAnswer{Text: "Yêu cầu này đã được xử lý trước đó."}, nil
		}
		return AdvisorAnswer{}, repository.ErrAdvisorBusy
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.registerCancellation(run.ID, cancel)
	defer s.unregisterCancellation(run.ID)
	if err := s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusRunning, s.now()); err != nil {
		return AdvisorAnswer{}, err
	}
	messages, err := s.Store.ListMessages(runCtx, principal.OwnerID, run.ConversationID, 0, 50)
	if err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	chat := BuildAdvisorContext(messages)
	if len(chat) == 0 {
		chat = []AdvisorChatMessage{{Role: entity.AdvisorRoleUser, Content: messageText(*userMessage)}}
	}
	answer, err := s.Orchestrator.Answer(runCtx, principal, chat)
	if err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if s.Orchestrator.Validator != nil {
		if err := s.Orchestrator.Validator.Validate(ctx, principal); err != nil {
			_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
			return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
		}
	}
	parts := BuildAdvisorParts(answer)
	if _, err := s.Store.CreateAssistantMessage(runCtx, principal.OwnerID, run.ID, run.Generation, parts); err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if _, err := s.Store.AppendEvent(runCtx, principal.OwnerID, run.ID, "run.completed", map[string]any{"text": answer.Text}); err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if err := s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusCompleted, s.now()); err != nil {
		return AdvisorAnswer{}, err
	}
	return answer, nil
}

// Cancel marks a run cancelled and interrupts the in-process provider call.
// The repository remains the source of truth, while the map only provides the
// best-effort immediate cancellation for work handled by this process.
func (s *AdvisorService) Cancel(ctx context.Context, ownerID, runID string) error {
	if s == nil || s.Store == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(runID) == "" {
		return repository.ErrNotFound
	}
	if err := s.Store.MarkRunStatus(ctx, ownerID, runID, entity.AdvisorStatusCancelled, s.now()); err != nil {
		return err
	}
	s.mu.Lock()
	cancel := s.cancellations[runID]
	delete(s.cancellations, runID)
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (s *AdvisorService) registerCancellation(runID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancellations[runID] = cancel
}

func (s *AdvisorService) unregisterCancellation(runID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancellations, runID)
}

func (s *AdvisorService) now() time.Time { return time.Now().UTC() }

func messageText(message entity.AdvisorMessage) string {
	for _, part := range message.Parts {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
			return part.Text
		}
	}
	return ""
}
