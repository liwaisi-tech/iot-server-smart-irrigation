package persistence

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
)

// TestDatabase setup and teardown helpers
func setupTestDB(t *testing.T) (*database.PostgresDB, func()) {
	// Skip tests if database is not available or not configured for testing
	if os.Getenv("POSTGRES_TEST_DSN") == "" {
		t.Skip("Skipping PostgreSQL tests: POSTGRES_TEST_DSN not set")
	}

	cfg := &config.DatabaseConfig{
		Host:            getTestEnv("DB_TEST_HOST", "localhost"),
		Port:            getTestEnvInt("DB_TEST_PORT", 5432),
		User:            getTestEnv("DB_TEST_USER", "postgres"),
		Password:        getTestEnv("DB_TEST_PASSWORD", "postgres"),
		Name:            getTestEnv("DB_TEST_NAME", "iot_test"),
		SSLMode:         getTestEnv("DB_TEST_SSL_MODE", "disable"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err, "Failed to create test database connection")

	// Run migrations for tests
	err = db.RunMigrations("../../../migrations")
	require.NoError(t, err, "Failed to run migrations for tests")

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		_, err := db.ExecContext(context.Background(), "DELETE FROM devices")
		if err != nil {
			t.Logf("Warning: Failed to clean up test data: %v", err)
		}
		db.Close()
	}

	return db, cleanup
}

func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTestEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func TestNewPostgresDeviceRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	assert.NotNil(t, repo, "NewPostgresDeviceRepository() returned nil")
}

func TestPostgresDeviceRepository_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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

	// Verify device was saved by finding it
	savedDevice, err := repo.FindByMACAddress(ctx, device.MACAddress)
	assert.NoError(t, err, "FindByMACAddress() after save unexpected error")
	require.NotNil(t, savedDevice, "Save() device was not found after save")

	assert.Equal(t, device.MACAddress, savedDevice.MACAddress, "Save() MAC address mismatch")
	assert.Equal(t, device.DeviceName, savedDevice.DeviceName, "Save() device name mismatch")
	assert.Equal(t, device.IPAddress, savedDevice.IPAddress, "Save() IP address mismatch")
	assert.Equal(t, device.LocationDescription, savedDevice.LocationDescription, "Save() location description mismatch")
	assert.Equal(t, device.Status, savedDevice.Status, "Save() status mismatch")
}

func TestPostgresDeviceRepository_Save_NilDevice(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	err := repo.Save(ctx, nil)
	assert.Error(t, err, "Save() expected error for nil device but got none")
}

func TestPostgresDeviceRepository_Save_DuplicateMAC(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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
	assert.Equal(t, domainerrors.ErrDeviceAlreadyExists, err, "Save() expected ErrDeviceAlreadyExists")

	// Verify only first device exists
	savedDevice, err := repo.FindByMACAddress(ctx, device1.MACAddress)
	assert.NoError(t, err, "FindByMACAddress() after duplicate save unexpected error")
	assert.Equal(t, device1.DeviceName, savedDevice.DeviceName, "Save() first device should remain unchanged")
}

func TestPostgresDeviceRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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
	device.DeviceName = "Updated Device Name"
	device.IPAddress = "192.168.1.101"
	err = device.UpdateStatus("online")
	require.NoError(t, err, "Failed to update device status")

	err = repo.Update(ctx, device)
	assert.NoError(t, err, "Update() unexpected error")

	// Verify device was updated
	updatedDevice, err := repo.FindByMACAddress(ctx, device.MACAddress)
	assert.NoError(t, err, "FindByMACAddress() after update unexpected error")
	assert.Equal(t, "Updated Device Name", updatedDevice.DeviceName, "Update() device name not updated")
	assert.Equal(t, "192.168.1.101", updatedDevice.IPAddress, "Update() device IP not updated")
	assert.Equal(t, "online", updatedDevice.Status, "Update() device status not updated")
}

func TestPostgresDeviceRepository_Update_NilDevice(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, nil)
	assert.Error(t, err, "Update() expected error for nil device but got none")
}

func TestPostgresDeviceRepository_Update_NonExistentDevice(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "Update() expected ErrDeviceNotFound")
}

func TestPostgresDeviceRepository_FindByMACAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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
	assert.Equal(t, originalDevice.DeviceName, foundDevice.DeviceName, "FindByMACAddress() device name mismatch")
	assert.Equal(t, originalDevice.IPAddress, foundDevice.IPAddress, "FindByMACAddress() IP address mismatch")
	assert.Equal(t, originalDevice.LocationDescription, foundDevice.LocationDescription, "FindByMACAddress() location description mismatch")
	assert.Equal(t, originalDevice.Status, foundDevice.Status, "FindByMACAddress() status mismatch")
}

func TestPostgresDeviceRepository_FindByMACAddress_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	device, err := repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	assert.Error(t, err, "FindByMACAddress() expected error for non-existent device but got none")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "FindByMACAddress() expected ErrDeviceNotFound")
	assert.Nil(t, device, "FindByMACAddress() expected nil device")
}

