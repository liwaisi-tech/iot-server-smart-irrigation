package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
)

// MockDeviceRepository implements a mock device repository for testing
type MockDeviceRepository struct {
	devices   map[string]*entities.Device
	saveError error
	mu        sync.RWMutex
}

func NewMockDeviceRepository() *MockDeviceRepository {
	return &MockDeviceRepository{
		devices: make(map[string]*entities.Device),
	}
}

func (m *MockDeviceRepository) Save(ctx context.Context, device *entities.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.saveError != nil {
		return m.saveError
	}

	m.devices[device.MACAddress] = device
	return nil
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *entities.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[device.MACAddress]; !exists {
		return fmt.Errorf("device not found")
	}

	m.devices[device.MACAddress] = device
	return nil
}

func (m *MockDeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if device, exists := m.devices[macAddress]; exists {
		return device, nil
	}

	return nil, nil
}

func (m *MockDeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.devices[macAddress]
	return exists, nil
}

func (m *MockDeviceRepository) Delete(ctx context.Context, macAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[macAddress]; !exists {
		return fmt.Errorf("device not found")
	}

	delete(m.devices, macAddress)
	return nil
}

func (m *MockDeviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*entities.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}

	// Simple pagination
	start := offset
	if start > len(devices) {
		return []*entities.Device{}, nil
	}

	end := offset + limit
	if end > len(devices) {
		end = len(devices)
	}

	return devices[start:end], nil
}

func (m *MockDeviceRepository) SetSaveError(err error) {
	m.saveError = err
}

// MockUseCase implements a mock use case for testing
type MockUseCase struct {
	callCount   int
	lastMessage *entities.DeviceRegistrationMessage
	returnError error
	mu          sync.RWMutex
}

// Ensure MockUseCase implements the DeviceRegistrationUseCase interface
var _ deviceregistration.DeviceRegistrationUseCase = (*MockUseCase)(nil)

func (m *MockUseCase) RegisterDevice(ctx context.Context, message *entities.DeviceRegistrationMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callCount++
	m.lastMessage = message
	return m.returnError
}

func (m *MockUseCase) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

func (m *MockUseCase) GetLastMessage() *entities.DeviceRegistrationMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastMessage
}

func (m *MockUseCase) SetReturnError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.returnError = err
}

func TestNewDeviceRegistrationHandler(t *testing.T) {
	// Create a real use case with a mock repository for testing
	mockRepo := NewMockDeviceRepository()
	realUseCase := deviceregistration.NewUseCase(mockRepo)
	handler := NewDeviceRegistrationHandler(realUseCase)

	if handler == nil {
		t.Errorf("NewDeviceRegistrationHandler() returned nil")
	}

	if handler.useCase == nil {
		t.Error("NewDeviceRegistrationHandler() did not set useCase")
	}
}

func TestDeviceRegistrationHandler_HandleMessage_ValidTopic(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := &MockUseCase{}
	handler := NewDeviceRegistrationHandler(mockUseCase)

	validPayload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	payload, err := json.Marshal(validPayload)
	if err != nil {
		t.Fatalf("Failed to marshal test payload: %v", err)
	}

	ctx := context.Background()
	err = handler.HandleMessage(ctx, "/liwaisi/iot/smart-irrigation/device/registration", payload)

	if err != nil {
		t.Errorf("HandleMessage() unexpected error: %v", err)
	}

	// Verify the use case was called
	if mockUseCase.GetCallCount() != 1 {
		t.Errorf("HandleMessage() expected use case to be called once, got %d calls", mockUseCase.GetCallCount())
	}

	// Verify the message was passed correctly
	lastMessage := mockUseCase.GetLastMessage()
	if lastMessage == nil {
		t.Errorf("HandleMessage() use case should have received a message")
		return
	}

	if lastMessage.MACAddress != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("HandleMessage() MAC address mismatch: expected %s, got %s", "AA:BB:CC:DD:EE:FF", lastMessage.MACAddress)
	}

	if lastMessage.DeviceName != "Test Device" {
		t.Errorf("HandleMessage() device name mismatch: expected %s, got %s", "Test Device", lastMessage.DeviceName)
	}
}

