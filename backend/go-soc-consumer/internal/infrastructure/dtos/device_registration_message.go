package dtos

type DeviceRegistrationMessage struct {
	EventType           string `json:"event_type"`
	MacAddress          string `json:"mac_address"`
	DeviceName          string `json:"device_name"`
	IPAddress           string `json:"ip_address"`
	LocationDescription string `json:"location_description"`
}
