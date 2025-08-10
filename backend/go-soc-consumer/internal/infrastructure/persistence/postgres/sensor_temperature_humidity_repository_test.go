package postgres

import (
	"context"
	"errors"
	"time"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks/stubs"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// setupSensorTestRepository initializes a test repository with a mock database
func setupSensorTestRepository(t *testing.T) (*sensorTemperatureHumidityRepository, sqlmock.Sqlmock) {
	gormMockDB, sqkmockDB := stubs.GetTestDB(t)
	assert.NotNil(t, gormMockDB)
	assert.NotNil(t, sqkmockDB)

	// Create test logger factory
	testLoggerFactory := createSensorTestLoggerFactory(t)

	postgresDB, err := database.NewGormPostgresDBWithoutConfig(gormMockDB, testLoggerFactory.Infrastructure())
	assert.NoError(t, err)
	assert.NotNil(t, postgresDB)

	repo := NewSensorTemperatureHumidityRepository(postgresDB, testLoggerFactory).(*sensorTemperatureHumidityRepository)
	assert.NotNil(t, repo)

	return repo, sqkmockDB
}

// createSensorTestLoggerFactory creates a test logger factory for use in tests
func createSensorTestLoggerFactory(t *testing.T) logger.LoggerFactory {
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	assert.NoError(t, err)
	assert.NotNil(t, loggerFactory)
	return loggerFactory
}

// createTestSensorData creates a valid SensorTemperatureHumidity instance for testing
func createTestSensorData() *entities.SensorTemperatureHumidity {
	sensor, _ := entities.NewSensorTemperatureHumidity(
		"00:11:22:33:44:55", // macAddress
		25.5,                // temperature
		60.0,                // humidity
	)
	return sensor
}

func TestNewSensorTemperatureHumidityRepository(t *testing.T) {
	gormDB, _ := stubs.GetTestDB(t)
	lf := createSensorTestLoggerFactory(t)
	postgresDB, err := database.NewGormPostgresDBWithoutConfig(gormDB, lf.Infrastructure())
	assert.NoError(t, err)
	repo := NewSensorTemperatureHumidityRepository(postgresDB, lf)
	assert.NotNil(t, repo)
}

func TestSensorTemperatureHumidityRepository_Create_NilSensorData(t *testing.T) {
	repo, _ := setupSensorTestRepository(t)

	err := repo.Create(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor data cannot be nil")
}

func TestSensorTemperatureHumidityRepository_Create_ValidationError(t *testing.T) {
	repo, _ := setupSensorTestRepository(t)

	// Create an invalid entity directly (zero value has empty mac and zeroed fields)
	sensor := &entities.SensorTemperatureHumidity{}

	err := repo.Create(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestSensorTemperatureHumidityRepository_Create_DatabaseError(t *testing.T) {
	repo, mock := setupSensorTestRepository(t)

	// Create valid sensor data
	sensor := createTestSensorData()

	// Expect an insert error (GORM uses a single INSERT ... RETURNING without explicit Begin here)
	mock.ExpectQuery(`INSERT INTO "sensor_temperature_humidity"`).
		WillReturnError(errors.New("insert failed"))

	err := repo.Create(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create sensor temperature humidity: insert failed")

	// Verify that all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSensorTemperatureHumidityRepository_Create_Success(t *testing.T) {
	repo, mock := setupSensorTestRepository(t)

	// Create valid sensor data
	sensor := createTestSensorData()

	// Expect the exact INSERT shape and RETURNING created_at, updated_at
	mock.ExpectQuery(
		`INSERT INTO "sensor_temperature_humidity" \("mac_address","temperature_celsius","humidity_percent","deleted_at","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "created_at","updated_at"`,
	).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(time.Now(), time.Now()))

	err := repo.Create(context.Background(), sensor)

	assert.NoError(t, err)

	// Verify that all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSensorTemperatureHumidityRepository_Create_ZeroRowsAffected(t *testing.T) {
	repo, mock := setupSensorTestRepository(t)

	// Create valid sensor data
	sensor := createTestSensorData()

	// Expect INSERT that returns no rows (RowsAffected = 0)
	mock.ExpectQuery(
		`INSERT INTO "sensor_temperature_humidity" \("mac_address","temperature_celsius","humidity_percent","deleted_at","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "created_at","updated_at"`,
	).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}))

	err := repo.Create(context.Background(), sensor)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrSensorTemperatureHumidityNotCreated)

	// Verify that all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
