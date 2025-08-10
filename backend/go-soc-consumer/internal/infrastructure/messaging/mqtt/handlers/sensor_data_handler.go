package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// SensorDataHandler handles temperature and humidity sensor data MQTT messages
// This is a logging-only handler that processes and logs sensor data without persistence
type SensorDataHandler struct {
	sensorLogger logger.SensorLogger
	coreLogger   logger.CoreLogger
}

// NewSensorDataHandler creates a new sensor data handler with domain-specific logger
func NewSensorDataHandler(sensorLogger logger.SensorLogger, coreLogger logger.CoreLogger) *SensorDataHandler {
	return &SensorDataHandler{
		sensorLogger: sensorLogger,
		coreLogger:   coreLogger,
	}
}

// NewSensorDataHandlerFromFactory creates a sensor data handler using LoggerFactory
func NewSensorDataHandlerFromFactory(loggerFactory logger.LoggerFactory) *SensorDataHandler {
	return &SensorDataHandler{
		sensorLogger: loggerFactory.Sensor(),
		coreLogger:   loggerFactory.Core(),
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
		h.sensorLogger.LogSensorDataProcessingError("unknown", payload, err, "json_unmarshal")
		return fmt.Errorf("failed to unmarshal sensor data message: %w", err)
	}

	// Validate event type
	if msgData.EventType != "sensor_data" {
		err := fmt.Errorf("invalid event type for sensor data: %s", msgData.EventType)
		h.sensorLogger.LogSensorDataProcessingError(msgData.MacAddress, payload, err, "event_type_validation")
		return err
	}

	// Create domain entity with validation
	sensorData, err := entities.NewSensorTemperatureHumidity(
		msgData.MacAddress,
		msgData.Temperature,
		msgData.Humidity,
	)
	if err != nil {
		h.sensorLogger.LogSensorDataProcessingError(msgData.MacAddress, payload, err, "entity_creation")
		return fmt.Errorf("failed to create sensor data entity: %w", err)
	}

	// Log the sensor data with appropriate level based on readings
	h.sensorLogger.LogSensorData(
		sensorData.MacAddress(),
		sensorData.Temperature(),
		sensorData.Humidity(),
		sensorData.HasAbnormalReadings(),
	)

	return nil
}
