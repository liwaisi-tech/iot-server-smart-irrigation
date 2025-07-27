package repositories

import (
	"context"
	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/domain/entities"
)

// DeviceRepository defines the interface for device data access operations
// Focused on device registration with mac_address as primary key
type DeviceRepository interface {
	// Create creates a new device in the repository
	Create(ctx context.Context, device *entities.Device) error
	
	// GetByMacAddress retrieves a device by its MAC address (primary key)
	GetByMacAddress(ctx context.Context, macAddress string) (*entities.Device, error)
	
	// Update updates an existing device
	Update(ctx context.Context, device *entities.Device) error
	
	// Delete removes a device from the repository by MAC address
	Delete(ctx context.Context, macAddress string) error
	
	// List retrieves devices with pagination support (limit and offset)
	List(ctx context.Context, limit, offset int) ([]*entities.Device, error)
	
	// Count returns the total number of devices
	Count(ctx context.Context) (int64, error)
	
	// Exists checks if a device with the given MAC address exists
	Exists(ctx context.Context, macAddress string) (bool, error)
}

// DeviceListOptions provides filtering and pagination options for device queries
type DeviceListOptions struct {
	Limit    int    // Maximum number of records to return
	Offset   int    // Number of records to skip for pagination
	Location string // Filter by location description (partial match)
}