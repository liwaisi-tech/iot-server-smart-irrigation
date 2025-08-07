package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string
	Format      string
	Environment string // production, development, testing
}

// IoTLogger wraps zap.Logger with IoT-specific structured logging methods
type IoTLogger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// NewLogger creates a new production-ready Zap logger with IoT-specific configuration
func NewLogger(config LoggerConfig) (*IoTLogger, error) {
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

	return &IoTLogger{
		Logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// NewDefaultLogger creates a logger with default production configuration
func NewDefaultLogger() (*IoTLogger, error) {
	return NewLogger(LoggerConfig{
		Level:       "info",
		Format:      "json",
		Environment: "production",
	})
}

// NewDevelopmentLogger creates a logger optimized for development
func NewDevelopmentLogger() (*IoTLogger, error) {
	return NewLogger(LoggerConfig{
		Level:       "debug",
		Format:      "console",
		Environment: "development",
	})
}

// Sugar returns the sugared logger for flexible logging
func (l *IoTLogger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Sync flushes any buffered log entries
func (l *IoTLogger) Sync() error {
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

// IoT-specific structured logging methods

// LogDeviceRegistration logs device registration events with structured fields
func (l *IoTLogger) LogDeviceRegistration(macAddress, deviceName, ipAddress, location string, isUpdate bool) {
	action := "device_registered"
	if isUpdate {
		action = "device_updated"
	}

	l.Info(action,
		zap.String("mac_address", macAddress),
		zap.String("device_name", deviceName),
		zap.String("ip_address", ipAddress),
		zap.String("location", location),
		zap.Bool("is_update", isUpdate),
		zap.String("component", "device_registration"),
	)
}

// LogMQTTMessage logs MQTT message processing with structured fields
func (l *IoTLogger) LogMQTTMessage(topic string, payloadSize int, processingDuration time.Duration, success bool) {
	level := l.Info
	message := "mqtt_message_processed"
	if !success {
		level = l.Error
		message = "mqtt_message_processing_failed"
	}

	level(message,
		zap.String("topic", topic),
		zap.Int("payload_size_bytes", payloadSize),
		zap.Duration("processing_duration", processingDuration),
		zap.Bool("success", success),
		zap.String("component", "mqtt_consumer"),
	)
}

// LogDatabaseOperation logs database operations with structured fields
func (l *IoTLogger) LogDatabaseOperation(operation, table string, duration time.Duration, recordsAffected int64, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.Int64("records_affected", recordsAffected),
		zap.String("component", "database"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("database_operation_failed", fields...)
	} else {
		l.Info("database_operation_completed", fields...)
	}
}

// LogEventPublishing logs event publishing operations
func (l *IoTLogger) LogEventPublishing(eventType, subject, eventID string, success bool, err error) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("subject", subject),
		zap.String("event_id", eventID),
		zap.Bool("success", success),
		zap.String("component", "event_publisher"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("event_publishing_failed", fields...)
	} else {
		l.Info("event_published", fields...)
	}
}

// LogDeviceHealthCheck logs device health checking operations
func (l *IoTLogger) LogDeviceHealthCheck(macAddress, ipAddress string, isAlive bool, responseTime time.Duration, err error) {
	fields := []zap.Field{
		zap.String("mac_address", macAddress),
		zap.String("ip_address", ipAddress),
		zap.Bool("is_alive", isAlive),
		zap.Duration("response_time", responseTime),
		zap.String("component", "device_health_checker"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Warn("device_health_check_error", fields...)
	} else {
		l.Debug("device_health_check_completed", fields...)
	}
}

// LogApplicationEvent logs application lifecycle events
func (l *IoTLogger) LogApplicationEvent(event string, component string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event", event),
		zap.String("component", component),
	}, fields...)

	l.Info("application_event", allFields...)
}

// LogPerformanceMetrics logs performance-related metrics
func (l *IoTLogger) LogPerformanceMetrics(operation string, duration time.Duration, throughput float64, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Float64("throughput", throughput),
		zap.String("component", "performance"),
	}, fields...)

	l.Info("performance_metrics", allFields...)
}