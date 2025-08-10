package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	domainerrors "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	ports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/mappers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/models"
	pkglogger "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// DeviceRepository implements the DeviceRepository interface using GORM PostgreSQL
type deviceRepository struct {
	db     *database.GormPostgresDB
	mapper *mappers.DeviceMapper
	logger pkglogger.CoreLogger
}

// NewDeviceRepository creates a new GORM-based PostgreSQL device repository
func NewDeviceRepository(db *database.GormPostgresDB, loggerFactory pkglogger.LoggerFactory) ports.DeviceRepository {
	return &deviceRepository{
		db:     db,
		mapper: mappers.NewDeviceMapper(),
		logger: loggerFactory.Core(),
	}
}

// Create persists a new device to the database using GORM
func (r *deviceRepository) Create(ctx context.Context, device *entities.Device) error {
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
			r.logger.Info("device_creation_failed", zap.String("operation", "create"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrDeviceAlreadyExists))
			return domainerrors.ErrDeviceAlreadyExists
		}
		r.logger.Info("device_creation_failed", zap.String("operation", "create"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return fmt.Errorf("failed to create device: %w", result.Error)
	}

	r.logger.Info("device_created_successfully", zap.String("mac_address", device.GetID()), zap.String("device_name", device.GetDeviceName()), zap.String("component", "device_repository"))
	return nil
}

// Update updates an existing device in the database using GORM
func (r *deviceRepository) Update(ctx context.Context, device *entities.Device) error {
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
		r.logger.Info("device_update_failed", zap.String("operation", "update"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return fmt.Errorf("failed to update device: %w", result.Error)
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		r.logger.Info("device_update_failed", zap.String("operation", "update"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrDeviceNotFound))
		return domainerrors.ErrDeviceNotFound
	}

	r.logger.Info("device_updated_successfully", zap.String("mac_address", device.GetID()), zap.String("device_name", device.GetDeviceName()), zap.String("component", "device_repository"))
	return nil
}

// FindByMACAddress retrieves a device by its MAC address using GORM
func (r *deviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address cannot be empty")
	}

	start := time.Now()
	var model models.DeviceModel
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).First(&model)
	duration := time.Since(start)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Info("device_not_found", zap.String("operation", "find_by_mac"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrDeviceNotFound))
			return nil, domainerrors.ErrDeviceNotFound
		}
		r.logger.Info("device_not_found", zap.String("operation", "find_by_mac"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return nil, fmt.Errorf("failed to find device by MAC address: %w", result.Error)
	}

	r.logger.Info("device_found_successfully", zap.String("mac_address", macAddress), zap.String("component", "device_repository"))
	// Convert GORM model to domain entity
	device := r.mapper.FromModel(&model)
	return device, nil
}

// Exists checks if a device with the given MAC address exists using GORM
func (r *deviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	if macAddress == "" {
		return false, fmt.Errorf("mac address cannot be empty")
	}

	start := time.Now()
	var count int64
	result := r.db.GetDB().WithContext(ctx).Model(&models.DeviceModel{}).
		Where("mac_address = ?", macAddress).Count(&count)
	duration := time.Since(start)

	if result.Error != nil {
		r.logger.Info("device_not_found", zap.String("operation", "exists"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return false, fmt.Errorf("failed to check device existence: %w", result.Error)
	}

	r.logger.Info("device_found_successfully", zap.String("mac_address", macAddress), zap.String("component", "device_repository"))
	return count > 0, nil
}

// List retrieves all devices with optional pagination using GORM
func (r *deviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
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
		r.logger.Info("device_not_found", zap.String("operation", "list"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return nil, fmt.Errorf("failed to list devices: %w", result.Error)
	}

	r.logger.Info("devices_listed_successfully", zap.Int("count", len(models)),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("component", "device_repository"),
	)

	// Convert GORM models to domain entities
	devices := r.mapper.FromModelSlice(models)
	return devices, nil
}

// Delete removes a device by MAC address using GORM soft delete
func (r *deviceRepository) Delete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	// GORM will perform soft delete by setting deleted_at timestamp
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	duration := time.Since(start)

	if result.Error != nil {
		r.logger.Info("device_not_found", zap.String("operation", "delete"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return fmt.Errorf("failed to delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Info("device_not_found", zap.String("operation", "delete"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrDeviceNotFound))
		return domainerrors.ErrDeviceNotFound
	}

	r.logger.Info("device_deleted_successfully", zap.String("mac_address", macAddress), zap.String("deletion_type", "soft"), zap.String("component", "device_repository"))
	return nil
}

// HardDelete permanently removes a device by MAC address (bypasses soft delete)
func (r *deviceRepository) HardDelete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	// Use Unscoped() to perform hard delete
	start := time.Now()
	result := r.db.GetDB().WithContext(ctx).Unscoped().Where("mac_address = ?", macAddress).Delete(&models.DeviceModel{})
	duration := time.Since(start)

	if result.Error != nil {
		r.logger.Info("device_not_found", zap.String("operation", "hard_delete"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(result.Error))
		return fmt.Errorf("failed to hard delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Info("device_not_found", zap.String("operation", "hard_delete"), zap.String("table", "devices"), zap.Duration("duration", duration), zap.Int64("records_affected", 0), zap.Error(domainerrors.ErrDeviceNotFound))
		return domainerrors.ErrDeviceNotFound
	}

	r.logger.Info("device_hard_deleted_successfully", zap.String("mac_address", macAddress), zap.String("deletion_type", "hard"), zap.String("component", "device_repository"))
	return nil
}
