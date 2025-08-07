# User Story: Device Aliveness Validation

## Story Overview
**As a** IoT system administrator  
**I want** the system to automatically validate device aliveness when devices are registered  
**So that** I can monitor device health and maintain an accurate status of my IoT infrastructure

## Acceptance Criteria

### Functional Requirements

1. **Event Generation on Device Registration**
   - When a device is successfully registered (new registration or update), the system MUST generate a `notify_device_detected` event
   - The event MUST only be published AFTER successful database insert or update operations
   - The event MUST contain the device MAC address and IP address
   - This should happen regardless of whether it's a new device or an update to an existing device
   - If database operations fail, NO event should be published

2. **NATS Event Publishing**
   - The system MUST publish the `notify_device_detected` event to NATS server
   - Event payload MUST include: MAC address and IP address
   - NATS subject should follow project naming conventions (to be determined by golang engineer)

3. **NATS Event Consumption**
   - The system MUST consume `notify_device_detected` events from NATS
   - Upon receiving an event, initiate device health check

4. **Device Health Check**
   - Make HTTP GET request to `http://{{received_ip}}/whoami` endpoint
   - HTTP request timeout: 15 seconds maximum (configurable via `HEALTH_CHECK_TIMEOUT`)
   - Success criteria: HTTP status code 200 only
   - Retry logic: 3 attempts with exponential backoff (3s → 6s → 12s)
   - Cooldown period: 2 minutes between health checks for the same device
   - Event deduplication: Process only the latest event per device, log warning for duplicates

5. **Device Status Updates**
   - **On Success**: Update device status to `online` in existing Device.Status field
   - **On Failure**: Update device status to `offline` in existing Device.Status field
   - Print the `/whoami` response to console when successful
   - Update Device.LastSeen timestamp on both success and failure

### Technical Requirements

1. **Architecture Compliance**
   - Follow Hexagonal Architecture pattern
   - Implement ports in `internal/domain/ports/`
   - Utilize `internal/domain/events/` for event definitions
   - Separate messaging implementations from business logic
   - HTTP client should be implemented as a port with infrastructure implementation

2. **NATS Integration**
   - Use existing NATS server from docker-compose.yml
   - Implement NATS producer for event publishing
   - Implement NATS consumer for event consumption
   - Follow project's messaging patterns but differentiate from MQTT implementation
   - Create separate NATS ports from existing MessageConsumer (which is for MQTT)
   - NATS subject naming: `liwaisi.iot.smart-irrigation.device.detected`

3. **Existing Entity Integration**
   - Use existing Device entity with Status field (supports "registered", "online", "offline")
   - Leverage existing DeviceRepository port for status updates
   - Maintain thread safety using existing sync.RWMutex in Device entity

4. **Error Handling**
   - Graceful handling of HTTP timeouts with exponential backoff retry
   - Proper structured logging with correlation IDs for event tracing
   - Database update failures should be logged but not block the process
   - **Critical**: Event publishing failures should be logged but MUST NOT fail the device registration process
   - NATS connection failures should be logged and handled gracefully
   - Duplicate event warnings should be logged for monitoring purposes

5. **Threading and Concurrency**
   - Health checks should not block device registration process
   - Use Go routines appropriately for async processing
   - Maintain thread safety as per existing codebase patterns

### Detailed Technical Specifications

#### Event Schema Definition
```go
type DeviceDetectedEvent struct {
    MACAddress string    `json:"mac_address"`
    IPAddress  string    `json:"ip_address"`
    DetectedAt time.Time `json:"detected_at"`
    EventID    string    `json:"event_id"`    // UUID for idempotency
    EventType  string    `json:"event_type"`  // "device.detected"
}
```

