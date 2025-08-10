package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/ports"
	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/pkg/logger"
)

// HealthClientConfig holds configuration for the health checker
type HealthClientConfig struct {
	Timeout       time.Duration
	RetryAttempts int
	InitialDelay  time.Duration
	UserAgent     string
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
	config        *HealthClientConfig
	client        *http.Client
	loggerFactory logger.LoggerFactory
}

// NewHealthClient creates a new HTTP health checker implementation
func NewHealthClient(config *HealthClientConfig, loggerFactory logger.LoggerFactory) ports.DeviceHealthChecker {
	if config == nil {
		config = DefaultHealthClientConfig()
	}

	if loggerFactory == nil {
		defaultLoggerFactory, err := logger.NewDefault()
		if err != nil {
			// Fallback to a basic implementation if default creation fails
			panic(fmt.Sprintf("failed to create default logger factory: %v", err))
		}
		loggerFactory = defaultLoggerFactory
	}

	return &healthClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		loggerFactory: loggerFactory,
	}
}

// CheckHealth performs a health check with retry logic and exponential backoff
func (hc *healthClient) CheckHealth(ctx context.Context, ipAddress string) (isAlive bool, err error) {
	url := fmt.Sprintf("http://%s/whoami", ipAddress)
	hc.loggerFactory.Core().Info("health_check_starting",
		zap.String("ip_address", ipAddress),
		zap.String("url", url),
		zap.String("component", "health_client"),
	)

	var lastErr error
	delay := hc.config.InitialDelay

	for attempt := 1; attempt <= hc.config.RetryAttempts; attempt++ {
		start := time.Now()
		success, statusCode, responseBody, err := hc.performHealthCheck(ctx, url)
		duration := time.Since(start)

		if success {
			hc.loggerFactory.Core().Info("health_check_succeeded",
				zap.String("ip_address", ipAddress),
				zap.Int("attempt", attempt),
				zap.Int("status_code", statusCode),
				zap.Duration("duration", duration),
				zap.String("response_body", responseBody),
				zap.String("component", "health_client"),
			)
			return true, nil
		}

		lastErr = err
		if err != nil {
			hc.loggerFactory.Core().Warn("health_check_attempt_failed",
				zap.String("ip_address", ipAddress),
				zap.Int("attempt", attempt),
				zap.Error(err),
				zap.Int("status_code", statusCode),
				zap.Duration("duration", duration),
				zap.String("component", "health_client"),
			)
		} else {
			hc.loggerFactory.Core().Warn("health_check_attempt_wrong_status",
				zap.String("ip_address", ipAddress),
				zap.Int("attempt", attempt),
				zap.Int("status_code", statusCode),
				zap.Int("expected_status", 200),
				zap.Duration("duration", duration),
				zap.String("component", "health_client"),
			)
		}

		// Don't wait after the last attempt
		if attempt < hc.config.RetryAttempts {
			hc.loggerFactory.Core().Debug("health_check_waiting_retry",
				zap.String("ip_address", ipAddress),
				zap.Duration("delay", delay),
				zap.Int("next_attempt", attempt+1),
				zap.String("component", "health_client"),
			)

			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(delay):
				// Exponential backoff: double the delay for next attempt
				delay *= 2
			}
		}
	}

	hc.loggerFactory.Core().Error("health_check_failed_all_attempts",
		zap.String("ip_address", ipAddress),
		zap.Int("total_attempts", hc.config.RetryAttempts),
		zap.Error(lastErr),
		zap.String("component", "health_client"),
	)

	return false, lastErr
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
			hc.loggerFactory.Core().Warn("response_body_close_failed",
				zap.Error(closeErr),
				zap.String("component", "health_client"),
			)
		}
	}()

	statusCode = resp.StatusCode

	// Read response body (limited to prevent memory exhaustion)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4096)) // Limit to 4KB
	if err != nil {
		hc.loggerFactory.Core().Warn("response_body_read_failed",
			zap.Error(err),
			zap.String("component", "health_client"),
		)
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
