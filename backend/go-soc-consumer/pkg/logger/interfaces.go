package logger

import (
	"time"

	"go.uber.org/zap"
)

// CoreLogger provides basic logging functionality that all domain loggers need
type CoreLogger interface {
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Sugar() *zap.SugaredLogger
	Sync() error
}

// DeviceLogger handles device-related logging operations
type DeviceLogger interface {
	LogDeviceRegistration(macAddress, deviceName, ipAddress, location string, isUpdate bool)
	LogDeviceHealthCheck(macAddress, ipAddress string, isAlive bool, responseTime time.Duration, err error)
	LogDeviceStatus(macAddress, status string, fields ...zap.Field)
}

// MessagingLogger handles MQTT and NATS messaging logging
type MessagingLogger interface {
	LogMQTTMessage(topic string, payloadSize int, processingDuration time.Duration, success bool)
	LogEventPublishing(eventType, subject, eventID string, success bool, err error)
	LogMessageProcessing(protocol, topic string, success bool, fields ...zap.Field)
}

// InfrastructureLogger handles database and external API logging
type InfrastructureLogger interface {
	LogDatabaseOperation(operation, table string, duration time.Duration, recordsAffected int64, err error)
	LogExternalAPICall(service, endpoint string, statusCode int, duration time.Duration, err error)
	LogCacheOperation(operation, key string, hit bool, duration time.Duration, err error)
}

// PerformanceLogger handles performance monitoring and metrics
type PerformanceLogger interface {
	LogPerformanceMetrics(operation string, duration time.Duration, throughput float64, fields ...zap.Field)
	LogResourceUsage(component string, cpuPercent, memoryMB float64, fields ...zap.Field)
	LogThroughputMetrics(component string, requestsPerSecond float64, fields ...zap.Field)
}

// ApplicationLogger handles application lifecycle and general events
type ApplicationLogger interface {
	LogApplicationEvent(event string, component string, fields ...zap.Field)
	LogStartupEvent(component string, duration time.Duration, fields ...zap.Field)
	LogShutdownEvent(component string, duration time.Duration, fields ...zap.Field)
}

// LoggerFactory provides access to domain-specific loggers
type LoggerFactory interface {
	Device() DeviceLogger
	Messaging() MessagingLogger
	Infrastructure() InfrastructureLogger
	Performance() PerformanceLogger
	Application() ApplicationLogger
	Core() CoreLogger
}
