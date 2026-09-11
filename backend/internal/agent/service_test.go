package agent

import (
	"context"
	"testing"
)

type storeStub struct{ creates int }

func (s *storeStub) CreateRun(_ context.Context, user, key string, kind Kind, text string) (Run, error) {
	s.creates++
	return Run{ID: "r", UserID: user, Kind: kind, RequestText: text}, nil
}
func (s *storeStub) CreateRunWithTool(_ context.Context, user, key string, kind Kind, text, receipt string) (Run, error) {
	s.creates++
	return Run{ID: "r", UserID: user, Kind: kind, RequestText: text}, nil
}
func (s *storeStub) GetRun(_ context.Context, user, id string) (Run, error) {
	return Run{ID: id, UserID: user}, nil
}
func (s *storeStub) CreateToolRun(_ context.Context, user, run, receipt string) (ToolRun, error) {
	return ToolRun{UserID: user, AgentRunID: run, ReceiptObjectID: receipt}, nil
}

func TestServiceRejectsBeforePersistence(t *testing.T) {
	store := &storeStub{}
	_, err := (Service{Store: store}).Submit(context.Background(), "u", "", KindAnalysis, "hello", "")
	if err == nil || store.creates != 0 {
		t.Fatalf("expected validation before persistence, calls=%d err=%v", store.creates, err)
	}
}
