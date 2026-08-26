package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
