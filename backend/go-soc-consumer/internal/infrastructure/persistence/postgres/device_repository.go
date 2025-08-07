package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/mappers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/models"
)

// DeviceRepository implements the DeviceRepository interface using GORM PostgreSQL
type DeviceRepository struct {
	db     *database.GormPostgresDB
	mapper *mappers.DeviceMapper
}

// NewDeviceRepository creates a new GORM-based PostgreSQL device repository
func NewDeviceRepository(db *database.GormPostgresDB) ports.DeviceRepository {
	return &DeviceRepository{
		db:     db,
		mapper: mappers.NewDeviceMapper(),
	}
}

// Save persists a new device to the database using GORM
func (r *DeviceRepository) Save(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	// Validate and normalize the domain entity before mapping
	device.Normalize()
	if err := device.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert domain entity to GORM model
	model := r.mapper.ToModel(device)

	// Use GORM's Create method which will trigger BeforeCreate hooks
	result := r.db.GetDB().WithContext(ctx).Create(model)
	if result.Error != nil {
		// Handle GORM-specific errors
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domainerrors.ErrDeviceAlreadyExists
		}
		return fmt.Errorf("failed to save device: %w", result.Error)
	}

	return nil
}

// Update updates an existing device in the database using GORM
func (r *DeviceRepository) Update(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	// Validate and normalize the domain entity before mapping
	device.Normalize()
	if err := device.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert domain entity to GORM model
	model := r.mapper.ToModel(device)

	// Use GORM's Save method which will trigger BeforeUpdate hooks
	// Save will update all fields, including zero values
	result := r.db.GetDB().WithContext(ctx).Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update device: %w", result.Error)
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		return domainerrors.ErrDeviceNotFound
	}

	return nil
}

// FindByMACAddress retrieves a device by its MAC address using GORM
func (r *DeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address cannot be empty")
	}

	var model models.DeviceModel
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to find device by MAC address: %w", result.Error)
	}

	// Convert GORM model to domain entity
	device := r.mapper.FromModel(&model)
	return device, nil
}

// Exists checks if a device with the given MAC address exists using GORM
func (r *DeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	if macAddress == "" {
		return false, fmt.Errorf("mac address cannot be empty")
	}

	var count int64
	result := r.db.GetDB().WithContext(ctx).Model(&models.DeviceModel{}).
		Where("mac_address = ?", macAddress).Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check device existence: %w", result.Error)
	}

	return count > 0, nil
}

// List retrieves all devices with optional pagination using GORM
func (r *DeviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative")
	}

	var models []*models.DeviceModel
	query := r.db.GetDB().WithContext(ctx).Order("registered_at DESC")

	// Apply pagination if specified
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list devices: %w", result.Error)
	}

	// Convert GORM models to domain entities
	devices := r.mapper.FromModelSlice(models)
	return devices, nil
}

// Delete removes a device by MAC address using GORM soft delete
func (r *DeviceRepository) Delete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	// GORM will perform soft delete by setting deleted_at timestamp
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrDeviceNotFound
	}

	return nil
}

// HardDelete permanently removes a device by MAC address (bypasses soft delete)
func (r *DeviceRepository) HardDelete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	// Use Unscoped() to perform hard delete
	result := r.db.GetDB().WithContext(ctx).Unscoped().Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domainerrors.ErrDeviceNotFound
	}

	return nil
}
