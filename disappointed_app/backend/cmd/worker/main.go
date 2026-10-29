package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mypocket/internal/agent"
	"mypocket/internal/audit"
	"mypocket/internal/lifecycle"
	"mypocket/internal/notification"
	"mypocket/internal/planning"
	"mypocket/internal/platform/config"
	"mypocket/internal/platform/db"
	"mypocket/internal/platform/logging"
	"mypocket/internal/platform/objectstore"
	"mypocket/internal/platform/ocr"
	openaiadapter "mypocket/internal/platform/openai"
	"mypocket/internal/portfolio"
	"mypocket/internal/worker"
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	portfolioRepo := portfolio.NewRepository(conn)
	auditRepo := audit.NewRepository(conn)
	runner := worker.RecurringRunner{Processor: planning.NewRepository(conn)}
	var pushDelivery notification.Delivery = notification.NoopDelivery{}
	if cfg.WebPushEnabled {
		pushDelivery, err = notification.NewWebPushDelivery(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject, nil)
		if err != nil {
			log.Fatalf("web push configuration error: %v", err)
		}
	}
	notificationProcessor := notification.Processor{Repo: notification.NewRepository(conn), Delivery: pushDelivery}
	portfolioPriceRunner, err := worker.NewPortfolioPriceRunnerFromEnv(portfolioRepo)
	if err != nil {
		log.Fatalf("portfolio price worker configuration error: %v", err)
	}
	auditRetentionRunner := worker.AuditRetentionRunner{Repo: auditRepo, RetentionDays: cfg.AuditRetentionDays}
	var lifecycleStore worker.LifecycleObjectStore
	if cfg.S3Endpoint != "" {
		lifecycleStore, err = objectstore.NewS3(cfg)
		if err != nil {
			log.Fatalf("lifecycle object store configuration error: %v", err)
		}
	}
	lifecycleRunner := worker.LifecycleRunner{Repo: lifecycle.NewRepository(conn), Store: lifecycleStore}
	var agentToolRunner worker.AgentToolRunner
	if cfg.OCR.Enabled {
		tool, toolErr := ocr.New(cfg.OCR.BaseURL, cfg.OCR.APIKey, cfg.OCR.Timeout)
		if toolErr != nil {
			log.Fatalf("OCR provider configuration error: %v", toolErr)
		}
		agentToolRunner = worker.AgentToolRunner{Repo: agent.NewRepository(conn), Store: lifecycleStore, Tool: tool, Owner: "worker-ocr", PollInterval: cfg.OCR.PollInterval, MaxProcessingTime: cfg.OCR.MaxProcessingTime}
	}
	var agentRunner worker.AgentRunner
	if cfg.AI.Enabled {
		model, modelErr := openaiadapter.New(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Model, cfg.AI.Timeout, cfg.AI.MaxRetries)
		if modelErr != nil {
			log.Fatalf("AI provider configuration error: %v", modelErr)
		}
		agentRunner = worker.AgentRunner{Repo: agent.NewRepository(conn), Model: model, Owner: "worker-agent"}
	}

	log.Printf("worker ready in %s", cfg.AppEnv)
	for {
		processed, err := runner.RunOnce(context.Background())
		if err != nil {
			log.Printf("recurring worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.recurring", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if processed > 0 {
			log.Printf("recurring worker created %d draft(s)", processed)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.recurring", Outcome: audit.OutcomeSuccess, Severity: audit.SeverityInfo, Source: audit.SourceWorker, Metadata: audit.SafeMetadata(map[string]any{"processed": processed})})
		}
		if _, err := notificationProcessor.RunOnce(context.Background()); err != nil {
			log.Printf("notification worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.notification", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		}
		if refreshed, err := portfolioPriceRunner.RunOnce(context.Background()); err != nil {
			log.Printf("portfolio price worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.portfolio_price_refresh", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if refreshed > 0 {
			log.Printf("portfolio price worker refreshed %d asset(s)", refreshed)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.portfolio_price_refresh", Outcome: audit.OutcomeSuccess, Severity: audit.SeverityInfo, Source: audit.SourceWorker, Metadata: audit.SafeMetadata(map[string]any{"refreshed": refreshed})})
		}
		if purged, err := auditRetentionRunner.RunOnce(context.Background()); err != nil {
			log.Printf("audit retention worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.audit_retention", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if purged > 0 {
			log.Printf("audit retention worker purged %d event(s)", purged)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.audit_retention", Outcome: audit.OutcomeSuccess, Severity: audit.SeverityInfo, Source: audit.SourceWorker, Metadata: audit.SafeMetadata(map[string]any{"count": purged})})
		}
		if processed, err := lifecycleRunner.RunOnce(context.Background()); err != nil {
			log.Printf("lifecycle worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.lifecycle", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if processed > 0 {
			log.Printf("lifecycle worker processed %d job(s)", processed)
		}
		if processed, err := agentToolRunner.RunOnce(context.Background()); err != nil {
			log.Printf("agent OCR worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.agent_ocr", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if processed {
			log.Printf("agent OCR worker processed one tool run")
		}
		if processed, err := agentRunner.RunOnce(context.Background()); err != nil {
			log.Printf("agent worker error: %v", err)
			_ = auditRepo.Append(context.Background(), audit.Event{CorrelationID: "worker", Action: "worker.agent", Outcome: audit.OutcomeFailure, Severity: audit.SeverityError, Source: audit.SourceWorker})
		} else if processed {
			log.Printf("agent worker processed one run")
		}
		select {
		case <-ticker.C:
		case <-stop:
			return
		}
	}
}
