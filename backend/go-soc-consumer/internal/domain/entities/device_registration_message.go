package entities

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// DeviceRegistrationMessage represents a device registration request message
type DeviceRegistrationMessage struct {
	MACAddress          string
	DeviceName          string
	IPAddress           string
	LocationDescription string
	ReceivedAt          time.Time
}

// NewDeviceRegistrationMessage creates a new device registration message with validation
func NewDeviceRegistrationMessage(macAddress, deviceName, ipAddress, locationDescription string) (*DeviceRegistrationMessage, error) {
	msg := &DeviceRegistrationMessage{
		MACAddress:          strings.ToUpper(strings.TrimSpace(macAddress)),
		DeviceName:          strings.TrimSpace(deviceName),
		IPAddress:           strings.TrimSpace(ipAddress),
		LocationDescription: strings.TrimSpace(locationDescription),
		ReceivedAt:          time.Now(),
	}

	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device registration message: %w", err)
	}

	return msg, nil
}

// Validate validates the device registration message fields
func (m *DeviceRegistrationMessage) Validate() error {
	if err := m.validateMacAddress(); err != nil {
		return err
	}

	if err := m.validateDeviceName(); err != nil {
		return err
	}

	if err := m.validateIPAddress(); err != nil {
		return err
	}

	if err := m.validateLocationDescription(); err != nil {
		return err
	}

	return nil
}

// validateMacAddress validates the MAC address format
func (m *DeviceRegistrationMessage) validateMacAddress() error {
	if m.MACAddress == "" {
		return fmt.Errorf("mac address is required")
	}

	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	macPattern := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	matched, err := regexp.MatchString(macPattern, m.MACAddress)
	if err != nil {
		return fmt.Errorf("error validating mac address: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid mac address format: %s", m.MACAddress)
	}

	return nil
}

// validateDeviceName validates the device name
func (m *DeviceRegistrationMessage) validateDeviceName() error {
	if m.DeviceName == "" {
		return fmt.Errorf("device name is required")
	}

	if len(m.DeviceName) > 100 {
		return fmt.Errorf("device name cannot exceed 100 characters")
	}

	return nil
}

// validateIPAddress validates the IP address format
func (m *DeviceRegistrationMessage) validateIPAddress() error {
	if m.IPAddress == "" {
		return fmt.Errorf("ip address is required")
	}

	if net.ParseIP(m.IPAddress) == nil {
		return fmt.Errorf("invalid ip address format: %s", m.IPAddress)
	}

	return nil
}

// validateLocationDescription validates the location description
func (m *DeviceRegistrationMessage) validateLocationDescription() error {
	if m.LocationDescription == "" {
		return fmt.Errorf("location description is required")
	}

	if len(m.LocationDescription) > 255 {
		return fmt.Errorf("location description cannot exceed 255 characters")
	}

	return nil
}

// ToDevice converts the registration message to a Device entity
func (m *DeviceRegistrationMessage) ToDevice() (*Device, error) {
	device := &Device{
		MACAddress:          m.MACAddress,
		DeviceName:          m.DeviceName,
		IPAddress:           m.IPAddress,
		LocationDescription: m.LocationDescription,
		RegisteredAt:        m.ReceivedAt,
		LastSeen:            m.ReceivedAt,
		Status:              "registered",
	}
	
	if err := device.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device created from registration message: %w", err)
	}
	
	return device, nil
}

// GetDeviceIdentifier returns the device identifier (MAC address)
func (m *DeviceRegistrationMessage) GetDeviceIdentifier() string {
	return m.MACAddress
}