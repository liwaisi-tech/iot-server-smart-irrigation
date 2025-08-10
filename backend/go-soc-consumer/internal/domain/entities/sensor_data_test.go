package entities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSensorData(t *testing.T) {
	tests := []struct {
		name        string
		macAddress  string
		temperature float64
		humidity    float64
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid sensor data with colons",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 25.5,
			humidity:    65.0,
			wantErr:     false,
		},
		{
			name:        "valid sensor data with dashes",
			macAddress:  "A0-A3-B3-AB-2F-D8",
			temperature: 28.8,
			humidity:    72.3,
			wantErr:     false,
		},
		{
			name:        "valid edge case temperature low",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: -40.0,
			humidity:    50.0,
			wantErr:     false,
		},
		{
			name:        "valid edge case temperature high",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 85.0,
			humidity:    50.0,
			wantErr:     false,
		},
		{
			name:        "valid edge case humidity low",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 25.0,
			humidity:    0.0,
			wantErr:     false,
		},
		{
			name:        "valid edge case humidity high",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 25.0,
			humidity:    100.0,
			wantErr:     false,
		},
		{
			name:        "empty mac address",
			macAddress:  "",
			temperature: 25.0,
			humidity:    65.0,
			wantErr:     true,
			errContains: "mac address is required",
		},
		{
			name:        "invalid mac address format",
			macAddress:  "invalid-mac",
			temperature: 25.0,
			humidity:    65.0,
			wantErr:     true,
			errContains: "invalid mac address format",
		},
		{
			name:        "mixed separators in mac address",
			macAddress:  "A0:A3-B3:AB-2F:D8",
			temperature: 25.0,
			humidity:    65.0,
			wantErr:     true,
			errContains: "mixed separators",
		},
		{
			name:        "temperature too low",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: -45.0,
			humidity:    65.0,
			wantErr:     true,
			errContains: "outside valid range",
		},
		{
			name:        "temperature too high",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 90.0,
			humidity:    65.0,
			wantErr:     true,
			errContains: "outside valid range",
		},
		{
			name:        "humidity too low",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 25.0,
			humidity:    -5.0,
			wantErr:     true,
			errContains: "outside valid range",
		},
		{
			name:        "humidity too high",
			macAddress:  "A0:A3:B3:AB:2F:D8",
			temperature: 25.0,
			humidity:    105.0,
			wantErr:     true,
			errContains: "outside valid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := NewSensorData(tt.macAddress, tt.temperature, tt.humidity)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, sensorData)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sensorData)
				
				// Verify normalized MAC address (should be uppercase and preserve original format)
				expectedMAC := strings.ToUpper(strings.TrimSpace(tt.macAddress))
				assert.Equal(t, expectedMAC, sensorData.MacAddress())
				assert.Equal(t, tt.temperature, sensorData.Temperature())
				assert.Equal(t, tt.humidity, sensorData.Humidity())
				assert.False(t, sensorData.Timestamp().IsZero())
			}
		})
	}
}

func TestSensorData_IsTemperatureNormal(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		want        bool
	}{
		{"normal temperature low", 0.0, true},
		{"normal temperature mid", 25.0, true},
		{"normal temperature high", 40.0, true},
		{"below normal", -5.0, false},
		{"above normal", 45.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := NewSensorData("A0:A3:B3:AB:2F:D8", tt.temperature, 50.0)
			require.NoError(t, err)
			
			assert.Equal(t, tt.want, sensorData.IsTemperatureNormal())
		})
	}
}

func TestSensorData_IsHumidityNormal(t *testing.T) {
	tests := []struct {
		name     string
		humidity float64
		want     bool
	}{
		{"normal humidity low", 30.0, true},
		{"normal humidity mid", 50.0, true},
		{"normal humidity high", 80.0, true},
		{"below normal", 25.0, false},
		{"above normal", 85.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := NewSensorData("A0:A3:B3:AB:2F:D8", 25.0, tt.humidity)
			require.NoError(t, err)
			
			assert.Equal(t, tt.want, sensorData.IsHumidityNormal())
		})
	}
}

func TestSensorData_HasAbnormalReadings(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		humidity    float64
		want        bool
	}{
		{"all normal", 25.0, 50.0, false},
		{"temperature abnormal", 45.0, 50.0, true},
		{"humidity abnormal", 25.0, 85.0, true},
		{"both abnormal", 45.0, 85.0, true},
		{"edge normal", 40.0, 80.0, false},
		{"edge abnormal temp", 40.1, 80.0, true},
		{"edge abnormal humidity", 40.0, 80.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := NewSensorData("A0:A3:B3:AB:2F:D8", tt.temperature, tt.humidity)
			require.NoError(t, err)
			
			assert.Equal(t, tt.want, sensorData.HasAbnormalReadings())
		})
	}
}

func TestSensorData_String(t *testing.T) {
	sensorData, err := NewSensorData("A0:A3:B3:AB:2F:D8", 28.8, 72.3)
	require.NoError(t, err)

	str := sensorData.String()
	assert.Contains(t, str, "A0:A3:B3:AB:2F:D8")
	assert.Contains(t, str, "28.80°C")
	assert.Contains(t, str, "72.30%")
	assert.Contains(t, str, "SensorData{")
}

func TestSensorData_Accessors(t *testing.T) {
	macAddr := "A0:A3:B3:AB:2F:D8"
	temp := 28.8
	humidity := 72.3
	
	sensorData, err := NewSensorData(macAddr, temp, humidity)
	require.NoError(t, err)

	assert.Equal(t, macAddr, sensorData.MacAddress())
	assert.Equal(t, temp, sensorData.Temperature())
	assert.Equal(t, humidity, sensorData.Humidity())
	assert.False(t, sensorData.Timestamp().IsZero())
}

func TestSensorData_MACAddressNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase with colons", "a0:a3:b3:ab:2f:d8", "A0:A3:B3:AB:2F:D8"},
		{"mixed case with colons", "a0:A3:b3:AB:2f:D8", "A0:A3:B3:AB:2F:D8"},
		{"with spaces", "  A0:A3:B3:AB:2F:D8  ", "A0:A3:B3:AB:2F:D8"},
		{"lowercase with dashes", "a0-a3-b3-ab-2f-d8", "A0-A3-B3-AB-2F-D8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := NewSensorData(tt.input, 25.0, 50.0)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, sensorData.MacAddress())
		})
	}
}