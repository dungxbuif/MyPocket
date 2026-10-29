package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type AdvisorService struct {
	Store          repository.AdvisorRepository
	Orchestrator   *AdvisorOrchestrator
	Runtime        repository.AdvisorRuntimeRepository
	RuntimePolicy  entity.AdvisorRuntimePolicy
	RuntimeEnabled bool
	mu             sync.Mutex
	cancellations  map[string]context.CancelFunc
}

func NewAdvisorService(store repository.AdvisorRepository, orchestrator *AdvisorOrchestrator) *AdvisorService {
	return &AdvisorService{Store: store, Orchestrator: orchestrator, RuntimePolicy: entity.DefaultAdvisorRuntimePolicy(), cancellations: make(map[string]context.CancelFunc)}
}

// NewAdvisorServiceWithRuntime enables the fenced executor path when the
// deployment has migrated the runtime lease schema. The legacy constructor
// remains intentionally safe for in-memory adapters and older test stores.
func NewAdvisorServiceWithRuntime(store repository.AdvisorRepository, orchestrator *AdvisorOrchestrator, enabled bool) *AdvisorService {
	service := NewAdvisorService(store, orchestrator)
	service.RuntimeEnabled = enabled
	if enabled {
		if runtime, ok := store.(repository.AdvisorRuntimeRepository); ok {
			service.Runtime = runtime
		}
	}
	return service
}

// Enabled reports whether the service is wired and, when the provider exposes
// an availability probe, configured for use. It intentionally does not make a
// network call: capability discovery must be cheap and side-effect free.
func (s *AdvisorService) Enabled() bool {
	if s == nil || s.Store == nil || s.Orchestrator == nil || s.Orchestrator.Provider == nil || s.Orchestrator.Tools == nil || s.RuntimeEnabled && s.Runtime == nil {
		return false
	}
	if probe, ok := s.Orchestrator.Provider.(interface{ Available() bool }); ok {
		return probe.Available()
	}
	return true
}

// ResetConversation clears the owner's assistant history and interrupts any
// provider calls for runs that were purged by the same transaction. The
// repository remains the source of truth; cancellation is best-effort and
// process-local, while generation/lease checks reject late writes everywhere.
func (s *AdvisorService) ResetConversation(ctx context.Context, ownerID, requestID, payloadHash string) (*entity.AdvisorResetResult, error) {
	if s == nil || s.Store == nil || strings.TrimSpace(ownerID) == "" {
		return nil, repository.ErrNotFound
	}
	lifecycle, ok := s.Store.(repository.AdvisorLifecycleRepository)
	if !ok {
		return nil, ErrAdvisorProviderUnavailable
	}
	result, err := lifecycle.ResetConversation(ctx, ownerID, requestID, payloadHash)
	if err != nil || result == nil || len(result.PurgedRunIDs) == 0 {
		return result, err
	}

	cancels := make([]context.CancelFunc, 0, len(result.PurgedRunIDs))
	s.mu.Lock()
	for _, runID := range result.PurgedRunIDs {
		if cancel := s.cancellations[runID]; cancel != nil {
			cancels = append(cancels, cancel)
		}
		delete(s.cancellations, runID)
	}
	s.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	return result, nil
}

