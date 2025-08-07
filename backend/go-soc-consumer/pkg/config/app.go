package config

import (
	"fmt"
	"time"
)

// AppConfig holds all application configuration
type AppConfig struct {
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	MQTT        MQTTConfig        `json:"mqtt"`
	NATS        NATSConfig        `json:"nats"`
	HealthCheck HealthCheckConfig `json:"health_check"`
	Logging     LoggingConfig     `json:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// MQTTConfig holds MQTT configuration
type MQTTConfig struct {
	BrokerURL            string        `json:"broker_url"`
	ClientID             string        `json:"client_id"`
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	CleanSession         bool          `json:"clean_session"`
	AutoReconnect        bool          `json:"auto_reconnect"`
	ConnectTimeout       time.Duration `json:"connect_timeout"`
	KeepAlive            time.Duration `json:"keep_alive"`
	MaxReconnectInterval time.Duration `json:"max_reconnect_interval"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URLs            []string      `json:"urls"`
	MaxReconnect    int           `json:"max_reconnect"`
	ReconnectWait   time.Duration `json:"reconnect_wait"`
	Timeout         time.Duration `json:"timeout"`
	DrainTimeout    time.Duration `json:"drain_timeout"`
	FlusherTimeout  time.Duration `json:"flusher_timeout"`
	PingInterval    time.Duration `json:"ping_interval"`
	MaxPingsOut     int           `json:"max_pings_out"`
	ReconnectBufSize int          `json:"reconnect_buf_size"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	InitialDelay  time.Duration `json:"initial_delay"`
	UserAgent     string        `json:"user_agent"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// NewAppConfig creates a new application configuration from environment variables
func NewAppConfig() (*AppConfig, error) {
	config := &AppConfig{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: *NewDatabaseConfig(),
		MQTT: MQTTConfig{
			BrokerURL:            getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
			ClientID:             getEnv("MQTT_CLIENT_ID", "iot-go-soc-consumer"),
			Username:             getEnv("MQTT_USERNAME", ""),
			Password:             getEnv("MQTT_PASSWORD", ""),
			CleanSession:         getEnvBool("MQTT_CLEAN_SESSION", true),
			AutoReconnect:        getEnvBool("MQTT_AUTO_RECONNECT", true),
			ConnectTimeout:       getEnvDuration("MQTT_CONNECT_TIMEOUT", 30*time.Second),
			KeepAlive:            getEnvDuration("MQTT_KEEP_ALIVE", 60*time.Second),
			MaxReconnectInterval: getEnvDuration("MQTT_MAX_RECONNECT_INTERVAL", 10*time.Minute),
		},
		NATS: NATSConfig{
			URLs:            getEnvStringSlice("NATS_URLS", []string{"nats://localhost:4222"}),
			MaxReconnect:    getEnvInt("NATS_MAX_RECONNECT", -1),
			ReconnectWait:   getEnvDuration("NATS_RECONNECT_WAIT", 2*time.Second),
			Timeout:         getEnvDuration("NATS_TIMEOUT", 5*time.Second),
			DrainTimeout:    getEnvDuration("NATS_DRAIN_TIMEOUT", 10*time.Second),
			FlusherTimeout:  getEnvDuration("NATS_FLUSHER_TIMEOUT", 5*time.Second),
			PingInterval:    getEnvDuration("NATS_PING_INTERVAL", 2*time.Minute),
			MaxPingsOut:     getEnvInt("NATS_MAX_PINGS_OUT", 2),
			ReconnectBufSize: getEnvInt("NATS_RECONNECT_BUF_SIZE", 8*1024*1024),
		},
		HealthCheck: HealthCheckConfig{
			Timeout:       getEnvDuration("HEALTH_CHECK_TIMEOUT", 15*time.Second),
			RetryAttempts: getEnvInt("HEALTH_CHECK_RETRY_ATTEMPTS", 3),
			InitialDelay:  getEnvDuration("HEALTH_CHECK_INITIAL_DELAY", 3*time.Second),
			UserAgent:     getEnv("HEALTH_CHECK_USER_AGENT", "iot-soc-consumer/1.0"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the entire application configuration
func (c *AppConfig) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}

	if err := c.validateServer(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.validateMQTT(); err != nil {
		return fmt.Errorf("mqtt config: %w", err)
	}

	if err := c.validateHealthCheck(); err != nil {
		return fmt.Errorf("health check config: %w", err)
	}

	return nil
}

func (c *AppConfig) validateServer() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	return nil
}

func (c *AppConfig) validateMQTT() error {
	if c.MQTT.BrokerURL == "" {
		return fmt.Errorf("MQTT broker URL is required")
	}
	if c.MQTT.ClientID == "" {
		return fmt.Errorf("MQTT client ID is required")
	}
	return nil
}

func (c *AppConfig) validateHealthCheck() error {
	if c.HealthCheck.Timeout <= 0 {
		return fmt.Errorf("health check timeout must be greater than 0")
	}
	if c.HealthCheck.RetryAttempts < 0 {
		return fmt.Errorf("health check retry attempts must be >= 0")
	}
	return nil
}

// GetServerAddress returns the full server address
func (c *AppConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}