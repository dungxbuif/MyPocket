package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"mypocket/internal/agent"
	"mypocket/internal/platform/config"
)

type AgentService interface {
	Submit(context.Context, string, string, agent.Kind, string, string) (agent.Run, error)
	SubmitSession(context.Context, string, string, string, agent.Kind, string, string) (agent.Run, agent.Session, agent.Message, error)
	Get(context.Context, string, string) (agent.Run, error)
	GetSession(context.Context, string, string, int) (agent.SessionHistory, error)
}

type agentRunResponse struct {
	Status        string    `json:"status"`
	Run           agent.Run `json:"run"`
	CorrelationID string    `json:"correlation_id"`
}

type agentSessionRunResponse struct {
	Status        string        `json:"status"`
	Run           agent.Run     `json:"run"`
	Session       agent.Session `json:"session"`
	Message       agent.Message `json:"message"`
	CorrelationID string        `json:"correlation_id"`
}

type agentSessionResponse struct {
	Status        string          `json:"status"`
	Session       agent.Session   `json:"session"`
	Messages      []agent.Message `json:"messages"`
	CorrelationID string          `json:"correlation_id"`
}

func agentMessages(cfg config.Config, service AgentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if service == nil {
				writeJSON(w, 503, ErrorEnvelope("CAPABILITY_UNAVAILABLE", "Agent is unavailable", correlationID(r.Context())))
				return
			}
			var request struct {
				Kind      agent.Kind `json:"kind"`
				Message   string     `json:"message"`
				ReceiptID string     `json:"receipt_id,omitempty"`
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&request) != nil {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			run, err := service.Submit(r.Context(), userID, strings.TrimSpace(r.Header.Get("Idempotency-Key")), request.Kind, request.Message, strings.TrimSpace(request.ReceiptID))
			if err != nil {
				writeAgentError(w, r, err)
				return
			}
			writeJSON(w, 202, agentRunResponse{Status: "ok", Run: run, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func agentIntakes(cfg config.Config, service AgentService) http.HandlerFunc {
	return fixedAgentMessage(cfg, service, agent.KindIntake, "")
}

func agentIntakeByID(cfg config.Config, service AgentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/messages") {
			sessionID := strings.TrimSuffix(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/agent/intakes/"), "/"), "/messages")
			if sessionID == "" || strings.Contains(sessionID, "/") {
				writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Agent session not found", correlationID(r.Context())))
				return
			}
			fixedAgentMessage(cfg, service, agent.KindIntake, sessionID).ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		sessionID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/agent/intakes/"), "/")
		if sessionID == "" || strings.Contains(sessionID, "/") {
			writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Agent session not found", correlationID(r.Context())))
			return
		}
		if service == nil {
			writeJSON(w, 503, ErrorEnvelope("CAPABILITY_UNAVAILABLE", "Agent is unavailable", correlationID(r.Context())))
			return
		}
		history, err := service.GetSession(r.Context(), userID, sessionID, 50)
		if err != nil {
			writeAgentError(w, r, err)
			return
		}
		writeJSON(w, 200, agentSessionResponse{Status: "ok", Session: history.Session, Messages: history.Messages, CorrelationID: correlationID(r.Context())})
	}
}

func agentAdvisorMessages(cfg config.Config, service AgentService) http.HandlerFunc {
	return fixedAgentMessage(cfg, service, agent.KindAdvisor, "")
}

func fixedAgentMessage(cfg config.Config, service AgentService, kind agent.Kind, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if service == nil {
				writeJSON(w, 503, ErrorEnvelope("CAPABILITY_UNAVAILABLE", "Agent is unavailable", correlationID(r.Context())))
				return
			}
			var request struct {
				Message   string `json:"message"`
				ReceiptID string `json:"receipt_id,omitempty"`
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&request) != nil {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			run, session, message, err := service.SubmitSession(r.Context(), userID, sessionID, strings.TrimSpace(r.Header.Get("Idempotency-Key")), kind, request.Message, strings.TrimSpace(request.ReceiptID))
			if err != nil {
				writeAgentError(w, r, err)
				return
			}
			writeJSON(w, 202, agentSessionRunResponse{Status: "ok", Run: run, Session: session, Message: message, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func agentRunByID(cfg config.Config, service AgentService) http.HandlerFunc {
	return agentRunByPrefix(cfg, service, "/api/v1/agent/runs/")
}

func agentRunByPrefix(cfg config.Config, service AgentService, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
		id = strings.TrimSuffix(id, "/messages")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Agent run not found", correlationID(r.Context())))
			return
		}
		if service == nil {
			writeJSON(w, 503, ErrorEnvelope("CAPABILITY_UNAVAILABLE", "Agent is unavailable", correlationID(r.Context())))
			return
		}
		run, err := service.Get(r.Context(), userID, id)
		if err != nil {
			writeAgentError(w, r, err)
			return
		}
		writeJSON(w, 200, agentRunResponse{Status: "ok", Run: run, CorrelationID: correlationID(r.Context())})
	}
}

func writeAgentError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, agent.ErrValidation):
		writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Agent request is invalid", correlationID(r.Context())))
	case errors.Is(err, agent.ErrNotFound):
		writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Agent run not found", correlationID(r.Context())))
	case errors.Is(err, agent.ErrConflict):
		writeJSON(w, 409, ErrorEnvelope("IDEMPOTENCY_CONFLICT", "Idempotency key was already used for another request", correlationID(r.Context())))
	default:
		writeJSON(w, 503, ErrorEnvelope("INTERNAL_RETRYABLE", "Agent request unavailable", correlationID(r.Context())))
	}
}
