package devicehealth

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// HealthCheckConfig holds configuration for the health check use case
type HealthCheckConfig struct {
	CooldownPeriod      time.Duration
	DeduplicationWindow time.Duration
	CleanupInterval     time.Duration
	MaxConcurrent       int
}

// DefaultHealthCheckConfig returns default configuration
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		CooldownPeriod:      2 * time.Minute,
		DeduplicationWindow: 5 * time.Minute,
		CleanupInterval:     10 * time.Minute,
		MaxConcurrent:       10,
	}
}

// DeviceHealthUseCase defines the interface for device health checking operations
type DeviceHealthUseCase interface {
	// ProcessDeviceDetectedEvent processes a device detected event and performs health check
	ProcessDeviceDetectedEvent(ctx context.Context, event *entities.DeviceDetectedEvent) error

	// StartCleanup starts periodic cleanup of internal state
	StartCleanup(ctx context.Context)

	// StopCleanup stops the cleanup process
	StopCleanup()
}

// useCaseImpl implements the DeviceHealthUseCase interface
type useCaseImpl struct {
	deviceRepo    ports.DeviceRepository
	healthChecker ports.DeviceHealthChecker
	config        *HealthCheckConfig
	logger        *slog.Logger
	semaphore     chan struct{} // For limiting concurrent health checks
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
	cleanupOnce   sync.Once
}

// NewDeviceHealthUseCase creates a new device health use case
func NewDeviceHealthUseCase(
	deviceRepo ports.DeviceRepository,
	healthChecker ports.DeviceHealthChecker,
	config *HealthCheckConfig,
	logger *slog.Logger,
) DeviceHealthUseCase {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &useCaseImpl{
		deviceRepo:    deviceRepo,
		healthChecker: healthChecker,
		config:        config,
		logger:        logger,
		semaphore:     make(chan struct{}, config.MaxConcurrent),
		cleanupDone:   make(chan struct{}),
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

	uc.logger.Info("Processing device detected event",
		"mac_address", event.MACAddress,
		"ip_address", event.IPAddress,
		"event_id", event.EventID)

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
		uc.logger.Warn("Context cancelled before acquiring semaphore",
			"mac_address", event.MACAddress)
		return
	}

	uc.logger.Info("Starting health check",
		"mac_address", event.MACAddress,
		"ip_address", event.IPAddress)

	// Perform the health check
	result, err := uc.healthChecker.CheckHealth(ctx, event.IPAddress)
	if err != nil {
		uc.logger.Error("Health check failed",
			"mac_address", event.MACAddress,
			"ip_address", event.IPAddress,
			"error", err)
		// Continue to update device status even if health check failed
	}

	// Update device status based on health check result
	if err := uc.updateDeviceStatus(ctx, event.MACAddress, result); err != nil {
		uc.logger.Error("Failed to update device status",
			"mac_address", event.MACAddress,
			"error", err)
	}
}

// updateDeviceStatus updates the device status based on health check results
func (uc *useCaseImpl) updateDeviceStatus(ctx context.Context, macAddress string, result *ports.HealthCheckResult) error {
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
	if result != nil && result.Success {
		newStatus = "online"
		uc.logger.Info("Device health check succeeded",
			"mac_address", macAddress,
			"ip_address", result.IPAddress,
			"status_code", result.StatusCode,
			"attempts", result.Attempts,
			"duration", result.Duration,
			"response_body", result.ResponseBody)

		// Print the /whoami response to console as required
		fmt.Printf("Device %s (/whoami response): %s\n", macAddress, result.ResponseBody)
	} else {
		newStatus = "offline"
		errorMsg := "unknown error"
		if result != nil && result.Error != "" {
			errorMsg = result.Error
		}
		uc.logger.Warn("Device health check failed",
			"mac_address", macAddress,
			"error", errorMsg,
			"attempts", func() int {
				if result != nil {
					return result.Attempts
				}
				return 0
			}())
	}

	// Update device status
	if err := device.UpdateStatus(newStatus); err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	// Save updated device to repository
	if err := uc.deviceRepo.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to save device status update: %w", err)
	}

	uc.logger.Info("Device status updated successfully",
		"mac_address", macAddress,
		"new_status", newStatus)

	return nil
}

// StartCleanup starts periodic cleanup of internal state
func (uc *useCaseImpl) StartCleanup(ctx context.Context) {
	uc.cleanupOnce.Do(func() {
		uc.cleanupTicker = time.NewTicker(uc.config.CleanupInterval)

		go func() {
			defer uc.cleanupTicker.Stop()

			for {
				select {
				case <-uc.cleanupTicker.C:
					uc.performCleanup()
				case <-uc.cleanupDone:
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		uc.logger.Info("Health check cleanup started",
			"cleanup_interval", uc.config.CleanupInterval)
	})
}

// StopCleanup stops the cleanup process
func (uc *useCaseImpl) StopCleanup() {
	close(uc.cleanupDone)
	if uc.cleanupTicker != nil {
		uc.cleanupTicker.Stop()
	}
	uc.logger.Info("Health check cleanup stopped")
}

// performCleanup cleans up internal state to prevent memory leaks
func (uc *useCaseImpl) performCleanup() {
	uc.logger.Debug("Performing health check cleanup")

	uc.logger.Debug("Health check cleanup completed")
}
