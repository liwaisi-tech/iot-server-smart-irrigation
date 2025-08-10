package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
	"go.uber.org/zap"
)

// DeviceRegistrationHandler handles device registration MQTT messages
type DeviceRegistrationHandler struct {
	coreLogger logger.CoreLogger
	useCase    deviceregistration.DeviceRegistrationUseCase
}

// NewDeviceRegistrationHandler creates a new device registration handler
func NewDeviceRegistrationHandler(loggerFactory logger.LoggerFactory, useCase deviceregistration.DeviceRegistrationUseCase) *DeviceRegistrationHandler {
	return &DeviceRegistrationHandler{
		coreLogger: loggerFactory.Core(),
		useCase:    useCase,
	}
}

// HandleMessage processes raw MQTT messages and converts them to domain logic
func (h *DeviceRegistrationHandler) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case "/liwaisi/iot/smart-irrigation/device/registration":
		return h.processDeviceRegistration(ctx, payload)
	default:
		h.coreLogger.Error("unknown_topic", zap.String("topic", topic), zap.String("component", "device_registration_handler"))
		return fmt.Errorf("unknown topic: %s", topic)
	}
}

// processDeviceRegistration processes device registration messages
func (h *DeviceRegistrationHandler) processDeviceRegistration(ctx context.Context, payload []byte) error {
	h.coreLogger.Info("device_registration_message_received", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"))
	// Parse JSON payload
	var msgData dtos.DeviceRegistrationMessage

	if err := json.Unmarshal(payload, &msgData); err != nil {
		h.coreLogger.Error("failed_to_unmarshal_device_registration_message", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"), zap.Error(err))
		return fmt.Errorf("failed to unmarshal device registration message: %w", err)
	}

	// Validate event type
	if msgData.EventType != "register" {
		h.coreLogger.Error("invalid_event_type_for_device_registration", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"), zap.String("event_type", msgData.EventType))
		return fmt.Errorf("invalid event type for device registration: %s", msgData.EventType)
	}

	// Create domain entity
	deviceRegMsg, err := entities.NewDeviceRegistrationMessage(
		msgData.MacAddress,
		msgData.DeviceName,
		msgData.IPAddress,
		msgData.LocationDescription,
	)
	if err != nil {
		h.coreLogger.Error("failed_to_create_device_registration_message", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"), zap.Error(err))
		return fmt.Errorf("failed to create device registration message: %w", err)
	}

	// Process the message using the use case
	if err := h.useCase.RegisterDevice(ctx, deviceRegMsg); err != nil {
		h.coreLogger.Error("failed_to_register_device", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"), zap.Error(err))
		return fmt.Errorf("failed to register device: %w", err)
	}
	h.coreLogger.Info("device_registered_successfully", zap.String("topic", "/liwaisi/iot/smart-irrigation/device/registration"), zap.String("component", "device_registration_handler"))
	return nil
}
