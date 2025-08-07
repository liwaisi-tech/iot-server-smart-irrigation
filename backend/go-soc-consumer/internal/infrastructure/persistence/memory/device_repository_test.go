package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

func TestNewDeviceRepository(t *testing.T) {
	repo := NewDeviceRepository()

	assert.NotNil(t, repo, "NewDeviceRepository() returned nil")
	
	// Cast to concrete type to check internal state
	memRepo := repo.(*DeviceRepository)
	assert.NotNil(t, memRepo.devices, "NewDeviceRepository() devices map not initialized")
	assert.Empty(t, memRepo.devices, "NewDeviceRepository() devices map should be empty initially")
}

func TestDeviceRepository_Save(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Save(ctx, device)
	assert.NoError(t, err, "Save() unexpected error")

	// Verify device was saved
	savedDevice, exists := repo.devices[device.MACAddress]
	assert.True(t, exists, "Save() device was not saved to repository")
	assert.Same(t, device, savedDevice, "Save() saved device is not the same reference")
}

func TestDeviceRepository_Save_NilDevice(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	err := repo.Save(ctx, nil)
	assert.Error(t, err, "Save() expected error for nil device but got none")
}

func TestDeviceRepository_Save_DuplicateMAC(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	device1, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device 1",
		"192.168.1.100",
		"Test Location 1",
	)
	require.NoError(t, err, "Failed to create device1")

	device2, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device 2",
		"192.168.1.101",
		"Test Location 2",
	)
	require.NoError(t, err, "Failed to create device2")

	// Save first device
	err = repo.Save(ctx, device1)
	assert.NoError(t, err, "Save() first device unexpected error")

	// Try to save second device with same MAC
	err = repo.Save(ctx, device2)
	assert.Error(t, err, "Save() expected error for duplicate MAC address but got none")
	assert.Equal(t, domainerrors.ErrDeviceAlreadyExists, err, "Save() should return ErrDeviceAlreadyExists")

	// Verify only first device exists
	assert.Len(t, repo.devices, 1, "Save() expected 1 device after duplicate attempt")

	savedDevice := repo.devices[device1.MACAddress]
	assert.Equal(t, device1.DeviceName, savedDevice.DeviceName, "Save() first device should remain unchanged")
}

func TestDeviceRepository_Update(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	// Create and save initial device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Save(ctx, device)
	require.NoError(t, err, "Failed to save initial device")

	// Update device
	device.SetDeviceName("Updated Device Name")
	device.SetIPAddress("192.168.1.101")

	err = repo.Update(ctx, device)
	assert.NoError(t, err, "Update() unexpected error")

	// Verify device was updated
	updatedDevice := repo.devices[device.MACAddress]
	assert.Equal(t, "Updated Device Name", updatedDevice.GetDeviceName(), "Update() device name not updated")
	assert.Equal(t, "192.168.1.101", updatedDevice.GetIPAddress(), "Update() device IP not updated")
}

func TestDeviceRepository_Update_NilDevice(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	err := repo.Update(ctx, nil)
	assert.Error(t, err, "Update() expected error for nil device but got none")
}

func TestDeviceRepository_Update_NonExistentDevice(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Update(ctx, device)
	assert.Error(t, err, "Update() expected error for non-existent device but got none")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "Update() should return ErrDeviceNotFound")
	assert.Empty(t, repo.devices, "Update() should not add device when updating non-existent device")
}

func TestDeviceRepository_FindByMACAddress(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	// Create and save device
	originalDevice, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Save(ctx, originalDevice)
	require.NoError(t, err, "Failed to save device")

	// Find device
	foundDevice, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "FindByMACAddress() unexpected error")
	require.NotNil(t, foundDevice, "FindByMACAddress() expected device but got nil")

	assert.Equal(t, originalDevice.MACAddress, foundDevice.MACAddress, "FindByMACAddress() MAC address mismatch")
	assert.Equal(t, originalDevice.GetDeviceName(), foundDevice.GetDeviceName(), "FindByMACAddress() device name mismatch")
}

