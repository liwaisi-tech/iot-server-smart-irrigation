package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/domain/entities"
)

// setupTestDB creates a mock database for testing
func setupTestDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

// createTestDevice creates a test device entity
func createTestDevice() *entities.Device {
	return &entities.Device{
		MacAddress:          "00:11:22:33:44:55",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Test Location",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

func TestDeviceRepository_Create(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)

	tests := []struct {
		name        string
		device      *entities.Device
		expectError bool
		setupMock   func(sqlmock.Sqlmock, *entities.Device)
	}{
		{
			name:        "Valid device",
			device:      createTestDevice(),
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock, device *entities.Device) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "devices"`)).
					WithArgs(device.MacAddress, device.DeviceName, device.IPAddress, device.LocationDescription, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:        "Nil device",
			device:      nil,
			expectError: true,
			setupMock:   func(mock sqlmock.Sqlmock, device *entities.Device) {},
		},
		{
			name: "Empty MAC address",
			device: &entities.Device{
				DeviceName: "Test Device",
			},
			expectError: true,
			setupMock:   func(mock sqlmock.Sqlmock, device *entities.Device) {},
		},
		{
			name:        "Database error when creating device",
			device:      createTestDevice(),
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock, device *entities.Device) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "devices"`)).
					WithArgs(device.MacAddress, device.DeviceName, device.IPAddress, device.LocationDescription, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock, tt.device)

			ctx := context.Background()
			err := repo.Create(ctx, tt.device)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceRepository_GetByMacAddress(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)

	testDevice := createTestDevice()
	ctx := context.Background()

	tests := []struct {
		name        string
		macAddress  string
		expectError bool
		setupMock   func(sqlmock.Sqlmock, string)
	}{
		{
			name:        "Existing device",
			macAddress:  testDevice.MacAddress,
			expectError: false,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string) {
				rows := sqlmock.NewRows([]string{"mac_address", "device_name", "ip_address", "location_description", "created_at", "updated_at"}).
					AddRow(testDevice.MacAddress, testDevice.DeviceName, testDevice.IPAddress, testDevice.LocationDescription, testDevice.CreatedAt, testDevice.UpdatedAt)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" WHERE mac_address = $1 ORDER BY "devices"."mac_address" LIMIT $2`)).
					WithArgs(macAddress, 1).
					WillReturnRows(rows)
			},
		},
		{
			name:        "Non-existing device",
			macAddress:  "00:11:22:33:44:99",
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" WHERE mac_address = $1 ORDER BY "devices"."mac_address" LIMIT $2`)).
					WithArgs(macAddress, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name:        "Empty MAC address",
			macAddress:  "",
			expectError: true,
			setupMock:   func(mock sqlmock.Sqlmock, macAddress string) {},
		},
		{
			name:        "Database error when getting device",
			macAddress:  testDevice.MacAddress,
			expectError: true,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" WHERE mac_address = $1 ORDER BY "devices"."mac_address" LIMIT $2`)).
					WithArgs(macAddress, 1).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock, tt.macAddress)

			device, err := repo.GetByMacAddress(ctx, tt.macAddress)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, device)
				assert.Equal(t, tt.macAddress, device.MacAddress)
			}
		})
	}
}

func TestDeviceRepository_Update(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)

	testDevice := createTestDevice()
	testDevice.DeviceName = "Updated Device Name"
	testDevice.IPAddress = "192.168.1.101"
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "devices" SET "mac_address"=$1,"device_name"=$2,"ip_address"=$3,"location_description"=$4,"created_at"=$5,"updated_at"=$6 WHERE mac_address = $7`)).
		WithArgs(testDevice.MacAddress, testDevice.DeviceName, testDevice.IPAddress, testDevice.LocationDescription, sqlmock.AnyArg(), sqlmock.AnyArg(), testDevice.MacAddress).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Update(ctx, testDevice)
	assert.NoError(t, err)
}

func TestDeviceRepository_Delete(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)

	testDevice := createTestDevice()
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "devices" WHERE mac_address = $1`)).
		WithArgs(testDevice.MacAddress).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Delete(ctx, testDevice.MacAddress)
	assert.NoError(t, err)
}

