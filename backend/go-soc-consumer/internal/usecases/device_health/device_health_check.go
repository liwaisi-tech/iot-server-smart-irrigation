package devicehealth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// HealthCheckConfig holds configuration for the health check use case
type HealthCheckConfig struct {
	MaxConcurrent int
}

// DefaultHealthCheckConfig returns default configuration
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		MaxConcurrent: 10,
	}
}

// DeviceHealthUseCase defines the interface for device health checking operations
type DeviceHealthUseCase interface {
	// ProcessDeviceDetectedEvent processes a device detected event and performs health check
	ProcessDeviceDetectedEvent(ctx context.Context, event *entities.DeviceDetectedEvent) error
}

// useCaseImpl implements the DeviceHealthUseCase interface
type useCaseImpl struct {
	deviceRepo    ports.DeviceRepository
	healthChecker ports.DeviceHealthChecker
	config        *HealthCheckConfig
	loggerFactory logger.LoggerFactory
	semaphore     chan struct{} // For limiting concurrent health checks
}

// NewDeviceHealthUseCase creates a new device health use case
func NewDeviceHealthUseCase(
	deviceRepo ports.DeviceRepository,
	healthChecker ports.DeviceHealthChecker,
	config *HealthCheckConfig,
	loggerFactory logger.LoggerFactory,
) DeviceHealthUseCase {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	if loggerFactory == nil {
		defaultLoggerFactory, err := logger.NewDefault()
		if err != nil {
			// Fallback to a basic logger if default creation fails
			panic(fmt.Sprintf("failed to create default logger factory: %v", err))
		}
		loggerFactory = defaultLoggerFactory
	}

	return &useCaseImpl{
		deviceRepo:    deviceRepo,
		healthChecker: healthChecker,
		config:        config,
		loggerFactory: loggerFactory,
		semaphore:     make(chan struct{}, config.MaxConcurrent),
	}
}

// ProcessDeviceDetectedEvent processes a device detected event
func (uc *useCaseImpl) ProcessDeviceDetectedEvent(ctx context.Context, event *entities.DeviceDetectedEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	uc.loggerFactory.Core().Info("device_detected_event_processing_started",
		zap.String("mac_address", event.MACAddress),
		zap.String("ip_address", event.IPAddress),
		zap.String("event_id", event.EventID),
		zap.String("component", "device_health_usecase"),
	)

	// Perform health check in a goroutine to avoid blocking
	go uc.performHealthCheck(context.Background(), event)

	return nil
}

// performHealthCheck performs the actual health check with concurrency control
func (uc *useCaseImpl) performHealthCheck(ctx context.Context, event *entities.DeviceDetectedEvent) {
	// Acquire semaphore for concurrency control
	select {
	case uc.semaphore <- struct{}{}:
		defer func() { <-uc.semaphore }()
	case <-ctx.Done():
		uc.loggerFactory.Core().Warn("health_check_cancelled_before_semaphore",
			zap.String("mac_address", event.MACAddress),
			zap.Error(ctx.Err()),
			zap.String("component", "device_health_usecase"),
		)
		return
	}

	uc.loggerFactory.Core().Debug("health_check_starting",
		zap.String("mac_address", event.MACAddress),
		zap.String("ip_address", event.IPAddress),
		zap.String("component", "device_health_usecase"),
	)

	// Perform the health check
	start := time.Now()
	isAlive, err := uc.healthChecker.CheckHealth(ctx, event.IPAddress)
	healthCheckDuration := time.Since(start)

	if err != nil {
		uc.loggerFactory.Device().LogDeviceHealthCheck(event.MACAddress, event.IPAddress, false, healthCheckDuration, err)
		uc.loggerFactory.Core().Error("health_check_error",
			zap.Error(err),
			zap.String("mac_address", event.MACAddress),
			zap.String("ip_address", event.IPAddress),
			zap.Duration("duration", healthCheckDuration),
			zap.String("component", "device_health_usecase"),
		)
		// Continue to update device status even if health check failed
	} else {
		uc.loggerFactory.Device().LogDeviceHealthCheck(event.MACAddress, event.IPAddress, isAlive, healthCheckDuration, nil)
	}

	// Update device status based on health check result
	if err := uc.updateDeviceStatus(ctx, event.MACAddress, isAlive); err != nil {
		uc.loggerFactory.Core().Error("device_status_update_failed",
			zap.Error(err),
			zap.String("mac_address", event.MACAddress),
			zap.String("component", "device_health_usecase"),
		)
	}
}

// updateDeviceStatus updates the device status based on health check results
func (uc *useCaseImpl) updateDeviceStatus(ctx context.Context, macAddress string, isAlive bool) error {
	// Retrieve the device from repository
	device, err := uc.deviceRepo.FindByMACAddress(ctx, macAddress)
	if err != nil {
		return fmt.Errorf("failed to find device %s: %w", macAddress, err)
	}

	if device == nil {
		return fmt.Errorf("device not found: %s", macAddress)
	}

	// Determine new status based on health check result
	var newStatus string
	if isAlive {
		newStatus = "online"
		uc.loggerFactory.Core().Info("device_health_check_succeeded",
			zap.String("mac_address", macAddress),
			zap.String("ip_address", device.GetIPAddress()),
			zap.String("component", "device_health_usecase"),
		)
	} else {
		newStatus = "offline"
		errorMsg := "unknown error"
		attempts := 0
		uc.loggerFactory.Core().Warn("device_health_check_failed",
			zap.String("mac_address", macAddress),
			zap.String("error", errorMsg),
			zap.Int("attempts", attempts),
			zap.String("component", "device_health_usecase"),
		)
	}

	// Update device status
	if err := device.UpdateStatus(newStatus); err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	// Save updated device to repository
	if err := uc.deviceRepo.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to save device status update: %w", err)
	}

	uc.loggerFactory.Core().Info("device_status_updated_successfully",
		zap.String("mac_address", macAddress),
		zap.String("new_status", newStatus),
		zap.String("component", "device_health_usecase"),
	)

	return nil
}
