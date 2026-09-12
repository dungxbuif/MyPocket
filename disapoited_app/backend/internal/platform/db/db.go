package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/platform/config"
)

func Open(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	conn, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return conn, nil
}