func TestDeviceRepository_List(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)
	ctx := context.Background()

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectError   bool
		setupMock     func(sqlmock.Sqlmock, int, int)
	}{
		{
			name:          "List with offset 0",
			limit:         3,
			offset:        0,
			expectedCount: 3,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows([]string{"mac_address", "device_name", "ip_address", "location_description", "created_at", "updated_at"}).
					AddRow("00:11:22:33:44:55", "Device 1", "192.168.1.100", "Location 1", time.Now(), time.Now()).
					AddRow("00:11:22:33:44:56", "Device 2", "192.168.1.101", "Location 2", time.Now(), time.Now()).
					AddRow("00:11:22:33:44:57", "Device 3", "192.168.1.102", "Location 3", time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" ORDER BY created_at DESC LIMIT $1`)).
					WithArgs(limit).
					WillReturnRows(rows)
			},
		},
		{
			name:          "List with offset > 0",
			limit:         2,
			offset:        1,
			expectedCount: 2,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows([]string{"mac_address", "device_name", "ip_address", "location_description", "created_at", "updated_at"}).
					AddRow("00:11:22:33:44:56", "Device 2", "192.168.1.101", "Location 2", time.Now(), time.Now()).
					AddRow("00:11:22:33:44:57", "Device 3", "192.168.1.102", "Location 3", time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:          "Database error when listing devices",
			limit:         3,
			offset:        0,
			expectedCount: 0,
			expectError:   true,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "devices" ORDER BY created_at DESC LIMIT $1`)).
					WithArgs(limit).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:          "Empty limit",
			limit:         0,
			offset:        0,
			expectedCount: 0,
			expectError:   true,
			setupMock:     func(mock sqlmock.Sqlmock, limit, offset int) {},
		},
		{
			name:          "Negative offset",
			limit:         3,
			offset:        -1,
			expectedCount: 0,
			expectError:   true,
			setupMock:     func(mock sqlmock.Sqlmock, limit, offset int) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock, tt.limit, tt.offset)

			devices, err := repo.List(ctx, tt.limit, tt.offset)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, devices, tt.expectedCount)
			}
		})
	}
}

func TestDeviceRepository_Count(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)
	ctx := context.Background()

	// Mock count query returning 5
	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "devices"`)).
		WillReturnRows(rows)

	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestDeviceRepository_Exists(t *testing.T) {
	db, mock, err := setupTestDB()
	require.NoError(t, err)
	defer assert.NoError(t, mock.ExpectationsWereMet())

	logger, _ := zap.NewDevelopment()
	repo := NewPostgresDeviceRepository(db, logger)
	ctx := context.Background()

	tests := []struct {
		name        string
		macAddress  string
		expectError bool
		exists      bool
		setupMock   func(sqlmock.Sqlmock, string, bool)
	}{
		{
			name:        "Existing device",
			macAddress:  "00:11:22:33:44:55",
			expectError: false,
			exists:      true,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string, exists bool) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "devices" WHERE mac_address = $1`)).
					WithArgs(macAddress).
					WillReturnRows(rows)
			},
		},
		{
			name:        "Non-existing device",
			macAddress:  "00:11:22:33:44:99",
			expectError: false,
			exists:      false,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string, exists bool) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "devices" WHERE mac_address = $1`)).
					WithArgs(macAddress).
					WillReturnRows(rows)
			},
		},
		{
			name:        "Empty MAC address",
			macAddress:  "",
			expectError: true,
			exists:      false,
			setupMock:   func(mock sqlmock.Sqlmock, macAddress string, exists bool) {},
		},
		{
			name:        "Database error when checking device existence",
			macAddress:  "00:11:22:33:44:55",
			expectError: true,
			exists:      false,
			setupMock: func(mock sqlmock.Sqlmock, macAddress string, exists bool) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "devices" WHERE mac_address = $1`)).
					WithArgs(macAddress).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock, tt.macAddress, tt.exists)

			exists, err := repo.Exists(ctx, tt.macAddress)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
			}
		})
	}
}
