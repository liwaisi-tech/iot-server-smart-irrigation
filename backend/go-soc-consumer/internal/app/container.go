package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	infrahttp "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/http"
	messagingmqtt "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/mqtt"
	messagingnats "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres"
	devicehealth "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_health"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// Container holds all the application dependencies
type Container struct {
	config        *config.AppConfig
	loggerFactory logger.LoggerFactory
	services      *Services
	cleanup       []func() error
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.AppConfig, loggerFactory logger.LoggerFactory) (*Container, error) {
	container := &Container{
		config:        cfg,
		loggerFactory: loggerFactory,
		cleanup:       make([]func() error, 0),
	}

	services, err := container.buildServices()
	if err != nil {
		loggerFactory.Core().Error("container_services_build_failed",
			zap.Error(err),
			zap.String("component", "container"),
		)
		return nil, fmt.Errorf("failed to build services: %w", err)
	}

	container.services = services
	loggerFactory.Application().LogApplicationEvent("container_initialized", "container")
	return container, nil
}

// GetServices returns the built services
func (c *Container) GetServices() *Services {
	return c.services
}

// Cleanup runs all cleanup functions
func (c *Container) Cleanup() error {
	c.loggerFactory.Application().LogApplicationEvent("container_cleanup_starting", "container")

	for i := len(c.cleanup) - 1; i >= 0; i-- {
		if err := c.cleanup[i](); err != nil {
			c.loggerFactory.Core().Error("container_cleanup_error",
				zap.Error(err),
				zap.Int("cleanup_step", i),
				zap.String("component", "container"),
			)
			return err
		}
	}

	c.loggerFactory.Application().LogApplicationEvent("container_cleanup_completed", "container")
	return nil
}

// buildServices constructs all the application services with proper dependency injection
func (c *Container) buildServices() (*Services, error) {
	services := &Services{}

	// Build infrastructure dependencies first
	if err := c.buildInfrastructure(services); err != nil {
		return nil, fmt.Errorf("failed to build infrastructure: %w", err)
	}

	// Build use cases
	if err := c.buildUseCases(services); err != nil {
		return nil, fmt.Errorf("failed to build use cases: %w", err)
	}

	return services, nil
}

// buildInfrastructure builds all infrastructure-layer dependencies
func (c *Container) buildInfrastructure(services *Services) error {
	// Build database repository
	if err := c.buildRepository(services); err != nil {
		return fmt.Errorf("failed to build repository: %w", err)
	}

	// Build messaging infrastructure
	if err := c.buildMessaging(services); err != nil {
		return fmt.Errorf("failed to build messaging: %w", err)
	}

	// Build external dependencies
	if err := c.buildExternalDependencies(services); err != nil {
		return fmt.Errorf("failed to build external dependencies: %w", err)
	}

	return nil
}

