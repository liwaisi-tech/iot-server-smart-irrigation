package dtos

// SensorDataMessage represents the JSON structure for temperature/humidity sensor data messages
type SensorDataMessage struct {
	EventType   string  `json:"event_type"`
	MacAddress  string  `json:"mac_address"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}