package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	ports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/mappers"
	pkglogger "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
	"go.uber.org/zap"
)

type sensorTemperatureHumidityRepository struct {
	db      *database.GormPostgresDB
	mapper  *mappers.SensorTemperatureHumidityMapper
	coreLog pkglogger.CoreLogger
}

// NewSensorTemperatureHumidityRepository creates a new GORM-based PostgreSQL sensor temperature humidity repository
func NewSensorTemperatureHumidityRepository(db *database.GormPostgresDB, loggerFactory pkglogger.LoggerFactory) ports.SensorTemperatureHumidityRepository {
	return &sensorTemperatureHumidityRepository{
		db:      db,
		mapper:  mappers.NewSensorTemperatureHumidityMapper(),
		coreLog: loggerFactory.Core(),
	}
}

// Create persists a new sensor temperature humidity reading to the database using GORM
func (r *sensorTemperatureHumidityRepository) Create(ctx context.Context, sensorData *entities.SensorTemperatureHumidity) error {
	if sensorData == nil {
		return fmt.Errorf("sensor data cannot be nil")
	}

	// Validate and normalize the domain entity before mapping
	sensorData.Normalize()
	if err := sensorData.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert domain entity to GORM model
	model := r.mapper.ToModel(sensorData)

	// Use GORM's Create method which will trigger BeforeCreate hooks
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Create(model)
	duration := time.Since(start)

	if result.Error != nil {
		r.coreLog.Error("sensor_temperature_humidity_not_created", zap.String("operation", "create"), zap.String("table", "sensor_temperature_humidities"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return fmt.Errorf("failed to create sensor temperature humidity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.coreLog.Error("sensor_temperature_humidity_not_created", zap.String("operation", "create"), zap.String("table", "sensor_temperature_humidities"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrSensorTemperatureHumidityNotFound))
		return domainerrors.ErrSensorTemperatureHumidityNotCreated
	}

	r.coreLog.Info("sensor_temperature_humidity_created_successfully", zap.String("mac_address", sensorData.MacAddress()), zap.String("component", "sensor_temperature_humidity_repository"))
	return nil
}
