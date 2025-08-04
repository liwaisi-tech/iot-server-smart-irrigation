package persistence

import (
	"context"
	"fmt"
	"sync"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

// MemoryDeviceRepository implements DeviceRepository using in-memory storage
type MemoryDeviceRepository struct {
	devices map[string]*entities.Device
	mu      sync.RWMutex
}

// NewMemoryDeviceRepository creates a new in-memory device repository
func NewMemoryDeviceRepository() *MemoryDeviceRepository {
	return &MemoryDeviceRepository{
		devices: make(map[string]*entities.Device),
	}
}

// Save saves a new device to the repository
func (r *MemoryDeviceRepository) Save(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device already exists
	if _, exists := r.devices[device.MACAddress]; exists {
		return fmt.Errorf("device with MAC address %s already exists", device.MACAddress)
	}

	// Save device
	r.devices[device.MACAddress] = device
	return nil
}

// Update updates an existing device in the repository
func (r *MemoryDeviceRepository) Update(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device exists
	if _, exists := r.devices[device.MACAddress]; !exists {
		return fmt.Errorf("device with MAC address %s not found", device.MACAddress)
	}

	// Update device
	r.devices[device.MACAddress] = device
	return nil
}

// FindByMACAddress finds a device by MAC address
func (r *MemoryDeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.devices[macAddress]
	if !exists {
		return nil, fmt.Errorf("device with MAC address %s not found", macAddress)
	}

	return device, nil
}

// Exists checks if a device with the given MAC address exists
func (r *MemoryDeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.devices[macAddress]
	return exists, nil
}

// Delete removes a device from the repository
func (r *MemoryDeviceRepository) Delete(ctx context.Context, macAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if device exists
	if _, exists := r.devices[macAddress]; !exists {
		return fmt.Errorf("device with MAC address %s not found", macAddress)
	}

	// Delete device
	delete(r.devices, macAddress)
	return nil
}

// List returns all devices with pagination
func (r *MemoryDeviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert map to slice
	allDevices := make([]*entities.Device, 0, len(r.devices))
	for _, device := range r.devices {
		allDevices = append(allDevices, device)
	}

	// Apply pagination
	start := offset
	if start >= len(allDevices) {
		return []*entities.Device{}, nil
	}

	end := start + limit
	if end > len(allDevices) {
		end = len(allDevices)
	}

	return allDevices[start:end], nil
}