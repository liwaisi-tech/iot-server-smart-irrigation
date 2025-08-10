package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string
	Format      string
	Environment string // production, development, testing
}

// coreLogger implements the CoreLogger interface and serves as the foundation for all domain loggers
type coreLogger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// NewCoreLogger creates a new core logger instance that serves as the foundation for domain loggers
func NewCoreLogger(config LoggerConfig) (CoreLogger, error) {
	// Parse log level
	level := parseLogLevel(config.Level)

	// Create encoder config based on environment
	var encoderConfig zapcore.EncoderConfig
	switch strings.ToLower(config.Environment) {
	case "production":
		encoderConfig = zap.NewProductionEncoderConfig()
	case "development":
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	// Configure time encoding for better readability
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder based on format
	var encoder zapcore.Encoder
	switch strings.ToLower(config.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console", "text":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core with console output
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Add caller information and stack traces for errors
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &coreLogger{
		Logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// NewDefaultCoreLogger creates a logger with default production configuration
func NewDefaultCoreLogger() (CoreLogger, error) {
	return NewCoreLogger(LoggerConfig{
		Level:       "info",
		Format:      "json",
		Environment: "production",
	})
}

// NewDevelopmentCoreLogger creates a logger optimized for development
func NewDevelopmentCoreLogger() (CoreLogger, error) {
	return NewCoreLogger(LoggerConfig{
		Level:       "debug",
		Format:      "console",
		Environment: "development",
	})
}

// Sugar returns the sugared logger for flexible logging
func (l *coreLogger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Sync flushes any buffered log entries
func (l *coreLogger) Sync() error {
	return l.Logger.Sync()
}

// parseLogLevel converts string level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}