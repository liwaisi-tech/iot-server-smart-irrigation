package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/dtos"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

func TestSensorDataHandler_HandleMessage(t *testing.T) {
	// Create test logger factory
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	handler := NewSensorDataHandlerFromFactory(loggerFactory)
	ctx := context.Background()

	tests := []struct {
		name        string
		topic       string
		payload     []byte
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid sensor data message",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "sensor_data",
				MacAddress:  "A0:A3:B3:AB:2F:D8",
				Temperature: 28.8,
				Humidity:    72.3,
			}),
			wantErr: false,
		},
		{
			name:  "valid sensor data with edge case values",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "sensor_data",
				MacAddress:  "B0:B1:B2:B3:B4:B5",
				Temperature: -40.0,
				Humidity:    0.0,
			}),
			wantErr: false,
		},
		{
			name:        "unknown topic",
			topic:       "/unknown/topic",
			payload:     []byte(`{"event_type":"sensor_data"}`),
			wantErr:     true,
			errContains: "unknown sensor topic",
		},
		{
			name:        "invalid JSON",
			topic:       "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload:     []byte(`{invalid json`),
			wantErr:     true,
			errContains: "failed to unmarshal",
		},
		{
			name:  "invalid event type",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "invalid_type",
				MacAddress:  "A0:A3:B3:AB:2F:D8",
				Temperature: 28.8,
				Humidity:    72.3,
			}),
			wantErr:     true,
			errContains: "invalid event type",
		},
		{
			name:  "invalid MAC address",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "sensor_data",
				MacAddress:  "invalid-mac",
				Temperature: 28.8,
				Humidity:    72.3,
			}),
			wantErr:     true,
			errContains: "failed to create sensor data entity",
		},
		{
			name:  "temperature out of range",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "sensor_data",
				MacAddress:  "A0:A3:B3:AB:2F:D8",
				Temperature: 100.0,
				Humidity:    72.3,
			}),
			wantErr:     true,
			errContains: "failed to create sensor data entity",
		},
		{
			name:  "humidity out of range",
			topic: "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity",
			payload: createValidSensorDataPayload(t, dtos.SensorDataMessage{
				EventType:   "sensor_data",
				MacAddress:  "A0:A3:B3:AB:2F:D8",
				Temperature: 28.8,
				Humidity:    105.0,
			}),
			wantErr:     true,
			errContains: "failed to create sensor data entity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.HandleMessage(ctx, tt.topic, tt.payload)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSensorDataHandler_processSensorData(t *testing.T) {
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	handler := NewSensorDataHandlerFromFactory(loggerFactory)
	ctx := context.Background()

	t.Run("valid processing", func(t *testing.T) {
		payload := createValidSensorDataPayload(t, dtos.SensorDataMessage{
			EventType:   "sensor_data",
			MacAddress:  "A0:A3:B3:AB:2F:D8",
			Temperature: 28.8,
			Humidity:    72.3,
		})

		err := handler.processSensorData(ctx, payload)
		assert.NoError(t, err)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		payload := []byte(`{malformed`)
		
		err := handler.processSensorData(ctx, payload)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	t.Run("missing fields", func(t *testing.T) {
		payload := []byte(`{"event_type":"sensor_data"}`)
		
		err := handler.processSensorData(ctx, payload)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create sensor data entity")
	})
}

func TestNewSensorDataHandler(t *testing.T) {
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	handler := NewSensorDataHandlerFromFactory(loggerFactory)
	assert.NotNil(t, handler)
	// Logger fields are private after refactoring - just test that handler was created
}

// Helper function to create valid sensor data payload
func createValidSensorDataPayload(t *testing.T, msg dtos.SensorDataMessage) []byte {
	payload, err := json.Marshal(msg)
	require.NoError(t, err)
	return payload
}

func TestSensorDataHandler_HandleMessage_Integration(t *testing.T) {
	// Test actual JSON payload that would come from IoT device
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	handler := NewSensorDataHandlerFromFactory(loggerFactory)
	ctx := context.Background()

	// Test with the exact JSON format specified in requirements
	jsonPayload := `{"event_type":"sensor_data","mac_address":"A0:A3:B3:AB:2F:D8","temperature":28.8000,"humidity":72.3}`
	topic := "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"

	err = handler.HandleMessage(ctx, topic, []byte(jsonPayload))
	assert.NoError(t, err)
}

func TestSensorDataHandler_AbnormalReadingsLogging(t *testing.T) {
	loggerFactory, err := logger.NewDevelopmentLoggerFactory()
	require.NoError(t, err)

	handler := NewSensorDataHandlerFromFactory(loggerFactory)
	ctx := context.Background()
	topic := "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"

	// Test normal readings
	normalPayload := createValidSensorDataPayload(t, dtos.SensorDataMessage{
		EventType:   "sensor_data",
		MacAddress:  "A0:A3:B3:AB:2F:D8",
		Temperature: 25.0, // Normal range
		Humidity:    50.0, // Normal range
	})

	err = handler.HandleMessage(ctx, topic, normalPayload)
	assert.NoError(t, err)

	// Test abnormal readings
	abnormalPayload := createValidSensorDataPayload(t, dtos.SensorDataMessage{
		EventType:   "sensor_data",
		MacAddress:  "A0:A3:B3:AB:2F:D8",
		Temperature: 45.0, // Above normal range (>40°C)
		Humidity:    80.0, // Above normal range (>70%)
	})

	err = handler.HandleMessage(ctx, topic, abnormalPayload)
	assert.NoError(t, err)
}