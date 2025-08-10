package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	temphumidityrepo "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// SensorDataHandler handles temperature and humidity sensor data MQTT messages
// This is a logging-only handler that processes and logs sensor data without persistence
type SensorDataHandler struct {
	coreLogger       logger.CoreLogger
	tempHumidityRepo temphumidityrepo.SensorTemperatureHumidityRepository
}

// NewSensorDataHandler creates a sensor data handler using LoggerFactory
func NewSensorDataHandler(loggerFactory logger.LoggerFactory, tempHumidityRepo temphumidityrepo.SensorTemperatureHumidityRepository) *SensorDataHandler {
	return &SensorDataHandler{
		coreLogger:       loggerFactory.Core(),
		tempHumidityRepo: tempHumidityRepo,
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

	// Create a database record for the sensor data
	if err := h.tempHumidityRepo.Create(ctx, sensorData); err != nil {
		h.coreLogger.Error("sensor_data_processing_error",
			zap.String("topic", "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"),
			zap.String("payload", string(payload)),
			zap.Error(err),
			zap.String("component", "sensor_data_handler"),
		)
		return fmt.Errorf("failed to create sensor data record: %w", err)
	}
	return nil
}
