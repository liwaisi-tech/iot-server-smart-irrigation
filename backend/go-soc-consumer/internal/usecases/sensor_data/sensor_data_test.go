package sensordata

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// createTestLoggerFactory creates a test logger factory for use in tests
func createTestLoggerFactory(t *testing.T) logger.LoggerFactory {
	loggerFactory, err := logger.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, loggerFactory)
	return loggerFactory
}
func TestSensorDataUseCase_StoreSensorData(t *testing.T) {
	mockRepo := mocks.NewMockSensorTemperatureHumidityRepository(t)
	loggerFactory := createTestLoggerFactory(t)
	useCase := NewSensorDataUseCase(loggerFactory, mockRepo)

	ctx := context.Background()
	sensorData, err := entities.NewSensorTemperatureHumidity("00:11:22:33:44:55", 25.5, 60.0)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Create", ctx, sensorData).Return(nil).Once()

		err := useCase.StoreSensorData(ctx, sensorData)

		assert.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("repo error")
		mockRepo.On("Create", ctx, sensorData).Return(expectedErr).Once()

		err := useCase.StoreSensorData(ctx, sensorData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store sensor data")
	})
}
