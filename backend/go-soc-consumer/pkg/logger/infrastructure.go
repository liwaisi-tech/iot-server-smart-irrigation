package logger

import (
	"time"

	"go.uber.org/zap"
)

// infrastructureLogger implements InfrastructureLogger interface
type infrastructureLogger struct {
	CoreLogger
}

// NewInfrastructureLogger creates a new infrastructure logger with the given core logger
func NewInfrastructureLogger(core CoreLogger) InfrastructureLogger {
	return &infrastructureLogger{
		CoreLogger: core,
	}
}

// LogDatabaseOperation logs database operations with structured fields
func (l *infrastructureLogger) LogDatabaseOperation(operation, table string, duration time.Duration, recordsAffected int64, err error) {
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

// LogExternalAPICall logs external API calls with structured fields
func (l *infrastructureLogger) LogExternalAPICall(service, endpoint string, statusCode int, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("service", service),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("component", "external_api"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("external_api_call_failed", fields...)
	} else {
		l.Info("external_api_call_completed", fields...)
	}
}

// LogCacheOperation logs cache operations with structured fields
func (l *infrastructureLogger) LogCacheOperation(operation, key string, hit bool, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("cache_hit", hit),
		zap.Duration("duration", duration),
		zap.String("component", "cache"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("cache_operation_failed", fields...)
	} else {
		l.Debug("cache_operation_completed", fields...)
	}
}