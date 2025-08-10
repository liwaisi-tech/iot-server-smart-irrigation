// Package logger provides domain-specific structured logging for IoT applications
// 
// This package implements a clean architecture approach to logging with:
// - Domain-specific logger interfaces (DeviceLogger, SensorLogger, etc.)
// - LoggerFactory for dependency injection and composition
// - Structured logging using Zap for high performance
// 
// Usage:
//   factory, _ := logger.NewLoggerFactory(config)
//   deviceLogger := factory.Device()
//   deviceLogger.LogDeviceRegistration(...)
//
//   Or use convenience functions:
//   factory, _ := logger.New(config)
//   factory, _ := logger.NewDefault()
//   factory, _ := logger.NewDevelopment()
//
package logger


// New creates a LoggerFactory with the given configuration
func New(config LoggerConfig) (LoggerFactory, error) {
	return NewLoggerFactory(config)
}

// NewDefault creates a LoggerFactory with default production configuration
func NewDefault() (LoggerFactory, error) {
	return NewDefaultLoggerFactory()
}

// NewDevelopment creates a LoggerFactory optimized for development
func NewDevelopment() (LoggerFactory, error) {
	return NewDevelopmentLoggerFactory()
}