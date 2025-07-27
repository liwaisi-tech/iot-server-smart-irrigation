package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.Level != "info" {
		t.Errorf("Expected default level to be 'info', got '%s'", config.Level)
	}
	
	if config.Environment != "production" {
		t.Errorf("Expected default environment to be 'production', got '%s'", config.Environment)
	}
	
	if config.Encoding != "json" {
		t.Errorf("Expected default encoding to be 'json', got '%s'", config.Encoding)
	}
	
	if len(config.OutputPaths) != 1 || config.OutputPaths[0] != "stdout" {
		t.Errorf("Expected default output paths to be ['stdout'], got %v", config.OutputPaths)
	}
}

func TestNewLoggerWithDefaults(t *testing.T) {
	logger, err := NewLoggerWithDefaults()
	if err != nil {
		t.Fatalf("Failed to create logger with defaults: %v", err)
	}
	
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
	
	// Test that logger can be used
	logger.Info("Test log message")
	logger.Sync()
}

func TestNewLoggerWithCustomConfig(t *testing.T) {
	config := LoggerConfig{
		Level:       "debug",
		Environment: "development",
		OutputPaths: []string{"stdout"},
		Encoding:    "console",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger with custom config: %v", err)
	}
	
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
	
	// Test that logger can be used
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Sync()
}

func TestNewLoggerWithInvalidLevel(t *testing.T) {
	config := LoggerConfig{
		Level:       "invalid",
		Environment: "production",
		OutputPaths: []string{"stdout"},
		Encoding:    "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Logger creation should not fail with invalid level: %v", err)
	}
	
	if logger == nil {
		t.Error("Expected logger to be non-nil even with invalid level")
	}
	
	// Should default to info level
	logger.Info("This should work with default level")
	logger.Sync()
}

func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		config := LoggerConfig{
			Level:       level,
			Environment: "production",
			OutputPaths: []string{"stdout"},
			Encoding:    "json",
		}
		
		logger, err := NewLogger(config)
		if err != nil {
			t.Fatalf("Failed to create logger with level '%s': %v", level, err)
		}
		
		if logger == nil {
			t.Errorf("Expected logger to be non-nil for level '%s'", level)
		}
		
		logger.Info("Test message", zapcore.Field{Key: "level", String: level})
		logger.Sync()
	}
}