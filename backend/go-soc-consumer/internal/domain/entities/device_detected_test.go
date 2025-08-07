package entities

import (
	"testing"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeviceDetectedEvent(t *testing.T) {
	tests := []struct {
		name        string
		macAddress  string
		ipAddress   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid event",
			macAddress:  "AA:BB:CC:DD:EE:FF",
			ipAddress:   "192.168.1.100",
			expectError: false,
		},
		{
			name:        "empty mac address",
			macAddress:  "",
			ipAddress:   "192.168.1.100",
			expectError: true,
			errorMsg:    "mac address is required",
		},
		{
			name:        "empty ip address",
			macAddress:  "AA:BB:CC:DD:EE:FF",
			ipAddress:   "",
			expectError: true,
			errorMsg:    "ip address is required",
		},
		{
			name:        "both empty",
			macAddress:  "",
			ipAddress:   "",
			expectError: true,
			errorMsg:    "mac address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewDeviceDetectedEvent(tt.macAddress, tt.ipAddress)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, event)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)

				assert.Equal(t, tt.macAddress, event.MACAddress)
				assert.Equal(t, tt.ipAddress, event.IPAddress)
				assert.Equal(t, events.DeviceDetectedEventType, event.EventType)
				assert.NotEmpty(t, event.EventID)
				assert.False(t, event.DetectedAt.IsZero())
				assert.Equal(t, events.DeviceDetectedSubject, event.GetSubject())
			}
		})
	}
}

func TestDeviceDetectedEvent_Validate(t *testing.T) {
	tests := []struct {
		name        string
		event       *DeviceDetectedEvent
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid event",
			event: &DeviceDetectedEvent{
				MACAddress: "AA:BB:CC:DD:EE:FF",
				IPAddress:  "192.168.1.100",
				DetectedAt: time.Now(),
				EventID:    "test-event-id",
				EventType:  events.DeviceDetectedEventType,
			},
			expectError: false,
		},
		{
			name: "empty mac address",
			event: &DeviceDetectedEvent{
				IPAddress:  "192.168.1.100",
				DetectedAt: time.Now(),
				EventID:    "test-event-id",
				EventType:  events.DeviceDetectedEventType,
			},
			expectError: true,
			errorMsg:    "mac address is required",
		},
		{
			name: "empty ip address",
			event: &DeviceDetectedEvent{
				MACAddress: "AA:BB:CC:DD:EE:FF",
				DetectedAt: time.Now(),
				EventID:    "test-event-id",
				EventType:  events.DeviceDetectedEventType,
			},
			expectError: true,
			errorMsg:    "ip address is required",
		},
		{
			name: "empty event id",
			event: &DeviceDetectedEvent{
				MACAddress: "AA:BB:CC:DD:EE:FF",
				IPAddress:  "192.168.1.100",
				DetectedAt: time.Now(),
				EventType:  events.DeviceDetectedEventType,
			},
			expectError: true,
			errorMsg:    "event ID is required",
		},
		{
			name: "empty event type",
			event: &DeviceDetectedEvent{
				MACAddress: "AA:BB:CC:DD:EE:FF",
				IPAddress:  "192.168.1.100",
				DetectedAt: time.Now(),
				EventID:    "test-event-id",
			},
			expectError: true,
			errorMsg:    "event type is required",
		},
		{
			name: "zero detected at",
			event: &DeviceDetectedEvent{
				MACAddress: "AA:BB:CC:DD:EE:FF",
				IPAddress:  "192.168.1.100",
				EventID:    "test-event-id",
				EventType:  events.DeviceDetectedEventType,
			},
			expectError: true,
			errorMsg:    "detected at timestamp is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeviceDetectedEvent_GetSubject(t *testing.T) {
	event, err := NewDeviceDetectedEvent("AA:BB:CC:DD:EE:FF", "192.168.1.100")
	require.NoError(t, err)
	require.NotNil(t, event)

	subject := event.GetSubject()
	assert.Equal(t, events.DeviceDetectedSubject, subject)
	assert.Equal(t, "liwaisi.iot.smart-irrigation.device.detected", subject)
}
