package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks"
)

func TestNewDeviceRegistrationHandler(t *testing.T) {
	// Create a real use case with a mock repository for testing
	mockRepo := mocks.NewMockDeviceRepository(t)
	realUseCase := deviceregistration.NewUseCase(mockRepo)
	handler := NewDeviceRegistrationHandler(realUseCase)

	assert.NotNil(t, handler, "NewDeviceRegistrationHandler() returned nil")
	assert.NotNil(t, handler.useCase, "NewDeviceRegistrationHandler() did not set useCase")
}

func TestDeviceRegistrationHandler_HandleMessage_ValidTopic(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	validPayload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	payload, err := json.Marshal(validPayload)
	require.NoError(t, err, "Failed to marshal test payload")

	mockUseCase.EXPECT().RegisterDevice(mock.Anything, mock.MatchedBy(func(msg *entities.DeviceRegistrationMessage) bool {
		return msg.MACAddress == "AA:BB:CC:DD:EE:FF" && msg.DeviceName == "Test Device"
	})).Return(nil).Once()

	ctx := context.Background()
	err = handler.HandleMessage(ctx, "/liwaisi/iot/smart-irrigation/device/registration", payload)

	assert.NoError(t, err, "HandleMessage() unexpected error")
}

func TestDeviceRegistrationHandler_HandleMessage_UnknownTopic(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	validPayload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	payload, err := json.Marshal(validPayload)
	require.NoError(t, err, "Failed to marshal test payload")

	ctx := context.Background()
	err = handler.HandleMessage(ctx, "/unknown/topic", payload)

	require.Error(t, err, "HandleMessage() expected error for unknown topic but got none")

	expectedError := "unknown topic: /unknown/topic"
	assert.Equal(t, expectedError, err.Error(), "HandleMessage() error message mismatch")
}

func TestDeviceRegistrationHandler_processDeviceRegistration_ValidPayload(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "valid registration payload",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "valid registration with lowercase MAC",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "aa:bb:cc:dd:ee:ff",
				"device_name":          "Test Device 2",
				"ip_address":           "192.168.1.101",
				"location_description": "Test Location 2",
			},
		},
		{
			name: "valid registration with dash-separated MAC",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA-BB-CC-DD-EE-FF",
				"device_name":          "Test Device 3",
				"ip_address":           "192.168.1.102",
				"location_description": "Test Location 3",
			},
		},
		{
			name: "valid registration with IPv6",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "IPv6 Device",
				"ip_address":           "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				"location_description": "IPv6 Location",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock use case for testing
			mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
			handler := NewDeviceRegistrationHandler(mockUseCase)

			expectedMAC := tt.payload["mac_address"].(string)
			if expectedMAC == "aa:bb:cc:dd:ee:ff" {
				expectedMAC = "AA:BB:CC:DD:EE:FF" // Should be normalized to uppercase
			} else if expectedMAC == "AA-BB-CC-DD-EE-FF" {
				expectedMAC = "AA-BB-CC-DD-EE-FF" // Dash format is preserved
			}

			mockUseCase.EXPECT().RegisterDevice(mock.Anything, mock.MatchedBy(func(msg *entities.DeviceRegistrationMessage) bool {
				return msg.MACAddress == expectedMAC &&
					msg.DeviceName == tt.payload["device_name"].(string) &&
					msg.IPAddress == tt.payload["ip_address"].(string) &&
					msg.LocationDescription == tt.payload["location_description"].(string)
			})).Return(nil).Once()

			payload, err := json.Marshal(tt.payload)
			require.NoError(t, err, "Failed to marshal test payload")

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payload)

			assert.NoError(t, err, "processDeviceRegistration() unexpected error")
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_MalformedJSON(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	malformedPayloads := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "invalid JSON syntax",
			payload: []byte(`{"event_type": "register", "mac_address": "AA:BB:CC:DD:EE:FF"`),
		},
		{
			name:    "empty payload",
			payload: []byte(""),
		},
		{
			name:    "null payload",
			payload: []byte("null"),
		},
		{
			name:    "non-JSON text",
			payload: []byte("this is not json"),
		},
	}

	for _, tt := range malformedPayloads {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := handler.processDeviceRegistration(ctx, tt.payload)

			assert.Error(t, err, "processDeviceRegistration() expected error for malformed JSON but got none")
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_InvalidEventType(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	invalidEventTypes := []struct {
		name      string
		eventType string
	}{
		{"empty event type", ""},
		{"unregister event", "unregister"},
		{"update event", "update"},
		{"delete event", "delete"},
		{"uppercase register", "REGISTER"},
		{"mixed case register", "Register"},
	}

	for _, tt := range invalidEventTypes {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"event_type":           tt.eventType,
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			}

			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err, "Failed to marshal test payload")

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payloadBytes)

			require.Error(t, err, "processDeviceRegistration() expected error for invalid event type but got none")

			expectedError := "invalid event type for device registration: " + tt.eventType
			assert.Equal(t, expectedError, err.Error(), "processDeviceRegistration() error message mismatch")
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_InvalidDeviceData(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	invalidPayloads := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "missing mac_address",
			payload: map[string]interface{}{
				"event_type":           "register",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "empty mac_address",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "invalid mac_address format",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "INVALID-MAC",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "missing device_name",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "empty device_name",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			},
		},
		{
			name: "missing ip_address",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"location_description": "Test Location",
			},
		},
		{
			name: "invalid ip_address",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"ip_address":           "invalid-ip",
				"location_description": "Test Location",
			},
		},
		{
			name: "missing location_description",
			payload: map[string]interface{}{
				"event_type":  "register",
				"mac_address": "AA:BB:CC:DD:EE:FF",
				"device_name": "Test Device",
				"ip_address":  "192.168.1.100",
			},
		},
		{
			name: "empty location_description",
			payload: map[string]interface{}{
				"event_type":           "register",
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "",
			},
		},
	}

	for _, tt := range invalidPayloads {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err, "Failed to marshal test payload")

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payloadBytes)

			assert.Error(t, err, "processDeviceRegistration() expected error for invalid device data but got none")
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_UseCaseError(t *testing.T) {
	// Create a mock use case that returns an error
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	mockUseCase.EXPECT().RegisterDevice(mock.Anything, mock.Anything).Return(errors.New("use case processing failed")).Once()

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err, "Failed to marshal test payload")

	ctx := context.Background()
	err = handler.processDeviceRegistration(ctx, payloadBytes)

	require.Error(t, err, "processDeviceRegistration() expected error from use case but got none")
	assert.Equal(t, "use case processing failed", err.Error(), "processDeviceRegistration() error message mismatch")
}