// Begin reserves a run and starts the bounded provider loop asynchronously.
// The returned run is visible to the caller before provider work completes, so
// the caller can cancel the first request through the run endpoint.
func (s *AdvisorService) Begin(ctx context.Context, principal Principal, input repository.AdvisorRunInput) (*entity.AdvisorRun, bool, error) {
	if s == nil || s.Store == nil || s.Orchestrator == nil || strings.TrimSpace(principal.OwnerID) == "" || !s.Enabled() {
		return nil, false, ErrAdvisorProviderUnavailable
	}
	terminalCtx := context.WithoutCancel(ctx)
	if s.Orchestrator.Validator != nil {
		if err := s.Orchestrator.Validator.Validate(ctx, principal); err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
		}
	}
	input.OwnerID = principal.OwnerID
	run, userMessage, replay, err := s.Store.StartRun(ctx, input)
	if err != nil {
		return nil, false, err
	}
	if replay {
		return run, true, nil
	}
	runCtx, cancel := context.WithCancel(terminalCtx)
	if s.runtimeActive() {
		var lease entity.AdvisorRunLease
		lease, err = s.claimRuntimeRun(terminalCtx, principal.OwnerID, run.ID)
		if err != nil {
			cancel()
			return nil, false, err
		}
		run.Status = entity.AdvisorStatusRunning
		runtimeRunCtx, runtimeCancel := s.runtimeContext(terminalCtx)
		runCtx, cancel = runtimeRunCtx, runtimeCancel
		s.registerCancellation(run.ID, cancel)
		stopWatch := s.watchRunCancellation(terminalCtx, principal.OwnerID, run.ID, cancel)
		go func() {
			defer stopWatch()
			defer s.unregisterCancellation(run.ID)
			_, _ = s.completeRunWithLease(runCtx, terminalCtx, principal, run, userMessage, lease)
		}()
		return run, false, nil
	}
	s.registerCancellation(run.ID, cancel)
	if err := s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusRunning, s.now()); err != nil {
		s.unregisterCancellation(run.ID)
		cancel()
		return nil, false, err
	}
	run.Status = entity.AdvisorStatusRunning
	stopWatch := s.watchRunCancellation(terminalCtx, principal.OwnerID, run.ID, cancel)
	go func() {
		defer stopWatch()
		defer s.unregisterCancellation(run.ID)
		if _, err := s.completeRun(runCtx, terminalCtx, principal, run, userMessage); err != nil {
			_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		}
	}()
	return run, false, nil
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
	runCtx, cancel := context.WithCancel(terminalCtx)
	if s.runtimeActive() {
		lease, claimErr := s.claimRuntimeRun(terminalCtx, principal.OwnerID, run.ID)
		if claimErr != nil {
			cancel()
			return AdvisorAnswer{}, claimErr
		}
		runtimeRunCtx, runtimeCancel := s.runtimeContext(terminalCtx)
		defer runtimeCancel()
		runCtx, cancel = runtimeRunCtx, runtimeCancel
		s.registerCancellation(run.ID, cancel)
		defer s.unregisterCancellation(run.ID)
		stopWatch := s.watchRunCancellation(terminalCtx, principal.OwnerID, run.ID, cancel)
		defer stopWatch()
		return s.completeRunWithLease(runCtx, terminalCtx, principal, run, userMessage, lease)
	}
	s.registerCancellation(run.ID, cancel)
	defer s.unregisterCancellation(run.ID)
	if err := s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusRunning, s.now()); err != nil {
		return AdvisorAnswer{}, err
	}
	stopWatch := s.watchRunCancellation(terminalCtx, principal.OwnerID, run.ID, cancel)
	defer stopWatch()
	return s.completeRun(runCtx, terminalCtx, principal, run, userMessage)
}

func (s *AdvisorService) runtimeActive() bool {
	return s != nil && s.RuntimeEnabled && s.Runtime != nil
}

func (s *AdvisorService) claimRuntimeRun(ctx context.Context, ownerID, runID string) (entity.AdvisorRunLease, error) {
	return s.Runtime.ClaimRun(ctx, ownerID, runID, uuid.NewString())
}

func (s *AdvisorService) runtimeContext(parent context.Context) (context.Context, context.CancelFunc) {
	policy := s.RuntimePolicy
	if policy.RunDeadline <= 0 {
		policy = entity.DefaultAdvisorRuntimePolicy()
	}
	return context.WithTimeout(context.WithoutCancel(parent), policy.RunDeadline)
}

