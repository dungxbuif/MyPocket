package main

import (
	"log"
	"net/http"

	"mypocket/internal/platform/config"
	"mypocket/internal/platform/httpapi"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{})
	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}
