package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/infrastructure/database/models"
)

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connection manages the database connection
type Connection struct {
	DB *gorm.DB
}

// NewConnection creates a new database connection with the given configuration
func NewPostgresConnection(config DatabaseConfig) (*Connection, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Connection{DB: db}, nil
}

// AutoMigrate runs GORM AutoMigrate for all models
// This creates tables and indexes automatically based on struct tags
func (c *Connection) AutoMigrate() error {
	log.Println("Starting database migration...")

	// Run AutoMigrate for Device model
	// This will create the devices table with proper indexes:
	// - Primary key on mac_address
	// - Index on ip_address (idx_device_ip)
	// - Index on event_type (idx_device_event_type)
	// - Index on created_at (idx_device_created_at)
	err := c.DB.AutoMigrate(&models.Device{})
	if err != nil {
		return fmt.Errorf("failed to migrate Device model: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// HealthCheck performs a simple health check on the database connection
func (c *Connection) HealthCheck() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
