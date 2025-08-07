package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/events"
	messaginghandlers "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/mqtt/handlers"
	natshandlers "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/handlers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/presentation/http/handlers"
)

// initializeServices initializes all application services using the container
func (a *Application) initializeServices() error {
	container, err := NewContainer(a.config, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	a.services = container.GetServices()
	
	// Store cleanup function
	a.cleanup = func() error {
		return container.Cleanup()
	}

	return nil
}

// initializeHTTPServer sets up the HTTP server with all routes
func (a *Application) initializeHTTPServer() error {
	// Initialize HTTP handlers
	pingHandler := handlers.NewPingHandler(a.services.PingUseCase)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler.Ping)

	// Create HTTP server
	a.server = &http.Server{
		Addr:         a.config.GetServerAddress(),
		Handler:      mux,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	return nil
}

// startMessageConsumers starts all message consumers and subscribes to topics
func (a *Application) startMessageConsumers(ctx context.Context) error {
	// Start MQTT consumer
	a.logger.Info("Starting MQTT consumer")
	if err := a.services.MQTTConsumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}

	// Subscribe to device registration topic
	deviceRegistrationHandler := messaginghandlers.NewDeviceRegistrationHandler(a.services.DeviceRegistrationUseCase)
	deviceRegistrationTopic := "/liwaisi/iot/smart-irrigation/device/registration"
	
	a.logger.Info("Subscribing to device registration topic", "topic", deviceRegistrationTopic)
	if err := a.services.MQTTConsumer.Subscribe(ctx, deviceRegistrationTopic, deviceRegistrationHandler.HandleMessage); err != nil {
		return fmt.Errorf("failed to subscribe to device registration topic: %w", err)
	}

	// Start NATS subscriber if available
	if a.services.NATSSubscriber != nil {
		a.logger.Info("Starting NATS subscriber")
		if err := a.services.NATSSubscriber.Start(ctx); err != nil {
			a.logger.Error("Failed to start NATS subscriber", "error", err)
		} else {
			// Subscribe to device detected events
			deviceHealthHandler := natshandlers.NewDeviceHealthHandler(a.services.DeviceHealthUseCase)
			deviceDetectedSubject := events.DeviceDetectedSubject
			
			a.logger.Info("Subscribing to device detected events", "subject", deviceDetectedSubject)
			if err := a.services.NATSSubscriber.Subscribe(ctx, deviceDetectedSubject, deviceHealthHandler.HandleMessage); err != nil {
				a.logger.Error("Failed to subscribe to device detected events", "error", err)
			}
		}
	}

	return nil
}

// startHTTPServer starts the HTTP server in a goroutine
func (a *Application) startHTTPServer() error {
	go func() {
		a.logger.Info("Starting HTTP server", 
			"address", a.server.Addr)
		a.logger.Info("Ping endpoint available", 
			"url", fmt.Sprintf("http://%s/ping", a.server.Addr))

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server failed to start", "error", err)
		}
	}()

	return nil
}

// startBackgroundServices starts any background services like health monitoring
func (a *Application) startBackgroundServices(ctx context.Context) error {
	// Start health monitoring if NATS subscriber is available
	if a.services.NATSSubscriber != nil && a.services.DeviceHealthUseCase != nil {
		a.logger.Info("Starting background health monitoring")
		a.services.DeviceHealthUseCase.StartCleanup(ctx)
	}

	return nil
}

// stopMessageConsumers stops all message consumers
func (a *Application) stopMessageConsumers(ctx context.Context) error {
	a.logger.Info("Stopping message consumers")

	// Stop NATS subscriber
	if a.services.NATSSubscriber != nil {
		if err := a.services.NATSSubscriber.Stop(ctx); err != nil {
			a.logger.Error("Error stopping NATS subscriber", "error", err)
		}
	}

	// Stop MQTT consumer
	if err := a.services.MQTTConsumer.Stop(ctx); err != nil {
		a.logger.Error("Error stopping MQTT consumer", "error", err)
		return err
	}

	return nil
}

// stopHTTPServer gracefully shuts down the HTTP server
func (a *Application) stopHTTPServer(ctx context.Context) error {
	a.logger.Info("Stopping HTTP server")
	
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server forced to shutdown", "error", err)
		return err
	}

	return nil
}

// stopBackgroundServices stops all background services
func (a *Application) stopBackgroundServices() {
	a.logger.Info("Stopping background services")
	
	// Stop health check cleanup
	if a.services.DeviceHealthUseCase != nil {
		a.services.DeviceHealthUseCase.StopCleanup()
	}
}