func TestDeviceRepository_FindByMACAddress_NotFound(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	device, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	assert.Error(t, err, "FindByMACAddress() expected error for non-existent device but got none")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "FindByMACAddress() should return ErrDeviceNotFound")
	assert.Nil(t, device, "FindByMACAddress() expected nil device")
}

func TestDeviceRepository_Exists(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	// Check non-existent device
	exists, err := repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Exists() unexpected error")
	assert.False(t, exists, "Exists() expected false for non-existent device")

	// Create and save device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Save(ctx, device)
	require.NoError(t, err, "Failed to save device")

	// Check existing device
	exists, err = repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Exists() unexpected error")
	assert.True(t, exists, "Exists() expected true for existing device")
}

func TestDeviceRepository_Delete(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	// Create and save device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	err = repo.Save(ctx, device)
	require.NoError(t, err, "Failed to save device")

	// Verify device exists
	assert.Len(t, repo.devices, 1, "Expected 1 device before delete")

	// Delete device
	err = repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Delete() unexpected error")

	// Verify device was deleted
	assert.Empty(t, repo.devices, "Expected 0 devices after delete")

	// Verify device is no longer accessible
	_, exists := repo.devices["AA:BB:CC:DD:EE:FF"]
	assert.False(t, exists, "Delete() device still exists in repository")
}

func TestDeviceRepository_Delete_NonExistent(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	assert.Error(t, err, "Delete() expected error for non-existent device but got none")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "Delete() should return ErrDeviceNotFound")
}

func TestDeviceRepository_List_Empty(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	devices, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err, "List() unexpected error")
	assert.NotNil(t, devices, "List() returned nil slice")
	assert.Empty(t, devices, "List() expected empty slice")
}

func TestDeviceRepository_List_AllDevices(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	// Create and save multiple devices
	devices := make([]*entities.Device, 3)
	for i := 0; i < 3; i++ {
		device, err := entities.NewDevice(
			fmt.Sprintf("AA:BB:CC:DD:EE:F%d", i),
			fmt.Sprintf("Test Device %d", i),
			fmt.Sprintf("192.168.1.10%d", i),
			fmt.Sprintf("Test Location %d", i),
		)
		require.NoError(t, err, "Failed to create device %d", i)
		devices[i] = device

		err = repo.Save(ctx, device)
		require.NoError(t, err, "Failed to save device %d", i)
	}

	// List all devices (no limit)
	listedDevices, err := repo.List(ctx, 0, 0)
	assert.NoError(t, err, "List() unexpected error")
	assert.Len(t, listedDevices, 3, "List() expected 3 devices")

	// Verify all devices are present (order may vary)
	deviceMACs := make(map[string]bool)
	for _, device := range listedDevices {
		deviceMACs[device.MACAddress] = true
	}

	for _, originalDevice := range devices {
		assert.True(t, deviceMACs[originalDevice.MACAddress], "List() missing device with MAC %s", originalDevice.MACAddress)
	}
}

func TestDeviceRepository_List_Pagination(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	// Create and save 5 devices
	for i := 0; i < 5; i++ {
		device, err := entities.NewDevice(
			fmt.Sprintf("AA:BB:CC:DD:EE:F%d", i),
			fmt.Sprintf("Test Device %d", i),
			fmt.Sprintf("192.168.1.10%d", i),
			fmt.Sprintf("Test Location %d", i),
		)
		require.NoError(t, err, "Failed to create device %d", i)

		err = repo.Save(ctx, device)
		require.NoError(t, err, "Failed to save device %d", i)
	}

	tests := []struct {
		name          string
		offset        int
		limit         int
		expectedCount int
	}{
		{"first page", 0, 2, 2},
		{"second page", 2, 2, 2},
		{"last page", 4, 2, 1},
		{"limit larger than remaining", 3, 5, 2},
		{"offset at end", 5, 2, 0},
		{"offset beyond end", 10, 2, 0},
		{"zero limit (no limit)", 0, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := repo.List(ctx, tt.offset, tt.limit)
			assert.NoError(t, err, "List() unexpected error")
			assert.Len(t, devices, tt.expectedCount, "List() expected device count mismatch")
		})
	}
}

