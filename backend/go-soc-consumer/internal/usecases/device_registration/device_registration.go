package deviceregistration

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	eventports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/events"
	repositoryports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// DeviceRegistrationUseCase defines the interface for device registration use case
type DeviceRegistrationUseCase interface {
	RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error
}

// UseCase handles device registration business logic
type useCaseImpl struct {
	deviceRepo     repositoryports.DeviceRepository
	eventPublisher eventports.EventPublisher
	loggerFactory  logger.LoggerFactory
}

// NewDeviceRegistrationUseCase creates a new device registration use case
func NewDeviceRegistrationUseCase(deviceRepo repositoryports.DeviceRepository, eventPublisher eventports.EventPublisher, loggerFactory logger.LoggerFactory) *useCaseImpl {
	return &useCaseImpl{
		deviceRepo:     deviceRepo,
		eventPublisher: eventPublisher,
		loggerFactory:  loggerFactory,
	}
}

// RegisterDevice processes a device registration message
func (uc *useCaseImpl) RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	start := time.Now()

	uc.loggerFactory.Core().Info("device_registration_started",
		zap.String("mac_address", message.MACAddress),
		zap.String("device_name", message.DeviceName),
		zap.String("ip_address", message.IPAddress),
		zap.String("location", message.LocationDescription),
		zap.String("component", "device_registration_usecase"),
	)

	// Check if device already exists
	existingDevice, err := uc.deviceRepo.FindByMACAddress(ctx, message.MACAddress)
	if err == nil && existingDevice != nil {
		// Device exists, update it
		uc.loggerFactory.Core().Debug("existing_device_found_for_update",
			zap.String("mac_address", message.MACAddress),
			zap.String("existing_name", existingDevice.GetDeviceName()),
			zap.String("new_name", message.DeviceName),
			zap.String("component", "device_registration_usecase"),
		)
		err := uc.updateExistingDevice(ctx, existingDevice, message)
		processingDuration := time.Since(start)

		if err != nil {
			uc.loggerFactory.Core().Error("device_update_failed",
				zap.Error(err),
				zap.String("mac_address", message.MACAddress),
				zap.Duration("processing_duration", processingDuration),
				zap.String("component", "device_registration_usecase"),
			)
		} else {
			uc.loggerFactory.Device().LogDeviceRegistration(message.MACAddress, message.DeviceName, message.IPAddress, message.LocationDescription, true)
		}
		return err
	}

	// Device doesn't exist, create new one
	uc.loggerFactory.Core().Debug("creating_new_device",
		zap.String("mac_address", message.MACAddress),
		zap.String("device_name", message.DeviceName),
		zap.String("component", "device_registration_usecase"),
	)
	err = uc.createNewDevice(ctx, message)
	processingDuration := time.Since(start)

	if err != nil {
		uc.loggerFactory.Core().Error("device_creation_failed",
			zap.Error(err),
			zap.String("mac_address", message.MACAddress),
			zap.Duration("processing_duration", processingDuration),
			zap.String("component", "device_registration_usecase"),
		)
	} else {
		uc.loggerFactory.Device().LogDeviceRegistration(message.MACAddress, message.DeviceName, message.IPAddress, message.LocationDescription, false)
	}
	return err
}

// createNewDevice creates a new device from registration message
func (uc *useCaseImpl) createNewDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	// Convert message to device entity
	device, err := message.ToDevice()
	if err != nil {
		return fmt.Errorf("failed to convert message to device: %w", err)
	}

	// Create device in repository
	if err := uc.deviceRepo.Create(ctx, device); err != nil {
		uc.loggerFactory.Core().Error("failed_to_create_new_device",
			zap.Error(err),
			zap.String("mac_address", device.GetID()),
			zap.String("device_name", device.GetDeviceName()),
			zap.String("component", "device_registration_usecase"),
		)
		return fmt.Errorf("failed to create new device: %w", err)
	}

	uc.loggerFactory.Core().Info("new_device_registered_successfully",
		zap.String("mac_address", device.GetID()),
		zap.String("device_name", device.GetDeviceName()),
		zap.String("ip_address", device.GetIPAddress()),
		zap.String("component", "device_registration_usecase"),
	)

	// Publish device detected event AFTER successful database operation
	// Event publishing failure should NOT fail the registration process
	uc.publishDeviceDetectedEvent(ctx, device.GetID(), device.GetIPAddress())

	return nil
}

