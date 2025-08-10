package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/models"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

func TestGormPostgresDB_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test database configuration using environment variables with defaults
	cfg := &config.DatabaseConfig{
		Host:            getTestEnv("TEST_DB_HOST", "localhost"),
		Port:            5432,
		User:            getTestEnv("TEST_DB_USER", "postgres"),
		Password:        getTestEnv("TEST_DB_PASSWORD", "password"),
		Name:            getTestEnv("TEST_DB_NAME", "test_iot_smart_irrigation"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// Create test logger factory
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	// Initialize GORM database
	gormDB, err := NewGormPostgresDB(cfg, loggerFactory)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}
	defer gormDB.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = gormDB.Ping(ctx)
	require.NoError(t, err, "Should be able to ping the database")

	// Test auto-migrations
	err = gormDB.AutoMigrate()
	require.NoError(t, err, "Auto-migrations should succeed")

	// Test health check
	err = gormDB.HealthCheck(ctx)
	assert.NoError(t, err, "Health check should pass")

	// Test basic CRUD operations with DeviceModel
	testDevice := &models.DeviceModel{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Test Location",
		Status:              "registered",
		RegisteredAt:        time.Now(),
		LastSeen:            time.Now(),
	}

	// Create device
	result := gormDB.GetDB().Create(testDevice)
	assert.NoError(t, result.Error, "Should be able to create device")

	// Find device
	var foundDevice models.DeviceModel
	result = gormDB.GetDB().Where("mac_address = ?", testDevice.MACAddress).First(&foundDevice)
	assert.NoError(t, result.Error, "Should be able to find device")
	assert.Equal(t, testDevice.MACAddress, foundDevice.MACAddress)
	assert.Equal(t, testDevice.DeviceName, foundDevice.DeviceName)

	// Update device
	foundDevice.Status = "online"
	result = gormDB.GetDB().Save(&foundDevice)
	assert.NoError(t, result.Error, "Should be able to update device")

	// Verify update
	var updatedDevice models.DeviceModel
	result = gormDB.GetDB().Where("mac_address = ?", testDevice.MACAddress).First(&updatedDevice)
	assert.NoError(t, result.Error, "Should be able to find updated device")
	assert.Equal(t, "online", updatedDevice.Status)

	// Soft delete device
	result = gormDB.GetDB().Delete(&updatedDevice)
	assert.NoError(t, result.Error, "Should be able to soft delete device")

	// Verify soft delete - device should not be found in normal queries
	result = gormDB.GetDB().Where("mac_address = ?", testDevice.MACAddress).First(&models.DeviceModel{})
	assert.Error(t, result.Error, "Soft deleted device should not be found in normal queries")

	// Find with Unscoped should still find it
	var softDeletedDevice models.DeviceModel
	result = gormDB.GetDB().Unscoped().Where("mac_address = ?", testDevice.MACAddress).First(&softDeletedDevice)
	assert.NoError(t, result.Error, "Should be able to find soft deleted device with Unscoped")
	assert.False(t, softDeletedDevice.DeletedAt.Time.IsZero(), "DeletedAt should be set")

	// Hard delete for cleanup
	result = gormDB.GetDB().Unscoped().Where("mac_address = ?", testDevice.MACAddress).Delete(&models.DeviceModel{})
	assert.NoError(t, result.Error, "Should be able to hard delete device")
}

func TestGormPostgresDB_ValidationHooks(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test database configuration using environment variables with defaults
	cfg := &config.DatabaseConfig{
		Host:            getTestEnv("TEST_DB_HOST", "localhost"),
		Port:            5432,
		User:            getTestEnv("TEST_DB_USER", "postgres"),
		Password:        getTestEnv("TEST_DB_PASSWORD", "password"),
		Name:            getTestEnv("TEST_DB_NAME", "test_iot_smart_irrigation"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// Create test logger factory
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	// Initialize GORM database
	gormDB, err := NewGormPostgresDB(cfg, loggerFactory)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}
	defer gormDB.Close()

	// Ensure migrations are run
	err = gormDB.AutoMigrate()
	require.NoError(t, err)

	// Test valid device model creation with hooks
	validDeviceModel := &models.DeviceModel{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Valid Device",
		IPAddress:           "192.168.1.101",
		LocationDescription: "Valid Location",
		Status:              "registered",
	}

	result := gormDB.GetDB().Create(validDeviceModel)
	assert.NoError(t, result.Error, "Should create valid device successfully")

	// Verify timestamps were set by hooks
	assert.False(t, validDeviceModel.RegisteredAt.IsZero(), "RegisteredAt should be set by BeforeCreate hook")
	assert.False(t, validDeviceModel.LastSeen.IsZero(), "LastSeen should be set by BeforeCreate hook")
	assert.False(t, validDeviceModel.CreatedAt.IsZero(), "CreatedAt should be set by GORM")
	assert.False(t, validDeviceModel.UpdatedAt.IsZero(), "UpdatedAt should be set by GORM")

	// Cleanup
	gormDB.GetDB().Unscoped().Where("mac_address = ?", validDeviceModel.MACAddress).Delete(&models.DeviceModel{})
}

// getTestEnv gets an environment variable with a fallback default value for testing
func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
