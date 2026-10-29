package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/notification"
	"mypocket/internal/platform/config"
)

type notificationsResponse struct {
	Status        string                `json:"status"`
	Notifications []notification.Notice `json:"notifications"`
	HasMore       bool                  `json:"has_more"`
	NextBefore    string                `json:"next_before,omitempty"`
	CorrelationID string                `json:"correlation_id"`
}

type pushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
	ExpirationTime *int64 `json:"expiration_time"`
}

type pushSubscriptionResponse struct {
	Status        string                        `json:"status"`
	Subscription  notification.PushSubscription `json:"subscription"`
	CorrelationID string                        `json:"correlation_id"`
}

func notifications(cfg config.Config, repo NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Notification repository unavailable", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > notification.MaxPageSize {
			limit = notification.MaxPageSize
		}
		var before time.Time
		if value := strings.TrimSpace(r.URL.Query().Get("before")); value != "" {
			parsed, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid notification cursor", correlationID(r.Context())))
				return
			}
			before = parsed
		}
		items, more, err := repo.List(r.Context(), userID, limit, before)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Notifications unavailable", correlationID(r.Context())))
			return
		}
		response := notificationsResponse{Status: "ok", Notifications: items, HasMore: more, CorrelationID: correlationID(r.Context())}
		if more && len(items) > 0 {
			response.NextBefore = items[len(items)-1].CreatedAt.Format(time.RFC3339Nano)
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func notificationByID(cfg config.Config, repo NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Notification repository unavailable", correlationID(r.Context())))
			return
		}
		id, action, valid := parseNotificationPath(r.URL.Path)
		if !valid || action != "read" || r.Method != http.MethodPatch {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Notification not found", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := repo.MarkRead(r.Context(), userID, id); err != nil {
				writePlanningError(w, r, err, "Notification unavailable")
				return
			}
			writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func parseNotificationPath(path string) (string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/notifications/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "read", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "read" {
		return parts[0], "read", true
	}
	return "", "", false
}

func pushSubscriptions(cfg config.Config, repo NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Notification repository unavailable", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req pushSubscriptionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			var expires *time.Time
			if req.ExpirationTime != nil && *req.ExpirationTime > 0 {
				value := time.UnixMilli(*req.ExpirationTime)
				expires = &value
			}
			subscription, err := repo.UpsertPushSubscription(r.Context(), userID, notification.CreatePushSubscriptionInput{Endpoint: req.Endpoint, P256DH: req.Keys.P256DH, Auth: req.Keys.Auth, ExpiresAt: expires})
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid push subscription", correlationID(r.Context())))
				return
			}
			writeJSON(w, http.StatusCreated, pushSubscriptionResponse{Status: "ok", Subscription: subscription, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func pushSubscriptionByID(cfg config.Config, repo NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Notification repository unavailable", correlationID(r.Context())))
			return
		}
		id, action, valid := parseBudgetPathWithPrefix(r.URL.Path, "/api/v1/push-subscriptions/")
		if !valid || action != "" || r.Method != http.MethodDelete {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Push subscription not found", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := repo.DeletePushSubscription(r.Context(), userID, id); err != nil {
				writePlanningError(w, r, err, "Push subscription unavailable")
				return
			}
			writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}
