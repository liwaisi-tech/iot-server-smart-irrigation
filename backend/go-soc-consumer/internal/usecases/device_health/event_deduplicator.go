package devicehealth

import (
	"sync"
	"time"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/domain/entities"
)

// EventDeduplicator manages event deduplication using a sliding window approach
// Only the latest event per device is processed, older events are logged as warnings
type EventDeduplicator struct {
	windowDuration time.Duration
	latestEvents   map[string]*entities.DeviceDetectedEvent
	mu             sync.RWMutex
}

// NewEventDeduplicator creates a new event deduplicator with the specified window duration
func NewEventDeduplicator(windowDuration time.Duration) *EventDeduplicator {
	return &EventDeduplicator{
		windowDuration: windowDuration,
		latestEvents:   make(map[string]*entities.DeviceDetectedEvent),
	}
}

// ShouldProcess determines if an event should be processed or is a duplicate
// Returns true if the event should be processed, false if it's a duplicate
func (ed *EventDeduplicator) ShouldProcess(event *entities.DeviceDetectedEvent) bool {
	if event == nil {
		return false
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	key := event.MACAddress
	existingEvent, exists := ed.latestEvents[key]

	if !exists {
		// First event for this device, should process
		ed.latestEvents[key] = event
		return true
	}

	// Compare timestamps to determine if this is a newer event
	if event.DetectedAt.After(existingEvent.DetectedAt) {
		// This is a newer event, update and process
		ed.latestEvents[key] = event
		return true
	}

	// This is an older or same-time event, it's a duplicate
	return false
}

// GetLatestEvent returns the latest event for a given device MAC address
func (ed *EventDeduplicator) GetLatestEvent(deviceMAC string) *entities.DeviceDetectedEvent {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return ed.latestEvents[deviceMAC]
}

// Cleanup removes old events from the deduplicator to prevent memory growth
// Should be called periodically
func (ed *EventDeduplicator) Cleanup() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-ed.windowDuration)

	for mac, event := range ed.latestEvents {
		if event.DetectedAt.Before(cutoff) {
			delete(ed.latestEvents, mac)
		}
	}
}

// GetEventCount returns the number of events currently tracked
func (ed *EventDeduplicator) GetEventCount() int {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return len(ed.latestEvents)
}
