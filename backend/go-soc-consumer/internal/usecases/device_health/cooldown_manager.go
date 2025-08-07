package devicehealth

import (
	"sync"
	"time"
)

// CooldownManager manages cooldown periods for device health checks
// to prevent excessive checking of the same device
type CooldownManager struct {
	cooldownPeriod time.Duration
	lastChecked    map[string]time.Time
	mu             sync.RWMutex
}

// NewCooldownManager creates a new cooldown manager with the specified cooldown period
func NewCooldownManager(cooldownPeriod time.Duration) *CooldownManager {
	return &CooldownManager{
		cooldownPeriod: cooldownPeriod,
		lastChecked:    make(map[string]time.Time),
	}
}

// CanCheck returns true if the device can be checked (not in cooldown)
func (cm *CooldownManager) CanCheck(deviceMAC string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	lastCheck, exists := cm.lastChecked[deviceMAC]
	if !exists {
		return true
	}

	return time.Since(lastCheck) >= cm.cooldownPeriod
}

// MarkChecked records that a health check was performed for the device
func (cm *CooldownManager) MarkChecked(deviceMAC string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.lastChecked[deviceMAC] = time.Now()
}

// GetTimeUntilNextCheck returns the duration until the next check is allowed
// Returns 0 if check is allowed immediately
func (cm *CooldownManager) GetTimeUntilNextCheck(deviceMAC string) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	lastCheck, exists := cm.lastChecked[deviceMAC]
	if !exists {
		return 0
	}

	elapsed := time.Since(lastCheck)
	if elapsed >= cm.cooldownPeriod {
		return 0
	}

	return cm.cooldownPeriod - elapsed
}

// GetLastChecked returns the last check time for a device, or zero time if never checked
func (cm *CooldownManager) GetLastChecked(deviceMAC string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.lastChecked[deviceMAC]
}

// Cleanup removes old entries to prevent memory growth
// Should be called periodically with a cleanup interval
func (cm *CooldownManager) Cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	// Remove entries older than 2x the cooldown period
	cleanupThreshold := now.Add(-2 * cm.cooldownPeriod)

	for mac, lastCheck := range cm.lastChecked {
		if lastCheck.Before(cleanupThreshold) {
			delete(cm.lastChecked, mac)
		}
	}
}