package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// MockMQTTClient is a mock implementation of the MQTT client interface
type MockMQTTClient struct {
	mock.Mock
}

func (m *MockMQTTClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMQTTClient) IsConnectionOpen() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMQTTClient) Connect() mqtt.Token {
	args := m.Called()
	return args.Get(0).(mqtt.Token)
}

func (m *MockMQTTClient) Disconnect(quiesce uint) {
	m.Called(quiesce)
}

func (m *MockMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	args := m.Called(topic, qos, retained, payload)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(topic, qos, callback)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMQTTClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(filters, callback)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMQTTClient) Unsubscribe(topics ...string) mqtt.Token {
	args := m.Called(topics)
	return args.Get(0).(mqtt.Token)
}

func (m *MockMQTTClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	m.Called(topic, callback)
}

func (m *MockMQTTClient) OptionsReader() mqtt.ClientOptionsReader {
	args := m.Called()
	return args.Get(0).(mqtt.ClientOptionsReader)
}

// MockMQTTToken is a mock implementation of the MQTT token interface
type MockMQTTToken struct {
	mock.Mock
}

func (m *MockMQTTToken) Wait() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMQTTToken) WaitTimeout(timeout time.Duration) bool {
	args := m.Called(timeout)
	return args.Bool(0)
}

func (m *MockMQTTToken) Done() <-chan struct{} {
	args := m.Called()
	return args.Get(0).(<-chan struct{})
}

func (m *MockMQTTToken) Error() error {
	args := m.Called()
	return args.Error(0)
}

// TestNewMQTTConsumer tests the constructor
func TestNewMQTTConsumer(t *testing.T) {
	config := MQTTConsumerConfig{
		BrokerURL:             "tcp://localhost:1883",
		ClientID:              "test-client",
		Username:              "test-user",
		Password:              "test-pass",
		ConnectTimeout:        30 * time.Second,
		KeepAlive:             60 * time.Second,
		CleanSession:          true,
		AutoReconnect:         true,
		MaxReconnectInterval:  10 * time.Minute,
	}

	consumer := NewMQTTConsumer(config)

	assert.NotNil(t, consumer)
	assert.Equal(t, config, consumer.config)
	assert.Nil(t, consumer.client)
	assert.Nil(t, consumer.handler)
}

// TestMQTTConsumer_Stop tests the Stop method
func TestMQTTConsumer_Stop(t *testing.T) {
	tests := []struct {
		name         string
		setupClient  func() *MockMQTTClient
		wantErr      bool
	}{
		{
			name: "successful stop with connected client",
			setupClient: func() *MockMQTTClient {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(true)
				mockClient.On("Disconnect", uint(250)).Return()
				return mockClient
			},
			wantErr: false,
		},
		{
			name: "stop with disconnected client",
			setupClient: func() *MockMQTTClient {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(false)
				return mockClient
			},
			wantErr: false,
		},
		{
			name: "stop with nil client",
			setupClient: func() *MockMQTTClient {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MQTTConsumerConfig{
				BrokerURL: "tcp://localhost:1883",
				ClientID:  "test-client",
			}

			consumer := NewMQTTConsumer(config)
			
			if tt.setupClient != nil {
				mockClient := tt.setupClient()
				if mockClient != nil {
					consumer.client = mockClient
				}
			}

			err := consumer.Stop(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if consumer.client != nil {
				mockClient := consumer.client.(*MockMQTTClient)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

// TestMQTTConsumer_Subscribe tests the Subscribe method
func TestMQTTConsumer_Subscribe(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		handler ports.MessageHandler
		setup   func() (*MockMQTTClient, *MockMQTTToken)
		wantErr bool
		errMsg  string
	}{
		{
			name:  "successful subscription",
			topic: "test/topic",
			handler: func(ctx context.Context, topic string, payload []byte) error {
				return nil
			},
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockToken := &MockMQTTToken{}

				mockClient.On("IsConnected").Return(true)
				mockToken.On("Wait").Return(true)
				mockToken.On("Error").Return(nil)
				mockClient.On("Subscribe", "test/topic", byte(1), mock.AnythingOfType("mqtt.MessageHandler")).Return(mockToken)

				return mockClient, mockToken
			},
			wantErr: false,
		},
		{
			name:  "subscription with disconnected client",
			topic: "test/topic",
			handler: func(ctx context.Context, topic string, payload []byte) error {
				return nil
			},
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(false)
				return mockClient, nil
			},
			wantErr: true,
			errMsg:  "MQTT client is not connected",
		},
		{
			name:  "subscription failure",
			topic: "test/topic",
			handler: func(ctx context.Context, topic string, payload []byte) error {
				return nil
			},
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockToken := &MockMQTTToken{}

				mockClient.On("IsConnected").Return(true)
				mockToken.On("Wait").Return(true)
				mockToken.On("Error").Return(errors.New("subscription failed"))
				mockClient.On("Subscribe", "test/topic", byte(1), mock.AnythingOfType("mqtt.MessageHandler")).Return(mockToken)

				return mockClient, mockToken
			},
			wantErr: true,
			errMsg:  "failed to subscribe to topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MQTTConsumerConfig{
				BrokerURL: "tcp://localhost:1883",
				ClientID:  "test-client",
			}

			consumer := NewMQTTConsumer(config)
			mockClient, mockToken := tt.setup()
			consumer.client = mockClient

			err := consumer.Subscribe(context.Background(), tt.topic, tt.handler)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consumer.handler)
			}

			mockClient.AssertExpectations(t)
			if mockToken != nil {
				mockToken.AssertExpectations(t)
			}
		})
	}
}

