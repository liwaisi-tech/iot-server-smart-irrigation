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
type UseCase struct {
	deviceRepo ports.DeviceRepository
}

// NewUseCase creates a new device registration use case
func NewUseCase(deviceRepo ports.DeviceRepository) *UseCase {
	return &UseCase{
		deviceRepo: deviceRepo,
	}
}

// RegisterDevice processes a device registration message
func (uc *UseCase) RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
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
func (uc *UseCase) createNewDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
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
	return nil
}

// updateExistingDevice updates an existing device with new information
func (uc *UseCase) updateExistingDevice(ctx context.Context, existingDevice *entities.Device, message *entities.DeviceRegistrationMessage) error {
	// Update device information
	existingDevice.DeviceName = message.DeviceName
	existingDevice.IPAddress = message.IPAddress
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

	log.Printf("Successfully updated existing device: %s (%s)", existingDevice.DeviceName, existingDevice.MACAddress)
	return nil
}

// MessageHandler implements the ports.MessageHandler interface
type MessageHandler struct {
	useCase *UseCase
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(useCase *UseCase) *MessageHandler {
	return &MessageHandler{
		useCase: useCase,
	}
}

// HandleDeviceRegistration processes device registration messages
func (h *MessageHandler) HandleDeviceRegistration(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	return h.useCase.RegisterDevice(ctx, message)
}
