package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"mypocket/internal/agent"
)

func TestClientUsesCompatibleContractAndRetriesRetryableStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer secret-value" {
			t.Fatalf("unexpected request %s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"response_format"`) {
			t.Fatalf("missing structured output contract: %s", body)
		}
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"answer\":\"ok\"}"}}]}`)
	}))
	defer server.Close()
	client, err := New(server.URL, "secret-value", "model-a", time.Second, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Generate(context.Background(), agent.ModelRequest{System: "safe", Input: "hello", Schema: agent.AnalysisSchema()})
	if err != nil || got.Text != `{"answer":"ok"}` || calls.Load() != 2 {
		t.Fatalf("unexpected response %#v calls=%d err=%v", got, calls.Load(), err)
	}
}

func TestClientRedactsKeyFromErrorsAndDoesNotRetryValidationStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"secret-value"}`)
	}))
	defer server.Close()
	client, _ := New(server.URL, "secret-value", "model-a", time.Second, 3)
	_, err := client.Generate(context.Background(), agent.ModelRequest{Schema: agent.AnalysisSchema()})
	if err == nil || strings.Contains(err.Error(), "secret-value") || calls.Load() != 1 {
		t.Fatalf("unexpected error %v calls=%d", err, calls.Load())
	}
}

func TestClientRejectsOversizedAndMalformedResponses(t *testing.T) {
	for _, response := range []string{`not-json`, `{"choices":[]}`, strings.Repeat("x", maxResponseBytes+1)} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, response) }))
		client, _ := New(server.URL, "secret", "model", time.Second, 0)
		if _, err := client.Generate(context.Background(), agent.ModelRequest{Schema: agent.AnalysisSchema()}); err == nil {
			t.Fatal("expected invalid response error")
		}
		server.Close()
	}
}

func TestClientTimeoutAndRefusalAreSafe(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(50 * time.Millisecond)
			_, _ = io.WriteString(w, `{}`)
		}))
		defer server.Close()
		client, _ := New(server.URL, "never-print-me", "model", time.Millisecond, 0)
		_, err := client.Generate(context.Background(), agent.ModelRequest{Schema: agent.AnalysisSchema()})
		if err == nil || strings.Contains(err.Error(), "never-print-me") {
			t.Fatalf("expected redacted timeout, got %v", err)
		}
	})
	t.Run("refusal", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"choices":[{"message":{"refusal":"private provider reason"}}]}`)
		}))
		defer server.Close()
		client, _ := New(server.URL, "secret", "model", time.Second, 0)
		got, err := client.Generate(context.Background(), agent.ModelRequest{Schema: agent.AnalysisSchema()})
		if err != nil || got.Refusal != "model refused request" {
			t.Fatalf("unexpected refusal %#v %v", got, err)
		}
	})
}
