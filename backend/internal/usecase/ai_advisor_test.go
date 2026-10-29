package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type blockingAdvisorProvider struct {
	started    chan struct{}
	cancelled  chan struct{}
	startOnce  sync.Once
	cancelOnce sync.Once
}

func (p *blockingAdvisorProvider) Chat(ctx context.Context, _ AdvisorRequest) (AdvisorResponse, error) {
	p.startOnce.Do(func() { close(p.started) })
	<-ctx.Done()
	p.cancelOnce.Do(func() { close(p.cancelled) })
	return AdvisorResponse{}, ctx.Err()
}

type cancellableAdvisorStore struct {
	mu      sync.Mutex
	run     *entity.AdvisorRun
	message *entity.AdvisorMessage
	status  string
}

// distributedCancelStore models two API processes sharing the same run
// repository. The runner service has no process-local cancellation function in
// the canceller service, so the test proves cancellation is observed through
// the persisted run state.
type distributedCancelStore struct{ cancellableAdvisorStore }

func (s *distributedCancelStore) GetRunStatus(_ context.Context, _, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == "" {
		return entity.AdvisorStatusQueued, nil
	}
	return s.status, nil
}

func (s *cancellableAdvisorStore) StartRun(context.Context, repository.AdvisorRunInput) (*entity.AdvisorRun, *entity.AdvisorMessage, bool, error) {
	return s.run, s.message, false, nil
}
func (s *cancellableAdvisorStore) MarkRunStatus(_ context.Context, _, _ string, status string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == entity.AdvisorStatusCancelled {
		return repository.ErrAdvisorLeaseLost
	}
	s.status = status
	return nil
}
func (s *cancellableAdvisorStore) GetRun(context.Context, string, string) (*entity.AdvisorRun, error) {
	return s.run, nil
}
func (s *cancellableAdvisorStore) GetRunByRequest(context.Context, string, string) (*entity.AdvisorRun, error) {
	return s.run, nil
}
func (s *cancellableAdvisorStore) ListMessages(context.Context, string, string, int64, int) ([]entity.AdvisorMessage, error) {
	return []entity.AdvisorMessage{*s.message}, nil
}
func (s *cancellableAdvisorStore) AppendEvent(context.Context, string, string, string, any) (*entity.AdvisorEvent, error) {
	return &entity.AdvisorEvent{}, nil
}
func (s *cancellableAdvisorStore) CreateAssistantMessage(context.Context, string, string, int64, []entity.AdvisorPart) (*entity.AdvisorMessage, error) {
	return &entity.AdvisorMessage{}, nil
}

type advisorStoreStub struct {
	run       *entity.AdvisorRun
	message   *entity.AdvisorMessage
	messages  []entity.AdvisorMessage
	assistant *entity.AdvisorMessage
	status    string
}

type advisorRuntimeStub struct {
	*advisorStoreStub
	lease        entity.AdvisorRunLease
	completed    int
	finished     int
	heartbeatErr error
}

func (s *advisorRuntimeStub) ClaimRun(_ context.Context, ownerID, runID, token string) (entity.AdvisorRunLease, error) {
	s.run.Status = entity.AdvisorStatusRunning
	s.lease = entity.AdvisorRunLease{OwnerID: ownerID, RunID: runID, Generation: s.run.Generation, Token: token, DeadlineAt: time.Now().Add(time.Minute), LeaseUntil: time.Now().Add(time.Minute)}
	return s.lease, nil
}
func (s *advisorRuntimeStub) HeartbeatRun(context.Context, entity.AdvisorRunLease) error {
	return s.heartbeatErr
}
func (s *advisorRuntimeStub) CompleteRun(_ context.Context, _ entity.AdvisorRunLease, completion entity.AdvisorCompletion) (entity.AdvisorCompletionResult, error) {
	s.completed++
	s.run.Status = entity.AdvisorStatusCompleted
	s.assistant = &entity.AdvisorMessage{Parts: completion.Parts}
	return entity.AdvisorCompletionResult{Message: s.assistant, Run: s.run}, nil
}
func (s *advisorRuntimeStub) FinishRun(context.Context, entity.AdvisorRunLease, string, string, map[string]any) error {
	s.finished++
	s.run.Status = entity.AdvisorStatusFailed
	return nil
}
func (s *advisorRuntimeStub) ReapExpiredRuns(context.Context, int) (int, error) { return 0, nil }

type advisorResetStore struct {
	advisorStoreStub
	result *entity.AdvisorResetResult
	err    error
}

