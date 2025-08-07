package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	infrahttp "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/http"
	messagingmqtt "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/mqtt"
	messagingnats "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres"
	devicehealth "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_health"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
)

// Container holds all the application dependencies
type Container struct {
	config   *config.AppConfig
	logger   *slog.Logger
	services *Services
	cleanup  []func() error
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.AppConfig, logger *slog.Logger) (*Container, error) {
	container := &Container{
		config:  cfg,
		logger:  logger,
		cleanup: make([]func() error, 0),
	}

	services, err := container.buildServices()
	if err != nil {
		return nil, fmt.Errorf("failed to build services: %w", err)
	}

	container.services = services
	return container, nil
}

// GetServices returns the built services
func (c *Container) GetServices() *Services {
	return c.services
}

// Cleanup runs all cleanup functions
func (c *Container) Cleanup() error {
	c.logger.Info("Running container cleanup")
	
	for i := len(c.cleanup) - 1; i >= 0; i-- {
		if err := c.cleanup[i](); err != nil {
			c.logger.Error("Error during cleanup", "error", err)
			return err
		}
	}
	
	c.logger.Info("Container cleanup completed")
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
	c.logger.Info("Initializing database repository")

	// Initialize GORM database
	gormDB, err := database.NewGormPostgresDB(&c.config.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	c.logger.Info("Running database migrations")
	if err := gormDB.AutoMigrate(); err != nil {
		gormDB.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repository
	services.DeviceRepository = postgres.NewDeviceRepository(gormDB)

	// Register cleanup
	c.cleanup = append(c.cleanup, func() error {
		c.logger.Info("Closing database connection")
		return gormDB.Close()
	})

	c.logger.Info("Database repository initialized successfully")
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
	c.logger.Info("Initializing MQTT consumer")

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

	services.MQTTConsumer = messagingmqtt.NewMQTTConsumer(mqttConfig)
	c.logger.Info("MQTT consumer initialized successfully")
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

	// Build NATS Publisher
	if natsPublisher, err := messagingnats.NewNATSPublisher(natsConfig, c.logger); err != nil {
		c.logger.Warn("Failed to initialize NATS publisher", "error", err)
		services.NATSPublisher = nil
	} else {
		services.NATSPublisher = natsPublisher
		c.cleanup = append(c.cleanup, func() error {
			return natsPublisher.Close(context.TODO())
		})
		c.logger.Info("NATS publisher initialized successfully")
	}

	// Build NATS Subscriber
	if natsSubscriber, err := messagingnats.NewNATSSubscriber(natsConfig, c.logger); err != nil {
		c.logger.Warn("Failed to initialize NATS subscriber", "error", err)
		services.NATSSubscriber = nil
	} else {
		services.NATSSubscriber = natsSubscriber
		c.logger.Info("NATS subscriber initialized successfully")
	}
}

// buildExternalDependencies builds external API clients
func (c *Container) buildExternalDependencies(services *Services) error {
	c.logger.Info("Initializing external dependencies")

	// Build health checker
	healthConfig := &infrahttp.HealthClientConfig{
		Timeout:       c.config.HealthCheck.Timeout,
		RetryAttempts: c.config.HealthCheck.RetryAttempts,
		InitialDelay:  c.config.HealthCheck.InitialDelay,
		UserAgent:     c.config.HealthCheck.UserAgent,
	}
	
	services.HealthChecker = infrahttp.NewHealthClient(healthConfig, c.logger)
	c.logger.Info("Health checker initialized successfully")

	return nil
}

// buildUseCases builds all use case implementations
func (c *Container) buildUseCases(services *Services) error {
	c.logger.Info("Initializing use cases")

	// Build Ping Use Case
	services.PingUseCase = ping.NewUseCase()

	// Build Device Registration Use Case
	services.DeviceRegistrationUseCase = deviceregistration.NewDeviceRegistrationUseCase(
		services.DeviceRepository,
		services.NATSPublisher,
	)

	// Build Device Health Use Case
	healthCheckConfig := devicehealth.DefaultHealthCheckConfig()
	services.DeviceHealthUseCase = devicehealth.NewDeviceHealthUseCase(
		services.DeviceRepository,
		services.HealthChecker,
		healthCheckConfig,
		c.logger,
	)

	c.logger.Info("Use cases initialized successfully")
	return nil
}