package mqtt

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	eventports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/events"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
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
	config        MQTTConsumerConfig
	client        mqtt.Client
	handlers      map[string]eventports.MessageHandler
	loggerFactory logger.LoggerFactory
}

// NewMQTTConsumer creates a new MQTT consumer
func NewMQTTConsumer(config MQTTConsumerConfig, loggerFactory logger.LoggerFactory) *MQTTConsumerImpl {
	return &MQTTConsumerImpl{
		config:        config,
		handlers:      make(map[string]eventports.MessageHandler),
		loggerFactory: loggerFactory,
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
		m.loggerFactory.Core().Error("mqtt_connection_lost",
			zap.Error(err),
			zap.String("broker_url", m.config.BrokerURL),
			zap.String("client_id", m.config.ClientID),
			zap.String("component", "mqtt_consumer"),
		)
	})

	// Set on connect handler
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		m.loggerFactory.Application().LogApplicationEvent("mqtt_connected", "mqtt_consumer",
			zap.String("broker_url", m.config.BrokerURL),
			zap.String("client_id", m.config.ClientID),
		)
	})

	// Create MQTT client
	m.client = mqtt.NewClient(opts)

	// Connect to broker
	start := time.Now()
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		m.loggerFactory.Core().Error("mqtt_connection_failed",
			zap.Error(token.Error()),
			zap.String("broker_url", m.config.BrokerURL),
			zap.String("client_id", m.config.ClientID),
			zap.Duration("connection_attempt_duration", time.Since(start)),
			zap.String("component", "mqtt_consumer"),
		)
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	m.loggerFactory.Application().LogApplicationEvent("mqtt_broker_connected", "mqtt_consumer",
		zap.String("broker_url", m.config.BrokerURL),
		zap.String("client_id", m.config.ClientID),
		zap.Duration("connection_duration", time.Since(start)),
	)
	return nil
}

// Stop gracefully stops the MQTT consumer
func (m *MQTTConsumerImpl) Stop(ctx context.Context) error {
	if m.client != nil && m.client.IsConnected() {
		start := time.Now()
		m.client.Disconnect(250) // Wait 250ms for graceful disconnect
		m.loggerFactory.Application().LogApplicationEvent("mqtt_consumer_stopped", "mqtt_consumer",
			zap.Duration("shutdown_duration", time.Since(start)),
			zap.String("client_id", m.config.ClientID),
		)
	}
	return nil
}

// Subscribe subscribes to a specific topic with a message handler
func (m *MQTTConsumerImpl) Subscribe(ctx context.Context, topic string, handler eventports.MessageHandler) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	// Store the handler for this specific topic
	m.handlers[topic] = handler

	// Create message handler function
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		start := time.Now()
		payloadSize := len(msg.Payload())

		m.loggerFactory.Core().Debug("mqtt_message_received",
			zap.String("topic", msg.Topic()),
			zap.Int("payload_size_bytes", payloadSize),
			zap.String("component", "mqtt_consumer"),
		)

		// Get the appropriate handler for this topic
		topicHandler, exists := m.handlers[msg.Topic()]
		if !exists {
			m.loggerFactory.Core().Error("no_handler_for_topic",
				zap.String("topic", msg.Topic()),
				zap.String("component", "mqtt_consumer"),
			)
			return
		}

		err := topicHandler(ctx, msg.Topic(), msg.Payload())
		processingDuration := time.Since(start)

		m.loggerFactory.Messaging().LogMQTTMessage(msg.Topic(), payloadSize, processingDuration, err == nil)

		if err != nil {
			m.loggerFactory.Core().Error("mqtt_message_processing_error",
				zap.Error(err),
				zap.String("topic", msg.Topic()),
				zap.Int("payload_size_bytes", payloadSize),
				zap.Duration("processing_duration", processingDuration),
				zap.String("component", "mqtt_consumer"),
			)
		}
	}

	// Subscribe to topic
	start := time.Now()
	if token := m.client.Subscribe(topic, 1, messageHandler); token.Wait() && token.Error() != nil {
		m.loggerFactory.Core().Error("mqtt_subscription_failed",
			zap.Error(token.Error()),
			zap.String("topic", topic),
			zap.String("client_id", m.config.ClientID),
			zap.Duration("subscription_attempt_duration", time.Since(start)),
			zap.String("component", "mqtt_consumer"),
		)
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}

	m.loggerFactory.Application().LogApplicationEvent("mqtt_topic_subscribed", "mqtt_consumer",
		zap.String("topic", topic),
		zap.String("client_id", m.config.ClientID),
		zap.Duration("subscription_duration", time.Since(start)),
		zap.Int("qos", 1),
	)
	return nil
}

// Unsubscribe stops consuming messages from the specified topic
func (m *MQTTConsumerImpl) Unsubscribe(topic string) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	start := time.Now()
	if token := m.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		m.loggerFactory.Core().Error("mqtt_unsubscription_failed",
			zap.Error(token.Error()),
			zap.String("topic", topic),
			zap.String("client_id", m.config.ClientID),
			zap.Duration("unsubscription_attempt_duration", time.Since(start)),
			zap.String("component", "mqtt_consumer"),
		)
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, token.Error())
	}

	// Remove the handler from the map
	delete(m.handlers, topic)

	m.loggerFactory.Application().LogApplicationEvent("mqtt_topic_unsubscribed", "mqtt_consumer",
		zap.String("topic", topic),
		zap.String("client_id", m.config.ClientID),
		zap.Duration("unsubscription_duration", time.Since(start)),
	)
	return nil
}

// IsConnected returns true if connected to MQTT broker
func (m *MQTTConsumerImpl) IsConnected() bool {
	return m.client != nil && m.client.IsConnected()
}
