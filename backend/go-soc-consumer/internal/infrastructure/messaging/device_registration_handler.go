package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
)

// DeviceRegistrationHandler handles device registration MQTT messages
type DeviceRegistrationHandler struct {
	useCase deviceregistration.DeviceRegistrationUseCase
}

// NewDeviceRegistrationHandler creates a new device registration handler
func NewDeviceRegistrationHandler(useCase deviceregistration.DeviceRegistrationUseCase) *DeviceRegistrationHandler {
	return &DeviceRegistrationHandler{
		useCase: useCase,
	}
}

// HandleMessage processes raw MQTT messages and converts them to domain logic
func (h *DeviceRegistrationHandler) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case "/liwaisi/iot/smart-irrigation/device/registration":
		return h.processDeviceRegistration(ctx, payload)
	default:
		log.Printf("Unknown topic: %s", topic)
		return fmt.Errorf("unknown topic: %s", topic)
	}
}

// processDeviceRegistration processes device registration messages
func (h *DeviceRegistrationHandler) processDeviceRegistration(ctx context.Context, payload []byte) error {
	// Parse JSON payload
	var msgData dtos.DeviceRegistrationMessage

	if err := json.Unmarshal(payload, &msgData); err != nil {
		return fmt.Errorf("failed to unmarshal device registration message: %w", err)
	}

	// Validate event type
	if msgData.EventType != "register" {
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
		return fmt.Errorf("failed to create device registration message: %w", err)
	}

	// Process the message using the use case
	return h.useCase.RegisterDevice(ctx, deviceRegMsg)
}
