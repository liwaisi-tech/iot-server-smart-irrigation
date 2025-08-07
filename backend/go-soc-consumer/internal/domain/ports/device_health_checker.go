package ports

import (
	"context"
	"time"
)

// HealthCheckResult represents the result of a device health check
type HealthCheckResult struct {
	Success      bool          `json:"success"`
	StatusCode   int           `json:"status_code"`
	ResponseBody string        `json:"response_body,omitempty"`
	Duration     time.Duration `json:"duration"`
	Attempts     int           `json:"attempts"`
	Error        string        `json:"error,omitempty"`
	IPAddress    string        `json:"ip_address"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// DeviceHealthChecker defines the contract for checking device health
type DeviceHealthChecker interface {
	// CheckHealth performs a health check on the device at the given IP address
	// It will make an HTTP GET request to http://ipAddress/whoami
	// Returns HealthCheckResult with success/failure details and retry information
	CheckHealth(ctx context.Context, ipAddress string) (*HealthCheckResult, error)
}