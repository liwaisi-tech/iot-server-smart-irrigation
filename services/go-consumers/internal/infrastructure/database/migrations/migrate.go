package migrations

import (
	"log"

	"github.com/liwaisi/iot-server-smart-irrigation/services/go-consumers/internal/infrastructure/database"
)

// RunMigrations executes the database migrations using GORM AutoMigrate
func RunMigrations(dbConn *database.Connection) error {
	log.Println("Executing database migrations...")

	if err := dbConn.AutoMigrate(); err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateFromConfig creates a database connection and runs migrations
func MigrateFromConfig(config database.DatabaseConfig) error {
	// Create database connection
	dbConn, err := database.NewPostgresConnection(config)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Run health check
	if err := dbConn.HealthCheck(); err != nil {
		log.Printf("Database health check failed: %v", err)
		return err
	}

	// Execute migrations
	return RunMigrations(dbConn)
}
