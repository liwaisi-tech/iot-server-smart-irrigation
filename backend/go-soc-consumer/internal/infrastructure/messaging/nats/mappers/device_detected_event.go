package mappers

import (
	"encoding/json"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/dtos"
)

type DeviceDetectedEventMapper struct {
}

func NewDeviceDetectedEventMapper() *DeviceDetectedEventMapper {
	return &DeviceDetectedEventMapper{}
}

func (m *DeviceDetectedEventMapper) ToDomainEventFromDTO(dto *dtos.DeviceDetectedEvent) *entities.DeviceDetectedEvent {
	if dto == nil {
		return nil
	}
	return &entities.DeviceDetectedEvent{
		MACAddress: dto.MACAddress,
		IPAddress:  dto.IPAddress,
		DetectedAt: dto.DetectedAt,
		EventID:    dto.EventID,
		EventType:  dto.EventType,
	}
}

func (m *DeviceDetectedEventMapper) ToDomainEventFromBytes(payload []byte) (*entities.DeviceDetectedEvent, error) {
	var dto dtos.DeviceDetectedEvent
	if err := json.Unmarshal(payload, &dto); err != nil {
		return nil, err
	}
	return m.ToDomainEventFromDTO(&dto), nil
}

func (m *DeviceDetectedEventMapper) ToDTOFromDomainEvent(event *entities.DeviceDetectedEvent) *dtos.DeviceDetectedEvent {
	if event == nil {
		return nil
	}
	return &dtos.DeviceDetectedEvent{
		MACAddress: event.MACAddress,
		IPAddress:  event.IPAddress,
		DetectedAt: event.DetectedAt,
		EventID:    event.EventID,
		EventType:  event.EventType,
	}
}
