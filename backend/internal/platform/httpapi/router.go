package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type Dependencies struct {
	ReadyCheck         func() error
	IdentityRepository IdentityRepository
}

func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/google", startGoogleAuth(cfg))
	mux.HandleFunc("/api/v1/auth/google/callback", googleCallback(cfg, deps.IdentityRepository))
	mux.Handle("/api/v1/auth/logout", requireCSRF(http.HandlerFunc(logout)))
	mux.HandleFunc("/api/v1/me", currentUser(cfg, deps.IdentityRepository))
	mux.HandleFunc("/api/v1/health/live", liveHealth)
	mux.HandleFunc("/api/v1/health/ready", readyHealth(deps))

	return correlationMiddleware(mux)
}

type IdentityRepository interface {
	FindOrCreateGoogleUser(ctx context.Context, profile identity.GoogleProfile) (identity.User, error)
	FindByID(ctx context.Context, id string) (identity.User, error)
}

func correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
		if correlationID == "" {
			correlationID = newCorrelationID()
		}
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(withCorrelationID(r.Context(), correlationID)))
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newCorrelationID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "req_unavailable"
	}
	return "req_" + hex.EncodeToString(bytes[:])
}
