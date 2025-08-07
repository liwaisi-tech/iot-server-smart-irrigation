package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks/stubs"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewDeviceRepository(t *testing.T) {
	gormMockDB, sqkmockDB := stubs.GetTestDB(t)
	assert.NotNil(t, gormMockDB)
	assert.NotNil(t, sqkmockDB)
	postgresDB, err := database.NewGormPostgresDBWithoutConfig(gormMockDB)
	assert.NoError(t, err)
	assert.NotNil(t, postgresDB)

	deviceRepository := NewDeviceRepository(postgresDB)
	assert.NotNil(t, deviceRepository)
}

func TestSave(t *testing.T) {
	gormMockDB, sqkmockDB := stubs.GetTestDB(t)
	assert.NotNil(t, gormMockDB)
	assert.NotNil(t, sqkmockDB)
	postgresDB, err := database.NewGormPostgresDBWithoutConfig(gormMockDB)
	assert.NoError(t, err)
	assert.NotNil(t, postgresDB)

	deviceRepository := NewDeviceRepository(postgresDB)
	assert.NotNil(t, deviceRepository)

	deviceEntity, err := entities.NewDevice("AA:BB:CC:DD:EE:FF", "test_device", "127.0.0.1", "In the very test code")
	assert.NoError(t, err)
	assert.NotNil(t, deviceEntity)

	t.Run("should return error due to device is nil", func(t *testing.T) {
		err := deviceRepository.Save(context.Background(), nil)

		assert.Error(t, err)
		assert.Equal(t, "device cannot be nil", err.Error())
	})

	t.Run("should return error due to device is invalid", func(t *testing.T) {
		device := &entities.Device{
			MACAddress: "invalid_mac_address",
		}
		err := deviceRepository.Save(context.Background(), device)

		assert.Error(t, err)
		assert.Equal(t, "validation failed: invalid mac address format: INVALID_MAC_ADDRESS (expected format: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX)", err.Error())
	})

	t.Run("should fail due to database raise error when inserting", func(t *testing.T) {
		sqkmockDB.ExpectQuery(`INSERT INTO "devices"`).WillReturnError(errors.New("insert failed"))

		err := deviceRepository.Save(context.Background(), deviceEntity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save device: insert failed")
	})

	t.Run("should fails due to the device is already exists", func(t *testing.T) {
		sqkmockDB.ExpectQuery(`INSERT INTO "devices"`).WillReturnError(gorm.ErrDuplicatedKey)

		err := deviceRepository.Save(context.Background(), deviceEntity)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrDeviceAlreadyExists)
	})

	t.Run("should success due to the device is saved successfully", func(t *testing.T) {
		sqkmockDB.ExpectQuery(
			`INSERT INTO "devices" \("mac_address","device_name","ip_address","location_description","status","deleted_at","registered_at","last_seen","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "registered_at","last_seen","created_at","updated_at"`).
			WillReturnRows(sqlmock.NewRows([]string{"registered_at", "last_seen", "created_at", "updated_at"}).
				AddRow(time.Now(), time.Now(), time.Now(), time.Now()))

		err := deviceRepository.Save(context.Background(), deviceEntity)
		assert.NoError(t, err)
	})

}
