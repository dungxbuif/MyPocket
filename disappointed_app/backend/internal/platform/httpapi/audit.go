package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/audit"
	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type auditEventsResponse struct {
	Status        string        `json:"status"`
	Events        []audit.Event `json:"events"`
	CorrelationID string        `json:"correlation_id"`
}

type auditAccessResponse struct {
	Status        string `json:"status"`
	Allowed       bool   `json:"allowed"`
	CorrelationID string `json:"correlation_id"`
}

func auditAccess(cfg config.Config, identityRepo IdentityRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if identityRepo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Audit unavailable", correlationID(r.Context())))
			return
		}
		allowed, _ := auditViewerAllowed(w, r, cfg, identityRepo)
		if !allowed {
			return
		}
		writeJSON(w, http.StatusOK, auditAccessResponse{Status: "ok", Allowed: true, CorrelationID: correlationID(r.Context())})
	}
}

func auditEvents(cfg config.Config, identityRepo IdentityRepository, auditRepo AuditRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if auditRepo == nil || identityRepo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Audit unavailable", correlationID(r.Context())))
			return
		}
		allowed, actorID := auditViewerAllowed(w, r, cfg, identityRepo)
		if !allowed {
			appendAudit(r.Context(), auditRepo, audit.Event{
				CorrelationID: correlationID(r.Context()),
				ActorUserID:   actorID,
				Action:        "audit.viewer.denied",
				EntityType:    "audit",
				Outcome:       audit.OutcomeDenied,
				Severity:      audit.SeveritySecurity,
				Source:        audit.SourceAPI,
				RequestMethod: r.Method,
				RequestPath:   safeRequestPath(r),
				IPHash:        audit.HashValue(cfg.AuditHashSecret, clientIP(r)),
				UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent()),
			})
			return
		}
		query, err := auditQueryFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid audit query", correlationID(r.Context())))
			return
		}
		events, err := auditRepo.List(r.Context(), query)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Audit unavailable", correlationID(r.Context())))
			return
		}
		writeJSON(w, http.StatusOK, auditEventsResponse{Status: "ok", Events: events, CorrelationID: correlationID(r.Context())})
	}
}

func auditViewerAllowed(w http.ResponseWriter, r *http.Request, cfg config.Config, repo IdentityRepository) (bool, string) {
	if cfg.AuditViewerEmail == "" {
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Audit viewer is not configured", correlationID(r.Context())))
		return false, ""
	}
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
		return false, ""
	}
	claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
		return false, ""
	}
	user, err := repo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		status := http.StatusServiceUnavailable
		code := "INTERNAL_RETRYABLE"
		message := "User unavailable"
		if errors.Is(err, identity.ErrUserNotFound) {
			status = http.StatusUnauthorized
			code = "AUTH_REQUIRED"
			message = "Authentication required"
		}
		writeJSON(w, status, ErrorEnvelope(code, message, correlationID(r.Context())))
		return false, claims.UserID
	}
	if !user.EmailVerified || strings.ToLower(strings.TrimSpace(user.Email)) != cfg.AuditViewerEmail {
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Forbidden", correlationID(r.Context())))
		return false, claims.UserID
	}
	return true, claims.UserID
}

func auditQueryFromRequest(r *http.Request) (audit.Query, error) {
	query := audit.Query{
		CorrelationID: strings.TrimSpace(r.URL.Query().Get("correlation_id")),
		Action:        strings.TrimSpace(r.URL.Query().Get("action")),
		Severity:      audit.Severity(strings.TrimSpace(r.URL.Query().Get("severity"))),
		Limit:         100,
	}
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 200 {
			return audit.Query{}, errors.New("invalid limit")
		}
		query.Limit = limit
	}
	var err error
	if value := strings.TrimSpace(r.URL.Query().Get("from")); value != "" {
		query.From, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return audit.Query{}, err
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("to")); value != "" {
		query.To, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return audit.Query{}, err
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("before")); value != "" {
		query.Before, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return audit.Query{}, err
		}
	}
	return query, nil
}
