package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/mappers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
	"github.com/nats-io/nats.go"
)

// publisher implements the EventPublisher port using NATS
type publisher struct {
	config        *NATSConfig
	conn          *nats.Conn
	loggerFactory logger.LoggerFactory
	mu            sync.RWMutex
	mapper        *mappers.DeviceDetectedEventMapper
}

// NewNATSPublisher creates a new NATS event publisher
func NewNATSPublisher(config *NATSConfig, loggerFactory logger.LoggerFactory) (ports.EventPublisher, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid NATS config: %w", err)
	}

	if loggerFactory == nil {
		defaultLoggerFactory, err := logger.NewDefault()
		if err != nil {
			return nil, fmt.Errorf("failed to create default logger factory: %w", err)
		}
		loggerFactory = defaultLoggerFactory
	}

	p := &publisher{
		config:        config,
		loggerFactory: loggerFactory,
		mapper:        mappers.NewDeviceDetectedEventMapper(),
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
				p.loggerFactory.Core().Error("nats_publisher_disconnected",
					zap.Error(err),
					zap.String("server_url", p.config.URL),
					zap.String("client_id", p.config.ClientID),
					zap.String("component", "nats_publisher"),
				)
			} else {
				p.loggerFactory.Application().LogApplicationEvent("nats_publisher_disconnected_gracefully", "nats_publisher",
					zap.String("server_url", p.config.URL),
					zap.String("client_id", p.config.ClientID),
				)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			p.loggerFactory.Application().LogApplicationEvent("nats_publisher_reconnected", "nats_publisher",
				zap.String("server_url", nc.ConnectedUrl()),
				zap.String("client_id", p.config.ClientID),
			)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			if nc.LastError() != nil {
				p.loggerFactory.Core().Error("nats_publisher_connection_closed",
					zap.Error(nc.LastError()),
					zap.String("server_url", p.config.URL),
					zap.String("client_id", p.config.ClientID),
					zap.String("component", "nats_publisher"),
				)
			} else {
				p.loggerFactory.Application().LogApplicationEvent("nats_publisher_connection_closed_gracefully", "nats_publisher",
					zap.String("server_url", p.config.URL),
					zap.String("client_id", p.config.ClientID),
				)
			}
		}),
	}

	start := time.Now()
	conn, err := nats.Connect(p.config.URL, opts...)
	connectionDuration := time.Since(start)
	
	if err != nil {
		p.loggerFactory.Core().Error("nats_publisher_connection_failed",
			zap.Error(err),
			zap.String("server_url", p.config.URL),
			zap.String("client_id", p.config.ClientID),
			zap.Duration("connection_attempt_duration", connectionDuration),
			zap.String("component", "nats_publisher"),
		)
		return fmt.Errorf("failed to connect to NATS server at %s: %w", p.config.URL, err)
	}

	p.conn = conn
	p.loggerFactory.Application().LogApplicationEvent("nats_publisher_connected", "nats_publisher",
		zap.String("server_url", conn.ConnectedUrl()),
		zap.String("client_id", p.config.ClientID),
		zap.Duration("connection_duration", connectionDuration),
	)

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
		p.loggerFactory.Core().Error("nats_event_marshaling_failed",
			zap.Error(err),
			zap.String("subject", subject),
			zap.String("component", "nats_publisher"),
		)
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	p.loggerFactory.Core().Debug("nats_event_publishing",
		zap.String("subject", subject),
		zap.Int("data_length_bytes", len(dataBytes)),
		zap.String("component", "nats_publisher"),
	)

	// Use a goroutine with done channel to handle context cancellation
	start := time.Now()
	done := make(chan error, 1)
	go func() {
		done <- conn.Publish(subject, dataBytes)
	}()

	select {
	case err := <-done:
		publishDuration := time.Since(start)
		if err != nil {
			p.loggerFactory.Messaging().LogEventPublishing("", subject, "", false, err)
			p.loggerFactory.Core().Error("nats_event_publishing_failed",
				zap.Error(err),
				zap.String("subject", subject),
				zap.Duration("publish_duration", publishDuration),
				zap.String("component", "nats_publisher"),
			)
			return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
		}

		p.loggerFactory.Messaging().LogEventPublishing("", subject, "", true, nil)
		p.loggerFactory.Core().Debug("nats_event_published_successfully",
			zap.String("subject", subject),
			zap.Duration("publish_duration", publishDuration),
			zap.String("component", "nats_publisher"),
		)
		return nil

	case <-ctx.Done():
		publishDuration := time.Since(start)
		p.loggerFactory.Core().Warn("nats_publish_operation_cancelled",
			zap.String("subject", subject),
			zap.Error(ctx.Err()),
			zap.Duration("cancelled_after", publishDuration),
			zap.String("component", "nats_publisher"),
		)
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

	p.loggerFactory.Application().LogApplicationEvent("nats_publisher_closing", "nats_publisher",
		zap.String("server_url", p.config.URL),
		zap.String("client_id", p.config.ClientID),
	)

	// Close the connection with context timeout
	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.conn.Close()
	}()

	select {
	case <-done:
		p.conn = nil
		p.loggerFactory.Application().LogApplicationEvent("nats_publisher_closed", "nats_publisher",
			zap.String("server_url", p.config.URL),
			zap.String("client_id", p.config.ClientID),
			zap.Duration("close_duration", time.Since(start)),
		)
		return nil

	case <-ctx.Done():
		// Force close if context timeout
		p.conn.Close()
		p.conn = nil
		p.loggerFactory.Core().Warn("nats_publisher_closed_timeout",
			zap.String("server_url", p.config.URL),
			zap.String("client_id", p.config.ClientID),
			zap.Duration("timeout_after", time.Since(start)),
			zap.String("component", "nats_publisher"),
		)
		return ctx.Err()
	}
}