func (s *advisorResetStore) ResetConversation(context.Context, string, string, string) (*entity.AdvisorResetResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := *s.result
	// A real repository only reports purged runs for the first reset. This
	// fixture mirrors replay semantics so the service test proves that a
	// replay cannot cancel a newer provider run with a reused ledger result.
	s.result = &entity.AdvisorResetResult{ConversationID: result.ConversationID, Generation: result.Generation, ServerTime: result.ServerTime, Replayed: true}
	return &result, nil
}

type advisorAvailabilityProvider struct{ available bool }

func (p advisorAvailabilityProvider) Available() bool { return p.available }
func (p advisorAvailabilityProvider) Chat(context.Context, AdvisorRequest) (AdvisorResponse, error) {
	return AdvisorResponse{Text: "ok"}, nil
}

func TestAdvisorServiceEnabledReflectsProviderConfiguration(t *testing.T) {
	store := &advisorStoreStub{}
	tools := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
	if NewAdvisorService(store, NewAdvisorOrchestrator(advisorAvailabilityProvider{available: false}, tools)).Enabled() {
		t.Fatal("unconfigured provider must disable capabilities")
	}
	if !NewAdvisorService(store, NewAdvisorOrchestrator(advisorAvailabilityProvider{available: true}, tools)).Enabled() {
		t.Fatal("configured provider should enable capabilities")
	}
}

func TestAdvisorServiceDoesNotStartRunWhenProviderIsDisabled(t *testing.T) {
	store := &advisorStoreStub{run: &entity.AdvisorRun{ID: "run-1"}, message: &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}}}}
	service := NewAdvisorService(store, NewAdvisorOrchestrator(advisorAvailabilityProvider{available: false}, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	if _, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1"}, repository.AdvisorRunInput{RequestID: "req-1", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}}}); !errors.Is(err, ErrAdvisorProviderUnavailable) {
		t.Fatalf("disabled provider should fail before persistence, got %v", err)
	}
	if store.status != "" {
		t.Fatalf("disabled provider must not start a run: %q", store.status)
	}
}

func (s *advisorStoreStub) StartRun(_ context.Context, _ repository.AdvisorRunInput) (*entity.AdvisorRun, *entity.AdvisorMessage, bool, error) {
	return s.run, s.message, false, nil
}
func (s *advisorStoreStub) MarkRunStatus(_ context.Context, _, _ string, status string, _ time.Time) error {
	s.status = status
	return nil
}
func (s *advisorStoreStub) GetRun(context.Context, string, string) (*entity.AdvisorRun, error) {
	return s.run, nil
}
func (s *advisorStoreStub) GetRunByRequest(context.Context, string, string) (*entity.AdvisorRun, error) {
	return s.run, nil
}
func (s *advisorStoreStub) ListMessages(context.Context, string, string, int64, int) ([]entity.AdvisorMessage, error) {
	if s.messages != nil {
		return s.messages, nil
	}
	return []entity.AdvisorMessage{*s.message}, nil
}
func (s *advisorStoreStub) AppendEvent(context.Context, string, string, string, any) (*entity.AdvisorEvent, error) {
	return &entity.AdvisorEvent{}, nil
}
func (s *advisorStoreStub) CreateAssistantMessage(_ context.Context, _, _ string, _ int64, parts []entity.AdvisorPart) (*entity.AdvisorMessage, error) {
	s.assistant = &entity.AdvisorMessage{Parts: parts}
	return s.assistant, nil
}

func TestAdvisorServiceSubmitPersistsCompletionAndAssistantMessage(t *testing.T) {
	store := &advisorStoreStub{run: &entity.AdvisorRun{ID: "run-1", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued}, message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "hello"}}}}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{Text: "grounded answer"}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
	service := NewAdvisorService(store, NewAdvisorOrchestrator(provider, registry))
	result, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{OwnerID: "owner-1", RequestID: "req-1", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "grounded answer" || store.status != entity.AdvisorStatusCompleted || store.assistant == nil {
		t.Fatalf("service did not complete run: result=%+v status=%s assistant=%+v", result, store.status, store.assistant)
	}
}

