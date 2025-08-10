package logger

import (
	"time"

	"go.uber.org/zap"
)

// messagingLogger implements MessagingLogger interface
type messagingLogger struct {
	CoreLogger
}

// NewMessagingLogger creates a new messaging logger with the given core logger
func NewMessagingLogger(core CoreLogger) MessagingLogger {
	return &messagingLogger{
		CoreLogger: core,
	}
}

// LogMQTTMessage logs MQTT message processing with structured fields
func (l *messagingLogger) LogMQTTMessage(topic string, payloadSize int, processingDuration time.Duration, success bool) {
	level := l.Info
	message := "mqtt_message_processed"
	if !success {
		level = l.Error
		message = "mqtt_message_processing_failed"
	}

	level(message,
		zap.String("topic", topic),
		zap.Int("payload_size_bytes", payloadSize),
		zap.Duration("processing_duration", processingDuration),
		zap.Bool("success", success),
		zap.String("component", "mqtt_consumer"),
	)
}

// LogEventPublishing logs event publishing operations
func (l *messagingLogger) LogEventPublishing(eventType, subject, eventID string, success bool, err error) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("subject", subject),
		zap.String("event_id", eventID),
		zap.Bool("success", success),
		zap.String("component", "event_publisher"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("event_publishing_failed", fields...)
	} else {
		l.Info("event_published", fields...)
	}
}

// LogMessageProcessing logs generic message processing operations
func (l *messagingLogger) LogMessageProcessing(protocol, topic string, success bool, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("protocol", protocol),
		zap.String("topic", topic),
		zap.Bool("success", success),
		zap.String("component", "message_processor"),
	}, fields...)

	message := "message_processed"
	level := l.Info
	if !success {
		message = "message_processing_failed"
		level = l.Error
	}

	level(message, allFields...)
}