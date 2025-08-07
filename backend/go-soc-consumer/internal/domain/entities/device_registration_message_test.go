package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeviceRegistrationMessage(t *testing.T) {
	tests := []struct {
		name                string
		macAddress          string
		deviceName          string
		ipAddress           string
		locationDescription string
		wantError           bool
	}{
		{
			name:                "valid message",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor 1",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           false,
		},
		{
			name:                "valid message with lowercase mac",
			macAddress:          "aa:bb:cc:dd:ee:ff",
			deviceName:          "Irrigation Sensor 2",
			ipAddress:           "192.168.1.101",
			locationDescription: "Garden Zone B",
			wantError:           false,
		},
		{
			name:                "empty mac address",
			macAddress:          "",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid mac address format",
			macAddress:          "INVALID-MAC",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "empty device name",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid IP address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "invalid-ip",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "empty location description",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "",
			wantError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewDeviceRegistrationMessage(
				tt.macAddress,
				tt.deviceName,
				tt.ipAddress,
				tt.locationDescription,
			)

			if tt.wantError {
				assert.Error(t, err, "NewDeviceRegistrationMessage() expected error but got none")
				assert.Nil(t, msg, "NewDeviceRegistrationMessage() expected nil message")
			} else {
				require.NoError(t, err, "NewDeviceRegistrationMessage() unexpected error")
				require.NotNil(t, msg, "NewDeviceRegistrationMessage() expected message but got nil")

				// Verify MAC address is normalized to uppercase
				assert.Equal(t, "AA:BB:CC:DD:EE:FF", msg.MACAddress, "NewDeviceRegistrationMessage() MAC address not normalized correctly")
			}
		})
	}
}

func TestDeviceRegistrationMessage_ToDevice(t *testing.T) {
	msg, err := NewDeviceRegistrationMessage(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create registration message")

	device, err := msg.ToDevice()
	require.NoError(t, err, "Failed to convert to device")

	assert.Equal(t, msg.MACAddress, device.MACAddress, "Device MAC address mismatch")
	assert.Equal(t, msg.DeviceName, device.DeviceName, "Device name mismatch")
	assert.Equal(t, msg.IPAddress, device.IPAddress, "Device IP address mismatch")
	assert.Equal(t, msg.LocationDescription, device.LocationDescription, "Device location description mismatch")
}

func TestDeviceRegistrationMessage_GetDeviceIdentifier(t *testing.T) {
	msg, err := NewDeviceRegistrationMessage(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create registration message")

	identifier := msg.GetDeviceIdentifier()
	expected := "AA:BB:CC:DD:EE:FF"

	assert.Equal(t, expected, identifier, "GetDeviceIdentifier() result mismatch")
}