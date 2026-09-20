package meta

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(level)}))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Fatal(logger *slog.Logger, message string, args ...any) {
	logger.Error(message, args...)
	os.Exit(1)
}
