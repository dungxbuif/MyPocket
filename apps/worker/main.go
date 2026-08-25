package main

import (
	"log"

	"mypocket/internal/platform/config"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	log.Printf("worker ready in %s", cfg.AppEnv)
}
