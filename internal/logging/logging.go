package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a console slog.Logger with provided level string.
func New(level string) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelFromString(level),
	})
	return slog.New(handler)
}

func levelFromString(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "error":
		return slog.LevelError
	case "warn", "warning":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}
