package mappers

import (
	"fmt"
	"reflect"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

func (m *DeviceDetectedEventMapper) ToDTOFromInterface(data interface{}) (dto interface{}, err error) {
	dataType := reflect.TypeOf(data)

	switch dataType {
	case reflect.TypeOf(&entities.DeviceDetectedEvent{}):
		return m.ToDTOFromDomainEvent(data.(*entities.DeviceDetectedEvent)), nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}
