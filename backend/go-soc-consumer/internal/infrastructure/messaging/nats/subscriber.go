package nats

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// subscriber implements the EventSubscriber port using NATS
type subscriber struct {
	config        *NATSConfig
	conn          *nats.Conn
	subscriptions map[string]*nats.Subscription
	logger        *slog.Logger
	mu            sync.RWMutex
	started       bool
}

// NewNATSSubscriber creates a new NATS event subscriber
func NewNATSSubscriber(config *NATSConfig, logger *slog.Logger) (ports.EventSubscriber, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid NATS config: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &subscriber{
		config:        config,
		subscriptions: make(map[string]*nats.Subscription),
		logger:        logger,
	}, nil
}

// Start establishes connection to NATS and starts the subscriber
func (s *subscriber) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("NATS subscriber is already started")
	}

	if err := s.connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	s.started = true
	s.logger.Info("NATS subscriber started successfully")
	return nil
}

// connect establishes a connection to the NATS server
func (s *subscriber) connect() error {
	opts := []nats.Option{
		nats.Name(s.config.ClientID + "-subscriber"),
		nats.Timeout(s.config.ConnectTimeout),
		nats.ReconnectWait(s.config.ReconnectWait),
		nats.MaxReconnects(s.config.MaxReconnectAttempts),
		nats.PingInterval(s.config.PingInterval),
		nats.MaxPingsOutstanding(s.config.MaxPingsOutstanding),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				s.logger.Error("NATS subscriber disconnected", "error", err)
			} else {
				s.logger.Info("NATS subscriber disconnected gracefully")
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			s.logger.Info("NATS subscriber reconnected", "server", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			if nc.LastError() != nil {
				s.logger.Error("NATS subscriber connection closed", "error", nc.LastError())
			} else {
				s.logger.Info("NATS subscriber connection closed gracefully")
			}
		}),
	}

	conn, err := nats.Connect(s.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server at %s: %w", s.config.URL, err)
	}

	s.conn = conn
	s.logger.Info("NATS subscriber connected successfully", "server", conn.ConnectedUrl())
	
	return nil
}

// Subscribe subscribes to events from the specified subject
func (s *subscriber) Subscribe(ctx context.Context, subject string, handler ports.MessageHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("NATS subscriber not started")
	}

	if s.conn == nil || !s.conn.IsConnected() {
		return fmt.Errorf("NATS subscriber not connected")
	}

	if _, exists := s.subscriptions[subject]; exists {
		return fmt.Errorf("already subscribed to subject: %s", subject)
	}

	s.logger.Info("Subscribing to NATS subject", "subject", subject)

	// Create a wrapper handler that adapts NATS message to our MessageHandler interface
	natsHandler := func(msg *nats.Msg) {
		s.logger.Debug("Received NATS message", 
			"subject", msg.Subject, 
			"data_length", len(msg.Data))

		// Create a background context for message processing
		// Individual handlers should implement their own timeouts if needed
		msgCtx := context.Background()

		if err := handler(msgCtx, msg.Subject, msg.Data); err != nil {
			s.logger.Error("Error handling NATS message", 
				"subject", msg.Subject, 
				"error", err)
		}
	}

	sub, err := s.conn.Subscribe(subject, natsHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	s.subscriptions[subject] = sub
	s.logger.Info("Successfully subscribed to NATS subject", "subject", subject)
	
	return nil
}

// Unsubscribe stops consuming events from the specified subject
func (s *subscriber) Unsubscribe(ctx context.Context, subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, exists := s.subscriptions[subject]
	if !exists {
		return fmt.Errorf("not subscribed to subject: %s", subject)
	}

	s.logger.Info("Unsubscribing from NATS subject", "subject", subject)

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("failed to unsubscribe from subject %s: %w", subject, err)
	}

	delete(s.subscriptions, subject)
	s.logger.Info("Successfully unsubscribed from NATS subject", "subject", subject)
	
	return nil
}

// IsConnected returns true if the subscriber is connected to NATS
func (s *subscriber) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.conn != nil && s.conn.IsConnected()
}

// Stop gracefully shuts down the NATS subscriber
func (s *subscriber) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	s.logger.Info("Stopping NATS subscriber")

	// Unsubscribe from all subjects
	for subject, sub := range s.subscriptions {
		s.logger.Debug("Unsubscribing from subject during shutdown", "subject", subject)
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warn("Error unsubscribing from subject during shutdown", 
				"subject", subject, 
				"error", err)
		}
	}
	s.subscriptions = make(map[string]*nats.Subscription)

	// Close the connection
	if s.conn != nil {
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.conn.Close()
		}()

		select {
		case <-done:
			s.logger.Info("NATS subscriber connection closed successfully")
		case <-ctx.Done():
			// Force close if context timeout
			s.conn.Close()
			s.logger.Warn("NATS subscriber connection closed due to context timeout")
		}

		s.conn = nil
	}

	s.started = false
	s.logger.Info("NATS subscriber stopped successfully")
	
	return ctx.Err()
}