package deviceregistration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// createTestLoggerFactory creates a test logger factory for use in tests
func createTestLoggerFactory(t *testing.T) logger.LoggerFactory {
	loggerFactory, err := logger.NewDevelopment()
	assert.NoError(t, err)
	assert.NotNil(t, loggerFactory)
	return loggerFactory
}

func TestNewUseCase(t *testing.T) {
	mockRepo := mocks.NewMockDeviceRepository(t)

	useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))

	assert.NotNil(t, useCase)
	// Note: Cannot directly access private fields in the updated implementation
}

func TestUseCase_RegisterDevice_NewDevice(t *testing.T) {
	tests := []struct {
		name    string
		message *entities.DeviceRegistrationMessage
		setup   func(*mocks.MockDeviceRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful new device registration",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device not found (new device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(nil, errors.New("device not found")).
					Once()

				// Save new device successfully
				mockRepo.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "save fails for new device",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device not found (new device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(nil, errors.New("device not found")).
					Once()

				// Save fails
				mockRepo.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
			errMsg:  "failed to save new device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockDeviceRepository(t)
			tt.setup(mockRepo)

			useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))
			err := useCase.RegisterDevice(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_RegisterDevice_ExistingDevice(t *testing.T) {
	tests := []struct {
		name           string
		message        *entities.DeviceRegistrationMessage
		existingDevice *entities.Device
		setup          func(*mocks.MockDeviceRepository)
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful existing device update",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Updated Device",
				IPAddress:           "192.168.1.101",
				LocationDescription: "Garden Zone 2",
				ReceivedAt:          time.Now(),
			},
			existingDevice: &entities.Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Old Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				RegisteredAt:        time.Now().Add(-24 * time.Hour),
				LastSeen:            time.Now().Add(-1 * time.Hour),
				Status:              "offline",
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device found (existing device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(&entities.Device{
						MACAddress:          "AA:BB:CC:DD:EE:FF",
						DeviceName:          "Old Device",
						IPAddress:           "192.168.1.100",
						LocationDescription: "Garden Zone 1",
						RegisteredAt:        time.Now().Add(-24 * time.Hour),
						LastSeen:            time.Now().Add(-1 * time.Hour),
						Status:              "offline",
					}, nil).
					Once()

				// Update device successfully
				mockRepo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(device *entities.Device) bool {
						return device.MACAddress == "AA:BB:CC:DD:EE:FF" &&
							device.DeviceName == "Updated Device" &&
							device.IPAddress == "192.168.1.101" &&
							device.LocationDescription == "Garden Zone 2" &&
							device.Status == "online"
					})).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "update fails for existing device",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Updated Device",
				IPAddress:           "192.168.1.101",
				LocationDescription: "Garden Zone 2",
				ReceivedAt:          time.Now(),
			},
			existingDevice: &entities.Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Old Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				RegisteredAt:        time.Now().Add(-24 * time.Hour),
				LastSeen:            time.Now().Add(-1 * time.Hour),
				Status:              "offline",
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device found (existing device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(&entities.Device{
						MACAddress:          "AA:BB:CC:DD:EE:FF",
						DeviceName:          "Old Device",
						IPAddress:           "192.168.1.100",
						LocationDescription: "Garden Zone 1",
						RegisteredAt:        time.Now().Add(-24 * time.Hour),
						LastSeen:            time.Now().Add(-1 * time.Hour),
						Status:              "offline",
					}, nil).
					Once()

				// Update fails
				mockRepo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
			errMsg:  "failed to update existing device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockDeviceRepository(t)
			tt.setup(mockRepo)

			useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))
			err := useCase.RegisterDevice(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_createNewDevice(t *testing.T) {
	tests := []struct {
		name    string
		message *entities.DeviceRegistrationMessage
		setup   func(*mocks.MockDeviceRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful device creation",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				mockRepo.EXPECT().
					Save(mock.Anything, mock.MatchedBy(func(device *entities.Device) bool {
						return device.MACAddress == "AA:BB:CC:DD:EE:FF" &&
							device.DeviceName == "Test Device" &&
							device.IPAddress == "192.168.1.100" &&
							device.LocationDescription == "Garden Zone 1" &&
							device.Status == "registered"
					})).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "invalid message - ToDevice fails",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "", // Invalid MAC address
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// No expectations - ToDevice should fail before calling Save
			},
			wantErr: true,
			errMsg:  "failed to convert message to device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockDeviceRepository(t)
			tt.setup(mockRepo)

			useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))
			err := useCase.createNewDevice(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_updateExistingDevice(t *testing.T) {
	tests := []struct {
		name           string
		existingDevice *entities.Device
		message        *entities.DeviceRegistrationMessage
		setup          func(*mocks.MockDeviceRepository)
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful device update",
			existingDevice: &entities.Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Old Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				RegisteredAt:        time.Now().Add(-24 * time.Hour),
				LastSeen:            time.Now().Add(-1 * time.Hour),
				Status:              "offline",
			},
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Updated Device",
				IPAddress:           "192.168.1.101",
				LocationDescription: "Garden Zone 2",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				mockRepo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(device *entities.Device) bool {
						return device.MACAddress == "AA:BB:CC:DD:EE:FF" &&
							device.DeviceName == "Updated Device" &&
							device.IPAddress == "192.168.1.101" &&
							device.LocationDescription == "Garden Zone 2" &&
							device.Status == "online"
					})).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "update repository error",
			existingDevice: &entities.Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Old Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				RegisteredAt:        time.Now().Add(-24 * time.Hour),
				LastSeen:            time.Now().Add(-1 * time.Hour),
				Status:              "offline",
			},
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Updated Device",
				IPAddress:           "192.168.1.101",
				LocationDescription: "Garden Zone 2",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				mockRepo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
			errMsg:  "failed to update existing device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockDeviceRepository(t)
			tt.setup(mockRepo)

			useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))
			err := useCase.updateExistingDevice(context.Background(), tt.existingDevice, tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNewMessageHandler(t *testing.T) {
	mockRepo := mocks.NewMockDeviceRepository(t)
	useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))

	handler := NewMessageHandler(useCase)

	assert.NotNil(t, handler)
	assert.Equal(t, useCase, handler.useCase)
}

