package entities

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
)

// SensorData represents temperature and humidity sensor measurements from IoT devices
type SensorData struct {
	macAddress  string
	temperature float64
	humidity    float64
	timestamp   time.Time
}

// NewSensorData creates a new SensorData entity with validation
func NewSensorData(macAddress string, temperature, humidity float64) (*SensorData, error) {
	// Normalize and validate MAC address
	normalizedMac := strings.ToUpper(strings.TrimSpace(macAddress))
	if err := validateMACAddressFormat(normalizedMac); err != nil {
		return nil, fmt.Errorf("invalid mac address: %w", err)
	}

	// Validate temperature range (-40°C to 85°C typical for IoT sensors)
	if temperature < -40.0 || temperature > 85.0 {
		return nil, errors.NewDomainError("INVALID_TEMPERATURE_RANGE",
			fmt.Sprintf("temperature %.2f°C is outside valid range (-40.0 to 85.0)", temperature))
	}

	// Validate humidity range (0% to 100%)
	if humidity < 0.0 || humidity > 100.0 {
		return nil, errors.NewDomainError("INVALID_HUMIDITY_RANGE",
			fmt.Sprintf("humidity %.2f%% is outside valid range (0.0 to 100.0)", humidity))
	}

	return &SensorData{
		macAddress:  normalizedMac,
		temperature: temperature,
		humidity:    humidity,
		timestamp:   time.Now().UTC(),
	}, nil
}

// validateMACAddressFormat validates the MAC address format (copied from Device entity pattern)
func validateMACAddressFormat(macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address is required")
	}

	// Check for consistent separator (either all colons or all dashes)
	hasColon := strings.Contains(macAddress, ":")
	hasDash := strings.Contains(macAddress, "-")

	if hasColon && hasDash {
		return fmt.Errorf("invalid mac address format: mixed separators (use either colons or dashes)")
	}

	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	macPattern := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	matched, err := regexp.MatchString(macPattern, macAddress)
	if err != nil {
		return fmt.Errorf("error validating mac address: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid mac address format: %s (expected format: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX)", macAddress)
	}

	return nil
}

// MacAddress returns the MAC address of the sensor device
func (s *SensorData) MacAddress() string {
	return s.macAddress
}

// Temperature returns the temperature measurement in Celsius
func (s *SensorData) Temperature() float64 {
	return s.temperature
}

// Humidity returns the humidity measurement as percentage
func (s *SensorData) Humidity() float64 {
	return s.humidity
}

// Timestamp returns when the sensor data was created
func (s *SensorData) Timestamp() time.Time {
	return s.timestamp
}

// String provides a human-readable representation of the sensor data
func (s *SensorData) String() string {
	return fmt.Sprintf("SensorData{MAC: %s, Temp: %.2f°C, Humidity: %.2f%%, Time: %s}",
		s.macAddress, s.temperature, s.humidity, s.timestamp.Format(time.RFC3339))
}

// IsTemperatureNormal checks if temperature is within normal operating range (0°C to 40°C)
func (s *SensorData) IsTemperatureNormal() bool {
	return s.temperature >= 0.0 && s.temperature <= 40.0
}

// IsHumidityNormal checks if humidity is within normal range (30% to 80%)
func (s *SensorData) IsHumidityNormal() bool {
	return s.humidity >= 30.0 && s.humidity <= 80.0
}

// HasAbnormalReadings returns true if either temperature or humidity is outside normal ranges
func (s *SensorData) HasAbnormalReadings() bool {
	return !s.IsTemperatureNormal() || !s.IsHumidityNormal()
}