func TestDeviceRegistrationHandler_HandleMessage_UnknownTopic(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := &MockUseCase{}
	handler := NewDeviceRegistrationHandler(mockUseCase)

	validPayload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	payload, err := json.Marshal(validPayload)
	if err != nil {
		t.Fatalf("Failed to marshal test payload: %v", err)
	}

	ctx := context.Background()
	err = handler.HandleMessage(ctx, "/unknown/topic", payload)

	if err == nil {
		t.Errorf("HandleMessage() expected error for unknown topic but got none")
	}

	expectedError := "unknown topic: /unknown/topic"
	if err.Error() != expectedError {
		t.Errorf("HandleMessage() expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Verify the use case was not called
	if mockUseCase.GetCallCount() != 0 {
		t.Errorf("HandleMessage() use case should not be called for unknown topic, got %d calls", mockUseCase.GetCallCount())
	}
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
			mockUseCase := &MockUseCase{}
			handler := NewDeviceRegistrationHandler(mockUseCase)

			payload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test payload: %v", err)
			}

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payload)

			if err != nil {
				t.Errorf("processDeviceRegistration() unexpected error: %v", err)
			}

			// Verify the use case was called
			if mockUseCase.GetCallCount() != 1 {
				t.Errorf("processDeviceRegistration() expected use case to be called once, got %d calls", mockUseCase.GetCallCount())
			}

			// Verify the message data
			lastMessage := mockUseCase.GetLastMessage()
			if lastMessage == nil {
				t.Errorf("processDeviceRegistration() use case should have received a message")
				return
			}

			expectedMAC := tt.payload["mac_address"].(string)
			if expectedMAC == "aa:bb:cc:dd:ee:ff" {
				expectedMAC = "AA:BB:CC:DD:EE:FF" // Should be normalized to uppercase
			}

			if lastMessage.MACAddress != expectedMAC {
				t.Errorf("processDeviceRegistration() MAC address mismatch")
			}

			if lastMessage.DeviceName != tt.payload["device_name"].(string) {
				t.Errorf("processDeviceRegistration() device name mismatch")
			}

			if lastMessage.IPAddress != tt.payload["ip_address"].(string) {
				t.Errorf("processDeviceRegistration() IP address mismatch")
			}

			if lastMessage.LocationDescription != tt.payload["location_description"].(string) {
				t.Errorf("processDeviceRegistration() location description mismatch")
			}
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_MalformedJSON(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := &MockUseCase{}
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
			// Reset mock
			mockUseCase.callCount = 0

			ctx := context.Background()
			err := handler.processDeviceRegistration(ctx, tt.payload)

			if err == nil {
				t.Errorf("processDeviceRegistration() expected error for malformed JSON but got none")
			}

			// Verify the use case was not called
			if mockUseCase.GetCallCount() != 0 {
				t.Errorf("processDeviceRegistration() use case should not be called for malformed JSON, got %d calls", mockUseCase.GetCallCount())
			}
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_InvalidEventType(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := &MockUseCase{}
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
			// Reset mock
			mockUseCase.callCount = 0

			payload := map[string]interface{}{
				"event_type":           tt.eventType,
				"mac_address":          "AA:BB:CC:DD:EE:FF",
				"device_name":          "Test Device",
				"ip_address":           "192.168.1.100",
				"location_description": "Test Location",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal test payload: %v", err)
			}

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payloadBytes)

			if err == nil {
				t.Errorf("processDeviceRegistration() expected error for invalid event type but got none")
			}

			expectedError := "invalid event type for device registration: " + tt.eventType
			if err.Error() != expectedError {
				t.Errorf("processDeviceRegistration() expected error '%s', got '%s'", expectedError, err.Error())
			}

			// Verify the use case was not called
			if mockUseCase.GetCallCount() != 0 {
				t.Errorf("processDeviceRegistration() use case should not be called for invalid event type, got %d calls", mockUseCase.GetCallCount())
			}
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_InvalidDeviceData(t *testing.T) {
	// Create a mock use case for testing
	mockUseCase := &MockUseCase{}
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
			// Reset mock
			mockUseCase.callCount = 0

			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test payload: %v", err)
			}

			ctx := context.Background()
			err = handler.processDeviceRegistration(ctx, payloadBytes)

			if err == nil {
				t.Errorf("processDeviceRegistration() expected error for invalid device data but got none")
			}

			// Verify the use case was not called
			if mockUseCase.GetCallCount() != 0 {
				t.Errorf("processDeviceRegistration() use case should not be called for invalid device data, got %d calls", mockUseCase.GetCallCount())
			}
		})
	}
}

