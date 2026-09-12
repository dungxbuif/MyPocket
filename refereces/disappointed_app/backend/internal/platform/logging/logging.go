package logging

import (
	"log/slog"
	"os"
	"strings"

	"mypocket/internal/platform/config"
)

func Configure(cfg config.Config) {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	options := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, options)))
		return
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, options)))
}
