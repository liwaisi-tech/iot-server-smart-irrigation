package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/mappers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/models"
	pkglogger "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// DeviceRepository implements the DeviceRepository interface using GORM PostgreSQL
type DeviceRepository struct {
	db       *database.GormPostgresDB
	mapper   *mappers.DeviceMapper
	infraLog pkglogger.InfrastructureLogger
	coreLog  pkglogger.CoreLogger
}

// NewDeviceRepository creates a new GORM-based PostgreSQL device repository
func NewDeviceRepository(db *database.GormPostgresDB, loggerFactory pkglogger.LoggerFactory) ports.DeviceRepository {
	return &DeviceRepository{
		db:       db,
		mapper:   mappers.NewDeviceMapper(),
		infraLog: loggerFactory.Infrastructure(),
		coreLog:  loggerFactory.Core(),
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
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Create(model)
	duration := time.Since(start)
	
	if result.Error != nil {
		// Handle GORM-specific errors
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			r.infraLog.LogDatabaseOperation("create", "devices", duration, 0, domainerrors.ErrDeviceAlreadyExists)
			return domainerrors.ErrDeviceAlreadyExists
		}
		r.infraLog.LogDatabaseOperation("create", "devices", duration, 0, result.Error)
		return fmt.Errorf("failed to save device: %w", result.Error)
	}

	r.infraLog.LogDatabaseOperation("create", "devices", duration, result.RowsAffected, nil)
	r.coreLog.Debug("device_saved_successfully",
		zap.String("mac_address", device.GetID()),
		zap.String("device_name", device.GetDeviceName()),
		zap.String("component", "device_repository"),
	)
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
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Save(model)
	duration := time.Since(start)
	
	if result.Error != nil {
		r.infraLog.LogDatabaseOperation("update", "devices", duration, 0, result.Error)
		return fmt.Errorf("failed to update device: %w", result.Error)
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		r.infraLog.LogDatabaseOperation("update", "devices", duration, 0, domainerrors.ErrDeviceNotFound)
		return domainerrors.ErrDeviceNotFound
	}

	r.infraLog.LogDatabaseOperation("update", "devices", duration, result.RowsAffected, nil)
	r.coreLog.Debug("device_updated_successfully",
		zap.String("mac_address", device.GetID()),
		zap.String("device_name", device.GetDeviceName()),
		zap.String("component", "device_repository"),
	)
	return nil
}

// FindByMACAddress retrieves a device by its MAC address using GORM
func (r *DeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address cannot be empty")
	}

	start := time.Now()
	var model models.DeviceModel
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).First(&model)
	duration := time.Since(start)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.infraLog.LogDatabaseOperation("find_by_mac", "devices", duration, 0, domainerrors.ErrDeviceNotFound)
			return nil, domainerrors.ErrDeviceNotFound
		}
		r.infraLog.LogDatabaseOperation("find_by_mac", "devices", duration, 0, result.Error)
		return nil, fmt.Errorf("failed to find device by MAC address: %w", result.Error)
	}

	r.infraLog.LogDatabaseOperation("find_by_mac", "devices", duration, 1, nil)
	// Convert GORM model to domain entity
	device := r.mapper.FromModel(&model)
	return device, nil
}

// Exists checks if a device with the given MAC address exists using GORM
func (r *DeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	if macAddress == "" {
		return false, fmt.Errorf("mac address cannot be empty")
	}

	start := time.Now()
	var count int64
	result := r.db.GetDB().WithContext(ctx).Model(&models.DeviceModel{}).
		Where("mac_address = ?", macAddress).Count(&count)
	duration := time.Since(start)

	if result.Error != nil {
		r.infraLog.LogDatabaseOperation("exists", "devices", duration, 0, result.Error)
		return false, fmt.Errorf("failed to check device existence: %w", result.Error)
	}

	r.infraLog.LogDatabaseOperation("exists", "devices", duration, count, nil)
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

	start := time.Now()
	result := query.Find(&models)
	duration := time.Since(start)
	
	if result.Error != nil {
		r.infraLog.LogDatabaseOperation("list", "devices", duration, 0, result.Error)
		return nil, fmt.Errorf("failed to list devices: %w", result.Error)
	}

	r.infraLog.LogDatabaseOperation("list", "devices", duration, result.RowsAffected, nil)
	r.coreLog.Debug("devices_listed_successfully",
		zap.Int("count", len(models)),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("component", "device_repository"),
	)
	
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
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	duration := time.Since(start)
	
	if result.Error != nil {
		r.infraLog.LogDatabaseOperation("delete", "devices", duration, 0, result.Error)
		return fmt.Errorf("failed to delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.infraLog.LogDatabaseOperation("delete", "devices", duration, 0, domainerrors.ErrDeviceNotFound)
		return domainerrors.ErrDeviceNotFound
	}

	r.infraLog.LogDatabaseOperation("delete", "devices", duration, result.RowsAffected, nil)
	r.coreLog.Debug("device_deleted_successfully",
		zap.String("mac_address", macAddress),
		zap.String("deletion_type", "soft"),
		zap.String("component", "device_repository"),
	)
	return nil
}

// HardDelete permanently removes a device by MAC address (bypasses soft delete)
func (r *DeviceRepository) HardDelete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	// Use Unscoped() to perform hard delete
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Unscoped().Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	duration := time.Since(start)
	
	if result.Error != nil {
		r.infraLog.LogDatabaseOperation("hard_delete", "devices", duration, 0, result.Error)
		return fmt.Errorf("failed to hard delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.infraLog.LogDatabaseOperation("hard_delete", "devices", duration, 0, domainerrors.ErrDeviceNotFound)
		return domainerrors.ErrDeviceNotFound
	}

	r.infraLog.LogDatabaseOperation("hard_delete", "devices", duration, result.RowsAffected, nil)
	r.coreLog.Debug("device_hard_deleted_successfully",
		zap.String("mac_address", macAddress),
		zap.String("deletion_type", "hard"),
		zap.String("component", "device_repository"),
	)
	return nil
}
