package logger

import (
	"time"

	"go.uber.org/zap"
)

// performanceLogger implements PerformanceLogger interface
type performanceLogger struct {
	CoreLogger
}

// NewPerformanceLogger creates a new performance logger with the given core logger
func NewPerformanceLogger(core CoreLogger) PerformanceLogger {
	return &performanceLogger{
		CoreLogger: core,
	}
}

// LogPerformanceMetrics logs performance-related metrics
func (l *performanceLogger) LogPerformanceMetrics(operation string, duration time.Duration, throughput float64, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Float64("throughput", throughput),
		zap.String("component", "performance"),
	}, fields...)

	l.Info("performance_metrics", allFields...)
}

// LogResourceUsage logs system resource usage metrics
func (l *performanceLogger) LogResourceUsage(component string, cpuPercent, memoryMB float64, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("component", component),
		zap.Float64("cpu_percent", cpuPercent),
		zap.Float64("memory_mb", memoryMB),
		zap.String("metric_type", "resource_usage"),
	}, fields...)

	l.Debug("resource_usage", allFields...)
}

// LogThroughputMetrics logs throughput-related metrics
func (l *performanceLogger) LogThroughputMetrics(component string, requestsPerSecond float64, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("component", component),
		zap.Float64("requests_per_second", requestsPerSecond),
		zap.String("metric_type", "throughput"),
	}, fields...)

	l.Info("throughput_metrics", allFields...)
}