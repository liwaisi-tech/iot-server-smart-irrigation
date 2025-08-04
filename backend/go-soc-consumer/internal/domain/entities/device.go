package entities

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// Device represents an IoT device in the smart irrigation system
type Device struct {
	MACAddress          string
	DeviceName          string
	IPAddress           string
	LocationDescription string
	RegisteredAt        time.Time
	LastSeen            time.Time
	Status              string // "registered", "online", "offline"
}

// NewDevice creates a new device with validation
func NewDevice(macAddress, deviceName, ipAddress, locationDescription string) (*Device, error) {
	device := &Device{
		MACAddress:          strings.ToUpper(strings.TrimSpace(macAddress)),
		DeviceName:          strings.TrimSpace(deviceName),
		IPAddress:           strings.TrimSpace(ipAddress),
		LocationDescription: strings.TrimSpace(locationDescription),
		RegisteredAt:        time.Now(),
		LastSeen:            time.Now(),
		Status:              "registered",
	}

	if err := device.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device: %w", err)
	}

	return device, nil
}

// Validate validates the device fields
func (d *Device) Validate() error {
	if err := d.validateMacAddress(); err != nil {
		return err
	}

	if err := d.validateDeviceName(); err != nil {
		return err
	}

	if err := d.validateIPAddress(); err != nil {
		return err
	}

	if err := d.validateLocationDescription(); err != nil {
		return err
	}

	if err := d.validateStatus(); err != nil {
		return err
	}

	return nil
}

// validateMacAddress validates the MAC address format
func (d *Device) validateMacAddress() error {
	if d.MACAddress == "" {
		return fmt.Errorf("mac address is required")
	}

	// Check for consistent separator (either all colons or all dashes)
	hasColon := strings.Contains(d.MACAddress, ":")
	hasDash := strings.Contains(d.MACAddress, "-")

	if hasColon && hasDash {
		return fmt.Errorf("invalid mac address format: mixed separators (use either colons or dashes)")
	}

	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	macPattern := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	matched, err := regexp.MatchString(macPattern, d.MACAddress)
	if err != nil {
		return fmt.Errorf("error validating mac address: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid mac address format: %s (expected format: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX)", d.MACAddress)
	}

	return nil
}

// validateDeviceName validates the device name
func (d *Device) validateDeviceName() error {
	if d.DeviceName == "" {
		return fmt.Errorf("device name is required")
	}

	// Check if device name contains only whitespace
	if strings.TrimSpace(d.DeviceName) == "" {
		return fmt.Errorf("device name cannot be empty or whitespace only")
	}

	if len(d.DeviceName) > 100 {
		return fmt.Errorf("device name cannot exceed 100 characters")
	}

	return nil
}

// validateIPAddress validates the IP address format
func (d *Device) validateIPAddress() error {
	if d.IPAddress == "" {
		return fmt.Errorf("ip address is required")
	}

	if net.ParseIP(d.IPAddress) == nil {
		return fmt.Errorf("invalid ip address format: %s", d.IPAddress)
	}

	return nil
}

// validateLocationDescription validates the location description
func (d *Device) validateLocationDescription() error {
	if d.LocationDescription == "" {
		return fmt.Errorf("location description is required")
	}

	// Check if location description contains only whitespace
	if strings.TrimSpace(d.LocationDescription) == "" {
		return fmt.Errorf("location description cannot be empty or whitespace only")
	}

	if len(d.LocationDescription) > 255 {
		return fmt.Errorf("location description cannot exceed 255 characters")
	}

	return nil
}

// validateStatus validates the device status
func (d *Device) validateStatus() error {
	validStatuses := map[string]bool{
		"registered": true,
		"online":     true,
		"offline":    true,
	}

	if !validStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s. Valid statuses: registered, online, offline", d.Status)
	}

	return nil
}

// UpdateStatus updates the device status and last seen timestamp
func (d *Device) UpdateStatus(status string) error {
	// Save the current status in case we need to roll back
	originalStatus := d.Status
	
	// Update the status for validation
	d.Status = status
	
	// Validate the new status
	if err := d.validateStatus(); err != nil {
		// Roll back the status on validation error
		d.Status = originalStatus
		return fmt.Errorf("invalid status update: %w", err)
	}
	
	// Only update LastSeen if the status is valid
	d.LastSeen = time.Now()
	return nil
}

// MarkOnline marks the device as online
func (d *Device) MarkOnline() {
	d.Status = "online"
	d.LastSeen = time.Now()
}

// MarkOffline marks the device as offline
func (d *Device) MarkOffline() {
	d.Status = "offline"
	d.LastSeen = time.Now()
}

// IsOnline returns true if the device is currently online
func (d *Device) IsOnline() bool {
	return d.Status == "online"
}

// IsOffline returns true if the device is currently offline
func (d *Device) IsOffline() bool {
	return d.Status == "offline"
}

// GetID returns a unique identifier for the device (MAC address)
func (d *Device) GetID() string {
	return d.MACAddress
}