func TestAdvisorServiceLegacyPathUsesBoundedRequestContext(t *testing.T) {
	messages := make([]entity.AdvisorMessage, 0, 30)
	for index := 1; index <= 30; index++ {
		role := entity.AdvisorRoleUser
		if index%2 == 0 {
			role = entity.AdvisorRoleAssistant
		}
		messages = append(messages, entity.AdvisorMessage{
			Seq:  int64(index),
			Role: role,
			Parts: []entity.AdvisorPart{{
				Type: "text",
				Text: fmt.Sprintf("legacy-%02d-%s", index, strings.Repeat("x", 3000)),
			}},
		})
	}
	store := &advisorStoreStub{
		run:      &entity.AdvisorRun{ID: "run-bounded", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued},
		message:  &messages[len(messages)-1],
		messages: messages,
	}
	provider := &capturingAdvisorProvider{response: AdvisorResponse{Text: "bounded legacy answer"}}
	service := NewAdvisorService(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	if _, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-bounded", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "current"}}}); err != nil {
		t.Fatal(err)
	}
	var historyBytes int
	foundOld := false
	foundLatest := false
	for _, message := range provider.request.Messages {
		if message.Role == "system" {
			continue
		}
		historyBytes += len([]byte(message.Content))
		if strings.Contains(message.Content, "legacy-01-") {
			foundOld = true
		}
		if strings.Contains(message.Content, "legacy-30-") {
			foundLatest = true
		}
	}
	if historyBytes > advisorContextTextBudget {
		t.Fatalf("legacy path exceeded bounded context: %d bytes", historyBytes)
	}
	if foundOld || !foundLatest {
		t.Fatalf("legacy path must drop oldest messages and retain latest: old=%v latest=%v", foundOld, foundLatest)
	}
}

func TestAdvisorServiceRuntimePathUsesAtomicCompletion(t *testing.T) {
	store := &advisorRuntimeStub{advisorStoreStub: &advisorStoreStub{
		run:     &entity.AdvisorRun{ID: "run-runtime", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued},
		message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}},
	}}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{Text: "bounded answer"}}}
	service := NewAdvisorServiceWithRuntime(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))), true)
	answer, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-runtime", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Text != "bounded answer" || store.completed != 1 || store.finished != 0 || store.assistant == nil || store.run.Status != entity.AdvisorStatusCompleted {
		t.Fatalf("runtime path must commit one atomic completion: answer=%+v completed=%d finished=%d status=%s assistant=%+v", answer, store.completed, store.finished, store.run.Status, store.assistant)
	}
}

func TestAdvisorServiceHeartbeatFailureCancelsProvider(t *testing.T) {
	store := &advisorRuntimeStub{advisorStoreStub: &advisorStoreStub{
		run:     &entity.AdvisorRun{ID: "run-heartbeat", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued},
		message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}},
	}, heartbeatErr: errors.New("lease lost")}
	provider := &blockingAdvisorProvider{started: make(chan struct{}), cancelled: make(chan struct{})}
	service := NewAdvisorServiceWithRuntime(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))), true)
	service.RuntimePolicy.HeartbeatEvery = time.Millisecond
	done := make(chan error, 1)
	go func() {
		_, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-heartbeat", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
		done <- err
	}()
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	select {
	case <-provider.cancelled:
	case <-time.After(time.Second):
		t.Fatal("heartbeat failure did not cancel provider")
	}
	if err := <-done; err == nil {
		t.Fatal("heartbeat failure must fail the run")
	}
}

func TestAdvisorServiceCancelInterruptsInFlightProvider(t *testing.T) {
	provider := &blockingAdvisorProvider{started: make(chan struct{}), cancelled: make(chan struct{})}
	store := &cancellableAdvisorStore{run: &entity.AdvisorRun{ID: "run-cancel", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued}, message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}}}
	service := NewAdvisorService(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	done := make(chan error, 1)
	go func() {
		_, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-cancel", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
		done <- err
	}()
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	if err := service.Cancel(context.Background(), "owner-1", "run-cancel"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-provider.cancelled:
	case <-time.After(time.Second):
		t.Fatal("cancel did not interrupt provider context")
	}
	if err := <-done; err == nil {
		t.Fatal("cancelled submit must return an error")
	}
}

func TestAdvisorServiceCancelFromAnotherProcessInterruptsProvider(t *testing.T) {
	provider := &blockingAdvisorProvider{started: make(chan struct{}), cancelled: make(chan struct{})}
	store := &distributedCancelStore{cancellableAdvisorStore: cancellableAdvisorStore{
		run:     &entity.AdvisorRun{ID: "run-cross-process-cancel", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued},
		message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}},
	}}
	runner := NewAdvisorService(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	canceller := NewAdvisorService(store, NewAdvisorOrchestrator(advisorAvailabilityProvider{available: true}, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	done := make(chan error, 1)
	go func() {
		_, err := runner.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-cross-process-cancel", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
		done <- err
	}()
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	if err := canceller.Cancel(context.Background(), "owner-1", "run-cross-process-cancel"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-provider.cancelled:
	case <-time.After(time.Second):
		t.Fatal("cross-process cancellation did not interrupt provider")
	}
	if err := <-done; err == nil {
		t.Fatal("cancelled submit must return an error")
	}
}

func TestAdvisorServiceBeginReturnsRunBeforeProviderCompletes(t *testing.T) {
	provider := &blockingAdvisorProvider{started: make(chan struct{}), cancelled: make(chan struct{})}
	store := &cancellableAdvisorStore{run: &entity.AdvisorRun{ID: "run-async", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued}, message: &entity.AdvisorMessage{ID: "msg-1", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}}}
	service := NewAdvisorService(store, NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))

	run, replay, err := service.Begin(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "req-async", PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "question"}}})
	if err != nil {
		t.Fatal(err)
	}
	if replay || run == nil || run.ID != "run-async" {
		t.Fatalf("begin must return the queued run immediately: run=%+v replay=%v", run, replay)
	}
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("async begin did not start provider")
	}
	if err := service.Cancel(context.Background(), "owner-1", run.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-provider.cancelled:
	case <-time.After(time.Second):
		t.Fatal("async cancellation did not interrupt provider")
	}
}

