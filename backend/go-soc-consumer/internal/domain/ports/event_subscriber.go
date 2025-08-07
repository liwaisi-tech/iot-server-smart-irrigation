package ports

import (
	"context"
)

// EventSubscriber defines the contract for subscribing to events from external messaging systems
// This interface is separate from MessageConsumer to distinguish between MQTT and NATS handling
type EventSubscriber interface {
	// Subscribe starts consuming events from the specified subject/topic
	Subscribe(ctx context.Context, subject string, handler MessageHandler) error

	// Unsubscribe stops consuming events from the specified subject/topic
	Unsubscribe(ctx context.Context, subject string) error

	// Start begins the event subscriber service
	Start(ctx context.Context) error

	// Stop gracefully shuts down the event subscriber
	Stop(ctx context.Context) error

	// IsConnected returns the connection status
	IsConnected() bool
}