#### Port Interface Definitions
```go
// Event publishing port (follows existing MessageConsumer pattern)
type EventPublisher interface {
    Publish(ctx context.Context, subject string, data []byte) error
    Close(ctx context.Context) error
}

// Event subscription port (follows existing MessageConsumer pattern)
type EventSubscriber interface {
    Subscribe(ctx context.Context, subject string, handler ports.MessageHandler) error
    Unsubscribe(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    IsConnected() bool
}

// HTTP health check port
type DeviceHealthChecker interface {
    CheckHealth(ctx context.Context, ipAddress string) (*HealthCheckResult, error)
}

// Health check result structure
type HealthCheckResult struct {
    Success      bool          `json:"success"`
    StatusCode   int           `json:"status_code"`
    ResponseBody string        `json:"response_body,omitempty"`
    Duration     time.Duration `json:"duration"`
    Attempts     int           `json:"attempts"`
    Error        string        `json:"error,omitempty"`
}

// NATS message handler (follows existing MQTT handler pattern)
type DeviceHealthHandler struct {
    useCase DeviceHealthUseCase
}

// HandleMessage processes NATS messages (similar to MQTT handler)
func (h *DeviceHealthHandler) HandleMessage(ctx context.Context, subject string, payload []byte) error
```

#### Configuration Environment Variables
```bash
# Health check settings with defaults
HEALTH_CHECK_TIMEOUT=15s           # HTTP request timeout
HEALTH_CHECK_RETRY_ATTEMPTS=3      # Number of retry attempts  
HEALTH_CHECK_INITIAL_DELAY=3s      # Initial backoff delay
HEALTH_CHECK_COOLDOWN=2m           # Cooldown between checks for same device
HEALTH_CHECK_MAX_CONCURRENT=10     # Max concurrent health checks

# NATS settings
NATS_URL=nats://localhost:4222     # NATS server URL
NATS_SUBJECT_PREFIX=liwaisi.iot.smart-irrigation
```

#### Thread Safety & Concurrency Strategy
- Device-level mutex for status updates to prevent race conditions
- Cooldown tracking using in-memory map with device MAC as key
- Event deduplication using sliding window with event timestamps
- Graceful shutdown handling for in-flight health checks

## Technical Implementation Guidance for Golang Engineer

### Architecture Decisions Made:
1. ✅ HTTP client implemented as `DeviceHealthChecker` port in domain layer
2. ✅ NATS subject naming: `liwaisi.iot.smart-irrigation.device.detected`
3. ✅ Event publishing integrated into existing device registration use case **AFTER** successful database operations
4. ✅ NATS consumer integrated into existing application (not separate service)
5. ✅ Event payload structure defined with DeviceDetectedEvent schema
6. ✅ Separate NATS ports (EventPublisher/EventSubscriber) distinct from MQTT MessageConsumer
7. ✅ Event publishing failures do not block device registration (fire-and-forget with logging)

### Implementation Areas:
```
internal/
├── domain/
│   ├── events/
│   │   ├── device_detected.go          # DeviceDetectedEvent definition
│   │   └── event_types.go              # Event type constants
│   ├── ports/
│   │   ├── event_publisher.go          # EventPublisher interface
│   │   ├── event_subscriber.go         # EventSubscriber interface
│   │   └── device_health_checker.go    # DeviceHealthChecker interface
├── infrastructure/
│   ├── messaging/
│   │   ├── nats/
│   │   │   ├── publisher.go            # NATS EventPublisher implementation
│   │   │   ├── subscriber.go           # NATS EventSubscriber implementation
│   │   │   ├── config.go               # NATS connection config
│   │   │   └── handlers/
│   │   │       └── device_health_handler.go # NATS message handler (follows MQTT pattern)
│   │   └── mqtt/                       # Existing MQTT code (unchanged)
│   │       └── handlers/               # Existing MQTT handlers
│   └── http/
│       └── health_client.go            # DeviceHealthChecker implementation
└── usecases/
    └── device_health/
        ├── health_check.go             # Health check use case
        ├── event_handler.go            # Event processing logic  
        └── cooldown_manager.go         # Device cooldown tracking
```

