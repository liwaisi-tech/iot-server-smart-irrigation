package app

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	eventports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/events"
	repositoryports "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports/repositories"
	devicehealth "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_health"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
	sensordata "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/sensor_data"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// Application represents the complete application with all its dependencies
type Application struct {
	config        *config.AppConfig
	loggerFactory logger.LoggerFactory
	services      *Services
	server        *http.Server
	cleanup       func() error
}

// Services holds all the business logic services
type Services struct {
	DeviceRepository                    repositoryports.DeviceRepository
	SensorTemperatureHumidityRepository repositoryports.SensorTemperatureHumidityRepository
	DeviceRegistrationUseCase           deviceregistration.DeviceRegistrationUseCase
	DeviceHealthUseCase                 devicehealth.DeviceHealthUseCase
	PingUseCase                         ping.PingUseCase
	SensorDataUseCase                   sensordata.SensorDataUseCase
	MQTTConsumer                        eventports.MessageConsumer
	NATSPublisher                       eventports.EventPublisher
	NATSSubscriber                      eventports.EventSubscriber
	HealthChecker                       ports.DeviceHealthChecker
}

// New creates a new application instance
func New(cfg *config.AppConfig, loggerFactory logger.LoggerFactory) (*Application, error) {
	app := &Application{
		config:        cfg,
		loggerFactory: loggerFactory,
	}

	// Initialize all dependencies
	if err := app.initializeServices(); err != nil {
		loggerFactory.Core().Error("services_initialization_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize HTTP server
	if err := app.initializeHTTPServer(); err != nil {
		loggerFactory.Core().Error("http_server_initialization_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	loggerFactory.Application().LogApplicationEvent("application_initialized", "application")
	return app, nil
}

// Start starts all application services
func (a *Application) Start(ctx context.Context) error {
	a.loggerFactory.Application().LogApplicationEvent("application_services_starting", "application")

	// Start message consumers
	if err := a.startMessageConsumers(ctx); err != nil {
		a.loggerFactory.Core().Error("message_consumers_start_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to start message consumers: %w", err)
	}

	// Start HTTP server
	if err := a.startHTTPServer(); err != nil {
		a.loggerFactory.Core().Error("http_server_start_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start background services
	if err := a.startBackgroundServices(ctx); err != nil {
		a.loggerFactory.Core().Error("background_services_start_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to start background services: %w", err)
	}

	a.loggerFactory.Application().LogApplicationEvent("application_services_started", "application")
	return nil
}

// Stop gracefully shuts down all application services
func (a *Application) Stop(ctx context.Context) error {
	a.loggerFactory.Application().LogApplicationEvent("application_services_stopping", "application")

	// Stop message consumers
	if err := a.stopMessageConsumers(ctx); err != nil {
		a.loggerFactory.Core().Error("message_consumers_stop_error",
			zap.Error(err),
			zap.String("component", "application"),
		)
	}

	// Stop HTTP server
	if err := a.stopHTTPServer(ctx); err != nil {
		a.loggerFactory.Core().Error("http_server_stop_error",
			zap.Error(err),
			zap.String("component", "application"),
		)
	}

	// Clean up resources
	if a.cleanup != nil {
		if err := a.cleanup(); err != nil {
			a.loggerFactory.Core().Error("cleanup_error",
				zap.Error(err),
				zap.String("component", "application"),
			)
		}
	}

	a.loggerFactory.Application().LogApplicationEvent("application_services_stopped", "application")
	return nil
}
