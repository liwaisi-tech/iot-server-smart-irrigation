package errors

// Device-specific domain errors
var (
	ErrDeviceNotFound      = NewDomainError("DEVICE_NOT_FOUND", "Device not found")
	ErrDeviceAlreadyExists = NewDomainError("DEVICE_ALREADY_EXISTS", "Device already exists")
	ErrInvalidDeviceStatus = NewDomainError("INVALID_DEVICE_STATUS", "Invalid device status")
)
