package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/planning"
	"mypocket/internal/platform/config"
	mysync "mypocket/internal/sync"
)

type Dependencies struct {
	ReadyCheck         func() error
	IdentityRepository IdentityRepository
	FinanceRepository  FinanceRepository
	PlanningRepository PlanningRepository
	SyncService        SyncService
}

func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/google", startGoogleAuth(cfg))
	mux.HandleFunc("/api/v1/auth/google/callback", googleCallback(cfg, deps.IdentityRepository))
	mux.Handle("/api/v1/auth/logout", requireCSRF(http.HandlerFunc(logout)))
	mux.HandleFunc("/api/v1/me", currentUser(cfg, deps.IdentityRepository))
	mux.HandleFunc("/api/v1/wallets", wallets(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/wallets/", walletByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/categories", categories(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/categories/", categoryByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/transactions", transactions(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/transactions/", transactionByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/budgets", budgets(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/budgets/", budgetByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/sync/mutations", syncMutations(cfg, deps.SyncService))
	mux.HandleFunc("/api/v1/sync/changes", syncChanges(cfg, deps.SyncService))
	mux.HandleFunc("/api/v1/sync/resync", syncResync(cfg, deps.SyncService))
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
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Correlation-ID, Idempotency-Key")
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
	UpdateWallet(ctx context.Context, userID string, walletID string, input finance.UpdateWalletInput) (finance.Wallet, error)
	ArchiveWallet(ctx context.Context, userID string, walletID string) error
	SetDefaultAIWallet(ctx context.Context, userID string, walletID string) error
	ListCategories(ctx context.Context, userID string) ([]finance.Category, error)
	CreateCategory(ctx context.Context, userID string, input finance.CreateCategoryInput) (finance.Category, error)
	UpdateCategory(ctx context.Context, userID string, categoryID string, input finance.UpdateCategoryInput) (finance.Category, error)
	ArchiveCategory(ctx context.Context, userID string, categoryID string) error
	SetWalletCategoryActive(ctx context.Context, userID string, walletID string, categoryID string, active bool) error
	ListTransactions(ctx context.Context, userID string, filters finance.TransactionFilters) ([]finance.Transaction, error)
	CreateTransaction(ctx context.Context, userID string, input finance.CreateTransactionInput) (finance.Transaction, error)
	UpdateTransaction(ctx context.Context, userID string, transactionID string, input finance.UpdateTransactionInput) (finance.Transaction, error)
	ArchiveTransaction(ctx context.Context, userID string, transactionID string) error
}

type SyncService interface {
	ApplyMutations(ctx context.Context, userID string, mutations []mysync.Mutation) ([]mysync.MutationResult, error)
	Changes(ctx context.Context, userID string, after int64, limit int) (mysync.ChangesResult, error)
	Resync(ctx context.Context, userID string) (mysync.Snapshot, error)
}

type PlanningRepository interface {
	ListBudgetProgress(ctx context.Context, userID string, now time.Time) ([]planning.BudgetProgress, error)
	CreateBudget(ctx context.Context, userID string, input planning.CreateBudgetInput) (planning.Budget, error)
	UpdateBudget(ctx context.Context, userID string, budgetID string, input planning.UpdateBudgetInput) (planning.Budget, error)
	ArchiveBudget(ctx context.Context, userID string, budgetID string) error
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
