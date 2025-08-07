package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/models"
)

func TestDeviceMapper_ToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *entities.Device
		expected *models.DeviceModel
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid device",
			input: &entities.Device{
				MACAddress:          "00:11:22:33:44:55",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.1",
				LocationDescription: "Test Location",
				RegisteredAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				LastSeen:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				Status:              "active",
			},
			expected: &models.DeviceModel{
				MACAddress:          "00:11:22:33:44:55",
				DeviceName:          "Test Device",
				IPAddress:           "192.168.1.1",
				LocationDescription: "Test Location",
				RegisteredAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				LastSeen:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				Status:              "active",
			},
		},
	}

	mapper := NewDeviceMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.ToModel(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.MACAddress, result.MACAddress)
			assert.Equal(t, tt.expected.DeviceName, result.DeviceName)
			assert.Equal(t, tt.expected.IPAddress, result.IPAddress)
			assert.Equal(t, tt.expected.LocationDescription, result.LocationDescription)
			assert.True(t, tt.expected.RegisteredAt.Equal(result.RegisteredAt))
			assert.True(t, tt.expected.LastSeen.Equal(result.LastSeen))
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.False(t, result.CreatedAt.IsZero())
			assert.False(t, result.UpdatedAt.IsZero())
		})
	}
}

func TestDeviceMapper_FromModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.DeviceModel
		expected *entities.Device
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid device model",
			input: &models.DeviceModel{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device From Model",
				IPAddress:           "10.0.0.1",
				LocationDescription: "Test Location From Model",
				RegisteredAt:        time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC),
				LastSeen:            time.Date(2023, 6, 2, 14, 30, 0, 0, time.UTC),
				Status:              "inactive",
			},
			expected: &entities.Device{
				MACAddress:          "AA:BB:CC:DD:EE:FF",
				DeviceName:          "Test Device From Model",
				IPAddress:           "10.0.0.1",
				LocationDescription: "Test Location From Model",
				RegisteredAt:        time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC),
				LastSeen:            time.Date(2023, 6, 2, 14, 30, 0, 0, time.UTC),
				Status:              "inactive",
			},
		},
	}

	mapper := NewDeviceMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.FromModel(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.MACAddress, result.MACAddress)
			assert.Equal(t, tt.expected.DeviceName, result.DeviceName)
			assert.Equal(t, tt.expected.IPAddress, result.IPAddress)
			assert.Equal(t, tt.expected.LocationDescription, result.LocationDescription)
			assert.True(t, tt.expected.RegisteredAt.Equal(result.RegisteredAt))
			assert.True(t, tt.expected.LastSeen.Equal(result.LastSeen))
			assert.Equal(t, tt.expected.Status, result.Status)
		})
	}
}
