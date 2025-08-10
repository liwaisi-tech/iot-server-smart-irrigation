package logger

import (
	"go.uber.org/zap"
)

// sensorLogger implements SensorLogger interface
type sensorLogger struct {
	CoreLogger
}

// NewSensorLogger creates a new sensor logger with the given core logger
func NewSensorLogger(core CoreLogger) SensorLogger {
	return &sensorLogger{
		CoreLogger: core,
	}
}

// LogSensorData logs temperature and humidity sensor data with structured fields
func (l *sensorLogger) LogSensorData(macAddress string, temperature, humidity float64, hasAbnormalReadings bool) {
	level := l.Info
	message := "sensor_data_received"
	
	// Use warning level for abnormal readings to aid monitoring
	if hasAbnormalReadings {
		level = l.Warn
		message = "sensor_data_abnormal_readings"
	}

	level(message,
		zap.String("mac_address", macAddress),
		zap.Float64("temperature_celsius", temperature),
		zap.Float64("humidity_percent", humidity),
		zap.Bool("has_abnormal_readings", hasAbnormalReadings),
		zap.Bool("temperature_normal", temperature >= 0.0 && temperature <= 40.0),
		zap.Bool("humidity_normal", humidity >= 30.0 && humidity <= 80.0),
		zap.String("component", "sensor_data_consumer"),
	)
}

// LogSensorDataProcessingError logs errors during sensor data processing
func (l *sensorLogger) LogSensorDataProcessingError(macAddress string, rawPayload []byte, err error, stage string) {
	l.Error("sensor_data_processing_error",
		zap.Error(err),
		zap.String("mac_address", macAddress),
		zap.String("processing_stage", stage), // e.g., "json_unmarshal", "validation", "entity_creation"
		zap.ByteString("raw_payload", rawPayload),
		zap.String("component", "sensor_data_consumer"),
	)
}

// LogSensorValidation logs sensor data validation results
func (l *sensorLogger) LogSensorValidation(macAddress string, validationResults map[string]bool, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("mac_address", macAddress),
		zap.Any("validation_results", validationResults),
		zap.String("component", "sensor_validation"),
	}, fields...)

	l.Debug("sensor_validation_completed", allFields...)
}