package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/db"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
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
	for {
		if err := runRecurringPass(database, runner, time.Now().UTC()); err != nil {
			log.Printf("recurring worker pass failed: %v", err)
		}
		if once {
			return
		}
		time.Sleep(interval)
	}
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
