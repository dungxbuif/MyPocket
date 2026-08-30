package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type Dependencies struct {
	ReadyCheck         func() error
	IdentityRepository IdentityRepository
	FinanceRepository  FinanceRepository
}

func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/google", startGoogleAuth(cfg))
	mux.HandleFunc("/api/v1/auth/google/callback", googleCallback(cfg, deps.IdentityRepository))
	mux.Handle("/api/v1/auth/logout", requireCSRF(http.HandlerFunc(logout)))
	mux.HandleFunc("/api/v1/me", currentUser(cfg, deps.IdentityRepository))
	mux.HandleFunc("/api/v1/wallets", wallets(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/categories", categories(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/health/live", liveHealth)
	mux.HandleFunc("/api/v1/health/ready", readyHealth(deps))

	return corsMiddleware(cfg, correlationMiddleware(mux))
}

func corsMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && origin == strings.TrimRight(cfg.PublicWebURL, "/") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Correlation-ID")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

type IdentityRepository interface {
	FindOrCreateGoogleUser(ctx context.Context, profile identity.GoogleProfile) (identity.User, error)
	FindByID(ctx context.Context, id string) (identity.User, error)
}

type FinanceRepository interface {
	ListWallets(ctx context.Context, userID string) ([]finance.Wallet, error)
	CreateWallet(ctx context.Context, userID string, input finance.CreateWalletInput) (finance.Wallet, error)
	ListCategories(ctx context.Context, userID string) ([]finance.Category, error)
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
