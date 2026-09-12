package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/platform/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("configuration error: missing required config: DATABASE_URL")
	}

	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer conn.Close()

	if err := conn.PingContext(context.Background()); err != nil {
		log.Fatalf("database error: %v", err)
	}
	if err := db.Migrate(context.Background(), conn, os.DirFS("migrations")); err != nil {
		log.Fatalf("migration error: %v", err)
	}
	fmt.Println("migrations applied")
}
