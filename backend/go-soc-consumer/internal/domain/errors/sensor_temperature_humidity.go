package errors

// SensorTemperatureHumidity-specific domain errors
var (
	ErrSensorTemperatureHumidityNotFound   = NewDomainError("SENSOR_TEMPERATURE_HUMIDITY_NOT_FOUND", "Sensor temperature humidity not found")
	ErrSensorTemperatureHumidityNotCreated = NewDomainError("SENSOR_TEMPERATURE_HUMIDITY_NOT_CREATED", "Sensor temperature humidity not created")
)
