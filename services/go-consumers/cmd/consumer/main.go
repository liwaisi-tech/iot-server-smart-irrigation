package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/infrastructure/config"
	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/infrastructure/database"
	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/pkg/logger"
)

func main() {
	// Initialize temporary logger for startup
	tempLogger, err := logger.NewLoggerWithDefaults()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize temporary logger: %v", err))
	}
	defer tempLogger.Sync()

	tempLogger.Info("Starting Go Consumer Service initialization")

	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		tempLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize configured logger
	loggerConfig := logger.LoggerConfig{
		Level:       cfg.Logger.Level,
		Environment: cfg.Logger.Environment,
		OutputPaths: cfg.Logger.OutputPaths,
		Encoding:    cfg.Logger.Encoding,
	}

	// If no logger config in file, use defaults
	if cfg.Logger.Level == "" {
		loggerConfig = logger.DefaultConfig()
		tempLogger.Warn("No logger configuration found, using defaults")
	}

	appLogger, err := logger.NewLogger(loggerConfig)
	if err != nil {
		tempLogger.Fatal("Failed to initialize application logger", zap.Error(err))
	}
	defer appLogger.Sync()

	appLogger.Info("Application logger initialized successfully",
		zap.String("level", loggerConfig.Level),
		zap.String("environment", loggerConfig.Environment),
		zap.String("encoding", loggerConfig.Encoding))

	// Convert config database config to database.DatabaseConfig
	dbConfig := database.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	// Initialize database connection
	appLogger.Info("Initializing database connection",
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("database", dbConfig.DBName),
		zap.String("ssl_mode", dbConfig.SSLMode))

	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			appLogger.Error("Error closing database connection", zap.Error(err))
		} else {
			appLogger.Info("Database connection closed successfully")
		}
	}()

	appLogger.Info("Database connection established successfully")

	// Run database migrations using GORM AutoMigrate
	appLogger.Info("Running database migrations...")
	if err := dbConn.AutoMigrate(); err != nil {
		appLogger.Fatal("Failed to run database migrations", zap.Error(err))
	}
	appLogger.Info("Database migrations completed successfully")

	// Perform database health check
	appLogger.Info("Performing database health check...")
	if err := dbConn.HealthCheck(); err != nil {
		appLogger.Fatal("Database health check failed", zap.Error(err))
	}
	appLogger.Info("Database health check passed")

	// TODO: Initialize NATS connection and message handlers
	// TODO: Start consuming messages from NATS JetStream
	// TODO: Initialize device registration handlers

	appLogger.Info("Go Consumer Service is running",
		zap.String("database_host", fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)),
		zap.String("database_name", dbConfig.DBName),
		zap.String("database_user", dbConfig.User))

	appLogger.Info("Service capabilities",
		zap.String("primary_key", "mac_address"),
		zap.Bool("pagination_support", true),
		zap.String("pagination_type", "limit_offset"))
	
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	appLogger.Info("Service started successfully, waiting for shutdown signal...")

	// Block until a signal is received
	<-c
	appLogger.Info("Shutdown signal received, beginning graceful shutdown...")
	
	appLogger.Info("Go Consumer Service shutdown completed")
}
