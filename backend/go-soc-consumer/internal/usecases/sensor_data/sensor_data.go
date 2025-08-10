package sensordata

import (
	"context"
	"fmt"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	ports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
	"go.uber.org/zap"
)

// SensorDataUseCase defines the interface for sensor data operations
type SensorDataUseCase interface {
	StoreSensorData(ctx context.Context, data *entities.SensorTemperatureHumidity) error
}

// sensorDataUseCase is the implementation of SensorDataUseCase
type sensorDataUseCase struct {
	coreLogger logger.CoreLogger
	repo       ports.SensorTemperatureHumidityRepository
}

// NewSensorDataUseCase creates a new sensor data use case
func NewSensorDataUseCase(loggerFactory logger.LoggerFactory, repo ports.SensorTemperatureHumidityRepository) SensorDataUseCase {
	return &sensorDataUseCase{
		coreLogger: loggerFactory.Core(),
		repo:       repo,
	}
}

// StoreSensorData stores the sensor data using the repository
func (uc *sensorDataUseCase) StoreSensorData(ctx context.Context, data *entities.SensorTemperatureHumidity) error {
	uc.coreLogger.Info("storing_sensor_data", zap.String("mac_address", data.MacAddress()), zap.String("component", "sensor_data_use_case"))

	if err := uc.repo.Create(ctx, data); err != nil {
		uc.coreLogger.Error("failed_to_store_sensor_data", zap.Error(err), zap.String("component", "sensor_data_use_case"))
		return fmt.Errorf("failed to store sensor data: %w", err)
	}

	uc.coreLogger.Info("sensor_data_stored_successfully", zap.String("mac_address", data.MacAddress()), zap.String("component", "sensor_data_use_case"))
	return nil
}