func TestMessageHandler_HandleDeviceRegistration(t *testing.T) {
	tests := []struct {
		name    string
		message *entities.DeviceRegistrationMessage
		setup   func(*mocks.MockDeviceRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful message handling",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device not found (new device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(nil, errors.New("device not found")).
					Once()

				// Save new device successfully
				mockRepo.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "message handling with error",
			message: &entities.DeviceRegistrationMessage{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.100",
				LocationDescription: "Garden Zone 1",
				ReceivedAt:          time.Now(),
			},
			setup: func(mockRepo *mocks.MockDeviceRepository) {
				// Device not found (new device)
				mockRepo.EXPECT().
					FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
					Return(nil, errors.New("device not found")).
					Once()

				// Save fails
				mockRepo.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
			errMsg:  "failed to save new device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockDeviceRepository(t)
			tt.setup(mockRepo)

			useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))
			handler := NewMessageHandler(useCase)

			err := handler.HandleDeviceRegistration(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test edge cases and error scenarios
func TestUseCase_RegisterDevice_EdgeCases(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		mockRepo := mocks.NewMockDeviceRepository(t)
		useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))

		// This should panic or be handled gracefully depending on implementation
		// Since the current implementation doesn't check for nil, this is more of a documentation test
		// We're intentionally not checking the error return value here to test the panic behavior
		//nolint:errcheck // This is an intentional test of panic behavior with nil message
		assert.Panics(t, func() {
			_ = useCase.RegisterDevice(context.Background(), nil) // This should panic
		})
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockRepo := mocks.NewMockDeviceRepository(t)

		// Setup mock to respect context cancellation
		mockRepo.EXPECT().
			FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").
			Return(nil, context.Canceled).
			Once()

		// The use case will still try to save since it treats any FindByMACAddress error as "not found"
		mockRepo.EXPECT().
			Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
			Return(context.Canceled).
			Once()

		useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(t))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		message := &entities.DeviceRegistrationMessage{
			MACAddress:          "AA:BB:CC:DD:EE:FF",
			DeviceName:          "Test Device",
			IPAddress:           "192.168.1.100",
			LocationDescription: "Garden Zone 1",
			ReceivedAt:          time.Now(),
		}

		err := useCase.RegisterDevice(ctx, message)
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkUseCase_RegisterDevice_NewDevice(b *testing.B) {
	mockRepo := mocks.NewMockDeviceRepository(&testing.T{})

	// Setup mock for all iterations
	mockRepo.EXPECT().
		FindByMACAddress(mock.Anything, mock.AnythingOfType("string")).
		Return(nil, errors.New("device not found")).
		Times(b.N)

	mockRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*entities.Device")).
		Return(nil).
		Times(b.N)

	useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(&testing.T{}))
	message := &entities.DeviceRegistrationMessage{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Test Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Garden Zone 1",
		ReceivedAt:          time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = useCase.RegisterDevice(context.Background(), message) // Ignore error in benchmark
	}
}

func BenchmarkUseCase_RegisterDevice_ExistingDevice(b *testing.B) {
	mockRepo := mocks.NewMockDeviceRepository(&testing.T{})

	existingDevice := &entities.Device{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Old Device",
		IPAddress:           "192.168.1.100",
		LocationDescription: "Garden Zone 1",
		RegisteredAt:        time.Now().Add(-24 * time.Hour),
		LastSeen:            time.Now().Add(-1 * time.Hour),
		Status:              "offline",
	}

	// Setup mock for all iterations
	mockRepo.EXPECT().
		FindByMACAddress(mock.Anything, mock.AnythingOfType("string")).
		Return(existingDevice, nil).
		Times(b.N)

	mockRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*entities.Device")).
		Return(nil).
		Times(b.N)

	useCase := NewDeviceRegistrationUseCase(mockRepo, nil, createTestLoggerFactory(&testing.T{}))
	message := &entities.DeviceRegistrationMessage{
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		DeviceName:          "Updated Device",
		IPAddress:           "192.168.1.101",
		LocationDescription: "Garden Zone 2",
		ReceivedAt:          time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = useCase.RegisterDevice(context.Background(), message) // Ignore error in benchmark
	}
}
