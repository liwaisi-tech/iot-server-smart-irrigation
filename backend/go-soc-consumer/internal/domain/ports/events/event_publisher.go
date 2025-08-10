package ports

import (
	"context"
)

// EventPublisher defines the contract for publishing events to external messaging systems
// This interface follows the existing MessageConsumer pattern but for publishing events
type EventPublisher interface {
	// Publish publishes an event to the specified subject/topic
	Publish(ctx context.Context, subject string, data interface{}) error

	// Close gracefully shuts down the event publisher
	Close(ctx context.Context) error

	// IsConnected returns the connection status
	IsConnected() bool
}