func TestDeviceRegistrationHandler_Integration(t *testing.T) {
	// This test verifies the full integration from HandleMessage to processDeviceRegistration
	mockUseCase := mocks.NewMockDeviceRegistrationUseCase(t)
	handler := NewDeviceRegistrationHandler(mockUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Integration Test Device",
		"ip_address":           "192.168.1.200",
		"location_description": "Integration Test Location",
	}

	mockUseCase.EXPECT().RegisterDevice(mock.Anything, mock.MatchedBy(func(msg *entities.DeviceRegistrationMessage) bool {
		return msg.MACAddress == "AA:BB:CC:DD:EE:FF" &&
			msg.DeviceName == "Integration Test Device" &&
			msg.IPAddress == "192.168.1.200" &&
			msg.LocationDescription == "Integration Test Location"
	})).Return(nil).Once()

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err, "Failed to marshal payload")

	err = handler.HandleMessage(context.Background(), "/liwaisi/iot/smart-irrigation/device/registration", payloadBytes)
	require.NoError(t, err, "HandleMessage() returned error")
}

func TestDeviceRegistrationHandler_RealUseCaseIntegration(t *testing.T) {
	// This test uses a real use case with mock repository to test full integration
	mockRepo := mocks.NewMockDeviceRepository(t)
	realUseCase := deviceregistration.NewUseCase(mockRepo)
	handler := NewDeviceRegistrationHandler(realUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Real Integration Device",
		"ip_address":           "192.168.1.250",
		"location_description": "Real Integration Location",
	}

	// Setup mock expectations
	mockRepo.EXPECT().FindByMACAddress(mock.Anything, "AA:BB:CC:DD:EE:FF").Return(nil, nil).Once()
	mockRepo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(device *entities.Device) bool {
		return device.MACAddress == "AA:BB:CC:DD:EE:FF" &&
			device.DeviceName == "Real Integration Device" &&
			device.IPAddress == "192.168.1.250" &&
			device.Status == "registered"
	})).Return(nil).Once()

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err, "Failed to marshal payload")

	err = handler.HandleMessage(context.Background(), "/liwaisi/iot/smart-irrigation/device/registration", payloadBytes)
	require.NoError(t, err, "HandleMessage() returned error")
}