func TestDeviceRepository_List_NegativeValues(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	// Test negative offset
	devices, err := repo.List(ctx, -1, 10)
	assert.Error(t, err, "List() expected error for negative offset")
	assert.Nil(t, devices, "List() should return nil for negative offset")

	// Test negative limit
	devices, err = repo.List(ctx, 0, -1)
	assert.Error(t, err, "List() expected error for negative limit")
	assert.Nil(t, devices, "List() should return nil for negative limit")
}

func TestDeviceRepository_Transaction(t *testing.T) {
	repo := NewDeviceRepository().(*DeviceRepository)
	ctx := context.Background()

	// Test successful transaction
	err := repo.Transaction(ctx, func(txRepo ports.DeviceRepository) error {
		device, err := entities.NewDevice(
			"AA:BB:CC:DD:EE:FF",
			"Test Device",
			"192.168.1.100",
			"Test Location",
		)
		if err != nil {
			return err
		}

		return txRepo.Save(ctx, device)
	})

	assert.NoError(t, err, "Transaction() unexpected error")

	// Verify device was saved
	exists, err := repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Exists() unexpected error")
	assert.True(t, exists, "Device should exist after successful transaction")
}

// Concurrent access tests
func TestDeviceRepository_ConcurrentAccess_SaveAndRead(t *testing.T) {
	repo := NewDeviceRepository()
	ctx := context.Background()

	const numGoroutines = 10
	const devicesPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*devicesPerGoroutine)

	// Concurrent saves
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < devicesPerGoroutine; j++ {
				device, err := entities.NewDevice(
					fmt.Sprintf("AA:BB:CC:DD:%02X:%02X", goroutineID, j),
					fmt.Sprintf("Device-%d-%d", goroutineID, j),
					fmt.Sprintf("192.168.%d.%d", goroutineID, j),
					fmt.Sprintf("Location-%d-%d", goroutineID, j),
				)
				if err != nil {
					errors <- err
					return
				}

				if err := repo.Save(ctx, device); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < devicesPerGoroutine; j++ {
				macAddress := fmt.Sprintf("AA:BB:CC:DD:%02X:%02X", goroutineID, j)

				// Try to find device (may not exist yet)
				_, err := repo.FindByMACAddress(ctx, macAddress)
				if err != nil {
					// Only report unexpected errors, not "not found" errors
					if err != domainerrors.ErrDeviceNotFound {
						errors <- fmt.Errorf("unexpected error finding device: %w", err)
						return
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err, "Concurrent access error")
	}

	// Verify total device count
	devices, err := repo.List(ctx, 0, 1000)
	assert.NoError(t, err, "List() after concurrent access error")

	expectedCount := numGoroutines * devicesPerGoroutine
	assert.Len(t, devices, expectedCount, "Expected device count after concurrent saves")
}

func TestDeviceRepository_ConcurrentAccess_RaceCondition(t *testing.T) {
	// This test specifically checks for race conditions using go test -race
	repo := NewDeviceRepository()
	ctx := context.Background()

	const numGoroutines = 20
	var wg sync.WaitGroup

	// Mix of operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			macAddress := fmt.Sprintf("AA:BB:CC:DD:EE:%02X", goroutineID)

			// Create device
			device, err := entities.NewDevice(
				macAddress,
				fmt.Sprintf("Device %d", goroutineID),
				fmt.Sprintf("192.168.1.%d", goroutineID),
				fmt.Sprintf("Location %d", goroutineID),
			)
			if err != nil {
				return
			}

			// Save
			repo.Save(ctx, device)

			// Check existence
			repo.Exists(ctx, macAddress)

			// Find
			repo.FindByMACAddress(ctx, macAddress)

			// Update
			device.MarkOnline()
			repo.Update(ctx, device)

			// List (with different pagination)
			repo.List(ctx, goroutineID%5, 2)

			// Delete some devices
			if goroutineID%3 == 0 {
				repo.Delete(ctx, macAddress)
			}
		}(i)
	}

	wg.Wait()

	// If we reach here without race detector complaints, the test passes
}