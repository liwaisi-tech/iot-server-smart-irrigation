package models

import (
	"time"

	"gorm.io/gorm"
)

// SensorTemperatureHumidityModel represents the GORM model for temperature and humidity sensor data persistence
// This model contains only data persistence concerns and GORM-specific annotations
type SensorTemperatureHumidityModel struct {
	// Foreign Key to Device
	MACAddress string `gorm:"size:17;not null;index" json:"mac_address"`

	// Sensor readings
	TemperatureCelsius float64 `gorm:"type:decimal(5,2);not null;index" json:"temperature_celsius"`
	HumidityPercent    float64 `gorm:"type:decimal(5,2);not null;check:humidity_percent >= 0 AND humidity_percent <= 100;index" json:"humidity_percent"`

	// Audit fields (GORM will handle these automatically)
	CreatedAt time.Time      `gorm:"not null;default:now();index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (SensorTemperatureHumidityModel) TableName() string {
	return "sensor_temperature_humidity"
}
