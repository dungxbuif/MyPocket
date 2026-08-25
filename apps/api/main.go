package main

import (
	"context"
	"log"
	"net/http"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
	"mypocket/internal/platform/db"
	"mypocket/internal/platform/httpapi"
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

	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{
		IdentityRepository: identity.NewRepository(conn),
		ReadyCheck: func() error {
			return conn.PingContext(context.Background())
		},
	})
	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}