func TestPostgresDeviceRepository_FindByMACAddress_EmptyMAC(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	device, err := repo.FindByMACAddress(ctx, "")
	assert.Error(t, err, "FindByMACAddress() expected error for empty MAC but got none")
	assert.Nil(t, device, "FindByMACAddress() expected nil device")
}

func TestPostgresDeviceRepository_Exists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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

func TestPostgresDeviceRepository_Exists_EmptyMAC(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, "")
	assert.Error(t, err, "Exists() expected error for empty MAC but got none")
	assert.False(t, exists, "Exists() expected false for empty MAC")
}

func TestPostgresDeviceRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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
	exists, err := repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Exists() unexpected error")
	assert.True(t, exists, "Expected device to exist before delete")

	// Delete device
	err = repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Delete() unexpected error")

	// Verify device was deleted
	exists, err = repo.Exists(ctx, "AA:BB:CC:DD:EE:FF")
	assert.NoError(t, err, "Exists() after delete unexpected error")
	assert.False(t, exists, "Expected device to not exist after delete")

	// Verify device is no longer findable
	_, err = repo.FindByMACAddress(ctx, "AA:BB:CC:DD:EE:FF")
	assert.Error(t, err, "FindByMACAddress() after delete should return error")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "FindByMACAddress() after delete expected ErrDeviceNotFound")
}

func TestPostgresDeviceRepository_Delete_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "AA:BB:CC:DD:EE:FF")
	assert.Error(t, err, "Delete() expected error for non-existent device but got none")
	assert.Equal(t, domainerrors.ErrDeviceNotFound, err, "Delete() expected ErrDeviceNotFound")
}

func TestPostgresDeviceRepository_Delete_EmptyMAC(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "")
	assert.Error(t, err, "Delete() expected error for empty MAC but got none")
}

func TestPostgresDeviceRepository_List_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	devices, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err, "List() unexpected error")
	assert.NotNil(t, devices, "List() returned nil slice")
	assert.Empty(t, devices, "List() expected empty slice")
}

func TestPostgresDeviceRepository_List_AllDevices(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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

	// List all devices
	listedDevices, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err, "List() unexpected error")
	assert.Len(t, listedDevices, 3, "List() expected 3 devices")

	// Verify all devices are present (order is by registered_at DESC)
	deviceMACs := make(map[string]bool)
	for _, device := range listedDevices {
		deviceMACs[device.MACAddress] = true
	}

	for _, originalDevice := range devices {
		assert.True(t, deviceMACs[originalDevice.MACAddress], "List() missing device with MAC %s", originalDevice.MACAddress)
	}
}

func TestPostgresDeviceRepository_List_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
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

		// Add small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
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
		{"zero limit with offset", 1, 0, 4}, // Default limit applies
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := repo.List(ctx, tt.offset, tt.limit)
			assert.NoError(t, err, "List() unexpected error")
			assert.Len(t, devices, tt.expectedCount, "List() expected device count mismatch")
		})
	}
}

func TestPostgresDeviceRepository_List_NegativeValues(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	// Test negative offset
	devices, err := repo.List(ctx, -1, 10)
	assert.Error(t, err, "List() expected error for negative offset")
	assert.Nil(t, devices, "List() expected nil devices for negative offset")

	// Test negative limit
	devices, err = repo.List(ctx, 0, -1)
	assert.Error(t, err, "List() expected error for negative limit")
	assert.Nil(t, devices, "List() expected nil devices for negative limit")
}

// Concurrent access tests
func TestPostgresDeviceRepository_ConcurrentAccess_SaveAndRead(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	const numGoroutines = 10
	const devicesPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*devicesPerGoroutine*2)

	// Concurrent saves
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < devicesPerGoroutine; j++ {
				device, err := entities.NewDevice(
					fmt.Sprintf("AA:BB:CC:DD:%02X:%02X", goroutineID, j),
					fmt.Sprintf("Device-%d-%d", goroutineID, j),
					fmt.Sprintf("192.168.%d.%d", goroutineID, j+1), // j+1 to avoid .0
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
				if err != nil && err.Error() != domainerrors.ErrDeviceNotFound.Error() {
					errors <- fmt.Errorf("unexpected error finding device: %w", err)
					return
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

func TestPostgresDeviceRepository_Transaction_Rollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	// Create device within transaction
	device, err := entities.NewDevice(
		"AA:BB:CC:DD:EE:FF",
		"Transaction Test Device",
		"192.168.1.100",
		"Transaction Test Location",
	)
	require.NoError(t, err, "Failed to create device")

	// Save device within transaction
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO devices (
			mac_address, device_name, ip_address, location_description,
			registered_at, last_seen, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		device.MACAddress,
		device.DeviceName,
		device.IPAddress,
		device.LocationDescription,
		device.RegisteredAt,
		device.LastSeen,
		device.Status,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err, "Failed to insert device in transaction")

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err, "Failed to rollback transaction")

	// Verify device was not saved
	repo := NewPostgresDeviceRepository(db)
	exists, err := repo.Exists(ctx, device.MACAddress)
	assert.NoError(t, err, "Exists() after rollback unexpected error")
	assert.False(t, exists, "Device should not exist after transaction rollback")
}