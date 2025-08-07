package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/app"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
)

func main() {
	// Initialize structured logger
	logger := initializeLogger()

	// Load configuration
	cfg, err := config.NewAppConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create application
	application, err := app.New(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	logger.Info("Starting IoT SOC Consumer application")
	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	waitForShutdownSignal(logger, cancel)

	// Graceful shutdown
	logger.Info("Shutting down application")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := application.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Application shut down gracefully")
}

// initializeLogger creates and configures the structured logger
func initializeLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// waitForShutdownSignal waits for SIGINT or SIGTERM and triggers shutdown
func waitForShutdownSignal(logger *slog.Logger, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Received shutdown signal", "signal", sig.String())
	cancel()
}
