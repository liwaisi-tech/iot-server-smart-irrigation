package entities

import "time"

type Device struct {
	MacAddress          string    `json:"mac_address"` // Primary key
	DeviceName          string    `json:"device_name"`
	IPAddress           string    `json:"ip_address"`
	LocationDescription string    `json:"location_description"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusUnknown DeviceStatus = "unknown"
)

type DeviceRegistrationEvent struct {
	Device
	EventType DeviceEventType `json:"event_type"`
}

// DeviceEventType represents the type of device events
type DeviceEventType string

const (
	DeviceEventRegister  DeviceEventType = "register"
	DeviceEventUpdate    DeviceEventType = "update"
	DeviceEventHeartbeat DeviceEventType = "heartbeat"
	DeviceEventOffline   DeviceEventType = "offline"
)
