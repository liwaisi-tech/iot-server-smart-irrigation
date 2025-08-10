package mappers

import (
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/infrastructure/persistence/postgres/models"
)

type SensorTemperatureHumidityMapper struct{}

func NewSensorTemperatureHumidityMapper() *SensorTemperatureHumidityMapper {
	return &SensorTemperatureHumidityMapper{}
}

func (m *SensorTemperatureHumidityMapper) ToModel(sensorData *entities.SensorTemperatureHumidity) *models.SensorTemperatureHumidityModel {
	if sensorData == nil {
		return nil
	}

	return &models.SensorTemperatureHumidityModel{
		MACAddress:         sensorData.MacAddress(),
		TemperatureCelsius: sensorData.Temperature(),
		HumidityPercent:    sensorData.Humidity(),
		CreatedAt:          sensorData.Timestamp(),
		UpdatedAt:          sensorData.Timestamp(),
	}
}

func (m *SensorTemperatureHumidityMapper) FromModel(model *models.SensorTemperatureHumidityModel) (*entities.SensorTemperatureHumidity, error) {
	if model == nil {
		return nil, nil
	}
	return entities.NewSensorTemperatureHumidity(model.MACAddress, model.TemperatureCelsius, model.HumidityPercent)
}

func (m *SensorTemperatureHumidityMapper) FromModelSlice(models []*models.SensorTemperatureHumidityModel) ([]*entities.SensorTemperatureHumidity, error) {
	if models == nil {
		return nil, nil
	}

	entitiesSlice := make([]*entities.SensorTemperatureHumidity, len(models))
	for i, model := range models {
		mapped, err := m.FromModel(model)
		if err != nil {
			return nil, err
		}
		entitiesSlice[i] = mapped
	}
	return entitiesSlice, nil
}
