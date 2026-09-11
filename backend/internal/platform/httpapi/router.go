package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/analytics"
	"mypocket/internal/audit"
	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/lifecycle"
	"mypocket/internal/notification"
	"mypocket/internal/planning"
	"mypocket/internal/platform/config"
	"mypocket/internal/portfolio"
	mysync "mypocket/internal/sync"
)

type Dependencies struct {
	ReadyCheck             func() error
	AuditRepository        AuditRepository
	AuthCache              AuthCache
	IdentityRepository     IdentityRepository
	APIKeyRepository       APIKeyRepository
	FinanceRepository      FinanceRepository
	ReceiptRepository      ReceiptRepository
	ObjectStore            ObjectStore
	PlanningRepository     PlanningRepository
	NotificationRepository NotificationRepository
	AnalyticsRepository    AnalyticsRepository
	PortfolioRepository    PortfolioRepository
	SyncService            SyncService
	LifecycleRepository    LifecycleRepository
	LifecycleObjectStore   LifecycleObjectStore
}

func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/google", startGoogleAuth(cfg, deps.AuditRepository))
	mux.HandleFunc("/api/v1/auth/google/callback", googleCallback(cfg, deps.IdentityRepository, deps.AuditRepository))
	mux.Handle("/api/v1/auth/logout", requireCSRF(http.HandlerFunc(logout)))
	mux.HandleFunc("/api/v1/me", currentUser(cfg, deps.IdentityRepository))
	mux.HandleFunc("/api/v1/api-keys", apiKeys(cfg, deps.APIKeyRepository, deps.AuditRepository, deps.AuthCache))
	mux.HandleFunc("/api/v1/api-keys/", apiKeyByID(cfg, deps.APIKeyRepository, deps.AuditRepository, deps.AuthCache))
	mux.HandleFunc("/api/v1/wallets", wallets(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/wallets/", walletByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/categories", categories(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/categories/", categoryByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/transactions", transactions(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/transactions/", transactionByID(cfg, deps.FinanceRepository))
	mux.HandleFunc("/api/v1/files/presign", filePresign(cfg, deps.ReceiptRepository, deps.ObjectStore))
	mux.HandleFunc("/api/v1/files/", fileDownload(cfg, deps.ReceiptRepository, deps.ObjectStore))
	mux.HandleFunc("/api/v1/budgets", budgets(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/budgets/", budgetByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/events", events(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/events/", eventByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/obligations", obligations(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/obligations/", obligationByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/recurring-schedules", recurringSchedules(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/recurring-schedules/", recurringScheduleByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/transaction-drafts", transactionDrafts(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/transaction-drafts/", transactionDraftByID(cfg, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/notifications", notifications(cfg, deps.NotificationRepository))
	mux.HandleFunc("/api/v1/notifications/", notificationByID(cfg, deps.NotificationRepository))
	mux.HandleFunc("/api/v1/push-subscriptions", pushSubscriptions(cfg, deps.NotificationRepository))
	mux.HandleFunc("/api/v1/push-subscriptions/", pushSubscriptionByID(cfg, deps.NotificationRepository))
	mux.HandleFunc("/api/v1/dashboard", dashboard(cfg, deps.FinanceRepository, deps.AnalyticsRepository))
	mux.HandleFunc("/api/v1/assets", assets(cfg, deps.PortfolioRepository))
	mux.HandleFunc("/api/v1/assets/", assetByID(cfg, deps.PortfolioRepository))
	mux.HandleFunc("/api/v1/portfolio/summary", portfolioSummary(cfg, deps.PortfolioRepository))
	mux.HandleFunc("/api/v1/reports/", reports(cfg, deps.AnalyticsRepository))
	mux.HandleFunc("/api/v1/search", search(cfg, deps.FinanceRepository, deps.PlanningRepository))
	mux.HandleFunc("/api/v1/sync/mutations", syncMutations(cfg, deps.SyncService))
	mux.HandleFunc("/api/v1/sync/changes", syncChanges(cfg, deps.SyncService))
	mux.HandleFunc("/api/v1/sync/resync", syncResync(cfg, deps.SyncService))
	mux.HandleFunc("/api/v1/audit/events", auditEvents(cfg, deps.IdentityRepository, deps.AuditRepository))
	mux.HandleFunc("/api/v1/audit/access", auditAccess(cfg, deps.IdentityRepository))
	mux.HandleFunc("/api/v1/imports", imports(cfg, deps.LifecycleRepository, deps.LifecycleObjectStore))
	mux.HandleFunc("/api/v1/imports/", importByID(cfg, deps.LifecycleRepository))
	mux.HandleFunc("/api/v1/exports", exports(cfg, deps.LifecycleRepository))
	mux.HandleFunc("/api/v1/exports/", exportByID(cfg, deps.LifecycleRepository, deps.LifecycleObjectStore))
	mux.HandleFunc("/api/v1/account/reset", destructiveAccount(cfg, deps.LifecycleRepository, lifecycle.KindReset))
	mux.HandleFunc("/api/v1/account/delete", destructiveAccount(cfg, deps.LifecycleRepository, lifecycle.KindDelete))
	mux.HandleFunc("/api/v1/account/jobs/", accountJob(cfg, deps.LifecycleRepository))
	mux.HandleFunc("/api/v1/health/live", liveHealth)
	mux.HandleFunc("/api/v1/health/ready", readyHealth(deps))
	mux.HandleFunc("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently).ServeHTTP)
	mux.HandleFunc("/docs/", DocsHandler(cfg))

	return corsMiddleware(cfg, correlationMiddleware(authContextMiddleware(cfg, deps.IdentityRepository, deps.APIKeyRepository, deps.AuthCache, deps.AuditRepository, observabilityMiddleware(cfg, deps.AuditRepository, mux))))
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
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-CSRF-Token, X-Correlation-ID, Idempotency-Key")
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

type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, userID string, name string, secret string) (identity.CreatedAPIKey, error)
	ListAPIKeys(ctx context.Context, userID string) ([]identity.APIKey, error)
	RevokeAPIKey(ctx context.Context, userID string, keyID string) (identity.APIKey, error)
	AuthenticateAPIKey(ctx context.Context, plaintext string, secret string) (identity.User, identity.APIKey, error)
}

type AuthCache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type ReceiptRepository interface {
	CreateReceiptObject(ctx context.Context, userID string, input finance.CreateReceiptObjectInput) (finance.ReceiptObject, error)
	GetReceiptObject(ctx context.Context, userID, receiptID string) (finance.ReceiptObject, error)
}

type ObjectStore interface {
	PresignPut(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

type LifecycleObjectStore interface {
	PutObject(context.Context, string, string, io.Reader, int64) error
	GetObject(context.Context, string, int64) (io.ReadCloser, int64, error)
	PresignGet(context.Context, string, time.Duration) (string, error)
}

type LifecycleRepository interface {
	CreateJob(context.Context, string, lifecycle.Kind, string, any) (lifecycle.Job, error)
	GetJob(context.Context, string, string) (lifecycle.Job, error)
	ConfirmImport(context.Context, string, string, int64) (lifecycle.Job, error)
	Counts(context.Context, string) (map[string]int, error)
	DisableUser(context.Context, string) error
}

type AuditRepository interface {
	Append(ctx context.Context, event audit.Event) error
	List(ctx context.Context, query audit.Query) ([]audit.Event, error)
	PurgeExpired(ctx context.Context, retentionDays int, limit int) (int, error)
}

type FinanceRepository interface {
	ListWallets(ctx context.Context, userID string) ([]finance.Wallet, error)
	CreateWallet(ctx context.Context, userID string, input finance.CreateWalletInput) (finance.Wallet, error)
	UpdateWallet(ctx context.Context, userID string, walletID string, input finance.UpdateWalletInput) (finance.Wallet, error)
	ArchiveWallet(ctx context.Context, userID string, walletID string, baseVersion int64) error
	SetDefaultAIWallet(ctx context.Context, userID string, walletID string, baseVersion int64) error
	ListCategories(ctx context.Context, userID string) ([]finance.Category, error)
	CreateCategory(ctx context.Context, userID string, input finance.CreateCategoryInput) (finance.Category, error)
	UpdateCategory(ctx context.Context, userID string, categoryID string, input finance.UpdateCategoryInput) (finance.Category, error)
	ArchiveCategory(ctx context.Context, userID string, categoryID string, baseVersion int64) error
	SetWalletCategoryActive(ctx context.Context, userID string, walletID string, categoryID string, active bool) error
	ListWalletCategorySettings(ctx context.Context, userID string, walletID string) ([]finance.WalletCategorySetting, error)
	ListTransactions(ctx context.Context, userID string, filters finance.TransactionFilters) ([]finance.Transaction, error)
	CreateTransaction(ctx context.Context, userID string, input finance.CreateTransactionInput) (finance.Transaction, error)
	UpdateTransaction(ctx context.Context, userID string, transactionID string, input finance.UpdateTransactionInput) (finance.Transaction, error)
	ArchiveTransaction(ctx context.Context, userID string, transactionID string, baseVersion int64) error
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
	ArchiveBudget(ctx context.Context, userID string, budgetID string, baseVersion int64) error
	ListEvents(ctx context.Context, userID string) ([]planning.EventSummary, error)
	CreateEvent(ctx context.Context, userID string, input planning.CreateEventInput) (planning.EventSummary, error)
	UpdateEvent(ctx context.Context, userID string, eventID string, input planning.UpdateEventInput) (planning.EventSummary, error)
	ArchiveEvent(ctx context.Context, userID string, eventID string, baseVersion int64) error
	LinkEventTransaction(ctx context.Context, userID string, eventID string, transactionID string) error
	ListObligations(ctx context.Context, userID string) ([]planning.ObligationSummary, error)
	CreateObligation(ctx context.Context, userID string, input planning.CreateObligationInput) (planning.ObligationSummary, error)
	UpdateObligation(ctx context.Context, userID string, obligationID string, input planning.UpdateObligationInput) (planning.ObligationSummary, error)
	ArchiveObligation(ctx context.Context, userID string, obligationID string, baseVersion int64) error
	LinkObligationRepayment(ctx context.Context, userID string, obligationID string, transactionID string) error
	ListRecurringSchedules(ctx context.Context, userID string) ([]planning.RecurringSchedule, error)
	CreateRecurringSchedule(ctx context.Context, userID string, input planning.CreateRecurringScheduleInput) (planning.RecurringSchedule, error)
	ArchiveRecurringSchedule(ctx context.Context, userID string, scheduleID string, baseVersion int64) error
	ListTransactionDrafts(ctx context.Context, userID string) ([]planning.TransactionDraft, error)
	ConfirmTransactionDraft(ctx context.Context, userID string, draftID string, input planning.ConfirmTransactionDraftInput) (planning.TransactionDraftDecision, error)
	RejectTransactionDraft(ctx context.Context, userID string, draftID string, input planning.RejectTransactionDraftInput) (planning.TransactionDraftDecision, error)
}

type NotificationRepository interface {
	List(ctx context.Context, userID string, limit int, before time.Time) ([]notification.Notice, bool, error)
	MarkRead(ctx context.Context, userID, id string) error
	UpsertPushSubscription(ctx context.Context, userID string, input notification.CreatePushSubscriptionInput) (notification.PushSubscription, error)
	DeletePushSubscription(ctx context.Context, userID, id string) error
}

type AnalyticsRepository interface {
	Summary(ctx context.Context, userID string, filter analytics.Filter) (analytics.Summary, error)
	Categories(ctx context.Context, userID string, filter analytics.Filter) ([]analytics.CategoryTotal, error)
	Daily(ctx context.Context, userID string, filter analytics.Filter) ([]analytics.DailyTotal, error)
	Dashboard(ctx context.Context, userID string, filter analytics.Filter) (analytics.Dashboard, error)
	Insider(ctx context.Context, userID string, filter analytics.Filter) (analytics.InsiderReport, error)
}

type PortfolioRepository interface {
	ListPositions(ctx context.Context, userID string, includeArchived bool) ([]portfolio.Position, error)
	CreatePosition(ctx context.Context, userID string, input portfolio.CreatePositionInput) (portfolio.Position, error)
	GetPosition(ctx context.Context, userID, assetID string) (portfolio.Position, error)
	ArchivePosition(ctx context.Context, userID, assetID string, baseVersion int64) error
	AddTrade(ctx context.Context, userID, assetID string, input portfolio.AddTradeInput) (portfolio.Position, error)
	UpdateTrade(ctx context.Context, userID, assetID, tradeID string, input portfolio.UpdateTradeInput) (portfolio.Position, error)
	ArchiveTrade(ctx context.Context, userID, assetID, tradeID string, baseVersion int64) (portfolio.Position, error)
	AddPrice(ctx context.Context, userID, assetID string, input portfolio.AddPriceInput) (portfolio.Position, error)
	Summary(ctx context.Context, userID string) (portfolio.PortfolioSummary, error)
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

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

func observabilityMiddleware(cfg config.Config, auditRepo AuditRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		correlation := correlationID(r.Context())
		defer func() {
			if recovered := recover(); recovered != nil {
				recorder.status = http.StatusInternalServerError
				slog.Error("http panic recovered", "correlation_id", correlation, "method", r.Method, "path", safeRequestPath(r), "panic", "redacted")
				writeJSON(recorder, http.StatusInternalServerError, ErrorEnvelope("INTERNAL_FAILURE", "A temporary internal error occurred", correlation))
				appendAudit(r.Context(), auditRepo, audit.Event{
					CorrelationID: correlation,
					ActorUserID:   authenticatedUserIDFromRequest(r, cfg),
					Action:        "http.panic",
					EntityType:    entityTypeFromPath(r.URL.Path),
					EntityID:      entityIDFromPath(r.URL.Path),
					Outcome:       audit.OutcomeFailure,
					Severity:      audit.SeverityError,
					Source:        audit.SourceAPI,
					ErrorCode:     "INTERNAL_FAILURE",
					RequestMethod: r.Method,
					RequestPath:   safeRequestPath(r),
					IPHash:        audit.HashValue(cfg.AuditHashSecret, clientIP(r)),
					UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent()),
				})
			}
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			duration := time.Since(start)
			actorUserID := authenticatedUserIDFromRequest(r, cfg)
			slog.Info("http request", "correlation_id", correlation, "method", r.Method, "path", safeRequestPath(r), "status", status, "duration_ms", duration.Milliseconds(), "bytes", recorder.bytes, "actor_hash", audit.HashValue(cfg.AuditHashSecret, actorUserID))
			if shouldAuditRequest(r) {
				appendAudit(r.Context(), auditRepo, audit.Event{
					CorrelationID: correlation,
					ActorUserID:   actorUserID,
					Action:        requestAction(r),
					EntityType:    entityTypeFromPath(r.URL.Path),
					EntityID:      entityIDFromPath(r.URL.Path),
					Outcome:       outcomeFromStatus(status),
					Severity:      severityFromStatus(status),
					Source:        audit.SourceAPI,
					RequestMethod: r.Method,
					RequestPath:   safeRequestPath(r),
					IPHash:        audit.HashValue(cfg.AuditHashSecret, clientIP(r)),
					UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent()),
					Metadata:      audit.SafeMetadata(map[string]any{"status": status}),
				})
			}
		}()
		next.ServeHTTP(recorder, r)
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
