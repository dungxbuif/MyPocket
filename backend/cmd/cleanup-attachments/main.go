package main

import (
	"context"
	"log"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/infrastructure/db"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/infrastructure/storage"
	"github.com/mypocket/backend/internal/usecase"
)

func main() {
	if err := config.LoadLocalEnv(".env.local"); err != nil {
		log.Fatal("load local environment failed")
	}
	cfg := config.Load()
	database, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("connect database failed")
	}
	s3, err := storage.NewS3(storage.S3Config{Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket, Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle})
	if err != nil {
		log.Fatal("configure private attachment storage failed")
	}
	cleaner := usecase.AIEntryAttachmentCleanup{Entries: repo.NewAIEntryPostgresRepository(database), Storage: s3}
	deleted, err := cleaner.Run(context.Background(), 100)
	if err != nil {
		log.Fatalf("attachment cleanup finished with errors; successfully deleted %d object(s)", deleted)
	}
	log.Printf("attachment cleanup deleted %d expired unlinked object(s)", deleted)
}