// updateExistingDevice updates an existing device with new information
func (uc *useCaseImpl) updateExistingDevice(ctx context.Context, existingDevice *entities.Device, message *entities.DeviceRegistrationMessage) error {
	// Update device information
	existingDevice.SetDeviceName(message.DeviceName)
	existingDevice.SetIPAddress(message.IPAddress)
	existingDevice.LocationDescription = message.LocationDescription
	existingDevice.LastSeen = message.ReceivedAt

	// Update status to online when device registers again
	if err := existingDevice.UpdateStatus("online"); err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	// Validate updated device
	if err := existingDevice.Validate(); err != nil {
		return fmt.Errorf("updated device validation failed: %w", err)
	}

	// Update existing device
	if err := uc.deviceRepo.Update(ctx, existingDevice); err != nil {
		uc.loggerFactory.Core().Error("failed_to_update_existing_device",
			zap.Error(err),
			zap.String("mac_address", existingDevice.GetID()),
			zap.String("device_name", existingDevice.GetDeviceName()),
			zap.String("component", "device_registration_usecase"),
		)
		return fmt.Errorf("failed to update existing device: %w", err)
	}

	uc.loggerFactory.Core().Info("existing_device_updated_successfully",
		zap.String("mac_address", existingDevice.GetID()),
		zap.String("device_name", existingDevice.GetDeviceName()),
		zap.String("ip_address", existingDevice.GetIPAddress()),
		zap.String("component", "device_registration_usecase"),
	)

	// Publish device detected event AFTER successful database operation
	uc.publishDeviceDetectedEvent(ctx, existingDevice.GetID(), existingDevice.GetIPAddress())

	return nil
}

// publishDeviceDetectedEvent publishes a device detected event
// This method logs errors but does not return them to avoid breaking the registration flow
func (uc *useCaseImpl) publishDeviceDetectedEvent(ctx context.Context, macAddress, ipAddress string) {
	// Skip if no event publisher is configured
	if uc.eventPublisher == nil {
		uc.loggerFactory.Core().Warn("no_event_publisher_configured",
			zap.String("mac_address", macAddress),
			zap.String("component", "device_registration_usecase"),
		)
		return
	}

	// Check if publisher is connected
	if !uc.eventPublisher.IsConnected() {
		uc.loggerFactory.Core().Warn("event_publisher_not_connected",
			zap.String("mac_address", macAddress),
			zap.String("component", "device_registration_usecase"),
		)
		return
	}

	// Create device detected event
	event, err := entities.NewDeviceDetectedEvent(macAddress, ipAddress)
	if err != nil {
		uc.loggerFactory.Core().Error("failed_to_create_device_detected_event",
			zap.Error(err),
			zap.String("mac_address", macAddress),
			zap.String("ip_address", ipAddress),
			zap.String("component", "device_registration_usecase"),
		)
		return
	}

	// Publish event (fire-and-forget with logging)
	subject := event.GetSubject()
	if err := uc.eventPublisher.Publish(ctx, subject, event); err != nil {
		uc.loggerFactory.Messaging().LogEventPublishing("device_detected", subject, event.EventID, false, err)
		return
	}

	uc.loggerFactory.Messaging().LogEventPublishing("device_detected", subject, event.EventID, true, nil)
	uc.loggerFactory.Core().Debug("device_detected_event_published",
		zap.String("mac_address", macAddress),
		zap.String("event_id", event.EventID),
		zap.String("subject", subject),
		zap.String("component", "device_registration_usecase"),
	)
}

// MessageHandler implements the ports.MessageHandler interface
type MessageHandler struct {
	useCase *useCaseImpl
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(useCase *useCaseImpl) *MessageHandler {
	return &MessageHandler{
		useCase: useCase,
	}
}

// HandleDeviceRegistration processes device registration messages
func (h *MessageHandler) HandleDeviceRegistration(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	return h.useCase.RegisterDevice(ctx, message)
}
