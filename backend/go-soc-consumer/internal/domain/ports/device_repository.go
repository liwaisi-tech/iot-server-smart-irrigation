package ports

import (
	"context"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

// DeviceRepository defines the contract for device persistence operations
type DeviceRepository interface {
	// Save persists a new device
	Save(ctx context.Context, device *entities.Device) error

	// Update updates an existing device
	Update(ctx context.Context, device *entities.Device) error

	// FindByMACAddress retrieves a device by its MAC address
	FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error)

	// Exists checks if a device with the given MAC address exists
	Exists(ctx context.Context, macAddress string) (bool, error)

	// List retrieves all devices with optional pagination
	List(ctx context.Context, offset, limit int) ([]*entities.Device, error)

	// Delete removes a device by MAC address
	Delete(ctx context.Context, macAddress string) error
}