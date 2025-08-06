package entities

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name                string
		macAddress          string
		deviceName          string
		ipAddress           string
		locationDescription string
		wantError           bool
		expectedMAC         string // Expected normalized MAC address
	}{
		{
			name:                "valid device with colon-separated MAC",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor 1",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           false,
			expectedMAC:         "AA:BB:CC:DD:EE:FF",
		},
		{
			name:                "valid device with dash-separated MAC",
			macAddress:          "AA-BB-CC-DD-EE-FF",
			deviceName:          "Irrigation Sensor 2",
			ipAddress:           "192.168.1.101",
			locationDescription: "Garden Zone B",
			wantError:           false,
			expectedMAC:         "AA-BB-CC-DD-EE-FF",
		},
		{
			name:                "valid device with lowercase MAC",
			macAddress:          "aa:bb:cc:dd:ee:ff",
			deviceName:          "Irrigation Sensor 3",
			ipAddress:           "192.168.1.102",
			locationDescription: "Garden Zone C",
			wantError:           false,
			expectedMAC:         "AA:BB:CC:DD:EE:FF",
		},
		{
			name:                "valid device with IPv6 address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "IPv6 Sensor",
			ipAddress:           "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			locationDescription: "IPv6 Zone",
			wantError:           false,
			expectedMAC:         "AA:BB:CC:DD:EE:FF",
		},
		{
			name:                "valid device with boundary device name (100 chars)",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          strings.Repeat("A", 100),
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           false,
			expectedMAC:         "AA:BB:CC:DD:EE:FF",
		},
		{
			name:                "valid device with boundary location description (255 chars)",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Test Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: strings.Repeat("A", 255),
			wantError:           false,
			expectedMAC:         "AA:BB:CC:DD:EE:FF",
		},
		{
			name:                "empty MAC address",
			macAddress:          "",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "whitespace-only MAC address",
			macAddress:          "   ",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid MAC address format - too short",
			macAddress:          "AA:BB:CC:DD:EE",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid MAC address format - too long",
			macAddress:          "AA:BB:CC:DD:EE:FF:GG",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid MAC address format - invalid characters",
			macAddress:          "ZZ:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid MAC address format - mixed separators",
			macAddress:          "AA:BB-CC:DD:EE:FF",
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
			name:                "whitespace-only device name",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "   ",
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "device name too long (101 chars)",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          strings.Repeat("A", 101),
			ipAddress:           "192.168.1.100",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "empty IP address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "whitespace-only IP address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "   ",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "invalid IPv4 address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "256.256.256.256",
			locationDescription: "Garden Zone A",
			wantError:           true,
		},
		{
			name:                "malformed IP address",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "not-an-ip",
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
		{
			name:                "whitespace-only location description",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: "   ",
			wantError:           true,
		},
		{
			name:                "location description too long (256 chars)",
			macAddress:          "AA:BB:CC:DD:EE:FF",
			deviceName:          "Irrigation Sensor",
			ipAddress:           "192.168.1.100",
			locationDescription: strings.Repeat("A", 256),
			wantError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeTime := time.Now()
			
			device, err := NewDevice(
				tt.macAddress,
				tt.deviceName,
				tt.ipAddress,
				tt.locationDescription,
			)
			
			afterTime := time.Now()

			if tt.wantError {
				assert.Error(t, err, "NewDevice() expected error but got none")
				assert.Nil(t, device, "NewDevice() expected nil device")
			} else {
				require.NoError(t, err, "NewDevice() unexpected error")
				require.NotNil(t, device, "NewDevice() expected device but got nil")

				// Verify MAC address normalization
				assert.Equal(t, tt.expectedMAC, device.MACAddress, "NewDevice() MAC address mismatch")

				// Verify other fields are trimmed and set correctly
				assert.Equal(t, strings.TrimSpace(tt.deviceName), device.DeviceName, "NewDevice() device name mismatch")
				assert.Equal(t, strings.TrimSpace(tt.ipAddress), device.IPAddress, "NewDevice() IP address mismatch")
				assert.Equal(t, strings.TrimSpace(tt.locationDescription), device.LocationDescription, "NewDevice() location description mismatch")

				// Verify timestamps are set correctly
				assert.False(t, device.RegisteredAt.Before(beforeTime) || device.RegisteredAt.After(afterTime), "NewDevice() RegisteredAt timestamp not within expected range")
				assert.False(t, device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime), "NewDevice() LastSeen timestamp not within expected range")

				// Verify initial status
				assert.Equal(t, "registered", device.Status, "NewDevice() expected initial status 'registered'")
			}
		})
	}
}

