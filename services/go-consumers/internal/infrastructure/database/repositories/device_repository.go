package repositories

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/domain/entities"
	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/domain/repositories"
	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/infrastructure/database/models"
)

// deviceRepository implements the DeviceRepository interface using GORM
type deviceRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDeviceRepository creates a new GORM-based device repository
func NewPostgresDeviceRepository(db *gorm.DB, logger *zap.Logger) repositories.DeviceRepository {
	return &deviceRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new device in the repository
func (r *deviceRepository) Create(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	if device.MacAddress == "" {
		return fmt.Errorf("device MAC address cannot be empty")
	}

	// Convert domain entity to database model
	deviceModel := r.entityToModel(device)

	// Create device in database
	if err := r.db.WithContext(ctx).Create(deviceModel).Error; err != nil {
		r.logger.Error("failed to create device",
			zap.String("mac_address", device.MacAddress),
			zap.Error(err))

		// Handle duplicate key error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("device with MAC address %s already exists", device.MacAddress)
		}

		return fmt.Errorf("failed to create device: %w", err)
	}

	r.logger.Info("device created successfully",
		zap.String("mac_address", device.MacAddress),
		zap.String("device_name", device.DeviceName))

	return nil
}

// GetByMacAddress retrieves a device by its MAC address (primary key)
func (r *deviceRepository) GetByMacAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("MAC address cannot be empty")
	}

	var deviceModel models.Device

	if err := r.db.WithContext(ctx).Where("mac_address = ?", macAddress).First(&deviceModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("device with MAC address %s not found", macAddress)
		}

		r.logger.Error("failed to get device by MAC address",
			zap.String("mac_address", macAddress),
			zap.Error(err))

		return nil, fmt.Errorf("failed to get device by MAC address: %w", err)
	}

	// Convert database model to domain entity
	device := r.modelToEntity(&deviceModel)

	r.logger.Debug("device retrieved successfully",
		zap.String("mac_address", macAddress))

	return device, nil
}

// Update updates an existing device
func (r *deviceRepository) Update(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	if device.MacAddress == "" {
		return fmt.Errorf("device MAC address cannot be empty")
	}

	// Convert domain entity to database model
	deviceModel := r.entityToModel(device)

	// Update device in database
	result := r.db.WithContext(ctx).Model(&models.Device{}).
		Where("mac_address = ?", device.MacAddress).
		Updates(deviceModel)

	if result.Error != nil {
		r.logger.Error("failed to update device",
			zap.String("mac_address", device.MacAddress),
			zap.Error(result.Error))

		return fmt.Errorf("failed to update device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("device with MAC address %s not found", device.MacAddress)
	}

	r.logger.Info("device updated successfully",
		zap.String("mac_address", device.MacAddress),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// Delete removes a device from the repository by MAC address
func (r *deviceRepository) Delete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("MAC address cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&models.Device{}, "mac_address = ?", macAddress)

	if result.Error != nil {
		r.logger.Error("failed to delete device",
			zap.String("mac_address", macAddress),
			zap.Error(result.Error))

		return fmt.Errorf("failed to delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("device with MAC address %s not found", macAddress)
	}

	r.logger.Info("device deleted successfully",
		zap.String("mac_address", macAddress),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// List retrieves devices with pagination support (limit and offset)
func (r *deviceRepository) List(ctx context.Context, limit, offset int) ([]*entities.Device, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}

	var deviceModels []models.Device

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&deviceModels).Error; err != nil {

		r.logger.Error("failed to list devices",
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Error(err))

		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Convert database models to domain entities
	devices := make([]*entities.Device, len(deviceModels))
	for i, model := range deviceModels {
		devices[i] = r.modelToEntity(&model)
	}

	r.logger.Debug("devices listed successfully",
		zap.Int("count", len(devices)),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	return devices, nil
}

// Count returns the total number of devices
func (r *deviceRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&models.Device{}).Count(&count).Error; err != nil {
		r.logger.Error("failed to count devices", zap.Error(err))
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}

	r.logger.Debug("devices counted successfully", zap.Int64("count", count))

	return count, nil
}

// Exists checks if a device with the given MAC address exists
func (r *deviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	if macAddress == "" {
		return false, fmt.Errorf("MAC address cannot be empty")
	}

	var count int64

	if err := r.db.WithContext(ctx).Model(&models.Device{}).
		Where("mac_address = ?", macAddress).
		Count(&count).Error; err != nil {

		r.logger.Error("failed to check device existence",
			zap.String("mac_address", macAddress),
			zap.Error(err))

		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	exists := count > 0

	r.logger.Debug("device existence checked",
		zap.String("mac_address", macAddress),
		zap.Bool("exists", exists))

	return exists, nil
}

// entityToModel converts a domain entity to a database model
func (r *deviceRepository) entityToModel(entity *entities.Device) *models.Device {
	if entity == nil {
		return nil
	}

	return &models.Device{
		MacAddress:          entity.MacAddress,
		DeviceName:          entity.DeviceName,
		IPAddress:           entity.IPAddress,
		LocationDescription: entity.LocationDescription,
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
}

// modelToEntity converts a database model to a domain entity
func (r *deviceRepository) modelToEntity(model *models.Device) *entities.Device {
	if model == nil {
		return nil
	}

	return &entities.Device{
		MacAddress:          model.MacAddress,
		DeviceName:          model.DeviceName,
		IPAddress:           model.IPAddress,
		LocationDescription: model.LocationDescription,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
