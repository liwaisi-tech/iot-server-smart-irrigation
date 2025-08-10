package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerFactory(t *testing.T) {
	t.Run("NewLoggerFactory should create domain-specific loggers", func(t *testing.T) {
		config := LoggerConfig{
			Level:       "info",
			Format:      "json",
			Environment: "production",
		}

		factory, err := NewLoggerFactory(config)
		require.NoError(t, err)
		require.NotNil(t, factory)

		// Test all domain-specific loggers are created
		assert.NotNil(t, factory.Core())
		assert.NotNil(t, factory.Device())
		assert.NotNil(t, factory.Sensor())
		assert.NotNil(t, factory.Messaging())
		assert.NotNil(t, factory.Infrastructure())
		assert.NotNil(t, factory.Performance())
		assert.NotNil(t, factory.Application())
	})

	t.Run("NewDefault should create factory with production config", func(t *testing.T) {
		factory, err := NewDefault()
		require.NoError(t, err)
		require.NotNil(t, factory)

		// Test all domain-specific loggers are created
		assert.NotNil(t, factory.Core())
		assert.NotNil(t, factory.Device())
		assert.NotNil(t, factory.Sensor())
		assert.NotNil(t, factory.Messaging())
		assert.NotNil(t, factory.Infrastructure())
		assert.NotNil(t, factory.Performance())
		assert.NotNil(t, factory.Application())
	})

	t.Run("NewDevelopment should create factory with dev config", func(t *testing.T) {
		factory, err := NewDevelopment()
		require.NoError(t, err)
		require.NotNil(t, factory)

		// Test all domain-specific loggers are created
		assert.NotNil(t, factory.Core())
		assert.NotNil(t, factory.Device())
		assert.NotNil(t, factory.Sensor())
		assert.NotNil(t, factory.Messaging())
		assert.NotNil(t, factory.Infrastructure())
		assert.NotNil(t, factory.Performance())
		assert.NotNil(t, factory.Application())
	})

	t.Run("Domain loggers should have expected methods", func(t *testing.T) {
		factory, err := NewDevelopment()
		require.NoError(t, err)

		// Test device logger methods
		deviceLogger := factory.Device()
		assert.NotNil(t, deviceLogger)
		deviceLogger.LogDeviceRegistration("00:11:22:33:44:55", "TestDevice", "192.168.1.100", "Living Room", false)

		// Test sensor logger methods
		sensorLogger := factory.Sensor()
		assert.NotNil(t, sensorLogger)
		sensorLogger.LogSensorData("00:11:22:33:44:55", 25.5, 60.2, false)

		// Test messaging logger methods
		messagingLogger := factory.Messaging()
		assert.NotNil(t, messagingLogger)
		messagingLogger.LogMQTTMessage("/test/topic", 100, 1000000, true)

		// Test infrastructure logger methods
		infraLogger := factory.Infrastructure()
		assert.NotNil(t, infraLogger)
		infraLogger.LogDatabaseOperation("SELECT", "devices", 1500000, 1, nil)

		// Test performance logger methods
		perfLogger := factory.Performance()
		assert.NotNil(t, perfLogger)
		perfLogger.LogPerformanceMetrics("test_operation", 2000000, 1000.0)

		// Test application logger methods
		appLogger := factory.Application()
		assert.NotNil(t, appLogger)
		appLogger.LogApplicationEvent("test_event", "test_component")

		// Test core logger methods
		coreLogger := factory.Core()
		assert.NotNil(t, coreLogger)
		coreLogger.Info("test message")
	})
}