package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// DeviceRepository implements DeviceRepository using in-memory storage
type DeviceRepository struct {
	devices map[string]*entities.Device
	mu      sync.RWMutex
}

// NewDeviceRepository creates a new in-memory device repository
func NewDeviceRepository() ports.DeviceRepository {
	return &DeviceRepository{
		devices: make(map[string]*entities.Device),
	}
}

// Save saves a new device to the repository
func (r *DeviceRepository) Save(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device already exists
	if _, exists := r.devices[device.MACAddress]; exists {
		return errors.ErrDeviceAlreadyExists
	}

	// Save device
	r.devices[device.MACAddress] = device
	return nil
}

// Update updates an existing device in the repository
func (r *DeviceRepository) Update(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device exists
	if _, exists := r.devices[device.MACAddress]; !exists {
		return errors.ErrDeviceNotFound
	}

	// Update device
	r.devices[device.MACAddress] = device
	return nil
}

// FindByMACAddress finds a device by MAC address
func (r *DeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.devices[macAddress]
	if !exists {
		return nil, errors.ErrDeviceNotFound
	}

	return device, nil
}

// Exists checks if a device with the given MAC address exists
func (r *DeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.devices[macAddress]
	return exists, nil
}

// Delete removes a device from the repository
func (r *DeviceRepository) Delete(ctx context.Context, macAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device exists
	if _, exists := r.devices[macAddress]; !exists {
		return errors.ErrDeviceNotFound
	}

	// Delete device
	delete(r.devices, macAddress)
	return nil
}

// List returns all devices with pagination
func (r *DeviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert map to slice
	allDevices := make([]*entities.Device, 0, len(r.devices))
	for _, device := range r.devices {
		allDevices = append(allDevices, device)
	}

	// Handle empty results
	if len(allDevices) == 0 {
		return []*entities.Device{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(allDevices) {
		return []*entities.Device{}, nil
	}

	end := len(allDevices)
	if limit > 0 {
		end = start + limit
		if end > len(allDevices) {
			end = len(allDevices)
		}
	}

	return allDevices[start:end], nil
}

// Transaction executes multiple repository operations (no-op for memory implementation)
// Since this is in-memory, transactions are essentially atomic by default due to the mutex
func (r *DeviceRepository) Transaction(ctx context.Context, fn func(repo ports.DeviceRepository) error) error {
	// For in-memory implementation, we just execute the function directly
	// The mutex in each method provides thread safety
	return fn(r)
}