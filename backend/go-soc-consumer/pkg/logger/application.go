package logger

import (
	"time"

	"go.uber.org/zap"
)

// applicationLogger implements ApplicationLogger interface
type applicationLogger struct {
	CoreLogger
}

// NewApplicationLogger creates a new application logger with the given core logger
func NewApplicationLogger(core CoreLogger) ApplicationLogger {
	return &applicationLogger{
		CoreLogger: core,
	}
}

// LogApplicationEvent logs application lifecycle events
func (l *applicationLogger) LogApplicationEvent(event string, component string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event", event),
		zap.String("component", component),
	}, fields...)

	l.Info("application_event", allFields...)
}

// LogStartupEvent logs application startup events with timing
func (l *applicationLogger) LogStartupEvent(component string, duration time.Duration, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("component", component),
		zap.Duration("startup_duration", duration),
		zap.String("event_type", "startup"),
	}, fields...)

	l.Info("application_startup", allFields...)
}

// LogShutdownEvent logs application shutdown events with timing
func (l *applicationLogger) LogShutdownEvent(component string, duration time.Duration, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("component", component),
		zap.Duration("shutdown_duration", duration),
		zap.String("event_type", "shutdown"),
	}, fields...)

	l.Info("application_shutdown", allFields...)
}