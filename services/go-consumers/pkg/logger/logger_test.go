package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "json", config.Encoding)
	assert.Len(t, config.OutputPaths, 1)
	assert.Equal(t, "stdout", config.OutputPaths[0])
}

func TestNewLoggerWithDefaults(t *testing.T) {
	logger, err := NewLoggerWithDefaults()
	require.NoError(t, err)
	assert.NotNil(t, logger)
	
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
	require.NoError(t, err)
	assert.NotNil(t, logger)
	
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
	require.NoError(t, err, "Logger creation should not fail with invalid level")
	assert.NotNil(t, logger, "Expected logger to be non-nil even with invalid level")
	
	// Should default to info level
	logger.Info("This should work with default level")
	logger.Sync()
}

func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			config := LoggerConfig{
				Level:       level,
				Environment: "production",
				OutputPaths: []string{"stdout"},
				Encoding:    "json",
			}
			
			logger, err := NewLogger(config)
			require.NoError(t, err, "Failed to create logger with level '%s'", level)
			assert.NotNil(t, logger, "Expected logger to be non-nil for level '%s'", level)
			
			logger.Info("Test message", zap.String("level", level))
			logger.Sync()
		})
	}
}