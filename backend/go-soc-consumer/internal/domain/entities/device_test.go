package entities

import (
	"strings"
	"testing"
	"time"
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
				if err == nil {
					t.Errorf("NewDevice() expected error but got none")
				}
				if device != nil {
					t.Errorf("NewDevice() expected nil device but got %v", device)
				}
			} else {
				if err != nil {
					t.Errorf("NewDevice() unexpected error: %v", err)
				}
				if device == nil {
					t.Errorf("NewDevice() expected device but got nil")
					return
				}

				// Verify MAC address normalization
				if device.MACAddress != tt.expectedMAC {
					t.Errorf("NewDevice() MAC address expected %s, got %s", tt.expectedMAC, device.MACAddress)
				}

				// Verify other fields are trimmed and set correctly
				if device.DeviceName != strings.TrimSpace(tt.deviceName) {
					t.Errorf("NewDevice() device name expected %s, got %s", strings.TrimSpace(tt.deviceName), device.DeviceName)
				}

				if device.IPAddress != strings.TrimSpace(tt.ipAddress) {
					t.Errorf("NewDevice() IP address expected %s, got %s", strings.TrimSpace(tt.ipAddress), device.IPAddress)
				}

				if device.LocationDescription != strings.TrimSpace(tt.locationDescription) {
					t.Errorf("NewDevice() location description expected %s, got %s", strings.TrimSpace(tt.locationDescription), device.LocationDescription)
				}

				// Verify timestamps are set correctly
				if device.RegisteredAt.Before(beforeTime) || device.RegisteredAt.After(afterTime) {
					t.Errorf("NewDevice() RegisteredAt timestamp not within expected range")
				}

				if device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime) {
					t.Errorf("NewDevice() LastSeen timestamp not within expected range")
				}

				// Verify initial status
				if device.Status != "registered" {
					t.Errorf("NewDevice() expected initial status 'registered', got '%s'", device.Status)
				}
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
				if err == nil {
					t.Errorf("validateMacAddress() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateMacAddress() unexpected error: %v", err)
				}
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
				if err == nil {
					t.Errorf("validateDeviceName() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateDeviceName() unexpected error: %v", err)
				}
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
				if err == nil {
					t.Errorf("validateIPAddress() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateIPAddress() unexpected error: %v", err)
				}
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
				if err == nil {
					t.Errorf("validateLocationDescription() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateLocationDescription() unexpected error: %v", err)
				}
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
				if err == nil {
					t.Errorf("validateStatus() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateStatus() unexpected error: %v", err)
				}
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
				if err == nil {
					t.Errorf("UpdateStatus() expected error but got none")
				}
				// Status and LastSeen should not be updated on error
				if device.LastSeen.After(originalLastSeen) {
					t.Errorf("UpdateStatus() LastSeen should not be updated on error")
				}
			} else {
				if err != nil {
					t.Errorf("UpdateStatus() unexpected error: %v", err)
				}
				if device.Status != tt.status {
					t.Errorf("UpdateStatus() expected status %s, got %s", tt.status, device.Status)
				}
				// LastSeen should be updated
				if device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime) {
					t.Errorf("UpdateStatus() LastSeen not updated correctly")
				}
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

	if device.Status != "online" {
		t.Errorf("MarkOnline() expected status 'online', got '%s'", device.Status)
	}

	if device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime) {
		t.Errorf("MarkOnline() LastSeen not updated correctly")
	}
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

	if device.Status != "offline" {
		t.Errorf("MarkOffline() expected status 'offline', got '%s'", device.Status)
	}

	if device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime) {
		t.Errorf("MarkOffline() LastSeen not updated correctly")
	}
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

			if result != tt.expected {
				t.Errorf("IsOnline() expected %t, got %t", tt.expected, result)
			}
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

			if result != tt.expected {
				t.Errorf("IsOffline() expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestDevice_GetID(t *testing.T) {
	device := &Device{MACAddress: "AA:BB:CC:DD:EE:FF"}
	id := device.GetID()

	if id != device.MACAddress {
		t.Errorf("GetID() expected %s, got %s", device.MACAddress, id)
	}
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
				if err == nil {
					t.Errorf("Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}