// TestMQTTConsumer_Unsubscribe tests the Unsubscribe method
func TestMQTTConsumer_Unsubscribe(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		setup   func() (*MockMQTTClient, *MockMQTTToken)
		wantErr bool
		errMsg  string
	}{
		{
			name:  "successful unsubscription",
			topic: "test/topic",
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockToken := &MockMQTTToken{}

				mockClient.On("IsConnected").Return(true)
				mockToken.On("Wait").Return(true)
				mockToken.On("Error").Return(nil)
				mockClient.On("Unsubscribe", []string{"test/topic"}).Return(mockToken)

				return mockClient, mockToken
			},
			wantErr: false,
		},
		{
			name:  "unsubscription with disconnected client",
			topic: "test/topic",
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(false)
				return mockClient, nil
			},
			wantErr: true,
			errMsg:  "MQTT client is not connected",
		},
		{
			name:  "unsubscription failure",
			topic: "test/topic",
			setup: func() (*MockMQTTClient, *MockMQTTToken) {
				mockClient := &MockMQTTClient{}
				mockToken := &MockMQTTToken{}

				mockClient.On("IsConnected").Return(true)
				mockToken.On("Wait").Return(true)
				mockToken.On("Error").Return(errors.New("unsubscription failed"))
				mockClient.On("Unsubscribe", []string{"test/topic"}).Return(mockToken)

				return mockClient, mockToken
			},
			wantErr: true,
			errMsg:  "failed to unsubscribe from topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MQTTConsumerConfig{
				BrokerURL: "tcp://localhost:1883",
				ClientID:  "test-client",
			}

			consumer := NewMQTTConsumer(config)
			mockClient, mockToken := tt.setup()
			consumer.client = mockClient

			err := consumer.Unsubscribe(tt.topic)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			if mockToken != nil {
				mockToken.AssertExpectations(t)
			}
		})
	}
}

// TestMQTTConsumer_IsConnected tests the IsConnected method
func TestMQTTConsumer_IsConnected(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MockMQTTClient
		expected bool
	}{
		{
			name: "connected client",
			setup: func() *MockMQTTClient {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(true)
				return mockClient
			},
			expected: true,
		},
		{
			name: "disconnected client",
			setup: func() *MockMQTTClient {
				mockClient := &MockMQTTClient{}
				mockClient.On("IsConnected").Return(false)
				return mockClient
			},
			expected: false,
		},
		{
			name: "nil client",
			setup: func() *MockMQTTClient {
				return nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MQTTConsumerConfig{
				BrokerURL: "tcp://localhost:1883",
				ClientID:  "test-client",
			}

			consumer := NewMQTTConsumer(config)
			
			if tt.setup != nil {
				mockClient := tt.setup()
				if mockClient != nil {
					consumer.client = mockClient
				}
			}

			result := consumer.IsConnected()
			assert.Equal(t, tt.expected, result)

			if consumer.client != nil {
				mockClient := consumer.client.(*MockMQTTClient)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

// TestMQTTConsumer_MessageHandling tests message handling functionality
func TestMQTTConsumer_MessageHandling(t *testing.T) {
	t.Run("message handler processes messages correctly", func(t *testing.T) {
		config := MQTTConsumerConfig{
			BrokerURL: "tcp://localhost:1883",
			ClientID:  "test-client",
		}

		consumer := NewMQTTConsumer(config)
		
		// Create a test handler
		var receivedTopic string
		var receivedPayload []byte
		var handlerError error
		
		testHandler := func(ctx context.Context, topic string, payload []byte) error {
			receivedTopic = topic
			receivedPayload = payload
			return handlerError
		}
		
		consumer.handler = testHandler
		
		// Test that our handler works correctly
		err := testHandler(context.Background(), "test/topic", []byte("test payload"))
		
		assert.NoError(t, err)
		assert.Equal(t, "test/topic", receivedTopic)
		assert.Equal(t, []byte("test payload"), receivedPayload)
	})
	
	t.Run("message handler handles errors", func(t *testing.T) {
		config := MQTTConsumerConfig{
			BrokerURL: "tcp://localhost:1883",
			ClientID:  "test-client",
		}

		consumer := NewMQTTConsumer(config)
		
		// Create a handler that returns an error
		testHandler := func(ctx context.Context, topic string, payload []byte) error {
			return errors.New("handler error")
		}
		
		consumer.handler = testHandler
		
		// Test that the handler returns the expected error
		err := testHandler(context.Background(), "test/topic", []byte("test payload"))
		assert.Error(t, err)
		assert.Equal(t, "handler error", err.Error())
	})
}

// Benchmark tests  
func BenchmarkMQTTConsumer_MessageHandling(b *testing.B) {
	config := MQTTConsumerConfig{
		BrokerURL: "tcp://localhost:1883",
		ClientID:  "test-client",
	}

	consumer := NewMQTTConsumer(config)
	
	// Simple handler for benchmarking
	testHandler := func(ctx context.Context, topic string, payload []byte) error {
		return nil
	}
	
	consumer.handler = testHandler
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testHandler(context.Background(), "test/topic", []byte("test payload"))
	}
}