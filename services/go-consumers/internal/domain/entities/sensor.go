package entities

import "time"

type SensorReading struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	SensorType  string    `json:"sensor_type"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
}
