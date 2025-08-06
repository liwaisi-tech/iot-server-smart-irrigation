package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/messaging"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/presentation/http/handlers"
	deviceregistration "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/device_registration"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	// TODO: Configuration will be improved later with proper config management
	port := "8080"
	host := "0.0.0.0"
	
	// MQTT Configuration (using environment variables with defaults)
	mqttConfig := messaging.MQTTConsumerConfig{
		BrokerURL:            getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		ClientID:             getEnv("MQTT_CLIENT_ID", "iot-go-soc-consumer"),
		Username:             getEnv("MQTT_USERNAME", ""),
		Password:             getEnv("MQTT_PASSWORD", ""),
		CleanSession:         true,
		AutoReconnect:        true,
		ConnectTimeout:       30 * time.Second,
		KeepAlive:            60 * time.Second,
		MaxReconnectInterval: 10 * time.Minute,
	}
	
	// Initialize repository based on configuration
	deviceRepo, dbCleanup, err := initializeRepository(logger)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer func() {
		if dbCleanup != nil {
			dbCleanup()
		}
	}()

	// Initialize MQTT consumer
	mqttConsumer := messaging.NewMQTTConsumer(mqttConfig)
	
	// Initialize use cases
	pingUseCase := ping.NewUseCase()
	deviceRegistrationUseCase := deviceregistration.NewUseCase(deviceRepo)
	
	// Initialize message handler
	messageHandler := messaging.NewDeviceRegistrationHandler(deviceRegistrationUseCase)
	
	// Initialize handlers
	pingHandler := handlers.NewPingHandler(pingUseCase)
	
	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start MQTT consumer
	logger.Info("Starting MQTT consumer")
	if err := mqttConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start MQTT consumer: %v", err)
	}
	
	// Subscribe to device registration topic
	deviceRegistrationTopic := "/liwaisi/iot/smart-irrigation/device/registration"
	logger.Info("Subscribing to device registration topic", slog.String("topic", deviceRegistrationTopic))
	if err := mqttConsumer.Subscribe(ctx, deviceRegistrationTopic, messageHandler.HandleMessage); err != nil {
		log.Fatalf("Failed to subscribe to device registration topic: %v", err)
	}
	
	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler.Ping)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", 
			slog.String("host", host), 
			slog.String("port", port))
		logger.Info("Ping endpoint available", 
			slog.String("url", fmt.Sprintf("http://%s:%s/ping", host, port)))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed to start", slog.String("error", err.Error()))
			cancel() // Cancel context to trigger shutdown
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-quit:
		logger.Info("Received shutdown signal")
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}
	
	logger.Info("Shutting down services...")
	
	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Shutdown MQTT consumer first
	if err := mqttConsumer.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping MQTT consumer", slog.String("error", err.Error()))
	}
	
	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", slog.String("error", err.Error()))
	}
	
	logger.Info("All services stopped gracefully")
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initializeRepository initializes the device repository based on environment configuration
func initializeRepository(logger *slog.Logger) (ports.DeviceRepository, func(), error) {
	// Check if PostgreSQL should be used
	dbType := strings.ToLower(getEnv("DB_TYPE", "memory"))
	
	switch dbType {
	case "postgres", "postgresql":
		logger.Info("Initializing PostgreSQL repository")
		
		// Create database configuration
		dbConfig := config.NewDatabaseConfig()
		if err := dbConfig.Validate(); err != nil {
			return nil, nil, fmt.Errorf("invalid database configuration: %w", err)
		}
		
		logger.Info("Connecting to PostgreSQL database", 
			slog.String("host", dbConfig.Host),
			slog.Int("port", dbConfig.Port),
			slog.String("database", dbConfig.Name))
		
		// Create database connection
		db, err := database.NewPostgresDB(dbConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create PostgreSQL connection: %w", err)
		}
		
		// Run database migrations
		migrationsPath := getEnv("MIGRATIONS_PATH", "./migrations")
		logger.Info("Running database migrations", slog.String("path", migrationsPath))
		
		if err := db.RunMigrations(migrationsPath); err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("failed to run database migrations: %w", err)
		}
		
		// Create PostgreSQL repository
		repo := persistence.NewPostgresDeviceRepository(db)
		
		// Return repository and cleanup function
		cleanup := func() {
			logger.Info("Closing database connection")
			if err := db.Close(); err != nil {
				logger.Error("Error closing database connection", slog.String("error", err.Error()))
			}
		}
		
		logger.Info("PostgreSQL repository initialized successfully")
		return repo, cleanup, nil
		
	case "memory":
		logger.Info("Initializing in-memory repository")
		repo := persistence.NewMemoryDeviceRepository()
		logger.Info("In-memory repository initialized successfully")
		return repo, nil, nil
		
	default:
		return nil, nil, fmt.Errorf("unsupported database type: %s. Supported types: memory, postgres", dbType)
	}
}