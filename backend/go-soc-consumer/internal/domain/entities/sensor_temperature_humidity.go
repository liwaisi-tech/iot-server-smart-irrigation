package entities

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/validation"
)

// SensorTemperatureHumidity represents temperature and humidity sensor measurements from IoT devices
type SensorTemperatureHumidity struct {
	mu          sync.RWMutex
	macAddress  string
	temperature float64
	humidity    float64
	timestamp   time.Time
}

// NewSensorTemperatureHumidity creates a new SensorTemperatureHumidity entity with validation
func NewSensorTemperatureHumidity(macAddress string, temperature, humidity float64) (*SensorTemperatureHumidity, error) {
	sensor := &SensorTemperatureHumidity{
		macAddress:  macAddress,
		temperature: temperature,
		humidity:    humidity,
		timestamp:   time.Now().UTC(),
	}

	// Normalize all fields
	sensor.Normalize()

	// Validate the sensor data
	if err := sensor.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return sensor, nil
}

// MacAddress returns the MAC address of the sensor device
func (s *SensorTemperatureHumidity) MacAddress() string {
	return s.macAddress
}

// Temperature returns the temperature measurement in Celsius
func (s *SensorTemperatureHumidity) Temperature() float64 {
	return s.temperature
}

// Humidity returns the humidity measurement as percentage
func (s *SensorTemperatureHumidity) Humidity() float64 {
	return s.humidity
}

// Timestamp returns when the sensor data was created
func (s *SensorTemperatureHumidity) Timestamp() time.Time {
	return s.timestamp
}

// String provides a human-readable representation of the sensor data
func (s *SensorTemperatureHumidity) String() string {
	return fmt.Sprintf("SensorTemperatureHumidity{MAC: %s, Temp: %.2f°C, Humidity: %.2f%%, Time: %s}",
		s.macAddress, s.temperature, s.humidity, s.timestamp.Format(time.RFC3339))
}

// Normalize ensures all fields are properly formatted and trimmed
func (s *SensorTemperatureHumidity) Normalize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Normalize MAC address (uppercase and trim spaces)
	s.macAddress = strings.ToUpper(strings.TrimSpace(s.macAddress))
	// Round temperature and humidity to 2 decimal places for consistency
	s.temperature = math.Round(s.temperature*100) / 100
	s.humidity = math.Round(s.humidity*100) / 100
}

// Validate performs validation of the sensor data
func (s *SensorTemperatureHumidity) Validate() error {
	// Validate MAC address
	if err := validation.ValidateMACAddress(s.macAddress); err != nil {
		return fmt.Errorf("invalid mac address: %w", err)
	}

	// Validate temperature range (-40°C to 85°C typical for IoT sensors)
	if s.temperature < -40.0 || s.temperature > 85.0 {
		return fmt.Errorf("temperature %.2f°C is outside valid range (-40.0 to 85.0)", s.temperature)
	}

	// Validate humidity range (0% to 100%)
	if s.humidity < 0.0 || s.humidity > 100.0 {
		return fmt.Errorf("humidity %.2f%% is outside valid range (0.0 to 100.0)", s.humidity)
	}

	// Validate timestamp (should not be zero or in the future)
	if s.timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	if s.timestamp.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("timestamp cannot be in the future")
	}

	return nil
}

// IsTemperatureNormal checks if temperature is within normal operating range (0°C to 40°C)
func (s *SensorTemperatureHumidity) IsTemperatureNormal() bool {
	return s.temperature >= 0.0 && s.temperature <= 40.0
}

// IsHumidityNormal checks if humidity is within normal range (30% to 80%)
func (s *SensorTemperatureHumidity) IsHumidityNormal() bool {
	return s.humidity >= 30.0 && s.humidity <= 80.0
}

// HasAbnormalReadings returns true if either temperature or humidity is outside normal ranges
func (s *SensorTemperatureHumidity) HasAbnormalReadings() bool {
	return !s.IsTemperatureNormal() || !s.IsHumidityNormal()
}
