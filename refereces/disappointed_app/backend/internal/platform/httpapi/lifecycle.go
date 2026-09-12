package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/lifecycle"
	"mypocket/internal/platform/config"
)

type lifecycleResponse struct {
	Status        string                        `json:"status"`
	Job           lifecycle.Job                 `json:"job"`
	DownloadURL   string                        `json:"download_url,omitempty"`
	Preview       *lifecycle.DestructivePreview `json:"preview,omitempty"`
	CorrelationID string                        `json:"correlation_id"`
}

func imports(cfg config.Config, repo LifecycleRepository, store LifecycleObjectStore) http.HandlerFunc {
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
			key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
			if key == "" {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Idempotency-Key is required", correlationID(r.Context())))
				return
			}
			var request lifecycle.ImportRequest
			if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "text/csv") {
				if store == nil {
					writeJSON(w, 503, ErrorEnvelope("INTERNAL_RETRYABLE", "Import storage unavailable", correlationID(r.Context())))
					return
				}
				body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, lifecycle.MaxImportBytes))
				if err != nil {
					writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "CSV is too large", correlationID(r.Context())))
					return
				}
				request.ObjectKey = fmt.Sprintf("users/%s/imports/%s.csv", userID, newCorrelationID())
				if err = store.PutObject(r.Context(), request.ObjectKey, "text/csv; charset=utf-8", strings.NewReader(string(body)), int64(len(body))); err != nil {
					writeJSON(w, 503, ErrorEnvelope("INTERNAL_RETRYABLE", "Import upload failed", correlationID(r.Context())))
					return
				}
			} else if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&request); err != nil {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			if !strings.HasPrefix(request.ObjectKey, "users/"+userID+"/imports/") {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Import object is not owned", correlationID(r.Context())))
				return
			}
			job, err := repo.CreateJob(r.Context(), userID, lifecycle.KindImport, key, request)
			if err != nil {
				writeLifecycleError(w, r, err)
				return
			}
			writeJSON(w, 202, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func importByID(cfg config.Config, repo LifecycleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/imports/"), "/"), "/")
		if len(parts) < 1 || parts[0] == "" {
			writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Import not found", correlationID(r.Context())))
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			job, err := repo.GetJob(r.Context(), userID, parts[0])
			if err != nil {
				writeLifecycleError(w, r, err)
				return
			}
			writeJSON(w, 200, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
			return
		}
		if len(parts) == 2 && parts[1] == "confirm" && r.Method == http.MethodPost {
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					Version int64 `json:"version"`
				}
				if json.NewDecoder(r.Body).Decode(&req) != nil {
					writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
					return
				}
				job, err := repo.ConfirmImport(r.Context(), userID, parts[0], req.Version)
				if err != nil {
					writeLifecycleError(w, r, err)
					return
				}
				writeJSON(w, 202, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
			return
		}
		writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Import not found", correlationID(r.Context())))
	}
}

func exports(cfg config.Config, repo LifecycleRepository) http.HandlerFunc {
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
			var req lifecycle.ExportRequest
			if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req) != nil {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
			if key == "" {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Idempotency-Key is required", correlationID(r.Context())))
				return
			}
			if req.SnapshotAt.IsZero() {
				req.SnapshotAt = time.Now().UTC()
			}
			job, err := repo.CreateJob(r.Context(), userID, lifecycle.KindExport, key, req)
			if err != nil {
				writeLifecycleError(w, r, err)
				return
			}
			writeJSON(w, 202, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func exportByID(cfg config.Config, repo LifecycleRepository, store LifecycleObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/exports/"), "/"), "/")
		if len(parts) < 1 || parts[0] == "" {
			writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Export not found", correlationID(r.Context())))
			return
		}
		job, err := repo.GetJob(r.Context(), userID, parts[0])
		if err != nil || job.Kind != lifecycle.KindExport {
			writeLifecycleError(w, r, lifecycle.ErrNotFound)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, 200, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
			return
		}
		if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet && job.Status == lifecycle.StatusCompleted && store != nil {
			url, err := store.PresignGet(r.Context(), job.ResultObjectKey, 5*time.Minute)
			if err != nil {
				writeJSON(w, 503, ErrorEnvelope("INTERNAL_RETRYABLE", "Download unavailable", correlationID(r.Context())))
				return
			}
			writeJSON(w, 200, lifecycleResponse{Status: "ok", Job: job, DownloadURL: url, CorrelationID: correlationID(r.Context())})
			return
		}
		writeJSON(w, 404, ErrorEnvelope("NOT_FOUND", "Export not found", correlationID(r.Context())))
	}
}

func accountJob(cfg config.Config, repo LifecycleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/account/jobs/"), "/")
		job, err := repo.GetJob(r.Context(), userID, id)
		if err != nil {
			writeLifecycleError(w, r, err)
			return
		}
		writeJSON(w, 200, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
	}
}

func destructiveAccount(cfg config.Config, repo LifecycleRepository, kind lifecycle.Kind) http.HandlerFunc {
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
			if !recentCookieAuth(r, cfg, 10*time.Minute) {
				writeJSON(w, 403, ErrorEnvelope("RECENT_AUTH_REQUIRED", "Recent browser authentication is required", correlationID(r.Context())))
				return
			}
			var req lifecycle.DestructiveRequest
			_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&req)
			expected := strings.ToUpper(string(kind))
			counts, err := repo.Counts(r.Context(), userID)
			if err != nil {
				writeLifecycleError(w, r, err)
				return
			}
			if req.PreviewToken == "" || req.Confirmation == "" {
				preview := lifecycle.DestructivePreview{AffectedCounts: counts, ExpiresAt: time.Now().Add(5 * time.Minute)}
				preview.Token = signPreview(cfg, userID, kind, preview.ExpiresAt)
				writeJSON(w, 200, lifecycleResponse{Status: "preview", Preview: &preview, CorrelationID: correlationID(r.Context())})
				return
			}
			if req.Confirmation != expected || !verifyPreview(cfg, req.PreviewToken, userID, kind) {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Confirmation or preview token is invalid", correlationID(r.Context())))
				return
			}
			key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
			if key == "" {
				writeJSON(w, 400, ErrorEnvelope("VALIDATION_FAILED", "Idempotency-Key is required", correlationID(r.Context())))
				return
			}
			job, err := repo.CreateJob(r.Context(), userID, kind, key, map[string]any{"affected_counts": counts})
			if err != nil {
				writeLifecycleError(w, r, err)
				return
			}
			if kind == lifecycle.KindDelete {
				if err = repo.DisableUser(r.Context(), userID); err != nil {
					writeLifecycleError(w, r, err)
					return
				}
			}
			writeJSON(w, 202, lifecycleResponse{Status: "ok", Job: job, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func recentCookieAuth(r *http.Request, cfg config.Config, maxAge time.Duration) bool {
	if authenticatedMethod(r.Context()) != "cookie" {
		return false
	}
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil {
		return false
	}
	claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
	return err == nil && time.Since(time.Unix(claims.IssuedAt, 0)) <= maxAge
}
func signPreview(cfg config.Config, user string, kind lifecycle.Kind, expires time.Time) string {
	payload := fmt.Sprintf("%s|%s|%d", user, kind, expires.Unix())
	mac := hmac.New(sha256.New, []byte(cfg.CSRFSecret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func verifyPreview(cfg config.Config, token, user string, kind lifecycle.Kind) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	payload := string(payloadBytes)
	var tokenUser, tokenKind string
	var expires int64
	if _, err = fmt.Sscanf(payload, "%s", &tokenUser); err != nil {
		return false
	}
	fields := strings.Split(payload, "|")
	if len(fields) != 3 {
		return false
	}
	tokenUser, tokenKind = fields[0], fields[1]
	if _, err = fmt.Sscan(fields[2], &expires); err != nil {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(cfg.CSRFSecret))
	mac.Write([]byte(payload))
	return hmac.Equal(signature, mac.Sum(nil)) && tokenUser == user && tokenKind == string(kind) && time.Now().Unix() < expires
}
func writeLifecycleError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusServiceUnavailable, "INTERNAL_RETRYABLE", "Lifecycle operation unavailable"
	if errors.Is(err, lifecycle.ErrNotFound) {
		status, code, message = 404, "NOT_FOUND", "Lifecycle job not found"
	} else if errors.Is(err, lifecycle.ErrConflict) {
		status, code, message = 409, "VERSION_CONFLICT", "Lifecycle job changed"
	} else if errors.Is(err, lifecycle.ErrInvalid) {
		status, code, message = 400, "VALIDATION_FAILED", "Lifecycle request is invalid"
	}
	writeJSON(w, status, ErrorEnvelope(code, message, correlationID(r.Context())))
}
