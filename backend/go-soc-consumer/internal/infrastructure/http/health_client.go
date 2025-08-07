package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
)

// HealthClientConfig holds configuration for the health checker
type HealthClientConfig struct {
	Timeout        time.Duration
	RetryAttempts  int
	InitialDelay   time.Duration
	UserAgent      string
}

// DefaultHealthClientConfig returns default configuration for the health client
func DefaultHealthClientConfig() *HealthClientConfig {
	return &HealthClientConfig{
		Timeout:       15 * time.Second,
		RetryAttempts: 3,
		InitialDelay:  3 * time.Second,
		UserAgent:     "iot-soc-consumer/1.0",
	}
}

// healthClient implements the DeviceHealthChecker port
type healthClient struct {
	config *HealthClientConfig
	client *http.Client
	logger *slog.Logger
}

// NewHealthClient creates a new HTTP health checker implementation
func NewHealthClient(config *HealthClientConfig, logger *slog.Logger) ports.DeviceHealthChecker {
	if config == nil {
		config = DefaultHealthClientConfig()
	}
	
	if logger == nil {
		logger = slog.Default()
	}

	return &healthClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// CheckHealth performs a health check with retry logic and exponential backoff
func (hc *healthClient) CheckHealth(ctx context.Context, ipAddress string) (*ports.HealthCheckResult, error) {
	result := &ports.HealthCheckResult{
		IPAddress: ipAddress,
		CheckedAt: time.Now(),
		Attempts:  0,
		Success:   false,
	}

	url := fmt.Sprintf("http://%s/whoami", ipAddress)
	hc.logger.Info("Starting health check", "ip", ipAddress, "url", url)

	var lastErr error
	delay := hc.config.InitialDelay

	for attempt := 1; attempt <= hc.config.RetryAttempts; attempt++ {
		result.Attempts = attempt
		
		start := time.Now()
		success, statusCode, responseBody, err := hc.performHealthCheck(ctx, url)
		duration := time.Since(start)
		
		result.Duration = duration
		result.StatusCode = statusCode
		result.ResponseBody = responseBody

		if success {
			result.Success = true
			hc.logger.Info("Health check succeeded", 
				"ip", ipAddress,
				"attempt", attempt,
				"status_code", statusCode,
				"duration", duration,
				"response_body", responseBody)
			return result, nil
		}

		lastErr = err
		if err != nil {
			result.Error = err.Error()
			hc.logger.Warn("Health check attempt failed",
				"ip", ipAddress,
				"attempt", attempt,
				"error", err,
				"status_code", statusCode,
				"duration", duration)
		} else {
			result.Error = fmt.Sprintf("HTTP status %d (expected 200)", statusCode)
			hc.logger.Warn("Health check attempt failed - wrong status code",
				"ip", ipAddress,
				"attempt", attempt,
				"status_code", statusCode,
				"expected_status", 200,
				"duration", duration)
		}

		// Don't wait after the last attempt
		if attempt < hc.config.RetryAttempts {
			hc.logger.Info("Waiting before next attempt",
				"ip", ipAddress,
				"delay", delay,
				"next_attempt", attempt+1)
			
			select {
			case <-ctx.Done():
				result.Error = fmt.Sprintf("context cancelled after %d attempts: %v", attempt, ctx.Err())
				return result, ctx.Err()
			case <-time.After(delay):
				// Exponential backoff: double the delay for next attempt
				delay *= 2
			}
		}
	}

	hc.logger.Error("Health check failed after all attempts",
		"ip", ipAddress,
		"attempts", hc.config.RetryAttempts,
		"final_error", lastErr)
	
	return result, lastErr
}

// performHealthCheck makes a single HTTP request to the device
func (hc *healthClient) performHealthCheck(ctx context.Context, url string) (success bool, statusCode int, responseBody string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", hc.config.UserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := hc.client.Do(req)
	if err != nil {
		return false, 0, "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			hc.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	statusCode = resp.StatusCode
	
	// Read response body (limited to prevent memory exhaustion)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4096)) // Limit to 4KB
	if err != nil {
		hc.logger.Warn("Failed to read response body", "error", err)
		responseBody = "<failed to read response>"
	} else {
		responseBody = string(bodyBytes)
	}

	// Success is determined by HTTP status code 200 only
	success = statusCode == http.StatusOK
	
	if !success && err == nil {
		err = fmt.Errorf("HTTP status %s (%s)", 
			strconv.Itoa(statusCode), 
			http.StatusText(statusCode))
	}

	return success, statusCode, responseBody, err
}