func (s *AdvisorService) completeRunWithLease(runCtx, terminalCtx context.Context, principal Principal, run *entity.AdvisorRun, userMessage *entity.AdvisorMessage, lease entity.AdvisorRunLease) (AdvisorAnswer, error) {
	providerCtx, cancelProvider := context.WithCancel(runCtx)
	defer cancelProvider()
	stopHeartbeat := s.startRuntimeHeartbeat(providerCtx, lease, cancelProvider)
	defer stopHeartbeat()
	answer, err := s.answerRun(providerCtx, terminalCtx, principal, run, userMessage)
	if err != nil {
		_ = s.Runtime.FinishRun(terminalCtx, lease, entity.AdvisorStatusFailed, advisorErrorCode(err), answer.Usage)
		return AdvisorAnswer{}, err
	}
	parts := BuildAdvisorPartsForRun(answer, run.ConversationID, run.Generation, run.ID)
	previous, err := s.loadContextState(providerCtx, principal, run)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	state, err := BuildAdvisorContextState(previous, run.Generation, userMessage.Seq, answer.Results)
	if err != nil {
		_ = s.Runtime.FinishRun(terminalCtx, lease, entity.AdvisorStatusFailed, "context_state_invalid", answer.Usage)
		return AdvisorAnswer{}, err
	}
	completion := entity.AdvisorCompletion{Parts: parts, Usage: answer.Usage, ContextState: map[string]any{}}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	if err := json.Unmarshal(stateJSON, &completion.ContextState); err != nil {
		return AdvisorAnswer{}, err
	}
	if bundle := advisorFactBundle(answer.Results); bundle != nil {
		completion.FactBundle = bundle
	}
	if _, err := s.Runtime.CompleteRun(terminalCtx, lease, completion); err != nil {
		return AdvisorAnswer{}, err
	}
	return answer, nil
}

func (s *AdvisorService) answerRun(runCtx, terminalCtx context.Context, principal Principal, run *entity.AdvisorRun, userMessage *entity.AdvisorMessage) (AdvisorAnswer, error) {
	messages, err := s.Store.ListMessages(runCtx, principal.OwnerID, run.ConversationID, 0, 50)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	state, err := s.loadContextState(runCtx, principal, run)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	chat, err := BuildAdvisorRequestContext(messages, state)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	if len(chat) == 0 {
		chat = []AdvisorChatMessage{{Role: entity.AdvisorRoleUser, Content: messageText(*userMessage)}}
	}
	answer, err := s.Orchestrator.Answer(runCtx, principal, chat)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	if s.Orchestrator.Validator != nil {
		if err := s.Orchestrator.Validator.Validate(terminalCtx, principal); err != nil {
			return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
		}
	}
	return answer, nil
}

func (s *AdvisorService) loadContextState(ctx context.Context, principal Principal, run *entity.AdvisorRun) (entity.AdvisorContextState, error) {
	state := entity.AdvisorContextState{Generation: run.Generation}
	if store, ok := s.Store.(repository.AdvisorContextRepository); ok && s.runtimeActive() {
		loaded, err := store.LoadAdvisorContextState(ctx, principal.OwnerID, run.ConversationID, run.Generation)
		if err != nil {
			return state, err
		}
		if loaded.Generation == run.Generation && loaded.Version == 1 {
			return revalidateAdvisorContext(ctx, principal, s.Orchestrator.Tools, loaded)
		}
	}
	return state, nil
}

func (s *AdvisorService) startRuntimeHeartbeat(ctx context.Context, lease entity.AdvisorRunLease, cancelProvider context.CancelFunc) func() {
	interval := s.RuntimePolicy.HeartbeatEvery
	if interval <= 0 {
		interval = entity.DefaultAdvisorRuntimePolicy().HeartbeatEvery
	}
	heartbeatCtx, stop := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				dbTimeout := s.RuntimePolicy.DBOperation
				if dbTimeout <= 0 {
					dbTimeout = entity.DefaultAdvisorRuntimePolicy().DBOperation
				}
				probeCtx, cancel := context.WithTimeout(heartbeatCtx, dbTimeout)
				if err := s.Runtime.HeartbeatRun(probeCtx, lease); err != nil {
					cancel()
					cancelProvider()
					stop()
					return
				}
				cancel()
			}
		}
	}()
	return stop
}

