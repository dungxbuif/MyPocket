package usecase

import (
	"context"
	"errors"
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
	assistant *entity.AdvisorMessage
	status    string
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
