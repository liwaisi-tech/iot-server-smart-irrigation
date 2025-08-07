package models

import (
	"time"

	"gorm.io/gorm"
)

// DeviceModel represents the GORM model for device persistence
// This model contains only data persistence concerns and GORM-specific annotations
type DeviceModel struct {
	// Primary fields
	MACAddress          string    `gorm:"primaryKey;size:17;not null" json:"mac_address"`
	DeviceName          string    `gorm:"size:150;not null" json:"device_name"`
	IPAddress           string    `gorm:"size:45;not null" json:"ip_address"`
	LocationDescription string    `gorm:"size:250;not null" json:"location_description"`
	RegisteredAt        time.Time `gorm:"not null;default:now();index" json:"registered_at"`
	LastSeen            time.Time `gorm:"not null;default:now();index" json:"last_seen"`
	Status              string    `gorm:"size:20;not null;default:'registered';check:status IN ('registered', 'online', 'offline');index" json:"status"`

	// Audit fields (GORM will handle these automatically)
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (DeviceModel) TableName() string {
	return "devices"
}

// BeforeCreate GORM hook called before creating a record
func (dm *DeviceModel) BeforeCreate(tx *gorm.DB) error {
	// Set timestamps if not already set
	now := time.Now()
	if dm.RegisteredAt.IsZero() {
		dm.RegisteredAt = now
	}
	if dm.LastSeen.IsZero() {
		dm.LastSeen = now
	}
	if dm.Status == "" {
		dm.Status = "registered"
	}

	return nil
}