func TestDevice_validateMacAddress(t *testing.T) {
	tests := []struct {
		name       string
		macAddress string
		wantError  bool
	}{
		{"valid colon-separated uppercase", "AA:BB:CC:DD:EE:FF", false},
		{"valid colon-separated lowercase", "aa:bb:cc:dd:ee:ff", false},
		{"valid colon-separated mixed case", "Aa:Bb:Cc:Dd:Ee:Ff", false},
		{"valid dash-separated uppercase", "AA-BB-CC-DD-EE-FF", false},
		{"valid dash-separated lowercase", "aa-bb-cc-dd-ee-ff", false},
		{"valid dash-separated mixed case", "Aa-Bb-Cc-Dd-Ee-Ff", false},
		{"empty MAC", "", true},
		{"too short", "AA:BB:CC:DD:EE", true},
		{"too long", "AA:BB:CC:DD:EE:FF:GG", true},
		{"invalid characters", "ZZ:BB:CC:DD:EE:FF", true},
		{"mixed separators", "AA:BB-CC:DD:EE:FF", true},
		{"no separators", "AABBCCDDEEFF", true},
		{"wrong separator", "AA.BB.CC.DD.EE.FF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{MACAddress: tt.macAddress}
			err := device.validateMacAddress()

			if tt.wantError {
				assert.Error(t, err, "validateMacAddress() expected error but got none")
			} else {
				assert.NoError(t, err, "validateMacAddress() unexpected error")
			}
		})
	}
}

func TestDevice_validateDeviceName(t *testing.T) {
	tests := []struct {
		name       string
		deviceName string
		wantError  bool
	}{
		{"valid short name", "Sensor", false},
		{"valid long name", "Very Long Irrigation Sensor Name", false},
		{"valid boundary (100 chars)", strings.Repeat("A", 100), false},
		{"empty name", "", true},
		{"too long (101 chars)", strings.Repeat("A", 101), true},
		{"only spaces", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{DeviceName: tt.deviceName}
			err := device.validateDeviceName()

			if tt.wantError {
				assert.Error(t, err, "validateDeviceName() expected error but got none")
			} else {
				assert.NoError(t, err, "validateDeviceName() unexpected error")
			}
		})
	}
}

func TestDevice_validateIPAddress(t *testing.T) {
	tests := []struct {
		name      string
		ipAddress string
		wantError bool
	}{
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv4 boundary", "255.255.255.255", false},
		{"valid IPv4 loopback", "127.0.0.1", false},
		{"valid IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"valid IPv6 compressed", "2001:db8:85a3::8a2e:370:7334", false},
		{"valid IPv6 loopback", "::1", false},
		{"empty IP", "", true},
		{"invalid IPv4 high octets", "256.256.256.256", true},
		{"invalid IPv4 format", "192.168.1", true},
		{"invalid IPv4 text", "not-an-ip", true},
		{"invalid IPv6", "2001:0db8:85a3::8a2e::7334", true},
		{"only spaces", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{IPAddress: tt.ipAddress}
			err := device.validateIPAddress()

			if tt.wantError {
				assert.Error(t, err, "validateIPAddress() expected error but got none")
			} else {
				assert.NoError(t, err, "validateIPAddress() unexpected error")
			}
		})
	}
}

func TestDevice_validateLocationDescription(t *testing.T) {
	tests := []struct {
		name        string
		location    string
		wantError   bool
	}{
		{"valid short location", "Garden", false},
		{"valid long location", "Very detailed location description with many words", false},
		{"valid boundary (255 chars)", strings.Repeat("A", 255), false},
		{"empty location", "", true},
		{"too long (256 chars)", strings.Repeat("A", 256), true},
		{"only spaces", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{LocationDescription: tt.location}
			err := device.validateLocationDescription()

			if tt.wantError {
				assert.Error(t, err, "validateLocationDescription() expected error but got none")
			} else {
				assert.NoError(t, err, "validateLocationDescription() unexpected error")
			}
		})
	}
}

