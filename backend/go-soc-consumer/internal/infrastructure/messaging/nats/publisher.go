package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/mappers"
	"github.com/nats-io/nats.go"
)

// publisher implements the EventPublisher port using NATS
type publisher struct {
	config *NATSConfig
	conn   *nats.Conn
	logger *slog.Logger
	mu     sync.RWMutex
	mapper *mappers.DeviceDetectedEventMapper
}

// NewNATSPublisher creates a new NATS event publisher
func NewNATSPublisher(config *NATSConfig, logger *slog.Logger) (ports.EventPublisher, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid NATS config: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	p := &publisher{
		config: config,
		logger: logger,
		mapper: mappers.NewDeviceDetectedEventMapper(),
	}

	// Establish connection
	if err := p.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return p, nil
}

// connect establishes a connection to the NATS server
func (p *publisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts := []nats.Option{
		nats.Name(p.config.ClientID + "-publisher"),
		nats.Timeout(p.config.ConnectTimeout),
		nats.ReconnectWait(p.config.ReconnectWait),
		nats.MaxReconnects(p.config.MaxReconnectAttempts),
		nats.PingInterval(p.config.PingInterval),
		nats.MaxPingsOutstanding(p.config.MaxPingsOutstanding),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				p.logger.Error("NATS publisher disconnected", "error", err)
			} else {
				p.logger.Info("NATS publisher disconnected gracefully")
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			p.logger.Info("NATS publisher reconnected", "server", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			if nc.LastError() != nil {
				p.logger.Error("NATS publisher connection closed", "error", nc.LastError())
			} else {
				p.logger.Info("NATS publisher connection closed gracefully")
			}
		}),
	}

	conn, err := nats.Connect(p.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server at %s: %w", p.config.URL, err)
	}

	p.conn = conn
	p.logger.Info("NATS publisher connected successfully", "server", conn.ConnectedUrl())

	return nil
}

// Publish publishes an event to the specified subject
func (p *publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	p.mu.RLock()
	conn := p.conn
	p.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("NATS publisher not connected")
	}

	if !conn.IsConnected() {
		return fmt.Errorf("NATS publisher connection lost")
	}

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before publish: %w", err)
	}

	dto, err := p.mapper.ToDTOFromInterface(data)
	if err != nil {
		return err
	}

	dataBytes, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	p.logger.Debug("Publishing event to NATS",
		"subject", subject,
		"data_length", len(dataBytes))

	// Use a goroutine with done channel to handle context cancellation
	done := make(chan error, 1)
	go func() {
		done <- conn.Publish(subject, dataBytes)
	}()

	select {
	case err := <-done:
		if err != nil {
			p.logger.Error("Failed to publish event to NATS",
				"subject", subject,
				"error", err)
			return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
		}

		p.logger.Debug("Successfully published event to NATS", "subject", subject)
		return nil

	case <-ctx.Done():
		p.logger.Warn("Publish operation cancelled by context",
			"subject", subject,
			"error", ctx.Err())
		return fmt.Errorf("publish cancelled: %w", ctx.Err())
	}
}

// IsConnected returns true if the publisher is connected to NATS
func (p *publisher) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.conn != nil && p.conn.IsConnected()
}

// Close gracefully closes the NATS publisher connection
func (p *publisher) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil {
		return nil
	}

	p.logger.Info("Closing NATS publisher connection")

	// Close the connection with context timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.conn.Close()
	}()

	select {
	case <-done:
		p.conn = nil
		p.logger.Info("NATS publisher connection closed successfully")
		return nil

	case <-ctx.Done():
		// Force close if context timeout
		p.conn.Close()
		p.conn = nil
		p.logger.Warn("NATS publisher connection closed due to context timeout")
		return ctx.Err()
	}
}
