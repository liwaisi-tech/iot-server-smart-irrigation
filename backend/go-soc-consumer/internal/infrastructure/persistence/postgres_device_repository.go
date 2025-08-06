package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/errors"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/database"
)

// PostgresDeviceRepository implements the DeviceRepository interface using PostgreSQL
type PostgresDeviceRepository struct {
	db *database.PostgresDB
}

// NewPostgresDeviceRepository creates a new PostgresDeviceRepository
func NewPostgresDeviceRepository(db *database.PostgresDB) ports.DeviceRepository {
	return &PostgresDeviceRepository{
		db: db,
	}
}

// Save persists a new device to the database
func (r *PostgresDeviceRepository) Save(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	query := `
		INSERT INTO devices (
			mac_address, device_name, ip_address, location_description,
			registered_at, last_seen, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		query,
		device.MACAddress,
		device.DeviceName,
		device.IPAddress,
		device.LocationDescription,
		device.RegisteredAt,
		device.LastSeen,
		device.Status,
		now,
		now,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return errors.ErrDeviceAlreadyExists
			case "23514": // check_violation
				return errors.ErrInvalidDeviceStatus
			}
		}
		return fmt.Errorf("failed to save device: %w", err)
	}

	return nil
}

// Update updates an existing device in the database
func (r *PostgresDeviceRepository) Update(ctx context.Context, device *entities.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	query := `
		UPDATE devices 
		SET device_name = $2, ip_address = $3, location_description = $4,
			last_seen = $5, status = $6, updated_at = $7
		WHERE mac_address = $1`

	result, err := r.db.ExecContext(
		ctx,
		query,
		device.MACAddress,
		device.DeviceName,
		device.IPAddress,
		device.LocationDescription,
		device.LastSeen,
		device.Status,
		time.Now(),
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // check_violation
				return errors.ErrInvalidDeviceStatus
			}
		}
		return fmt.Errorf("failed to update device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrDeviceNotFound
	}

	return nil
}

// FindByMACAddress retrieves a device by its MAC address
func (r *PostgresDeviceRepository) FindByMACAddress(ctx context.Context, macAddress string) (*entities.Device, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address cannot be empty")
	}

	query := `
		SELECT mac_address, device_name, ip_address, location_description,
			   registered_at, last_seen, status
		FROM devices 
		WHERE mac_address = $1`

	var device entities.Device
	err := r.db.QueryRowContext(ctx, query, macAddress).Scan(
		&device.MACAddress,
		&device.DeviceName,
		&device.IPAddress,
		&device.LocationDescription,
		&device.RegisteredAt,
		&device.LastSeen,
		&device.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to find device by MAC address: %w", err)
	}

	return &device, nil
}

// Exists checks if a device with the given MAC address exists
func (r *PostgresDeviceRepository) Exists(ctx context.Context, macAddress string) (bool, error) {
	if macAddress == "" {
		return false, fmt.Errorf("mac address cannot be empty")
	}

	query := `SELECT 1 FROM devices WHERE mac_address = $1 LIMIT 1`

	var exists int
	err := r.db.QueryRowContext(ctx, query, macAddress).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	return true, nil
}

// List retrieves all devices with optional pagination
func (r *PostgresDeviceRepository) List(ctx context.Context, offset, limit int) ([]*entities.Device, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative")
	}

	query := `
		SELECT mac_address, device_name, ip_address, location_description,
			   registered_at, last_seen, status
		FROM devices 
		ORDER BY registered_at DESC`

	args := []interface{}{}
	
	// Add LIMIT clause if limit is specified
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
		
		// Add OFFSET clause if offset is specified
		if offset > 0 {
			query += " OFFSET $2"
			args = append(args, offset)
		}
	} else if offset > 0 {
		// If only offset is specified, we need a reasonable default limit
		// to avoid performance issues
		query += " LIMIT $1 OFFSET $2"
		args = append(args, 1000, offset) // Default limit of 1000
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []*entities.Device
	for rows.Next() {
		var device entities.Device
		err := rows.Scan(
			&device.MACAddress,
			&device.DeviceName,
			&device.IPAddress,
			&device.LocationDescription,
			&device.RegisteredAt,
			&device.LastSeen,
			&device.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over device rows: %w", err)
	}

	return devices, nil
}

// Delete removes a device by MAC address
func (r *PostgresDeviceRepository) Delete(ctx context.Context, macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address cannot be empty")
	}

	query := `DELETE FROM devices WHERE mac_address = $1`

	result, err := r.db.ExecContext(ctx, query, macAddress)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrDeviceNotFound
	}

	return nil
}