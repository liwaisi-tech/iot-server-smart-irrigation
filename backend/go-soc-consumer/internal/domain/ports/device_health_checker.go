package ports

import (
	"context"
)

// DeviceHealthChecker defines the contract for checking device health
type DeviceHealthChecker interface {
	// CheckHealth performs a health check on the device at the given IP address
	// It will make an HTTP GET request to http://ipAddress/whoami
	// Returns HealthCheckResult with success/failure details and retry information
	CheckHealth(ctx context.Context, ipAddress string) (isAlive bool, err error)
}
