package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/events"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/mappers"
	devicehealth "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_health"
)

// DeviceHealthHandler handles device health check NATS messages
type DeviceHealthHandler struct {
	useCase devicehealth.DeviceHealthUseCase
	mapper  *mappers.DeviceDetectedEventMapper
}

// NewDeviceHealthHandler creates a new device health handler
func NewDeviceHealthHandler(useCase devicehealth.DeviceHealthUseCase) *DeviceHealthHandler {
	return &DeviceHealthHandler{
		useCase: useCase,
		mapper:  mappers.NewDeviceDetectedEventMapper(),
	}
}

// HandleMessage processes raw NATS messages and converts them to domain logic
// This follows the same pattern as the existing MQTT handler
func (h *DeviceHealthHandler) HandleMessage(ctx context.Context, subject string, payload []byte) error {
	switch subject {
	case events.DeviceDetectedSubject:
		return h.processDeviceDetectedEvent(ctx, payload)
	default:
		log.Printf("Unknown NATS subject: %s", subject)
		return fmt.Errorf("unknown NATS subject: %s", subject)
	}
}

// processDeviceDetectedEvent processes device detected events
func (h *DeviceHealthHandler) processDeviceDetectedEvent(ctx context.Context, payload []byte) error {
	log.Printf("Processing device detected event, payload size: %d bytes", len(payload))

	// Parse JSON payload into domain event
	event, err := h.mapper.ToDomainEventFromBytes(payload)
	if err != nil {
		log.Printf("Failed to parse device detected event: %v", err)
		return fmt.Errorf("failed to parse device detected event: %w", err)
	}

	// Validate event type
	if event.EventType != events.DeviceDetectedEventType {
		log.Printf("Invalid event type for device detected event: %s, expected: %s",
			event.EventType, events.DeviceDetectedEventType)
		return fmt.Errorf("invalid event type: %s", event.EventType)
	}

	log.Printf("Received device detected event for MAC: %s, IP: %s, Event ID: %s",
		event.MACAddress, event.IPAddress, event.EventID)

	// Process the event using the health check use case
	return h.useCase.ProcessDeviceDetectedEvent(ctx, event)
}
