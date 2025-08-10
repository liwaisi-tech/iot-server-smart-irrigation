package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	sensordata "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/sensor_data"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// SensorDataHandler handles temperature and humidity sensor data MQTT messages
type SensorDataHandler struct {
	coreLogger logger.CoreLogger
	useCase    sensordata.SensorDataUseCase
}

// NewSensorDataHandler creates a sensor data handler using LoggerFactory
func NewSensorDataHandler(loggerFactory logger.LoggerFactory, useCase sensordata.SensorDataUseCase) *SensorDataHandler {
	return &SensorDataHandler{
		coreLogger: loggerFactory.Core(),
		useCase:    useCase,
	}
}

// HandleMessage processes raw MQTT messages and logs sensor data
func (h *SensorDataHandler) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity":
		return h.processSensorData(ctx, payload)
	default:
		h.coreLogger.Warn("unknown_sensor_topic",
			zap.String("topic", topic),
			zap.String("component", "sensor_data_handler"),
		)
		return fmt.Errorf("unknown sensor topic: %s", topic)
	}
}

// processSensorData processes temperature and humidity sensor messages
func (h *SensorDataHandler) processSensorData(ctx context.Context, payload []byte) error {
	// Parse JSON payload
	var msgData dtos.SensorDataMessage
	if err := json.Unmarshal(payload, &msgData); err != nil {
		h.coreLogger.Error("sensor_data_processing_error",
			zap.String("topic", "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"),
			zap.String("payload", string(payload)),
			zap.Error(err),
			zap.String("component", "sensor_data_handler"),
		)
		return fmt.Errorf("failed to unmarshal sensor data message: %w", err)
	}

	// Validate event type
	if msgData.EventType != "sensor_data" {
		err := fmt.Errorf("invalid event type for sensor data: %s", msgData.EventType)
		h.coreLogger.Error("sensor_data_processing_error",
			zap.String("topic", "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"),
			zap.String("payload", string(payload)),
			zap.Error(err),
			zap.String("component", "sensor_data_handler"),
		)
		return err
	}

	// Create domain entity with validation
	sensorData, err := entities.NewSensorTemperatureHumidity(
		msgData.MacAddress,
		msgData.Temperature,
		msgData.Humidity,
	)
	if err != nil {
		h.coreLogger.Error("sensor_data_processing_error",
			zap.String("topic", "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"),
			zap.String("payload", string(payload)),
			zap.Error(err),
			zap.String("component", "sensor_data_handler"),
		)
		return fmt.Errorf("failed to create sensor data entity: %w", err)
	}

	// Process the message using the use case
	if err := h.useCase.StoreSensorData(ctx, sensorData); err != nil {
		h.coreLogger.Error("failed_to_store_sensor_data",
			zap.String("topic", "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"),
			zap.String("payload", string(payload)),
			zap.Error(err),
			zap.String("component", "sensor_data_handler"),
		)
		return fmt.Errorf("failed to store sensor data: %w", err)
	}
	return nil
}