**Architecture Notes**: 
- **NATS handlers**: Placed in `internal/infrastructure/messaging/nats/handlers/` following the existing MQTT pattern in `internal/infrastructure/messaging/mqtt/handlers/`
- **Presentation layer**: Only for HTTP/REST handlers that interact with users  
- **Infrastructure messaging**: For internal event handlers (MQTT/NATS) that adapt external messages to domain logic
- **Dependency injection**: Follow the same pattern as MQTT in `cmd/server/main.go`:
  ```go
  // Similar to existing MQTT pattern:
  // mqttConsumer := messaging.NewMQTTConsumer(mqttConfig)
  // messageHandler := messaginghandlers.NewDeviceRegistrationHandler(deviceRegistrationUseCase)
  
  natsSubscriber := nats.NewNATSSubscriber(natsConfig)
  natsPublisher := nats.NewNATSPublisher(natsConfig)
  healthHandler := natshandlers.NewDeviceHealthHandler(deviceHealthUseCase)
  
  // Event publishing integrated into existing device registration use case
  deviceRegistrationUseCase := deviceregistration.NewDeviceRegistrationUseCase(
      deviceRepo,
      natsPublisher, // Add event publisher dependency
  )
  ```

### Current Entity Structure Reference:
```go
type Device struct {
    mu                  sync.RWMutex
    MACAddress          string
    DeviceName          string
    IPAddress           string
    LocationDescription string
    RegisteredAt        time.Time
    LastSeen            time.Time
    Status              string // "registered", "online", "offline"
}
```

### Testing Strategy

#### Unit Tests Required
```go
// Event handling
func TestDeviceDetectedEvent_Validation(t *testing.T)
func TestEventPublisher_PublishDeviceDetected(t *testing.T)
func TestEventDeduplication_ProcessLatestOnly(t *testing.T)

// Health checking
func TestDeviceHealthChecker_SuccessScenario(t *testing.T)
func TestDeviceHealthChecker_RetryLogic(t *testing.T)  
func TestDeviceHealthChecker_TimeoutHandling(t *testing.T)
func TestDeviceHealthChecker_ExponentialBackoff(t *testing.T)

// Cooldown management
func TestCooldownManager_EnforcesCooldownPeriod(t *testing.T)
func TestCooldownManager_AllowsAfterCooldown(t *testing.T)

// Concurrency & thread safety
func TestDevice_ConcurrentStatusUpdates(t *testing.T)
func TestHealthCheck_DoesNotBlockRegistration(t *testing.T)

// Use case integration
func TestHealthCheckUseCase_UpdatesDeviceStatus(t *testing.T)
func TestHealthCheckUseCase_HandlesRepositoryFailure(t *testing.T)
```

#### Integration Tests Required
- NATS pub/sub with real NATS server
- HTTP client with mock HTTP server  
- End-to-end: device registration → event → health check → status update
- Graceful shutdown with pending health checks
- Database transaction handling

#### Mock Generation Strategy
Use `mockery` to generate mocks.

## Definition of Done
- [ ] Device registration triggers `notify_device_detected` event with proper schema **ONLY after successful database operations**
- [ ] Events are NOT published if device registration/update fails in database
- [ ] NATS producer publishes events successfully to `liwaisi.iot.smart-irrigation.device.detected`
- [ ] NATS event publishing failures are logged but do NOT break device registration process
- [ ] NATS consumer receives and processes events with deduplication logic
- [ ] HTTP client calls `/whoami` endpoint with 15s timeout and retry logic (3 attempts, exponential backoff)
- [ ] Health check success determined by HTTP status 200 only
- [ ] Device status updates to `online` on successful health check
- [ ] Device status updates to `offline` on failed health check (after all retries)
- [ ] Device.LastSeen timestamp updated on health check attempts
- [ ] Response body printed to console on successful health check
- [ ] Cooldown period (2 minutes) enforced between health checks for same device
- [ ] Duplicate events logged as warnings and only latest processed
- [ ] All settings configurable via environment variables with defaults
- [ ] All components follow hexagonal architecture principles
- [ ] Unit tests cover new functionality including retry logic and cooldown
- [ ] Integration tests validate NATS communication and HTTP health checks
- [ ] No blocking of existing device registration flow
- [ ] Thread safety maintained using existing Device mutex
- [ ] Graceful shutdown handling for in-flight health checks
- [ ] Proper error handling and structured logging implemented

## Dependencies
- Existing NATS server in docker-compose.yml (port 4222)
- Device registration functionality (existing MQTT-based system)
- PostgreSQL database with existing Device entity
- Existing DeviceRepository port for database operations
- NATS Go client library: `github.com/nats-io/nats.go`

## Estimated Effort
To be determined by golang engineer based on current codebase complexity and implementation approach.