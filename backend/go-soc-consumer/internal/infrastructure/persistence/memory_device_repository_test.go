package persistence

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

func TestNewMemoryDeviceRepository(t *testing.T) {
	repo := NewMemoryDeviceRepository()

	if repo == nil {
		t.Errorf("NewMemoryDeviceRepository() returned nil")
	}

	if repo.devices == nil {
		t.Errorf("NewMemoryDeviceRepository() devices map not initialized")
	}

	if len(repo.devices) != 0 {
		t.Errorf("NewMemoryDeviceRepository() devices map should be empty initially")
	}
}

func TestMemoryDeviceRepository_Save(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Errorf("Save() unexpected error: %v", err)
	}

	// Verify device was saved
	savedDevice, exists := repo.devices[device.MACAddress]
	if !exists {
		t.Errorf("Save() device was not saved to repository")
	}

	if savedDevice != device {
		t.Errorf("Save() saved device is not the same reference")
	}
}

func TestMemoryDeviceRepository_Save_NilDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.Save(ctx, nil)
	if err == nil {
		t.Errorf("Save() expected error for nil device but got none")
	}
}

func TestMemoryDeviceRepository_Save_DuplicateMAC(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device1, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device 1",
		"192.168.1.100",
		"Test Location 1",
	)
	if err != nil {
		t.Fatalf("Failed to create device1: %v", err)
	}

	device2, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device 2",
		"192.168.1.101",
		"Test Location 2",
	)
	if err != nil {
		t.Fatalf("Failed to create device2: %v", err)
	}

	// Save first device
	err = repo.Save(ctx, device1)
	if err != nil {
		t.Errorf("Save() first device unexpected error: %v", err)
	}

	// Try to save second device with same MAC
	err = repo.Save(ctx, device2)
	if err == nil {
		t.Errorf("Save() expected error for duplicate MAC address but got none")
	}

	// Verify only first device exists
	if len(repo.devices) != 1 {
		t.Errorf("Save() expected 1 device after duplicate attempt, got %d", len(repo.devices))
	}

	savedDevice := repo.devices[device1.MACAddress]
	if savedDevice.DeviceName != device1.DeviceName {
		t.Errorf("Save() first device should remain unchanged")
	}
}

func TestMemoryDeviceRepository_Update(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save initial device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Fatalf("Failed to save initial device: %v", err)
	}

	// Update device
	device.DeviceName = "Updated Device Name"
	device.IPAddress = "192.168.1.101"

	err = repo.Update(ctx, device)
	if err != nil {
		t.Errorf("Update() unexpected error: %v", err)
	}

	// Verify device was updated
	updatedDevice := repo.devices[device.MACAddress]
	if updatedDevice.DeviceName != "Updated Device Name" {
		t.Errorf("Update() device name not updated")
	}

	if updatedDevice.IPAddress != "192.168.1.101" {
		t.Errorf("Update() device IP not updated")
	}
}

func TestMemoryDeviceRepository_Update_NilDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.Update(ctx, nil)
	if err == nil {
		t.Errorf("Update() expected error for nil device but got none")
	}
}

func TestMemoryDeviceRepository_Update_NonExistentDevice(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Update(ctx, device)
	if err == nil {
		t.Errorf("Update() expected error for non-existent device but got none")
	}

	if len(repo.devices) != 0 {
		t.Errorf("Update() should not add device when updating non-existent device")
	}
}

func TestMemoryDeviceRepository_FindByMACAddress(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save device
	originalDevice, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, originalDevice)
	if err != nil {
		t.Fatalf("Failed to save device: %v", err)
	}

	// Find device
	foundDevice, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Errorf("FindByMACAddress() unexpected error: %v", err)
	}

	if foundDevice == nil {
		t.Errorf("FindByMACAddress() expected device but got nil")
		return
	}

	if foundDevice.MACAddress != originalDevice.MACAddress {
		t.Errorf("FindByMACAddress() MAC address mismatch")
	}

	if foundDevice.DeviceName != originalDevice.DeviceName {
		t.Errorf("FindByMACAddress() device name mismatch")
	}
}

func TestMemoryDeviceRepository_FindByMACAddress_NotFound(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	if err == nil {
		t.Errorf("FindByMACAddress() expected error for non-existent device but got none")
	}

	if device != nil {
		t.Errorf("FindByMACAddress() expected nil device but got %v", device)
	}
}

func TestMemoryDeviceRepository_FindByMACAddress_EmptyMAC(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	device, err := repo.FindByMACAddress(ctx, "")
	if err == nil {
		t.Errorf("FindByMACAddress() expected error for empty MAC but got none")
	}

	if device != nil {
		t.Errorf("FindByMACAddress() expected nil device but got %v", device)
	}
}

func TestMemoryDeviceRepository_Exists(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Check non-existent device
	exists, err := repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Errorf("Exists() unexpected error: %v", err)
	}

	if exists {
		t.Errorf("Exists() expected false for non-existent device")
	}

	// Create and save device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Fatalf("Failed to save device: %v", err)
	}

	// Check existing device
	exists, err = repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Errorf("Exists() unexpected error: %v", err)
	}

	if !exists {
		t.Errorf("Exists() expected true for existing device")
	}
}

func TestMemoryDeviceRepository_Delete(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Fatalf("Failed to save device: %v", err)
	}

	// Verify device exists
	if len(repo.devices) != 1 {
		t.Errorf("Expected 1 device before delete, got %d", len(repo.devices))
	}

	// Delete device
	err = repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Errorf("Delete() unexpected error: %v", err)
	}

	// Verify device was deleted
	if len(repo.devices) != 0 {
		t.Errorf("Expected 0 devices after delete, got %d", len(repo.devices))
	}

	// Verify device is no longer accessible
	_, exists := repo.devices["AA:BB:CC:DD:EE:FF"]
	if exists {
		t.Errorf("Delete() device still exists in repository")
	}
}

