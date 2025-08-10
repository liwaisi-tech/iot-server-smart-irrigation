package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/app"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

func main() {
	// Load configuration first for logger initialization
	cfg, err := config.NewAppConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger factory with config
	loggerFactory, err := initializeLoggerFactoryWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger factory: %v", err)
	}
	defer func() {
		if syncErr := loggerFactory.Core().Sync(); syncErr != nil {
			// Don't log sync errors for stdout/stderr
			if !strings.Contains(syncErr.Error(), "sync /dev/stdout") && !strings.Contains(syncErr.Error(), "sync /dev/stderr") {
				log.Printf("Error syncing logger: %v", syncErr)
			}
		}
	}()

	// Configuration already loaded above

	loggerFactory.Application().LogApplicationEvent("configuration_loaded", "main",
		zap.String("mqtt_broker_url", cfg.MQTT.BrokerURL),
		zap.String("db_host", cfg.Database.Host),
		zap.Int("db_port", cfg.Database.Port),
		zap.String("log_level", cfg.Logging.Level),
		zap.String("log_format", cfg.Logging.Format),
	)

	// Create application
	application, err := app.New(cfg, loggerFactory)
	if err != nil {
		loggerFactory.Core().Error("application_creation_failed",
			zap.Error(err),
			zap.String("component", "main"),
		)
		log.Fatalf("Failed to create application: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	loggerFactory.Application().LogApplicationEvent("application_starting", "main")
	start := time.Now()
	if err := application.Start(ctx); err != nil {
		loggerFactory.Core().Error("application_start_failed",
			zap.Error(err),
			zap.Duration("startup_duration", time.Since(start)),
			zap.String("component", "main"),
		)
		log.Fatalf("Failed to start application: %v", err)
	}

	loggerFactory.Application().LogApplicationEvent("application_started", "main",
		zap.Duration("startup_duration", time.Since(start)),
	)

	// Wait for shutdown signal
	waitForShutdownSignal(loggerFactory, cancel)

	// Graceful shutdown
	loggerFactory.Application().LogApplicationEvent("application_shutting_down", "main")
	shutdownStart := time.Now()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := application.Stop(shutdownCtx); err != nil {
		loggerFactory.Core().Error("application_shutdown_error",
			zap.Error(err),
			zap.Duration("shutdown_duration", time.Since(shutdownStart)),
			zap.String("component", "main"),
		)
		os.Exit(1)
	}

	loggerFactory.Application().LogApplicationEvent("application_shutdown_complete", "main",
		zap.Duration("shutdown_duration", time.Since(shutdownStart)),
	)
}

// initializeLoggerFactoryWithConfig creates and configures the logger factory using app config
func initializeLoggerFactoryWithConfig(cfg *config.AppConfig) (logger.LoggerFactory, error) {
	// Get environment configuration with fallback to config
	environment := getEnv("ENVIRONMENT", "production")

	// Create logger configuration from app config
	loggerConfig := logger.LoggerConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Environment: environment,
	}

	// Create and return the logger factory
	return logger.NewLoggerFactory(loggerConfig)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// waitForShutdownSignal waits for SIGINT or SIGTERM and triggers shutdown
func waitForShutdownSignal(loggerFactory logger.LoggerFactory, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	loggerFactory.Application().LogApplicationEvent("shutdown_signal_received", "main",
		zap.String("signal", sig.String()),
	)
	cancel()
}
