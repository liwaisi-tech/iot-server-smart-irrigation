package ports

import (
	"context"
)

// MessageHandler defines a function type for handling received messages
type MessageHandler func(ctx context.Context, topic string, payload []byte) error

// MessageConsumer defines the contract for consuming messages from external systems
type MessageConsumer interface {
	// Subscribe starts consuming messages from the specified topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error

	// Unsubscribe stops consuming messages from the specified topic
	Unsubscribe(topic string) error

	// Start begins the message consumer service
	Start(ctx context.Context) error

	// Stop gracefully shuts down the message consumer
	Stop(ctx context.Context) error

	// IsConnected returns the connection status
	IsConnected() bool
}