package main

import (
	"context"
	"log"

	"mypocket/internal/platform/config"
	"mypocket/internal/platform/db"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	conn, err := db.Open(context.Background(), cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer conn.Close()

	log.Printf("worker ready in %s", cfg.AppEnv)
}
