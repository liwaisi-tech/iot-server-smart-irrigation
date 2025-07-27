package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONB represents a PostgreSQL JSONB field
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Device represents the devices table with optimized indexing for IoT operations
// Using mac_address as primary key for device registration focused schema
type Device struct {
	MacAddress          string    `gorm:"type:varchar(17);primaryKey;not null;comment:Device MAC address as primary identifier" json:"mac_address"`
	DeviceName          string    `gorm:"type:varchar(255);not null;comment:Human readable device name" json:"device_name"`
	IPAddress           string    `gorm:"type:inet;index:idx_device_ip;comment:Device IP address" json:"ip_address"`
	LocationDescription string    `gorm:"type:text;comment:Physical location description" json:"location_description"`
	CreatedAt           time.Time `gorm:"autoCreateTime;index:idx_device_created_at;comment:Device registration timestamp" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime;comment:Last update timestamp" json:"updated_at"`
}

// BeforeCreate hook for device creation
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	// Validate MAC address format if needed
	// MAC address validation can be added here
	return nil
}

// TableName returns the table name for Device model
func (Device) TableName() string {
	return "devices"
}
