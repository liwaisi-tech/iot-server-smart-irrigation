package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/events"
)

// DeviceDetectedEvent represents an event triggered when a device is detected/registered
type DeviceDetectedEvent struct {
	MACAddress string
	IPAddress  string
	DetectedAt time.Time
	EventID    string
	EventType  string
}

// NewDeviceDetectedEvent creates a new device detected event with validation
func NewDeviceDetectedEvent(macAddress, ipAddress string) (*DeviceDetectedEvent, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address is required")
	}

	if ipAddress == "" {
		return nil, fmt.Errorf("ip address is required")
	}

	eventID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate event ID: %w", err)
	}

	return &DeviceDetectedEvent{
		MACAddress: macAddress,
		IPAddress:  ipAddress,
		DetectedAt: time.Now(),
		EventID:    eventID.String(),
		EventType:  events.DeviceDetectedEventType,
	}, nil
}

// Validate ensures the event has all required fields
func (e *DeviceDetectedEvent) Validate() error {
	if e.MACAddress == "" {
		return fmt.Errorf("mac address is required")
	}

	if e.IPAddress == "" {
		return fmt.Errorf("ip address is required")
	}

	if e.EventID == "" {
		return fmt.Errorf("event ID is required")
	}

	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	if e.DetectedAt.IsZero() {
		return fmt.Errorf("detected at timestamp is required")
	}

	return nil
}

// GetSubject returns the NATS subject for this event type
func (e *DeviceDetectedEvent) GetSubject() string {
	return events.DeviceDetectedSubject
}