func TestMemoryDeviceRepository_Delete_NonExistent(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	if err == nil {
		t.Errorf("Delete() expected error for non-existent device but got none")
	}
}

func TestMemoryDeviceRepository_List_Empty(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	devices, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Errorf("List() unexpected error: %v", err)
	}

	if devices == nil {
		t.Errorf("List() returned nil slice")
	}

	if len(devices) != 0 {
		t.Errorf("List() expected empty slice, got %d devices", len(devices))
	}
}

func TestMemoryDeviceRepository_List_AllDevices(t *testing.T) {
	repo := NewMemoryDeviceRepository()
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
		if err != nil {
			t.Fatalf("Failed to create device %d: %v", i, err)
		}
		devices[i] = device

		err = repo.Save(ctx, device)
		if err != nil {
			t.Fatalf("Failed to save device %d: %v", i, err)
		}
	}

	// List all devices (no pagination)
	listedDevices, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Errorf("List() unexpected error: %v", err)
	}

	if len(listedDevices) != 3 {
		t.Errorf("List() expected 3 devices, got %d", len(listedDevices))
	}

	// Verify all devices are present (order may vary)
	deviceMACs := make(map[string]bool)
	for _, device := range listedDevices {
		deviceMACs[device.MACAddress] = true
	}

	for _, originalDevice := range devices {
		if !deviceMACs[originalDevice.MACAddress] {
			t.Errorf("List() missing device with MAC %s", originalDevice.MACAddress)
		}
	}
}

func TestMemoryDeviceRepository_List_Pagination(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create and save 5 devices
	for i := 0; i < 5; i++ {
		device, err := entities.NewDevice(
			fmt.Sprintf("AA:BB:CC:DD:EE:F%d", i),
			fmt.Sprintf("Test Device %d", i),
			fmt.Sprintf("192.168.1.10%d", i),
			fmt.Sprintf("Test Location %d", i),
		)
		if err != nil {
			t.Fatalf("Failed to create device %d: %v", i, err)
		}

		err = repo.Save(ctx, device)
		if err != nil {
			t.Fatalf("Failed to save device %d: %v", i, err)
		}
	}

	tests := []struct {
		name           string
		offset         int
		limit          int
		expectedCount  int
	}{
		{"first page", 0, 2, 2},
		{"second page", 2, 2, 2},
		{"last page", 4, 2, 1},
		{"limit larger than remaining", 3, 5, 2},
		{"offset at end", 5, 2, 0},
		{"offset beyond end", 10, 2, 0},
		{"zero limit", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := repo.List(ctx, tt.offset, tt.limit)
			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
			}

			if len(devices) != tt.expectedCount {
				t.Errorf("List() expected %d devices, got %d", tt.expectedCount, len(devices))
			}
		})
	}
}

func TestMemoryDeviceRepository_List_OffsetGreaterThanTotal(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Save one device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Test Device",
		"192.168.1.100",
		"Test Location",
	)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Fatalf("Failed to save device: %v", err)
	}

	// Request with offset greater than total devices
	devices, err := repo.List(ctx, 5, 10)
	if err != nil {
		t.Errorf("List() unexpected error: %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("List() expected empty slice when offset > total, got %d devices", len(devices))
	}
}

// Concurrent access tests
func TestMemoryDeviceRepository_ConcurrentAccess_SaveAndRead(t *testing.T) {
	repo := NewMemoryDeviceRepository()
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
					errMsg := fmt.Sprintf("device with MAC address %s not found", macAddress)
					if err.Error() != errMsg {
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
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify total device count
	devices, err := repo.List(ctx, 0, 1000)
	if err != nil {
		t.Errorf("List() after concurrent access error: %v", err)
	}

	expectedCount := numGoroutines * devicesPerGoroutine
	if len(devices) != expectedCount {
		t.Errorf("Expected %d devices after concurrent saves, got %d", expectedCount, len(devices))
	}
}

func TestMemoryDeviceRepository_ConcurrentAccess_UpdateAndRead(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	ctx := context.Background()

	// Create initial device
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Initial Device",
		"192.168.1.100",
		"Initial Location",
	)
	if err != nil {
		t.Fatalf("Failed to create initial device: %v", err)
	}

	err = repo.Save(ctx, device)
	if err != nil {
		t.Fatalf("Failed to save initial device: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Concurrent updates
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			// Find and update device
			foundDevice, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
			if err != nil {
				errors <- err
				return
			}

			foundDevice.DeviceName = fmt.Sprintf("Updated Device %d", goroutineID)
			foundDevice.MarkOnline()

			if err := repo.Update(ctx, foundDevice); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			_, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent update/read error: %v", err)
	}

	// Verify device still exists and was updated
	finalDevice, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Errorf("FindByMACAddress() after concurrent updates error: %v", err)
	}

	if finalDevice == nil {
		t.Errorf("Device should still exist after concurrent updates")
	}

	// The final device name will be from one of the updates
	if finalDevice.DeviceName == "Initial Device" {
		t.Errorf("Device name should have been updated")
	}
}

func TestMemoryDeviceRepository_ConcurrentAccess_RaceCondition(t *testing.T) {
	// This test specifically checks for race conditions using go test -race
	repo := NewMemoryDeviceRepository()
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

