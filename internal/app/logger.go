package app

import (
	"log/slog"
	"os"
)

func InitLogger(logLevel string) {
	level := slog.LevelDebug

	switch logLevel {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: true}))
	slog.SetDefault(logger)
}
