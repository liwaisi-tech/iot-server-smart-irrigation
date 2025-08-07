package logger

import (
	"log/slog"
	"os"
	"strings"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string
	Format string
}

// NewLogger creates a new structured logger based on configuration
func NewLogger(config LoggerConfig) *slog.Logger {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Configure handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create handler based on format
	var handler slog.Handler
	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text", "console":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() *slog.Logger {
	return NewLogger(LoggerConfig{
		Level:  "info",
		Format: "text",
	})
}