// buildRepository builds the device repository
func (c *Container) buildRepository(services *Services) error {
	c.loggerFactory.Application().LogApplicationEvent("database_repository_initializing", "container")

	// Initialize GORM database with logger factory
	gormDB, err := database.NewGormPostgresDB(&c.config.Database, c.loggerFactory)
	if err != nil {
		c.loggerFactory.Core().Error("database_initialization_failed",
			zap.Error(err),
			zap.String("host", c.config.Database.Host),
			zap.Int("port", c.config.Database.Port),
			zap.String("component", "container"),
		)
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	c.loggerFactory.Application().LogApplicationEvent("database_migrations_running", "container")
	if err := gormDB.AutoMigrate(); err != nil {
		c.loggerFactory.Core().Error("database_migrations_failed",
			zap.Error(err),
			zap.String("component", "container"),
		)
		gormDB.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repository with logger factory
	services.DeviceRepository = postgres.NewDeviceRepository(gormDB, c.loggerFactory)
	services.SensorTemperatureHumidityRepository = postgres.NewSensorTemperatureHumidityRepository(gormDB, c.loggerFactory)

	// Register cleanup
	c.cleanup = append(c.cleanup, func() error {
		c.loggerFactory.Application().LogApplicationEvent("database_connection_closing", "container")
		return gormDB.Close()
	})

	c.loggerFactory.Application().LogApplicationEvent("database_repository_initialized", "container")
	return nil
}

// buildMessaging builds messaging infrastructure (MQTT and NATS)
func (c *Container) buildMessaging(services *Services) error {
	// Build MQTT Consumer
	if err := c.buildMQTTConsumer(services); err != nil {
		return fmt.Errorf("failed to build MQTT consumer: %w", err)
	}

	// Build NATS components (optional - warn if they fail)
	c.buildNATSComponents(services)

	return nil
}

// buildMQTTConsumer builds the MQTT consumer
func (c *Container) buildMQTTConsumer(services *Services) error {
	c.loggerFactory.Application().LogApplicationEvent("mqtt_consumer_initializing", "container",
		zap.String("broker_url", c.config.MQTT.BrokerURL),
		zap.String("client_id", c.config.MQTT.ClientID),
	)

	mqttConfig := messagingmqtt.MQTTConsumerConfig{
		BrokerURL:            c.config.MQTT.BrokerURL,
		ClientID:             c.config.MQTT.ClientID,
		Username:             c.config.MQTT.Username,
		Password:             c.config.MQTT.Password,
		CleanSession:         c.config.MQTT.CleanSession,
		AutoReconnect:        c.config.MQTT.AutoReconnect,
		ConnectTimeout:       c.config.MQTT.ConnectTimeout,
		KeepAlive:            c.config.MQTT.KeepAlive,
		MaxReconnectInterval: c.config.MQTT.MaxReconnectInterval,
	}

	services.MQTTConsumer = messagingmqtt.NewMQTTConsumer(mqttConfig, c.loggerFactory)
	c.loggerFactory.Application().LogApplicationEvent("mqtt_consumer_initialized", "container")
	return nil
}

// buildNATSComponents builds NATS publisher and subscriber (optional)
func (c *Container) buildNATSComponents(services *Services) {
	// Use existing NATS config with defaults
	natsConfig := messagingnats.DefaultNATSConfig()

	// Override with app config if provided
	if len(c.config.NATS.URLs) > 0 {
		natsConfig.URL = c.config.NATS.URLs[0] // Use first URL for now
	}
	// Configure other NATS settings
	natsConfig.MaxReconnectAttempts = c.config.NATS.MaxReconnect
	natsConfig.ReconnectWait = c.config.NATS.ReconnectWait
	natsConfig.ConnectTimeout = c.config.NATS.Timeout
	natsConfig.PingInterval = c.config.NATS.PingInterval
	natsConfig.MaxPingsOutstanding = c.config.NATS.MaxPingsOut

	// Build NATS Publisher
	if natsPublisher, err := messagingnats.NewNATSPublisher(natsConfig, c.loggerFactory); err != nil {
		c.loggerFactory.Core().Warn("nats_publisher_initialization_failed",
			zap.Error(err),
			zap.String("url", natsConfig.URL),
			zap.String("component", "container"),
		)
		services.NATSPublisher = nil
	} else {
		services.NATSPublisher = natsPublisher
		c.cleanup = append(c.cleanup, func() error {
			return natsPublisher.Close(context.TODO())
		})
		c.loggerFactory.Application().LogApplicationEvent("nats_publisher_initialized", "container",
			zap.String("url", natsConfig.URL),
		)
	}

	// Build NATS Subscriber
	if natsSubscriber, err := messagingnats.NewNATSSubscriber(natsConfig, c.loggerFactory); err != nil {
		c.loggerFactory.Core().Warn("nats_subscriber_initialization_failed",
			zap.Error(err),
			zap.String("url", natsConfig.URL),
			zap.String("component", "container"),
		)
		services.NATSSubscriber = nil
	} else {
		services.NATSSubscriber = natsSubscriber
		c.loggerFactory.Application().LogApplicationEvent("nats_subscriber_initialized", "container",
			zap.String("url", natsConfig.URL),
		)
	}
}

// buildExternalDependencies builds external API clients
func (c *Container) buildExternalDependencies(services *Services) error {
	c.loggerFactory.Application().LogApplicationEvent("external_dependencies_initializing", "container")

	// Build health checker
	healthConfig := &infrahttp.HealthClientConfig{
		Timeout:       c.config.HealthCheck.Timeout,
		RetryAttempts: c.config.HealthCheck.RetryAttempts,
		InitialDelay:  c.config.HealthCheck.InitialDelay,
		UserAgent:     c.config.HealthCheck.UserAgent,
	}

	services.HealthChecker = infrahttp.NewHealthClient(healthConfig, c.loggerFactory)
	c.loggerFactory.Application().LogApplicationEvent("health_checker_initialized", "container",
		zap.Duration("timeout", c.config.HealthCheck.Timeout),
		zap.Int("retry_attempts", c.config.HealthCheck.RetryAttempts),
	)

	return nil
}

// buildUseCases builds all use case implementations
func (c *Container) buildUseCases(services *Services) error {
	c.loggerFactory.Application().LogApplicationEvent("use_cases_initializing", "container")

	// Build Ping Use Case
	services.PingUseCase = ping.NewUseCase()

	// Build Device Registration Use Case
	services.DeviceRegistrationUseCase = deviceregistration.NewDeviceRegistrationUseCase(
		services.DeviceRepository,
		services.NATSPublisher,
		c.loggerFactory,
	)

	// Build Device Health Use Case
	healthCheckConfig := devicehealth.DefaultHealthCheckConfig()
	services.DeviceHealthUseCase = devicehealth.NewDeviceHealthUseCase(
		services.DeviceRepository,
		services.HealthChecker,
		healthCheckConfig,
		c.loggerFactory,
	)

	c.loggerFactory.Application().LogApplicationEvent("use_cases_initialized", "container")
	return nil
}
