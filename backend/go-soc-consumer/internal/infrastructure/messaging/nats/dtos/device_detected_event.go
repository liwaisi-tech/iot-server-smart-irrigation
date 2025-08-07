package dtos

import "time"

type DeviceDetectedEvent struct {
	MACAddress string    `json:"mac_address"`
	IPAddress  string    `json:"ip_address"`
	DetectedAt time.Time `json:"detected_at"`
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
}
