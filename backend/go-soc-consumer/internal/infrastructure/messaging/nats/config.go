package nats

import (
	"fmt"
	"os"
	"time"
)

// NATSConfig holds NATS connection configuration
type NATSConfig struct {
	URL                string
	ClientID           string
	SubjectPrefix      string
	ConnectTimeout     time.Duration
	ReconnectWait      time.Duration
	MaxReconnectAttempts int
	PingInterval       time.Duration
	MaxPingsOutstanding int
}

// DefaultNATSConfig returns default NATS configuration with environment variable overrides
func DefaultNATSConfig() *NATSConfig {
	config := &NATSConfig{
		URL:                  "nats://localhost:4222",
		ClientID:             "iot-go-soc-consumer",
		SubjectPrefix:        "liwaisi.iot.smart-irrigation",
		ConnectTimeout:       5 * time.Second,
		ReconnectWait:        2 * time.Second,
		MaxReconnectAttempts: 60, // Will keep trying for ~2 minutes
		PingInterval:         30 * time.Second,
		MaxPingsOutstanding:  2,
	}

	// Override with environment variables if present
	if url := os.Getenv("NATS_URL"); url != "" {
		config.URL = url
	}

	if clientID := os.Getenv("NATS_CLIENT_ID"); clientID != "" {
		config.ClientID = clientID
	}

	if prefix := os.Getenv("NATS_SUBJECT_PREFIX"); prefix != "" {
		config.SubjectPrefix = prefix
	}

	return config
}

// GetDeviceDetectedSubject returns the full subject name for device detected events
func (c *NATSConfig) GetDeviceDetectedSubject() string {
	return fmt.Sprintf("%s.device.detected", c.SubjectPrefix)
}

// Validate ensures the configuration is valid
func (c *NATSConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("NATS URL is required")
	}

	if c.ClientID == "" {
		return fmt.Errorf("NATS client ID is required")
	}

	if c.SubjectPrefix == "" {
		return fmt.Errorf("NATS subject prefix is required")
	}

	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect timeout must be positive")
	}

	if c.ReconnectWait <= 0 {
		return fmt.Errorf("reconnect wait must be positive")
	}

	return nil
}