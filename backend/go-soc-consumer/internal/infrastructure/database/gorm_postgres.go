package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/models"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/config"
	pkglogger "github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// GormPostgresDB wraps the GORM database connection and provides additional functionality
type GormPostgresDB struct {
	db     *gorm.DB
	config *config.DatabaseConfig
	logger pkglogger.InfrastructureLogger
}

var (
	instance  *GormPostgresDB
	once      sync.Once
	initError error
	initMutex sync.Mutex
)

// NewGormPostgresDBWithoutConfig creates a new GORM PostgreSQL database connection without a config for testing purposes
func NewGormPostgresDBWithoutConfig(db *gorm.DB, infraLogger pkglogger.InfrastructureLogger) (*GormPostgresDB, error) {
	if infraLogger == nil {
		return nil, fmt.Errorf("infrastructure logger cannot be nil")
	}

	return &GormPostgresDB{
		db:     db,
		config: nil,
		logger: infraLogger,
	}, nil
}

// initDatabase handles the actual database initialization
func initDatabase(cfg *config.DatabaseConfig, infraLogger pkglogger.InfrastructureLogger) (*GormPostgresDB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Use plural table names (devices, not device)
			NoLowerCase:   false, // Convert field names to lowercase
		},
	}

	// Open GORM connection
	start := time.Now()
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), gormConfig)
	connectionDuration := time.Since(start)

	if err != nil {
		infraLogger.LogExternalAPICall("postgres", "connection", 0, connectionDuration, err)
		return nil, fmt.Errorf("failed to open GORM database connection: %w", err)
	}

	infraLogger.LogExternalAPICall("postgres", "connection", 200, connectionDuration, nil)

	// Get the underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	gormDB := &GormPostgresDB{
		db:     db,
		config: cfg,
		logger: infraLogger,
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := gormDB.Ping(ctx); err != nil {
		infraLogger.LogExternalAPICall("postgres", "initial_ping", 0, time.Since(start), err)
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping GORM database: %w", err)
	}

	return gormDB, nil
}

// NewGormPostgresDB creates a new GORM PostgreSQL database connection using singleton pattern
func NewGormPostgresDB(cfg *config.DatabaseConfig, loggerFactory pkglogger.LoggerFactory) (*GormPostgresDB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database configuration cannot be nil")
	}
	if loggerFactory == nil {
		return nil, fmt.Errorf("logger factory cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	once.Do(func() {
		initMutex.Lock()
		defer initMutex.Unlock()

		// Check again in case another goroutine initialized the instance
		// while we were waiting for the lock
		if instance != nil {
			return
		}

		// Initialize the database with infrastructure logger
		var err error
		infraLogger := loggerFactory.Infrastructure()
		instance, err = initDatabase(cfg, infraLogger)
		if err != nil {
			initError = fmt.Errorf("failed to initialize database: %w", err)
		}
	})

	// If there was an error during initialization, return it
	if initError != nil {
		return nil, initError
	}

	// If we get here, the instance should be initialized
	if instance == nil {
		return nil, fmt.Errorf("database instance is nil after initialization")
	}

	return instance, nil
}

// GetDB returns the underlying *gorm.DB instance
func (g *GormPostgresDB) GetDB() *gorm.DB {
	return g.db
}

// Ping tests the database connection
func (g *GormPostgresDB) Ping(ctx context.Context) error {
	start := time.Now()
	sqlDB, err := g.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	err = sqlDB.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		g.logger.LogExternalAPICall("postgres", "ping", 0, duration, err)
		return fmt.Errorf("ping failed: %w", err)
	}

	g.logger.LogExternalAPICall("postgres", "ping", 200, duration, nil)
	return nil
}

// Close closes the database connection
func (g *GormPostgresDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// AutoMigrate runs GORM auto-migrations for all registered models
func (g *GormPostgresDB) AutoMigrate() error {
	start := time.Now()
	// Simple GORM AutoMigrate
	err := g.db.AutoMigrate(
		&models.DeviceModel{},
		&models.SensorTemperatureHumidityModel{},
	)
	duration := time.Since(start)

	if err != nil {
		g.logger.LogDatabaseOperation("auto_migrate", "devices", duration, 0, err)
		return fmt.Errorf("auto migration failed: %w", err)
	}

	g.logger.LogDatabaseOperation("auto_migrate", "devices", duration, 1, nil)
	return nil
}

// HealthCheck performs a basic health check on the database
func (g *GormPostgresDB) HealthCheck(ctx context.Context) error {
	start := time.Now()
	// Simple query to test database connectivity and basic functionality
	var result int
	err := g.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	duration := time.Since(start)

	if err != nil {
		g.logger.LogDatabaseOperation("health_check", "postgres", duration, 0, err)
		return fmt.Errorf("health check failed: %w", err)
	}

	if result != 1 {
		g.logger.LogDatabaseOperation("health_check", "postgres", duration, 0, fmt.Errorf("unexpected result %d", result))
		return fmt.Errorf("health check failed: unexpected result %d", result)
	}

	g.logger.LogDatabaseOperation("health_check", "postgres", duration, 1, nil)
	return nil
}

// GetStats returns database connection pool statistics
func (g *GormPostgresDB) GetStats() (interface{}, error) {
	start := time.Now()
	sqlDB, err := g.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	duration := time.Since(start)

	// Log connection pool statistics gathering
	g.logger.LogDatabaseOperation("get_stats", "connection_pool", duration, int64(stats.OpenConnections), nil)
	return stats, nil
}

// BeginTx starts a database transaction with GORM
func (g *GormPostgresDB) BeginTx(ctx context.Context) *gorm.DB {
	return g.db.WithContext(ctx).Begin()
}

// Transaction executes a function within a database transaction
func (g *GormPostgresDB) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return g.db.WithContext(ctx).Transaction(fn)
}

// GetConfig returns the database configuration
func (g *GormPostgresDB) GetConfig() *config.DatabaseConfig {
	return g.config
}
