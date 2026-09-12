package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/agent"
	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type agentServiceStub struct {
	submitted int
	user, key string
	kind      agent.Kind
	run       agent.Run
	session   agent.Session
	message   agent.Message
	history   agent.SessionHistory
	err       error
}

func (s *agentServiceStub) Submit(_ context.Context, user, key string, kind agent.Kind, _, _ string) (agent.Run, error) {
	s.submitted++
	s.user = user
	s.key = key
	s.kind = kind
	return s.run, s.err
}
func (s *agentServiceStub) SubmitSession(_ context.Context, user, sessionID, key string, kind agent.Kind, text, _ string) (agent.Run, agent.Session, agent.Message, error) {
	s.submitted++
	s.user = user
	s.key = key
	s.kind = kind
	if s.run.ID == "" {
		s.run = agent.Run{ID: "run-1", SessionID: sessionID, Kind: kind, Status: agent.StatusQueued, RequestText: text}
	}
	return s.run, s.session, s.message, s.err
}
func (s *agentServiceStub) Get(_ context.Context, user, id string) (agent.Run, error) {
	s.user = user
	return s.run, s.err
}
func (s *agentServiceStub) GetSession(_ context.Context, user, id string, _ int) (agent.SessionHistory, error) {
	s.user = user
	if s.history.Session.ID == "" {
		s.history.Session = agent.Session{ID: id, Kind: agent.KindIntake, Status: "active"}
	}
	return s.history, s.err
}

func TestAgentMessagesQueuesReviewFirstRun(t *testing.T) {
	service := &agentServiceStub{run: agent.Run{ID: "run-1", Status: agent.StatusQueued}}
	handler := agentMessages(agentTestConfig(), service)
	req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/messages", `{"kind":"transaction_draft","message":"Lunch 120k"}`)
	agentAddCSRF(req)
	req.Header.Set("Idempotency-Key", "agent-once")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 202 || service.submitted != 1 || service.user != "user_123" || service.key != "agent-once" {
		t.Fatalf("code=%d body=%s service=%#v", res.Code, res.Body.String(), service)
	}
}

func TestAgentIntakeEndpointQueuesIntakeKind(t *testing.T) {
	service := &agentServiceStub{
		run:     agent.Run{ID: "run-1", SessionID: "session-1", Kind: agent.KindIntake, Status: agent.StatusQueued},
		session: agent.Session{ID: "session-1", Kind: agent.KindIntake, Status: "active"},
		message: agent.Message{ID: "message-1", SessionID: "session-1", RunID: "run-1", Role: agent.MessageRoleUser, Text: "Ăn trưa 80k"},
	}
	handler := agentIntakes(agentTestConfig(), service)
	req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/intakes", `{"message":"Ăn trưa 80k"}`)
	agentAddCSRF(req)
	req.Header.Set("Idempotency-Key", "intake-once")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != 202 || service.submitted != 1 || service.user != "user_123" || service.key != "intake-once" || service.kind != agent.KindIntake {
		t.Fatalf("code=%d body=%s service=%#v", res.Code, res.Body.String(), service)
	}
	var body struct {
		Run     agent.Run     `json:"run"`
		Session agent.Session `json:"session"`
		Message agent.Message `json:"message"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Session.ID != "session-1" || body.Message.Text != "Ăn trưa 80k" || body.Run.SessionID != "session-1" {
		t.Fatalf("session response missing chat envelope: %#v", body)
	}
}

func TestAgentIntakeHistoryEndpointReturnsMessages(t *testing.T) {
	service := &agentServiceStub{history: agent.SessionHistory{
		Session: agent.Session{ID: "session-1", Kind: agent.KindIntake, Status: "active"},
		Messages: []agent.Message{
			{ID: "message-1", SessionID: "session-1", Role: agent.MessageRoleUser, Text: "Ăn trưa 80k"},
			{ID: "message-2", SessionID: "session-1", Role: agent.MessageRoleAssistant, Text: "Đã tạo nháp", Action: &agent.ActionCard{Type: agent.ActionDraftsCreated, DraftIDs: []string{"draft-1"}}},
		},
	}}
	handler := agentIntakeByID(agentTestConfig(), service)
	req := agentAuthenticatedRequest(t, http.MethodGet, "/api/v1/agent/intakes/session-1", "")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != 200 || service.user != "user_123" {
		t.Fatalf("code=%d body=%s service=%#v", res.Code, res.Body.String(), service)
	}
	var body struct {
		Session  agent.Session   `json:"session"`
		Messages []agent.Message `json:"messages"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Session.ID != "session-1" || len(body.Messages) != 2 || body.Messages[1].Action == nil || body.Messages[1].Action.DraftIDs[0] != "draft-1" {
		t.Fatalf("history response missing messages/action card: %#v", body)
	}
}

func TestAgentAdvisorEndpointQueuesAdvisorKind(t *testing.T) {
	service := &agentServiceStub{run: agent.Run{ID: "run-2", Kind: agent.KindAdvisor, Status: agent.StatusQueued}}
	handler := agentAdvisorMessages(agentTestConfig(), service)
	req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/advisor/messages", `{"message":"Tháng này tiêu gì nhiều?"}`)
	agentAddCSRF(req)
	req.Header.Set("Idempotency-Key", "advisor-once")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != 202 || service.submitted != 1 || service.user != "user_123" || service.key != "advisor-once" || service.kind != agent.KindAdvisor {
		t.Fatalf("code=%d body=%s service=%#v", res.Code, res.Body.String(), service)
	}
}

func TestAgentMessagesFailClosedWhenDisabled(t *testing.T) {
	req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/messages", `{"kind":"analysis","message":"summary"}`)
	agentAddCSRF(req)
	res := httptest.NewRecorder()
	agentMessages(agentTestConfig(), nil).ServeHTTP(res, req)
	if res.Code != 503 {
		t.Fatalf("expected 503 got %d: %s", res.Code, res.Body.String())
	}
}

func agentTestConfig() config.Config {
	return config.Config{CookieSecret: "01234567890123456789012345678901", CSRFSecret: "abcdefghijklmnopqrstuvwxyz123456"}
}

func agentAuthenticatedRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	value, err := identity.NewCookieSigner([]byte(agentTestConfig().CookieSecret)).Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: value})
	return req
}

func agentAddCSRF(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: identity.CSRFCookieName, Value: "token"})
	req.Header.Set(identity.CSRFHeaderName, "token")
}
