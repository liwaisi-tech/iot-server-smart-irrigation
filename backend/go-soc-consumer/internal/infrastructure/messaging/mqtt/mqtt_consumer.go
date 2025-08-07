package mqtt

import (
	"context"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// MQTTConsumerConfig holds configuration for MQTT consumer
type MQTTConsumerConfig struct {
	BrokerURL            string
	ClientID             string
	Username             string
	Password             string
	ConnectTimeout       time.Duration
	KeepAlive            time.Duration
	CleanSession         bool
	AutoReconnect        bool
	MaxReconnectInterval time.Duration
}

// MQTTConsumerImpl implements the MessageConsumer port
type MQTTConsumerImpl struct {
	config  MQTTConsumerConfig
	client  mqtt.Client
	handler ports.MessageHandler
}

// NewMQTTConsumer creates a new MQTT consumer
func NewMQTTConsumer(config MQTTConsumerConfig) *MQTTConsumerImpl {
	return &MQTTConsumerImpl{
		config: config,
	}
}

// Start begins consuming messages from MQTT broker
func (m *MQTTConsumerImpl) Start(ctx context.Context) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.config.BrokerURL)
	opts.SetClientID(m.config.ClientID)
	opts.SetUsername(m.config.Username)
	opts.SetPassword(m.config.Password)
	opts.SetConnectTimeout(m.config.ConnectTimeout)
	opts.SetKeepAlive(m.config.KeepAlive)
	opts.SetCleanSession(m.config.CleanSession)
	opts.SetAutoReconnect(m.config.AutoReconnect)
	opts.SetMaxReconnectInterval(m.config.MaxReconnectInterval)

	// Set connection lost handler
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	// Set on connect handler
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("MQTT connected successfully")
	})

	// Create MQTT client
	m.client = mqtt.NewClient(opts)

	// Connect to broker
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", m.config.BrokerURL)
	return nil
}

// Stop gracefully stops the MQTT consumer
func (m *MQTTConsumerImpl) Stop(ctx context.Context) error {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250) // Wait 250ms for graceful disconnect
		log.Println("MQTT consumer stopped")
	}
	return nil
}

// Subscribe subscribes to a specific topic with a message handler
func (m *MQTTConsumerImpl) Subscribe(ctx context.Context, topic string, handler ports.MessageHandler) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	m.handler = handler

	// Create message handler function
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))

		if err := m.handler(ctx, msg.Topic(), msg.Payload()); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}

	// Subscribe to topic
	if token := m.client.Subscribe(topic, 1, messageHandler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}

	log.Printf("Subscribed to MQTT topic: %s", topic)
	return nil
}

// Unsubscribe stops consuming messages from the specified topic
func (m *MQTTConsumerImpl) Unsubscribe(topic string) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	if token := m.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, token.Error())
	}

	log.Printf("Unsubscribed from MQTT topic: %s", topic)
	return nil
}

// IsConnected returns true if connected to MQTT broker
func (m *MQTTConsumerImpl) IsConnected() bool {
	return m.client != nil && m.client.IsConnected()
}
