package logger

// loggerFactory implements LoggerFactory interface
type loggerFactory struct {
	core           CoreLogger
	device         DeviceLogger
	messaging      MessagingLogger
	infrastructure InfrastructureLogger
	performance    PerformanceLogger
	application    ApplicationLogger
}

// NewLoggerFactory creates a new logger factory with all domain loggers
func NewLoggerFactory(config LoggerConfig) (LoggerFactory, error) {
	core, err := NewCoreLogger(config)
	if err != nil {
		return nil, err
	}

	return &loggerFactory{
		core:           core,
		device:         NewDeviceLogger(core),
		messaging:      NewMessagingLogger(core),
		infrastructure: NewInfrastructureLogger(core),
		performance:    NewPerformanceLogger(core),
		application:    NewApplicationLogger(core),
	}, nil
}

// Device returns the device logger
func (f *loggerFactory) Device() DeviceLogger {
	return f.device
}

// Messaging returns the messaging logger
func (f *loggerFactory) Messaging() MessagingLogger {
	return f.messaging
}

// Infrastructure returns the infrastructure logger
func (f *loggerFactory) Infrastructure() InfrastructureLogger {
	return f.infrastructure
}

// Performance returns the performance logger
func (f *loggerFactory) Performance() PerformanceLogger {
	return f.performance
}

// Application returns the application logger
func (f *loggerFactory) Application() ApplicationLogger {
	return f.application
}

// Core returns the core logger
func (f *loggerFactory) Core() CoreLogger {
	return f.core
}

// NewDefaultLoggerFactory creates a logger factory with default production configuration
func NewDefaultLoggerFactory() (LoggerFactory, error) {
	return NewLoggerFactory(LoggerConfig{
		Level:       "info",
		Format:      "json",
		Environment: "production",
	})
}

// NewDevelopmentLoggerFactory creates a logger factory optimized for development
func NewDevelopmentLoggerFactory() (LoggerFactory, error) {
	return NewLoggerFactory(LoggerConfig{
		Level:       "debug",
		Format:      "console",
		Environment: "development",
	})
}
