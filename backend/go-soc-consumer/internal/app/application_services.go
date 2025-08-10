package app

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/events"
	messaginghandlers "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/mqtt/handlers"
	natshandlers "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging/nats/handlers"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/presentation/http/handlers"
)

// initializeServices initializes all application services using the container
func (a *Application) initializeServices() error {
	container, err := NewContainer(a.config, a.loggerFactory)
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
	a.loggerFactory.Application().LogApplicationEvent("mqtt_consumer_starting", "application")
	if err := a.services.MQTTConsumer.Start(ctx); err != nil {
		a.loggerFactory.Core().Error("mqtt_consumer_start_failed",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}

	// Subscribe to device registration topic
	deviceRegistrationHandler := messaginghandlers.NewDeviceRegistrationHandler(a.loggerFactory, a.services.DeviceRegistrationUseCase)
	deviceRegistrationTopic := "/liwaisi/iot/smart-irrigation/device/registration"

	a.loggerFactory.Application().LogApplicationEvent("mqtt_topic_subscribing", "application",
		zap.String("topic", deviceRegistrationTopic),
		zap.String("handler", "device_registration"),
	)
	if err := a.services.MQTTConsumer.Subscribe(ctx, deviceRegistrationTopic, deviceRegistrationHandler.HandleMessage); err != nil {
		a.loggerFactory.Core().Error("mqtt_topic_subscription_failed",
			zap.Error(err),
			zap.String("topic", deviceRegistrationTopic),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to subscribe to device registration topic: %w", err)
	}

	// Subscribe to temperature and humidity sensor data topic
	sensorDataHandler := messaginghandlers.NewSensorDataHandler(a.loggerFactory, a.services.SensorDataUseCase)
	sensorDataTopic := "/liwaisi/iot/smart-irrigation/sensors/temperature-and-humidity"

	a.loggerFactory.Application().LogApplicationEvent("mqtt_topic_subscribing", "application",
		zap.String("topic", sensorDataTopic),
		zap.String("handler", "sensor_data"),
	)
	if err := a.services.MQTTConsumer.Subscribe(ctx, sensorDataTopic, sensorDataHandler.HandleMessage); err != nil {
		a.loggerFactory.Core().Error("mqtt_topic_subscription_failed",
			zap.Error(err),
			zap.String("topic", sensorDataTopic),
			zap.String("component", "application"),
		)
		return fmt.Errorf("failed to subscribe to sensor data topic: %w", err)
	}

	// Start NATS subscriber if available
	if a.services.NATSSubscriber != nil {
		a.loggerFactory.Application().LogApplicationEvent("nats_subscriber_starting", "application")
		if err := a.services.NATSSubscriber.Start(ctx); err != nil {
			a.loggerFactory.Core().Error("nats_subscriber_start_failed",
				zap.Error(err),
				zap.String("component", "application"),
			)
		} else {
			// Subscribe to device detected events
			deviceHealthHandler := natshandlers.NewDeviceHealthHandler(a.services.DeviceHealthUseCase)
			deviceDetectedSubject := events.DeviceDetectedSubject

			a.loggerFactory.Application().LogApplicationEvent("nats_subject_subscribing", "application",
				zap.String("subject", deviceDetectedSubject),
				zap.String("handler", "device_health"),
			)
			if err := a.services.NATSSubscriber.Subscribe(ctx, deviceDetectedSubject, deviceHealthHandler.HandleMessage); err != nil {
				a.loggerFactory.Core().Error("nats_subject_subscription_failed",
					zap.Error(err),
					zap.String("subject", deviceDetectedSubject),
					zap.String("component", "application"),
				)
			}
		}
	}

	return nil
}

// startHTTPServer starts the HTTP server in a goroutine
func (a *Application) startHTTPServer() error {
	go func() {
		a.loggerFactory.Application().LogApplicationEvent("http_server_starting", "application",
			zap.String("address", a.server.Addr),
		)
		a.loggerFactory.Core().Info("http_server_endpoints_available",
			zap.String("ping_url", fmt.Sprintf("http://%s/ping", a.server.Addr)),
			zap.String("component", "application"),
		)

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.loggerFactory.Core().Error("http_server_start_failed",
				zap.Error(err),
				zap.String("address", a.server.Addr),
				zap.String("component", "application"),
			)
		}
	}()

	return nil
}

// startBackgroundServices starts any background services like health monitoring
func (a *Application) startBackgroundServices(ctx context.Context) error {
	// Start health monitoring if NATS subscriber is available
	if a.services.NATSSubscriber != nil && a.services.DeviceHealthUseCase != nil {
		a.loggerFactory.Application().LogApplicationEvent("background_health_monitoring_starting", "application")
	}

	return nil
}

// stopMessageConsumers stops all message consumers
func (a *Application) stopMessageConsumers(ctx context.Context) error {
	a.loggerFactory.Application().LogApplicationEvent("message_consumers_stopping", "application")

	// Stop NATS subscriber
	if a.services.NATSSubscriber != nil {
		if err := a.services.NATSSubscriber.Stop(ctx); err != nil {
			a.loggerFactory.Core().Error("nats_subscriber_stop_error",
				zap.Error(err),
				zap.String("component", "application"),
			)
		}
	}

	// Stop MQTT consumer
	if err := a.services.MQTTConsumer.Stop(ctx); err != nil {
		a.loggerFactory.Core().Error("mqtt_consumer_stop_error",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return err
	}

	return nil
}

// stopHTTPServer gracefully shuts down the HTTP server
func (a *Application) stopHTTPServer(ctx context.Context) error {
	a.loggerFactory.Application().LogApplicationEvent("http_server_stopping", "application")

	if err := a.server.Shutdown(ctx); err != nil {
		a.loggerFactory.Core().Error("http_server_shutdown_error",
			zap.Error(err),
			zap.String("component", "application"),
		)
		return err
	}

	return nil
}
