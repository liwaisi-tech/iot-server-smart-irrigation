package nats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nats-io/nats.go"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// subscriber implements the EventSubscriber port using NATS
type subscriber struct {
	config        *NATSConfig
	conn          *nats.Conn
	subscriptions map[string]*nats.Subscription
	logger        *logger.IoTLogger
	mu            sync.RWMutex
	started       bool
}

// NewNATSSubscriber creates a new NATS event subscriber
func NewNATSSubscriber(config *NATSConfig, iotLogger *logger.IoTLogger) (ports.EventSubscriber, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid NATS config: %w", err)
	}

	if iotLogger == nil {
		defaultLogger, err := logger.NewDefaultLogger()
		if err != nil {
			return nil, fmt.Errorf("failed to create default logger: %w", err)
		}
		iotLogger = defaultLogger
	}

	return &subscriber{
		config:        config,
		subscriptions: make(map[string]*nats.Subscription),
		logger:        iotLogger,
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
	s.logger.LogApplicationEvent("nats_subscriber_started", "nats_subscriber",
		zap.String("server_url", s.config.URL),
		zap.String("client_id", s.config.ClientID),
	)
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
				s.logger.Error("nats_subscriber_disconnected",
					zap.Error(err),
					zap.String("server_url", s.config.URL),
					zap.String("client_id", s.config.ClientID),
					zap.String("component", "nats_subscriber"),
				)
			} else {
				s.logger.LogApplicationEvent("nats_subscriber_disconnected_gracefully", "nats_subscriber",
					zap.String("server_url", s.config.URL),
					zap.String("client_id", s.config.ClientID),
				)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			s.logger.LogApplicationEvent("nats_subscriber_reconnected", "nats_subscriber",
				zap.String("server_url", nc.ConnectedUrl()),
				zap.String("client_id", s.config.ClientID),
			)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			if nc.LastError() != nil {
				s.logger.Error("nats_subscriber_connection_closed",
					zap.Error(nc.LastError()),
					zap.String("server_url", s.config.URL),
					zap.String("client_id", s.config.ClientID),
					zap.String("component", "nats_subscriber"),
				)
			} else {
				s.logger.LogApplicationEvent("nats_subscriber_connection_closed_gracefully", "nats_subscriber",
					zap.String("server_url", s.config.URL),
					zap.String("client_id", s.config.ClientID),
				)
			}
		}),
	}

	conn, err := nats.Connect(s.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server at %s: %w", s.config.URL, err)
	}

	s.conn = conn
	s.logger.LogApplicationEvent("nats_subscriber_connected", "nats_subscriber",
		zap.String("server_url", conn.ConnectedUrl()),
		zap.String("client_id", s.config.ClientID),
	)
	
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

	s.logger.LogApplicationEvent("nats_subscribing_to_subject", "nats_subscriber",
		zap.String("subject", subject),
		zap.String("client_id", s.config.ClientID),
	)

	// Create a wrapper handler that adapts NATS message to our MessageHandler interface
	natsHandler := func(msg *nats.Msg) {
		start := time.Now()
		payloadSize := len(msg.Data)
		
		s.logger.Debug("nats_message_received",
			zap.String("subject", msg.Subject),
			zap.Int("data_length_bytes", payloadSize),
			zap.String("component", "nats_subscriber"),
		)

		// Create a background context for message processing
		// Individual handlers should implement their own timeouts if needed
		msgCtx := context.Background()

		err := handler(msgCtx, msg.Subject, msg.Data)
		processingDuration := time.Since(start)
		
		if err != nil {
			s.logger.Error("nats_message_processing_error",
				zap.Error(err),
				zap.String("subject", msg.Subject),
				zap.Int("payload_size_bytes", payloadSize),
				zap.Duration("processing_duration", processingDuration),
				zap.String("component", "nats_subscriber"),
			)
		} else {
			s.logger.Debug("nats_message_processed_successfully",
				zap.String("subject", msg.Subject),
				zap.Int("payload_size_bytes", payloadSize),
				zap.Duration("processing_duration", processingDuration),
				zap.String("component", "nats_subscriber"),
			)
		}
	}

	sub, err := s.conn.Subscribe(subject, natsHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	s.subscriptions[subject] = sub
	s.logger.LogApplicationEvent("nats_subscribed_to_subject", "nats_subscriber",
		zap.String("subject", subject),
		zap.String("client_id", s.config.ClientID),
	)
	
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

	s.logger.LogApplicationEvent("nats_subject_unsubscribing", "nats_subscriber",
		zap.String("subject", subject),
		zap.String("client_id", s.config.ClientID),
	)

	start := time.Now()
	if err := sub.Unsubscribe(); err != nil {
		s.logger.Error("nats_subject_unsubscription_failed",
			zap.Error(err),
			zap.String("subject", subject),
			zap.Duration("unsubscription_attempt_duration", time.Since(start)),
			zap.String("component", "nats_subscriber"),
		)
		return fmt.Errorf("failed to unsubscribe from subject %s: %w", subject, err)
	}

	delete(s.subscriptions, subject)
	s.logger.LogApplicationEvent("nats_subject_unsubscribed", "nats_subscriber",
		zap.String("subject", subject),
		zap.String("client_id", s.config.ClientID),
		zap.Duration("unsubscription_duration", time.Since(start)),
	)
	
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

	s.logger.LogApplicationEvent("nats_subscriber_stopping", "nats_subscriber")

	// Unsubscribe from all subjects
	for subject, sub := range s.subscriptions {
		s.logger.Debug("nats_subject_unsubscribing_shutdown",
			zap.String("subject", subject),
			zap.String("component", "nats_subscriber"),
		)
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warn("nats_subject_unsubscription_error_shutdown",
				zap.Error(err),
				zap.String("subject", subject),
				zap.String("component", "nats_subscriber"),
			)
		}
	}
	s.subscriptions = make(map[string]*nats.Subscription)

	// Close the connection
	if s.conn != nil {
		start := time.Now()
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.conn.Close()
		}()

		select {
		case <-done:
			s.logger.LogApplicationEvent("nats_subscriber_connection_closed", "nats_subscriber",
				zap.Duration("close_duration", time.Since(start)),
			)
		case <-ctx.Done():
			// Force close if context timeout
			s.conn.Close()
			s.logger.Warn("nats_subscriber_connection_timeout",
				zap.Duration("timeout_after", time.Since(start)),
				zap.String("component", "nats_subscriber"),
			)
		}

		s.conn = nil
	}

	s.started = false
	s.logger.LogApplicationEvent("nats_subscriber_stopped", "nats_subscriber")
	
	return ctx.Err()
}