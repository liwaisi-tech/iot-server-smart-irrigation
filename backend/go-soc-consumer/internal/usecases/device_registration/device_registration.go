package deviceregistration

import (
	"context"
	"fmt"
	"log"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// DeviceRegistrationUseCase defines the interface for device registration use case
type DeviceRegistrationUseCase interface {
	RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error
}

// UseCase handles device registration business logic
type useCaseImpl struct {
	deviceRepo     ports.DeviceRepository
	eventPublisher ports.EventPublisher
}

// NewDeviceRegistrationUseCase creates a new device registration use case
func NewDeviceRegistrationUseCase(deviceRepo ports.DeviceRepository, eventPublisher ports.EventPublisher) *useCaseImpl {
	return &useCaseImpl{
		deviceRepo:     deviceRepo,
		eventPublisher: eventPublisher,
	}
}

// RegisterDevice processes a device registration message
func (uc *useCaseImpl) RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	log.Printf("Processing device registration for MAC: %s, Name: %s",
		message.MACAddress, message.DeviceName)

	// Check if device already exists
	existingDevice, err := uc.deviceRepo.FindByMACAddress(ctx, message.MACAddress)
	if err == nil && existingDevice != nil {
		// Device exists, update it
		log.Printf("Device already exists, updating: %s", message.MACAddress)
		return uc.updateExistingDevice(ctx, existingDevice, message)
	}

	// Device doesn't exist, create new one
	log.Printf("Creating new device: %s", message.MACAddress)
	return uc.createNewDevice(ctx, message)
}

// createNewDevice creates a new device from registration message
func (uc *useCaseImpl) createNewDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	// Convert message to device entity
	device, err := message.ToDevice()
	if err != nil {
		return fmt.Errorf("failed to convert message to device: %w", err)
	}

	// Save device to repository
	if err := uc.deviceRepo.Save(ctx, device); err != nil {
		return fmt.Errorf("failed to save new device: %w", err)
	}

	log.Printf("Successfully registered new device: %s (%s)", device.DeviceName, device.MACAddress)

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

	// Validate updated device
	if err := existingDevice.Validate(); err != nil {
		return fmt.Errorf("updated device validation failed: %w", err)
	}

	// Update existing device
	if err := uc.deviceRepo.Update(ctx, existingDevice); err != nil {
		return fmt.Errorf("failed to update existing device: %w", err)
	}

	log.Printf("Successfully updated existing device: %s (%s)", existingDevice.GetDeviceName(), existingDevice.GetID())

	// Publish device detected event AFTER successful database operation
	uc.publishDeviceDetectedEvent(ctx, existingDevice.GetID(), existingDevice.GetIPAddress())

	return nil
}

// publishDeviceDetectedEvent publishes a device detected event
// This method logs errors but does not return them to avoid breaking the registration flow
func (uc *useCaseImpl) publishDeviceDetectedEvent(ctx context.Context, macAddress, ipAddress string) {
	// Skip if no event publisher is configured
	if uc.eventPublisher == nil {
		log.Printf("No event publisher configured, skipping event for device: %s", macAddress)
		return
	}

	// Check if publisher is connected
	if !uc.eventPublisher.IsConnected() {
		log.Printf("Event publisher not connected, skipping event for device: %s", macAddress)
		return
	}

	// Create device detected event
	event, err := entities.NewDeviceDetectedEvent(macAddress, ipAddress)
	if err != nil {
		log.Printf("Failed to create device detected event for %s: %v", macAddress, err)
		return
	}

	// Publish event (fire-and-forget with logging)
	subject := event.GetSubject()
	if err := uc.eventPublisher.Publish(ctx, subject, event); err != nil {
		log.Printf("Failed to publish device detected event for %s to subject %s: %v",
			macAddress, subject, err)
		return
	}

	log.Printf("Successfully published device detected event for device: %s (Event ID: %s)",
		macAddress, event.EventID)
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
