package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"mypocket/internal/platform/config"
	mysync "mypocket/internal/sync"
)

type syncMutationsRequest struct {
	Mutations []mysync.Mutation `json:"mutations"`
}

type syncMutationsResponse struct {
	Status        string                  `json:"status"`
	Results       []mysync.MutationResult `json:"results"`
	CorrelationID string                  `json:"correlation_id"`
}

type syncChangesResponse struct {
	Status        string          `json:"status"`
	Changes       []mysync.Change `json:"changes"`
	NextCursor    int64           `json:"next_cursor"`
	CorrelationID string          `json:"correlation_id"`
}

type syncResyncResponse struct {
	Status        string          `json:"status"`
	Snapshot      mysync.Snapshot `json:"snapshot"`
	CorrelationID string          `json:"correlation_id"`
}

func syncMutations(cfg config.Config, service SyncService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Sync service unavailable", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req syncMutationsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			results, err := service.ApplyMutations(r.Context(), userID, req.Mutations)
			if err != nil {
				writeSyncError(w, r, err)
				return
			}
			writeJSON(w, http.StatusOK, syncMutationsResponse{Status: "ok", Results: results, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func syncChanges(cfg config.Config, service SyncService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Sync service unavailable", correlationID(r.Context())))
			return
		}
		after, err := parseInt64Query(r, "after", 0)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid after cursor", correlationID(r.Context())))
			return
		}
		limit, err := parseIntQuery(r, "limit", 100)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid limit", correlationID(r.Context())))
			return
		}
		result, err := service.Changes(r.Context(), userID, after, limit)
		if err != nil {
			writeSyncError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, syncChangesResponse{Status: "ok", Changes: result.Changes, NextCursor: result.NextCursor, CorrelationID: correlationID(r.Context())})
	}
}

func syncResync(cfg config.Config, service SyncService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Sync service unavailable", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			snapshot, err := service.Resync(r.Context(), userID)
			if err != nil {
				writeSyncError(w, r, err)
				return
			}
			writeJSON(w, http.StatusOK, syncResyncResponse{Status: "ok", Snapshot: snapshot, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func parseInt64Query(r *http.Request, key string, fallback int64) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseIntQuery(r *http.Request, key string, fallback int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func writeSyncError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, mysync.ErrValidation) {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Sync unavailable", correlationID(r.Context())))
}
