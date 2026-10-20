package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/cache"
	"github.com/mypocket/backend/internal/infrastructure/db"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/infrastructure/storage"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	database, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("worker database connection failed: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("worker database handle failed: %v", err)
	}
	defer sqlDB.Close()

	interval := workerInterval()
	once := workerOnce()
	runner := repo.NewRecurringPostgresRepository(database)
	budgetSeries := repo.NewBudgetSeriesPostgresRepository(database)
	auditOutbox := repo.NewAuditOutboxPostgresRepository(database)
	auditRedis, redisErr := cache.NewRedis(cfg.RedisURL)
	if redisErr != nil {
		log.Printf("worker audit redis unavailable; durable outbox will retry: %v", redisErr)
	}
	defer func() {
		if auditRedis != nil {
			_ = auditRedis.Close()
		}
	}()
	advisorRepository := repo.NewAdvisorPostgresRepository(database)
	advisorExpiry, _ := advisorRepository.(repository.AdvisorExpiryRepository)
	advisorRuntime, _ := advisorRepository.(repository.AdvisorRuntimeRepository)
	reaperCtx, stopReaper := context.WithCancel(context.Background())
	defer stopReaper()
	stopAdvisorReaper := func() {}
	if cfg.AIAdvisorRunLeasesEnabled {
		stopAdvisorReaper = startAdvisorReaper(reaperCtx, advisorRuntime)
	} else {
		log.Printf("advisor runtime reaper disabled; set AI_ADVISOR_RUN_LEASES_ENABLED=true after migration")
	}
	defer stopAdvisorReaper()
	attachmentCleanup := attachmentCleanupRunner(nil)
	if privateStorage, storageErr := storage.NewS3(storage.S3Config{
		Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket,
		Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle,
	}); storageErr != nil {
		log.Printf("attachment cleanup unavailable; configure private storage to enable scheduled cleanup: %v", storageErr)
	} else {
		attachmentCleanup = &usecase.AIEntryAttachmentCleanup{
			Entries: repo.NewAIEntryPostgresRepository(database), Storage: privateStorage,
		}
	}
	for {
		if err := runRecurringPass(database, runner, time.Now().UTC()); err != nil {
			log.Printf("recurring worker pass failed: %v", err)
		}
		if err := runBudgetSeriesPass(database, budgetSeries, time.Now().UTC()); err != nil {
			log.Printf("budget series worker pass failed: %v", err)
		}
		if err := runBudgetNotificationPass(database, time.Now().UTC()); err != nil {
			log.Printf("budget notification worker pass failed: %v", err)
		}
		if advisorExpiry != nil {
			if _, err := runAdvisorExpiryPass(context.Background(), advisorExpiry, time.Now().UTC()); err != nil {
				log.Printf("advisor expiry worker pass failed: %v", err)
			}
		}
		if attachmentCleanup != nil {
			deleted, cleanupErr := runAttachmentCleanupPass(context.Background(), attachmentCleanup)
			if cleanupErr != nil {
				log.Printf("attachment cleanup pass failed after %d deletion(s): %v", deleted, cleanupErr)
			} else if deleted > 0 {
				log.Printf("attachment cleanup pass deleted %d expired unlinked object(s)", deleted)
			}
		}
		if auditRedis == nil {
			if candidate, reconnectErr := cache.NewRedis(cfg.RedisURL); reconnectErr == nil {
				auditRedis = candidate
			} else {
				log.Printf("audit outbox redis reconnect failed: %v", reconnectErr)
			}
		}
		if auditRedis != nil {
			if _, err := runAuditOutboxPass(context.Background(), auditOutbox, auditRedis, time.Now().UTC()); err != nil {
				log.Printf("audit outbox worker pass failed: %v", err)
			}
		}
		if once {
			return
		}
		time.Sleep(interval)
	}
}

type attachmentCleanupRunner interface {
	Run(context.Context, int) (int, error)
}

func runAttachmentCleanupPass(ctx context.Context, runner attachmentCleanupRunner) (int, error) {
	if runner == nil {
		return 0, nil
	}
	return runner.Run(ctx, 100)
}

func runBudgetSeriesPass(database *gorm.DB, runner repository.BudgetSeriesRepository, now time.Time) error {
	var owners []string
	if err := database.Model(&entity.BudgetSeries{}).Where("active = true").Distinct("owner_id").Pluck("owner_id", &owners).Error; err != nil {
		return err
	}
	var first error
	for _, owner := range owners {
		if _, err := runner.Ensure(owner, now); err != nil && first == nil {
			first = fmt.Errorf("one or more owner budget series failed: %w", err)
		}
	}
	return first
}

func runAdvisorExpiryPass(ctx context.Context, runner repository.AdvisorExpiryRepository, now time.Time) (int, error) {
	if runner == nil {
		return 0, nil
	}
	return runner.ExpireIdleBatch(ctx, now, 100)
}

func runRecurringPass(database *gorm.DB, runner interface {
	RunDue(string, time.Time) (int, error)
}, now time.Time) error {
	var owners []string
	if err := database.Model(&entity.RecurringSchedule{}).Where("active = true").Distinct("owner_id").Pluck("owner_id", &owners).Error; err != nil {
		return err
	}
	var firstErr error
	for _, ownerID := range owners {
		if _, err := runner.RunDue(ownerID, now); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("one or more owner schedules failed: %w", err)
			}
		}
	}
	return firstErr
}

func workerInterval() time.Duration {
	seconds, err := strconv.Atoi(os.Getenv("WORKER_INTERVAL_SECONDS"))
	if err != nil || seconds <= 0 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

func workerOnce() bool {
	value, err := strconv.ParseBool(os.Getenv("WORKER_ONCE"))
	return err == nil && value
}
