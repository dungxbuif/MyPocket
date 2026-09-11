package httpapi

import (
	"context"
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
	run       agent.Run
	err       error
}

func (s *agentServiceStub) Submit(_ context.Context, user, key string, _ agent.Kind, _, _ string) (agent.Run, error) {
	s.submitted++
	s.user = user
	s.key = key
	return s.run, s.err
}
func (s *agentServiceStub) Get(_ context.Context, user, id string) (agent.Run, error) {
	s.user = user
	return s.run, s.err
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
