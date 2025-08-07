package mappers

import (
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/models"
)

// DeviceMapper provides mapping functions between domain entities and GORM models
type DeviceMapper struct{}

// NewDeviceMapper creates a new device mapper
func NewDeviceMapper() *DeviceMapper {
	return &DeviceMapper{}
}

// ToModel converts a domain entity to a GORM model
func (m *DeviceMapper) ToModel(device *entities.Device) *models.DeviceModel {
	if device == nil {
		return nil
	}

	now := time.Now()
	return &models.DeviceModel{
		MACAddress:          device.MACAddress,
		DeviceName:          device.DeviceName,
		IPAddress:           device.IPAddress,
		LocationDescription: device.LocationDescription,
		RegisteredAt:        device.RegisteredAt,
		LastSeen:            device.LastSeen,
		Status:              device.Status,
		CreatedAt:           now, // Will be overridden by GORM if already set
		UpdatedAt:           now, // Will be overridden by GORM if already set
	}
}

// FromModel converts a GORM model to a domain entity
func (m *DeviceMapper) FromModel(model *models.DeviceModel) *entities.Device {
	if model == nil {
		return nil
	}

	// Create the device directly since we can't use NewDevice (it validates and normalizes)
	device := &entities.Device{}
	device.MACAddress = model.MACAddress
	device.DeviceName = model.DeviceName
	device.IPAddress = model.IPAddress
	device.LocationDescription = model.LocationDescription
	device.RegisteredAt = model.RegisteredAt
	device.LastSeen = model.LastSeen
	device.Status = model.Status

	return device
}

// ToModelSlice converts a slice of domain entities to GORM models
func (m *DeviceMapper) ToModelSlice(devices []*entities.Device) []*models.DeviceModel {
	if devices == nil {
		return nil
	}

	models := make([]*models.DeviceModel, len(devices))
	for i, device := range devices {
		models[i] = m.ToModel(device)
	}
	return models
}

// FromModelSlice converts a slice of GORM models to domain entities
func (m *DeviceMapper) FromModelSlice(models []*models.DeviceModel) []*entities.Device {
	if models == nil {
		return nil
	}

	entities := make([]*entities.Device, len(models))
	for i, model := range models {
		entities[i] = m.FromModel(model)
	}
	return entities
}

// UpdateModelFromEntity updates an existing GORM model with data from a domain entity
// This is useful for update operations where you want to preserve certain fields
func (m *DeviceMapper) UpdateModelFromEntity(model *models.DeviceModel, device *entities.Device) {
	if model == nil || device == nil {
		return
	}

	model.MACAddress = device.MACAddress
	model.DeviceName = device.DeviceName
	model.IPAddress = device.IPAddress
	model.LocationDescription = device.LocationDescription
	model.RegisteredAt = device.RegisteredAt
	model.LastSeen = device.LastSeen
	model.Status = device.Status
	// Note: CreatedAt, UpdatedAt, DeletedAt are managed by GORM
}

// ToModelForUpdate converts a domain entity to a GORM model specifically for update operations
// This preserves the original creation timestamp and other audit fields
func (m *DeviceMapper) ToModelForUpdate(device *entities.Device, originalModel *models.DeviceModel) *models.DeviceModel {
	if device == nil {
		return nil
	}

	model := m.ToModel(device)
	if originalModel != nil {
		// Preserve audit fields from the original model
		model.CreatedAt = originalModel.CreatedAt
		model.UpdatedAt = time.Now() // This will be updated by GORM anyway
		model.DeletedAt = originalModel.DeletedAt
	}
	
	return model
}