func advisorFactBundle(results []entity.FinanceResult) *entity.AdvisorFactBundle {
	if len(results) == 0 {
		return nil
	}
	facts := make([]entity.Fact, 0)
	refs := make([]string, 0, len(results))
	for _, result := range results {
		facts = append(facts, result.Facts...)
		if result.QueryKey != "" {
			refs = append(refs, result.QueryKey)
		}
	}
	return &entity.AdvisorFactBundle{Scope: map[string]any{"query_refs": refs}, AsOf: time.Now().UTC(), Facts: facts}
}

func advisorErrorCode(err error) string {
	if err == nil {
		return ""
	}
	code := strings.ReplaceAll(strings.ToLower(err.Error()), " ", "_")
	runes := []rune(code)
	if len(runes) > 120 {
		runes = runes[:120]
	}
	return string(runes)
}

func (s *AdvisorService) completeRun(runCtx, terminalCtx context.Context, principal Principal, run *entity.AdvisorRun, userMessage *entity.AdvisorMessage) (AdvisorAnswer, error) {
	messages, err := s.Store.ListMessages(runCtx, principal.OwnerID, run.ConversationID, 0, 50)
	if err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	chat, err := BuildAdvisorRequestContext(messages, entity.AdvisorContextState{Generation: run.Generation})
	if err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if len(chat) == 0 {
		chat = []AdvisorChatMessage{{Role: entity.AdvisorRoleUser, Content: messageText(*userMessage)}}
	}
	answer, err := s.Orchestrator.Answer(runCtx, principal, chat)
	if err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if s.Orchestrator.Validator != nil {
		if err := s.Orchestrator.Validator.Validate(terminalCtx, principal); err != nil {
			_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
			return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
		}
	}
	parts := BuildAdvisorPartsForRun(answer, run.ConversationID, run.Generation, run.ID)
	if _, err := s.Store.CreateAssistantMessage(runCtx, principal.OwnerID, run.ID, run.Generation, parts); err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if _, err := s.Store.AppendEvent(runCtx, principal.OwnerID, run.ID, "run.completed", map[string]any{"text": answer.Text}); err != nil {
		_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
		return AdvisorAnswer{}, err
	}
	if answer.Usage != nil {
		if store, ok := s.Store.(interface {
			StoreRunUsage(context.Context, string, string, map[string]any) error
		}); ok {
			if err := store.StoreRunUsage(terminalCtx, principal.OwnerID, run.ID, answer.Usage); err != nil {
				_ = s.Store.MarkRunStatus(terminalCtx, principal.OwnerID, run.ID, entity.AdvisorStatusFailed, s.now())
				return AdvisorAnswer{}, err
			}
		}
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

// watchRunCancellation bridges the persisted lifecycle state across API
// processes. The local cancellation map remains the fast path; this watcher
// is the durable fallback for a cancel request handled by another process.
// The probe is intentionally bounded and non-blocking with respect to run
// locks, and it stops as soon as the provider loop returns.
func (s *AdvisorService) watchRunCancellation(ctx context.Context, ownerID, runID string, cancel context.CancelFunc) func() {
	statusStore, ok := s.Store.(repository.AdvisorRunStatusRepository)
	if !ok {
		return func() {}
	}
	watchCtx, stop := context.WithCancel(ctx)
	var once sync.Once
	stopOnce := func() { once.Do(stop) }
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
				probeCtx, probeCancel := context.WithTimeout(watchCtx, time.Second)
				status, err := statusStore.GetRunStatus(probeCtx, ownerID, runID)
				probeCancel()
				if err == nil && (status == entity.AdvisorStatusCancelled || status == entity.AdvisorStatusInterrupted || status == entity.AdvisorStatusPurged) {
					cancel()
					return
				}
			}
		}
	}()
	return stopOnce
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
