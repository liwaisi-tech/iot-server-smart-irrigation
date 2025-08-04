package entities

import (
	"testing"
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
				if err == nil {
					t.Errorf("NewDeviceRegistrationMessage() expected error but got none")
				}
				if msg != nil {
					t.Errorf("NewDeviceRegistrationMessage() expected nil message but got %v", msg)
				}
			} else {
				if err != nil {
					t.Errorf("NewDeviceRegistrationMessage() unexpected error: %v", err)
				}
				if msg == nil {
					t.Errorf("NewDeviceRegistrationMessage() expected message but got nil")
				}

				// Verify MAC address is normalized to uppercase
				if msg != nil && msg.MACAddress != "AA:BB:CC:DD:EE:FF" {
					t.Errorf("NewDeviceRegistrationMessage() MAC address not normalized correctly, got %s", msg.MACAddress)
				}
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
	if err != nil {
		t.Fatalf("Failed to create registration message: %v", err)
	}

	device, err := msg.ToDevice()
	if err != nil {
		t.Fatalf("Failed to convert to device: %v", err)
	}

	if device.MACAddress != msg.MACAddress {
		t.Errorf("Device MAC address mismatch: expected %s, got %s", msg.MACAddress, device.MACAddress)
	}

	if device.DeviceName != msg.DeviceName {
		t.Errorf("Device name mismatch: expected %s, got %s", msg.DeviceName, device.DeviceName)
	}

	if device.IPAddress != msg.IPAddress {
		t.Errorf("Device IP address mismatch: expected %s, got %s", msg.IPAddress, device.IPAddress)
	}

	if device.LocationDescription != msg.LocationDescription {
		t.Errorf("Device location description mismatch: expected %s, got %s", msg.LocationDescription, device.LocationDescription)
	}
}

func TestDeviceRegistrationMessage_GetDeviceIdentifier(t *testing.T) {
	msg, err := NewDeviceRegistrationMessage(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create registration message: %v", err)
	}

	identifier := msg.GetDeviceIdentifier()
	expected := "AA:BB:CC:DD:EE:FF"

	if identifier != expected {
		t.Errorf("GetDeviceIdentifier() expected %s, got %s", expected, identifier)
	}
}