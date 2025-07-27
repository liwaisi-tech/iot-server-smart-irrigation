package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level       string `yaml:"level"`       // debug, info, warn, error
	Environment string `yaml:"environment"` // development, production
	OutputPaths []string `yaml:"output_paths,omitempty"`
	Encoding    string `yaml:"encoding"`    // json, console
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		Level:       "info",
		Environment: "production",
		OutputPaths: []string{"stdout"},
		Encoding:    "json",
	}
}

// NewLogger creates a new zap logger with the provided configuration
func NewLogger(config LoggerConfig) (*zap.Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create configuration based on environment
	var zapConfig zap.Config
	if config.Environment == "development" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Apply custom settings
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}

	if config.Encoding != "" {
		zapConfig.Encoding = config.Encoding
	}

	// Add caller information for better debugging
	zapConfig.DisableCaller = false
	zapConfig.DisableStacktrace = false

	// Build the logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // Skip one level to show actual caller
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewLoggerWithDefaults creates a new logger with default production settings
func NewLoggerWithDefaults() (*zap.Logger, error) {
	return NewLogger(DefaultConfig())
}
