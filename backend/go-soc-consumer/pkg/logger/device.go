package logger

import (
	"time"

	"go.uber.org/zap"
)

// deviceLogger implements DeviceLogger interface
type deviceLogger struct {
	CoreLogger
}

// NewDeviceLogger creates a new device logger with the given core logger
func NewDeviceLogger(core CoreLogger) DeviceLogger {
	return &deviceLogger{
		CoreLogger: core,
	}
}

// LogDeviceRegistration logs device registration events with structured fields
func (l *deviceLogger) LogDeviceRegistration(macAddress, deviceName, ipAddress, location string, isUpdate bool) {
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

// LogDeviceHealthCheck logs device health checking operations
func (l *deviceLogger) LogDeviceHealthCheck(macAddress, ipAddress string, isAlive bool, responseTime time.Duration, err error) {
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

// LogDeviceStatus logs general device status changes
func (l *deviceLogger) LogDeviceStatus(macAddress, status string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("mac_address", macAddress),
		zap.String("status", status),
		zap.String("component", "device_management"),
	}, fields...)

	l.Info("device_status_changed", allFields...)
}