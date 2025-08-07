package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	devicehealth "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_health"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
)

// Application represents the complete application with all its dependencies
type Application struct {
	config   *config.AppConfig
	logger   *slog.Logger
	services *Services
	server   *http.Server
	cleanup  func() error
}

// Services holds all the business logic services
type Services struct {
	DeviceRepository          ports.DeviceRepository
	DeviceRegistrationUseCase deviceregistration.DeviceRegistrationUseCase
	DeviceHealthUseCase       devicehealth.DeviceHealthUseCase
	PingUseCase               ping.PingUseCase
	MQTTConsumer              ports.MessageConsumer
	NATSPublisher             ports.EventPublisher
	NATSSubscriber            ports.EventSubscriber
	HealthChecker             ports.DeviceHealthChecker
}

// New creates a new application instance
func New(cfg *config.AppConfig, logger *slog.Logger) (*Application, error) {
	app := &Application{
		config: cfg,
		logger: logger,
	}

	// Initialize all dependencies
	if err := app.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize HTTP server
	if err := app.initializeHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	return app, nil
}

// Start starts all application services
func (a *Application) Start(ctx context.Context) error {
	a.logger.Info("Starting application services")

	// Start message consumers
	if err := a.startMessageConsumers(ctx); err != nil {
		return fmt.Errorf("failed to start message consumers: %w", err)
	}

	// Start HTTP server
	if err := a.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start background services
	if err := a.startBackgroundServices(ctx); err != nil {
		return fmt.Errorf("failed to start background services: %w", err)
	}

	a.logger.Info("All application services started successfully")
	return nil
}

// Stop gracefully shuts down all application services
func (a *Application) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application services")

	// Stop background services first
	a.stopBackgroundServices()

	// Stop message consumers
	if err := a.stopMessageConsumers(ctx); err != nil {
		a.logger.Error("Error stopping message consumers", "error", err)
	}

	// Stop HTTP server
	if err := a.stopHTTPServer(ctx); err != nil {
		a.logger.Error("Error stopping HTTP server", "error", err)
	}

	// Clean up resources
	if a.cleanup != nil {
		if err := a.cleanup(); err != nil {
			a.logger.Error("Error during cleanup", "error", err)
		}
	}

	a.logger.Info("All application services stopped")
	return nil
}

// Health returns the health status of all services
func (a *Application) Health(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})
	
	// Add individual service health checks here
	health["status"] = "ok"
	health["timestamp"] = time.Now()
	
	return health
}