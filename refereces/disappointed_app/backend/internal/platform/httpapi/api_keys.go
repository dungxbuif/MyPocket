package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/audit"
	"mypocket/internal/identity"
	"mypocket/internal/platform/authcache"
	"mypocket/internal/platform/config"
)

type apiKeyRequest struct {
	Name string `json:"name"`
}

type apiKeysResponse struct {
	Status        string            `json:"status"`
	Keys          []identity.APIKey `json:"keys"`
	CorrelationID string            `json:"correlation_id"`
}

type createdAPIKeyResponse struct {
	Status        string                 `json:"status"`
	Key           identity.CreatedAPIKey `json:"key"`
	CorrelationID string                 `json:"correlation_id"`
}

func apiKeys(cfg config.Config, repo APIKeyRepository, auditRepo AuditRepository, cache AuthCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cache == nil {
			cache = authcache.Noop{}
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "API keys unavailable", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			keys, err := repo.ListAPIKeys(r.Context(), userID)
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "API keys unavailable", correlationID(r.Context())))
				return
			}
			writeJSON(w, http.StatusOK, apiKeysResponse{Status: "ok", Keys: keys, CorrelationID: correlationID(r.Context())})
		case http.MethodPost:
			if authenticatedMethod(r.Context()) == "api_key" || !browserCSRFSatisfied(r) {
				writeJSON(w, http.StatusForbidden, ErrorEnvelope("CSRF_REQUIRED", "CSRF token is required", correlationID(r.Context())))
				return
			}
			var req apiKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			key, err := repo.CreateAPIKey(r.Context(), userID, req.Name, cfg.APIKeyHashSecret)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid API key input", correlationID(r.Context())))
				return
			}
			if key.KeyHash != "" {
				_ = cache.Set(r.Context(), apiKeyCacheKey(key.KeyHash), userID, 5*time.Minute)
			}
			appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), ActorUserID: userID, Action: "api_key.create", EntityType: "api_key", EntityID: key.ID, Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAPI, RequestMethod: r.Method, RequestPath: safeRequestPath(r)})
			writeJSON(w, http.StatusCreated, createdAPIKeyResponse{Status: "ok", Key: key, CorrelationID: correlationID(r.Context())})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func apiKeyByID(cfg config.Config, repo APIKeyRepository, auditRepo AuditRepository, cache AuthCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cache == nil {
			cache = authcache.Noop{}
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "API keys unavailable", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if authenticatedMethod(r.Context()) == "api_key" || !browserCSRFSatisfied(r) {
			writeJSON(w, http.StatusForbidden, ErrorEnvelope("CSRF_REQUIRED", "CSRF token is required", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		keyID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/api-keys/"), "/revoke")
		if keyID == "" || !strings.HasSuffix(r.URL.Path, "/revoke") {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "API key not found", correlationID(r.Context())))
			return
		}
		key, err := repo.RevokeAPIKey(r.Context(), userID, keyID)
		if err != nil {
			status := http.StatusServiceUnavailable
			code := "INTERNAL_RETRYABLE"
			message := "API keys unavailable"
			if errors.Is(err, identity.ErrUserNotFound) {
				status = http.StatusNotFound
				code = "NOT_FOUND"
				message = "API key not found"
			}
			writeJSON(w, status, ErrorEnvelope(code, message, correlationID(r.Context())))
			return
		}
		if key.KeyHash != "" {
			_ = cache.Delete(r.Context(), apiKeyCacheKey(key.KeyHash))
		}
		appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), ActorUserID: userID, Action: "api_key.revoke", EntityType: "api_key", EntityID: keyID, Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAPI, RequestMethod: r.Method, RequestPath: safeRequestPath(r)})
		writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
	}
}

func csrfSatisfiedForMethod(r *http.Request) bool {
	if authenticatedMethod(r.Context()) == "api_key" {
		return true
	}
	return browserCSRFSatisfied(r)
}

func browserCSRFSatisfied(r *http.Request) bool {
	cookie, err := r.Cookie(identity.CSRFCookieName)
	header := r.Header.Get(identity.CSRFHeaderName)
	return err == nil && header != "" && cookie.Value != "" && header == cookie.Value
}