func TestAdvisorServiceResetInterruptsPurgedProviderRunsOnlyOnce(t *testing.T) {
	cancelled := make(chan struct{})
	store := &advisorResetStore{result: &entity.AdvisorResetResult{ConversationID: "conversation-1", Generation: 2, PurgedRunIDs: []string{"run-reset"}}}
	service := NewAdvisorService(store, NewAdvisorOrchestrator(advisorAvailabilityProvider{available: true}, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))))
	service.cancellations["run-reset"] = func() { close(cancelled) }

	first, err := service.ResetConversation(context.Background(), "owner-1", "reset-1", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if first.Replayed || first.PurgedRunIDs == nil {
		t.Fatalf("first reset must carry internal purge IDs: %+v", first)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("reset did not interrupt the purged provider run")
	}

	second, err := service.ResetConversation(context.Background(), "owner-1", "reset-1", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if !second.Replayed || len(second.PurgedRunIDs) != 0 {
		t.Fatalf("reset replay must not report runs for cancellation: %+v", second)
	}
}

func TestBuildAdvisorContextKeepsOnlyTwelveRecentMessages(t *testing.T) {
	messages := make([]entity.AdvisorMessage, 0, 20)
	for index := 1; index <= 20; index++ {
		role := entity.AdvisorRoleUser
		if index%2 == 0 {
			role = entity.AdvisorRoleAssistant
		}
		messages = append(messages, entity.AdvisorMessage{Seq: int64(index), Role: role, Parts: []entity.AdvisorPart{{Type: "text", Text: fmt.Sprintf("message-%02d", index)}}})
	}
	chat := BuildAdvisorContext(messages)
	if len(chat) != 12 || chat[0].Content != "message-09" || chat[11].Content != "message-20" {
		t.Fatalf("context must keep the twelve latest messages in order: len=%d first=%q last=%q", len(chat), chat[0].Content, chat[11].Content)
	}
}

func TestBuildAdvisorSystemMessageUsesAccountLocalDate(t *testing.T) {
	now := time.Date(2026, time.October, 19, 17, 30, 0, 0, time.UTC)
	message := BuildAdvisorSystemMessageAt(Principal{Timezone: "Asia/Ho_Chi_Minh"}, now)
	if message.Role != "system" || !strings.Contains(message.Content, "Account-local today is 2026-10-20") {
		t.Fatalf("system message must use account-local calendar date: %+v", message)
	}
	if !strings.Contains(message.Content, "Account timezone: Asia/Ho_Chi_Minh") {
		t.Fatalf("system message must disclose the timezone used for relative dates: %s", message.Content)
	}
}

func TestBuildAdvisorSystemMessageUsesTimezoneAcrossDateBoundary(t *testing.T) {
	now := time.Date(2026, time.October, 19, 17, 30, 0, 0, time.UTC)
	message := BuildAdvisorSystemMessageAt(Principal{Timezone: "America/Los_Angeles"}, now)
	if !strings.Contains(message.Content, "Account-local today is 2026-10-19") {
		t.Fatalf("system message must calculate date in the account timezone: %s", message.Content)
	}
}

func TestBuildAdvisorSystemMessageDoesNotGuessInvalidAccountDate(t *testing.T) {
	now := time.Date(2026, time.October, 19, 17, 30, 0, 0, time.UTC)
	for _, timezone := range []string{"", "not/a-timezone"} {
		message := BuildAdvisorSystemMessageAt(Principal{Timezone: timezone}, now)
		if !strings.Contains(message.Content, "Account-local today is unavailable") || strings.Contains(message.Content, "Account-local today is 2026-") {
			t.Fatalf("invalid or missing timezone must not produce a guessed date (%q): %s", timezone, message.Content)
		}
	}
}
