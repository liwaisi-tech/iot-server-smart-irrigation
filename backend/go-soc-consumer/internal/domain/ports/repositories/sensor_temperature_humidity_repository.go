package ports

import (
	"context"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

// SensorTemperatureHumidityRepository defines the port for sensor temperature humidity data persistence operations
type SensorTemperatureHumidityRepository interface {
	// Create creates a new sensor temperature humidity reading record
	Create(ctx context.Context, sensorData *entities.SensorTemperatureHumidity) error
}
