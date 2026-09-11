package main

import (
	"context"
	"log"
	"net/http"

	"mypocket/internal/analytics"
	"mypocket/internal/audit"
	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/lifecycle"
	"mypocket/internal/notification"
	"mypocket/internal/planning"
	"mypocket/internal/platform/authcache"
	"mypocket/internal/platform/config"
	"mypocket/internal/platform/db"
	"mypocket/internal/platform/httpapi"
	"mypocket/internal/platform/logging"
	"mypocket/internal/platform/objectstore"
	"mypocket/internal/portfolio"
	mysync "mypocket/internal/sync"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}
	logging.Configure(cfg)

	conn, err := db.Open(context.Background(), cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer conn.Close()

	financeRepo := finance.NewRepository(conn)
	portfolioRepo := portfolio.NewRepository(conn)
	auditRepo := audit.NewRepository(conn)
	identityRepo := identity.NewRepository(conn)
	cache, err := authcache.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("auth cache configuration error: %v", err)
	}
	if cache != nil {
		defer cache.Close()
	}
	var objectStore httpapi.ObjectStore
	var lifecycleStore httpapi.LifecycleObjectStore
	if cfg.S3Endpoint != "" {
		store, storeErr := objectstore.NewS3(cfg)
		if storeErr != nil {
			log.Fatalf("object store configuration error: %v", storeErr)
		}
		objectStore = store
		lifecycleStore = store
	}
	syncService := mysync.NewService(mysync.NewRepository(conn), financeRepo, portfolioRepo)
	syncService.SetAuditSink(auditRepo)
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{
		AuditRepository:        auditRepo,
		AuthCache:              cache,
		IdentityRepository:     identityRepo,
		APIKeyRepository:       identityRepo,
		FinanceRepository:      financeRepo,
		ReceiptRepository:      financeRepo,
		ObjectStore:            objectStore,
		PlanningRepository:     planning.NewRepository(conn),
		NotificationRepository: notification.NewRepository(conn),
		AnalyticsRepository:    analytics.NewRepository(conn),
		PortfolioRepository:    portfolioRepo,
		SyncService:            syncService,
		LifecycleRepository:    lifecycle.NewRepository(conn),
		LifecycleObjectStore:   lifecycleStore,
		ReadyCheck: func() error {
			return conn.PingContext(context.Background())
		},
	})
	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}