func TestDeviceRegistrationHandler_processDeviceRegistration_UseCaseError(t *testing.T) {
	// Create a real use case with a mock repository that returns an error
	mockRepo := NewMockDeviceRepository()
	mockRepo.SetSaveError(errors.New("use case processing failed"))

	realUseCase := deviceregistration.NewUseCase(mockRepo)
	handler := NewDeviceRegistrationHandler(realUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Test Device",
		"ip_address":           "192.168.1.100",
		"location_description": "Test Location",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal test payload: %v", err)
	}

	ctx := context.Background()
	err = handler.processDeviceRegistration(ctx, payloadBytes)

	if err == nil {
		t.Errorf("processDeviceRegistration() expected error from use case but got none")
	}

	// Check that the error is related to the mock repository error we set
	expectedErrorSubstring := "use case processing failed"
	if err.Error() != "failed to save new device: use case processing failed" {
		t.Errorf("processDeviceRegistration() expected error containing '%s', got '%s'", expectedErrorSubstring, err.Error())
	}

	// For this test, we expect the error to be returned, which indicates the use case was called
	// (we can't verify the call count with a real use case, but the error confirms it was called)
}

func TestDeviceRegistrationHandler_Integration(t *testing.T) {
	// This test verifies the full integration from HandleMessage to processDeviceRegistration
	mockUseCase := &MockUseCase{}
	handler := NewDeviceRegistrationHandler(mockUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Integration Test Device",
		"ip_address":           "192.168.1.200",
		"location_description": "Integration Test Location",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	err = handler.HandleMessage(context.Background(), "/liwaisi/iot/smart-irrigation/device/registration", payloadBytes)
	if err != nil {
		t.Fatalf("HandleMessage() returned error: %v", err)
	}

	// Verify the complete flow worked
	if mockUseCase.GetCallCount() != 1 {
		t.Errorf("Integration test expected use case to be called once, got %d calls", mockUseCase.GetCallCount())
	}

	lastMessage := mockUseCase.GetLastMessage()
	if lastMessage == nil {
		t.Errorf("Integration test use case should have received a message")
		return
	}

	if lastMessage.MACAddress != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Integration test MAC address mismatch")
	}

	if lastMessage.DeviceName != "Integration Test Device" {
		t.Errorf("Integration test device name mismatch")
	}

	if lastMessage.IPAddress != "192.168.1.200" {
		t.Errorf("Integration test IP address mismatch")
	}

	if lastMessage.LocationDescription != "Integration Test Location" {
		t.Errorf("Integration test location description mismatch")
	}
}

func TestDeviceRegistrationHandler_RealUseCaseIntegration(t *testing.T) {
	// This test uses a real use case with mock repository to test full integration
	mockRepo := NewMockDeviceRepository()
	realUseCase := deviceregistration.NewUseCase(mockRepo)
	handler := NewDeviceRegistrationHandler(realUseCase)

	payload := map[string]interface{}{
		"event_type":           "register",
		"mac_address":          "AA:BB:CC:DD:EE:FF",
		"device_name":          "Real Integration Device",
		"ip_address":           "192.168.1.250",
		"location_description": "Real Integration Location",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	err = handler.HandleMessage(context.Background(), "/liwaisi/iot/smart-irrigation/device/registration", payloadBytes)
	if err != nil {
		t.Fatalf("HandleMessage() returned error: %v", err)
	}

	// Verify device was saved to the mock repository
	device, exists := mockRepo.devices["AA:BB:CC:DD:EE:FF"]
	if !exists {
		t.Errorf("Real integration test device was not saved to repository")
		return
	}

	if device.DeviceName != "Real Integration Device" {
		t.Errorf("Real integration test device name mismatch")
	}

	if device.IPAddress != "192.168.1.250" {
		t.Errorf("Real integration test IP address mismatch")
	}

	if device.Status != "registered" {
		t.Errorf("Real integration test device status should be 'registered'")
	}
}