func TestDevice_validateStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantError bool
	}{
		{"valid registered status", "registered", false},
		{"valid online status", "online", false},
		{"valid offline status", "offline", false},
		{"invalid status", "unknown", true},
		{"empty status", "", true},
		{"uppercase status", "ONLINE", true},
		{"mixed case status", "Online", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{Status: tt.status}
			err := device.validateStatus()

			if tt.wantError {
				assert.Error(t, err, "validateStatus() expected error but got none")
			} else {
				assert.NoError(t, err, "validateStatus() unexpected error")
			}
		})
	}
}

func TestDevice_UpdateStatus(t *testing.T) {
	device := &Device{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Test Location",
		RegisteredAt:        time.Now(),
		LastSeen:            time.Now().Add(-time.Hour),
		Status:              "registered",
	}

	tests := []struct {
		name      string
		status    string
		wantError bool
	}{
		{"update to online", "online", false},
		{"update to offline", "offline", false},
		{"update to registered", "registered", false},
		{"invalid status", "invalid", true},
		{"empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLastSeen := device.LastSeen
			beforeTime := time.Now()
			
			err := device.UpdateStatus(tt.status)
			
			afterTime := time.Now()

			if tt.wantError {
				assert.Error(t, err, "UpdateStatus() expected error but got none")
				// Status and LastSeen should not be updated on error
				assert.False(t, device.LastSeen.After(originalLastSeen), "UpdateStatus() LastSeen should not be updated on error")
			} else {
				assert.NoError(t, err, "UpdateStatus() unexpected error")
				assert.Equal(t, tt.status, device.Status, "UpdateStatus() status mismatch")
				// LastSeen should be updated
				assert.False(t, device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime), "UpdateStatus() LastSeen not updated correctly")
			}
		})
	}
}

func TestDevice_MarkOnline(t *testing.T) {
	device := &Device{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Test Location",
		RegisteredAt:        time.Now(),
		LastSeen:            time.Now().Add(-time.Hour),
		Status:              "registered",
	}

	beforeTime := time.Now()
	device.MarkOnline()
	afterTime := time.Now()

	assert.Equal(t, "online", device.Status, "MarkOnline() expected status 'online'")
	assert.False(t, device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime), "MarkOnline() LastSeen not updated correctly")
}

func TestDevice_MarkOffline(t *testing.T) {
	device := &Device{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Test Location",
		RegisteredAt:        time.Now(),
		LastSeen:            time.Now().Add(-time.Hour),
		Status:              "online",
	}

	beforeTime := time.Now()
	device.MarkOffline()
	afterTime := time.Now()

	assert.Equal(t, "offline", device.Status, "MarkOffline() expected status 'offline'")
	assert.False(t, device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime), "MarkOffline() LastSeen not updated correctly")
}

func TestDevice_IsOnline(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"online device", "online", true},
		{"offline device", "offline", false},
		{"registered device", "registered", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{Status: tt.status}
			result := device.IsOnline()

			assert.Equal(t, tt.expected, result, "IsOnline() result mismatch")
		})
	}
}

func TestDevice_IsOffline(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"offline device", "offline", true},
		{"online device", "online", false},
		{"registered device", "registered", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{Status: tt.status}
			result := device.IsOffline()

			assert.Equal(t, tt.expected, result, "IsOffline() result mismatch")
		})
	}
}

func TestDevice_GetID(t *testing.T) {
	device := &Device{MACAddress: "AA:BB:CC:DD:EE:FF"}
	id := device.GetID()

	assert.Equal(t, device.MACAddress, id, "GetID() result mismatch")
}

func TestDevice_Validate(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		wantError bool
	}{
		{
			name: "valid device",
			device: &Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Test Location",
				Status:              "registered",
			},
			wantError: false,
		},
		{
			name: "invalid MAC address",
			device: &Device{
				MACAddress:          "INVALID",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Test Location",
				Status:              "registered",
			},
			wantError: true,
		},
		{
			name: "invalid device name",
			device: &Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Test Location",
				Status:              "registered",
			},
			wantError: true,
		},
		{
			name: "invalid IP address",
			device: &Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "invalid-ip",
				LocationDescription: "Test Location",
				Status:              "registered",
			},
			wantError: true,
		},
		{
			name: "invalid location description",
			device: &Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "",
				Status:              "registered",
			},
			wantError: true,
		},
		{
			name: "invalid status",
			device: &Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Test Location",
				Status:              "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()

			if tt.wantError {
				assert.Error(t, err, "Validate() expected error but got none")
			} else {
				assert.NoError(t, err, "Validate() unexpected error")
			}
		})
	}
}