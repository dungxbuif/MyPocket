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
	Submit(context.Context, string, string, agent.Kind, string) (agent.Run, error)
	Get(context.Context, string, string) (agent.Run, error)
}

type agentRunResponse struct {
	Status        string    `json:"status"`
	Run           agent.Run `json:"run"`
	CorrelationID string    `json:"correlation_id"`
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
				Kind    agent.Kind `json:"kind"`
				Message string     `json:"message"`
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&request) != nil {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			run, err := service.Submit(r.Context(), userID, strings.TrimSpace(r.Header.Get("Idempotency-Key")), request.Kind, request.Message)
			if err != nil {
				writeAgentError(w, r, err)
				return
			}
			writeJSON(w, 202, agentRunResponse{Status: "ok", Run: run, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func agentRunByID(cfg config.Config, service AgentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/agent/runs/"), "/")
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
