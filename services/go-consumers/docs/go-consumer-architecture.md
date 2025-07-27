# Go-Consumer Service Architecture

## Overview

The `go-consumer` service is the central message processing component of the IoT Smart Irrigation system. It directly consumes MQTT messages from ESP32 devices and processes them through dedicated topic-specific handlers, storing processed data in PostgreSQL using GORM.

## System Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ESP32 Devices │    │   API Services   │    │   LLM Services  │
│                 │    │                  │    │                 │
└─────────┬───────┘    └────────┬─────────┘    └─────────┬───────┘
          │                     │                        │
          │ MQTT Messages       │ HTTP/WebSocket         │ HTTP/gRPC
          │                     │                        │
          └─────────────────────┼────────────────────────┘
                                │
                                ▼
                    ┌─────────────────────┐
                    │    MQTT Broker      │
                    │   (Eclipse Mosquitto│
                    │    or HiveMQ)       │
                    └─────────┬───────────┘
                              │ MQTT Subscribe
                              ▼
                    ┌─────────────────────┐
                    │    Go-Consumer      │
                    │    Service          │
                    │  (MQTT Subscriber)  │
                    └─────────┬───────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │    PostgreSQL       │
                    │    Database         │
                    └─────────────────────┘
```

## Domain-Driven Design Architecture

### Core Domains

#### 1. Message Processing Domain
- **Purpose**: Handle incoming messages from various IoT sources
- **Responsibilities**: Message validation, routing, and processing coordination

#### 2. Device Data Domain
- **Purpose**: Manage sensor data and device state information
- **Responsibilities**: Sensor reading processing, device health tracking

#### 3. Command Domain
- **Purpose**: Handle device command responses and status updates
- **Responsibilities**: Command acknowledgment processing, device state updates

### Bounded Contexts

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Go-Consumer Service                          │
├─────────────────┬─────────────────┬─────────────────┬───────────────┤
│  Message        │  Device Data    │  Command        │  Infrastructure│
│  Processing     │  Context        │  Context        │  Context      │
│  Context        │                 │                 │               │
│                 │                 │                 │               │
│ • Router        │ • SensorData    │ • CommandResp   │ • MQTT Client │
│ • Handler Mgmt  │ • DeviceHealth  │ • DeviceStatus  │ • DB Client   │
│ • Error Handling│ • Aggregates    │ • Aggregates    │ • Config      │
│                 │                 │                 │ • Monitoring  │
└─────────────────┴─────────────────┴─────────────────┴───────────────┘
```

## Component Architecture

### 1. Enhanced Message Router with Circuit Breaker
```go
type MessageRouter interface {
    Route(ctx context.Context, subject string, message []byte) error
    RegisterHandler(pattern string, handler MessageHandler) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck() error
}

type messageRouter struct {
    handlers     map[string]MessageHandler
    circuitBreaker *CircuitBreaker
    metrics      *Metrics
    logger       Logger
}
```

### 2. Enhanced Topic-Specific Handlers with DDD Patterns

#### Sensor Data Handler with Value Objects
```go
type SensorDataHandler struct {
    sensorService domain.SensorService
    batchProcessor *BatchProcessor
    validator     SensorDataValidator
    metrics       *HandlerMetrics
}

// Actual MQTT topic pattern based on ESP32 implementation:
// /liwaisi/iot/smart-irrigation/{message_type}/{sub_type}
// Examples: 
// - /liwaisi/iot/smart-irrigation/device/registration
// - /liwaisi/iot/smart-irrigation/sensor/temperature
// - /liwaisi/iot/smart-irrigation/sensor/humidity
// - /liwaisi/iot/smart-irrigation/device/status

type SensorValue struct {
    Value float64 `json:"value"`
    Unit  string  `json:"unit"`
    Quality Quality `json:"quality"`
}

type Quality string
const (
    QualityGood    Quality = "good"
    QualityWarning Quality = "warning"
    QualityError   Quality = "error"
)
```

#### Command Response Handler with Aggregate Pattern
```go
type CommandResponseHandler struct {
    commandService domain.CommandService
    validator     CommandValidator
    metrics       *HandlerMetrics
}

// MQTT topic pattern for command responses:
// /liwaisi/iot/smart-irrigation/command/{command_type}
// Examples:
// - /liwaisi/iot/smart-irrigation/command/irrigation
// - /liwaisi/iot/smart-irrigation/command/calibration
// - /liwaisi/iot/smart-irrigation/command/restart

type CommandAggregate struct {
    commandResponse *domain.CommandResponse
    deviceState     *domain.DeviceState
}
```

#### Health Status Handler with State Management
```go
type HealthStatusHandler struct {
    healthService  domain.HealthService
    stateManager   *DeviceStateManager
    alertService   AlertService
    validator      HealthValidator
    metrics        *HandlerMetrics
}

// MQTT topic pattern for device health:
// /liwaisi/iot/smart-irrigation/health/{metric}
// Examples:
// - /liwaisi/iot/smart-irrigation/health/status
// - /liwaisi/iot/smart-irrigation/health/battery
// - /liwaisi/iot/smart-irrigation/health/connectivity
// - /liwaisi/iot/smart-irrigation/health/diagnostics
```

### 3. Enhanced Message Handler Interface with Error Types
```go
type MessageHandler interface {
    Handle(ctx context.Context, msg *Message) error
    GetRetryPolicy() RetryPolicy
    GetSubject() string
    HealthCheck() error
    GetMetrics() HandlerMetrics
}

type Message struct {
    Subject     string            `json:"subject"`
    Data        []byte            `json:"data"`
    Headers     map[string]string `json:"headers"`
    Timestamp   time.Time         `json:"timestamp"`
    DeviceID    string            `json:"device_id"`
    Region      string            `json:"region"`
    Zone        string            `json:"zone"`
    MessageType string            `json:"message_type"`
    SequenceID  uint64            `json:"sequence_id"`
}

type RetryPolicy struct {
    MaxRetries      int               `yaml:"max_retries"`
    BackoffStrategy BackoffStrategy   `yaml:"backoff_strategy"`
    RetryableErrors []error          `yaml:"retryable_errors"`
}

type BackoffStrategy string
const (
    ExponentialBackoff BackoffStrategy = "exponential"
    LinearBackoff      BackoffStrategy = "linear"
    FixedBackoff       BackoffStrategy = "fixed"
)

// Enhanced Error Types
type ProcessingError struct {
    Code      ErrorCode `json:"code"`
    Message   string    `json:"message"`
    Retryable bool      `json:"retryable"`
    DeviceID  string    `json:"device_id"`
    Subject   string    `json:"subject"`
}

type ErrorCode string
const (
    ValidationError    ErrorCode = "VALIDATION_ERROR"
    DatabaseError      ErrorCode = "DATABASE_ERROR"
    BusinessLogicError ErrorCode = "BUSINESS_LOGIC_ERROR"
    InfrastructureError ErrorCode = "INFRASTRUCTURE_ERROR"
)
```

### 4. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    failureThreshold int
    successThreshold int
    timeout          time.Duration
    state           CircuitState
    failures        int
    lastFailureTime time.Time
    mu              sync.RWMutex
}

type CircuitState int
const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if !cb.allowRequest() {
        return ErrCircuitBreakerOpen
    }
    
    err := fn()
    cb.recordResult(err == nil)
    return err
}
```

## Data Models and Domain Design

### Domain Entities (Enhanced with DDD Patterns)

#### Sensor Reading Aggregate
```go
// Domain Entity
type SensorReading struct {
    id          SensorReadingID
    deviceID    DeviceID
    location    Location
    sensorValue SensorValue
    timestamp   time.Time
    quality     Quality
    metadata    map[string]interface{}
}

// Value Objects
type SensorReadingID struct {
    value uint
}

type DeviceID struct {
    value string
}

type Location struct {
    Region string `json:"region"`
    Zone   string `json:"zone"`
}

type SensorValue struct {
    Type  SensorType `json:"type"`
    Value float64    `json:"value"`
    Unit  string     `json:"unit"`
}

type SensorType string
const (
    Temperature   SensorType = "temperature"
    Humidity      SensorType = "humidity"
    SoilMoisture  SensorType = "soil_moisture"
    pH            SensorType = "ph"
    LightLevel    SensorType = "light_level"
)

// GORM Database Model (Infrastructure Layer)
type SensorReadingModel struct {
    ID            uint      `gorm:"primaryKey"`
    DeviceID      string    `gorm:"index;not null;size:50"`
    Region        string    `gorm:"index;not null;size:50"`
    Zone          string    `gorm:"index;not null;size:50"`
    SensorType    string    `gorm:"not null;size:20"`
    Value         float64   `gorm:"not null"`
    Unit          string    `gorm:"not null;size:10"`
    Quality       string    `gorm:"default:good;size:20"`
    Timestamp     time.Time `gorm:"index;not null"`
    MetadataJSON  string    `gorm:"type:jsonb"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// Index for time-series queries
func (SensorReadingModel) TableName() string {
    return "sensor_readings"
}
```

#### Command Response Aggregate
```go
// Domain Entity
type CommandResponse struct {
    id           CommandResponseID
    deviceID     DeviceID
    commandID    CommandID
    commandType  CommandType
    result       CommandResult
    executedAt   time.Time
    deviceState  *DeviceState
}

// Value Objects
type CommandType string
const (
    IrrigationStart   CommandType = "irrigation_start"
    IrrigationStop    CommandType = "irrigation_stop"
    CalibrateSensors  CommandType = "calibrate_sensors"
    UpdateFirmware    CommandType = "update_firmware"
    DeviceRestart     CommandType = "device_restart"
)

type CommandResult struct {
    Status    CommandStatus `json:"status"`
    Message   string        `json:"message"`
    Data      interface{}   `json:"data,omitempty"`
    ErrorCode string        `json:"error_code,omitempty"`
}

type CommandStatus string
const (
    CommandSuccess   CommandStatus = "success"
    CommandFailed    CommandStatus = "failed"
    CommandTimeout   CommandStatus = "timeout"
    CommandRetrying  CommandStatus = "retrying"
)

// GORM Database Model
type CommandResponseModel struct {
    ID            uint      `gorm:"primaryKey"`
    DeviceID      string    `gorm:"index;not null;size:50"`
    CommandID     string    `gorm:"not null;size:100"`
    CommandType   string    `gorm:"not null;size:50"`
    Status        string    `gorm:"not null;size:20"`
    Message       string    `gorm:"type:text"`
    ResponseData  string    `gorm:"type:jsonb"`
    ErrorCode     string    `gorm:"size:50"`
    ExecutedAt    time.Time `gorm:"not null"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### Device Health Aggregate
```go
// Domain Entity
type DeviceHealth struct {
    id               DeviceHealthID
    deviceID         DeviceID
    location         Location
    status           DeviceStatus
    connectivity     ConnectivityMetrics
    powerMetrics     PowerMetrics
    diagnostics      DiagnosticData
    lastSeen         time.Time
    alertsTriggered  []Alert
}

// Value Objects
type DeviceStatus string
const (
    DeviceOnline    DeviceStatus = "online"
    DeviceOffline   DeviceStatus = "offline"
    DeviceWarning   DeviceStatus = "warning"
    DeviceError     DeviceStatus = "error"
    DeviceMaintenance DeviceStatus = "maintenance"
)

type ConnectivityMetrics struct {
    SignalStrength   int     `json:"signal_strength"`
    PacketLoss       float64 `json:"packet_loss"`
    Latency          int     `json:"latency_ms"`
    ConnectionType   string  `json:"connection_type"`
}

type PowerMetrics struct {
    BatteryLevel     float64 `json:"battery_level"`
    Voltage          float64 `json:"voltage"`
    PowerConsumption float64 `json:"power_consumption_watts"`
    IsCharging       bool    `json:"is_charging"`
}

type DiagnosticData struct {
    FirmwareVersion  string            `json:"firmware_version"`
    UptimeSeconds    int64             `json:"uptime_seconds"`
    MemoryUsage      float64           `json:"memory_usage_percent"`
    CPUTemperature   float64           `json:"cpu_temperature"`
    ErrorCounts      map[string]int    `json:"error_counts"`
    LastErrors       []string          `json:"last_errors"`
}

// GORM Database Model
type DeviceHealthModel struct {
    ID               uint      `gorm:"primaryKey"`
    DeviceID         string    `gorm:"index;not null;size:50"`
    Region           string    `gorm:"index;not null;size:50"`
    Zone             string    `gorm:"index;not null;size:50"`
    Status           string    `gorm:"not null;size:20"`
    BatteryLevel     float64   `gorm:"default:0"`
    SignalStrength   int       `gorm:"default:0"`
    PacketLoss       float64   `gorm:"default:0"`
    Latency          int       `gorm:"default:0"`
    Voltage          float64   `gorm:"default:0"`
    PowerConsumption float64   `gorm:"default:0"`
    FirmwareVersion  string    `gorm:"size:50"`
    UptimeSeconds    int64     `gorm:"default:0"`
    MemoryUsage      float64   `gorm:"default:0"`
    CPUTemperature   float64   `gorm:"default:0"`
    DiagnosticsJSON  string    `gorm:"type:jsonb"`
    LastSeen         time.Time `gorm:"not null"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### Repository Interfaces (Domain Layer)
```go
type SensorReadingRepository interface {
    Save(ctx context.Context, reading *SensorReading) error
    SaveBatch(ctx context.Context, readings []*SensorReading) error
    FindByDeviceAndTimeRange(ctx context.Context, deviceID DeviceID, start, end time.Time) ([]*SensorReading, error)
    FindLatestByDevice(ctx context.Context, deviceID DeviceID) (*SensorReading, error)
    GetAggregatedData(ctx context.Context, deviceID DeviceID, aggregation Aggregation, timeRange TimeRange) (*AggregatedSensorData, error)
}

type CommandResponseRepository interface {
    Save(ctx context.Context, response *CommandResponse) error
    FindByCommandID(ctx context.Context, commandID CommandID) (*CommandResponse, error)
    FindByDeviceID(ctx context.Context, deviceID DeviceID, limit int) ([]*CommandResponse, error)
    UpdateStatus(ctx context.Context, id CommandResponseID, status CommandStatus) error
}

type DeviceHealthRepository interface {
    Save(ctx context.Context, health *DeviceHealth) error
    FindByDeviceID(ctx context.Context, deviceID DeviceID) (*DeviceHealth, error)
    FindUnhealthyDevices(ctx context.Context, threshold time.Duration) ([]*DeviceHealth, error)
    UpdateStatus(ctx context.Context, deviceID DeviceID, status DeviceStatus) error
}
```

## Service Architecture

### Directory Structure
```
services/go-consumer/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── sensor_reading.go
│   │   │   ├── command_response.go
│   │   │   └── device_health.go
│   │   ├── repositories/
│   │   │   ├── sensor_repository.go
│   │   │   ├── command_repository.go
│   │   │   └── health_repository.go
│   │   └── services/
│   │       ├── sensor_service.go
│   │       ├── command_service.go
│   │       └── health_service.go
│   ├── infrastructure/
│   │   ├── nats/
│   │   │   ├── client.go
│   │   │   └── subscriber.go
│   │   ├── database/
│   │   │   ├── connection.go
│   │   │   └── migrations/
│   │   └── config/
│   │       └── config.go
│   ├── application/
│   │   ├── handlers/
│   │   │   ├── sensor_handler.go
│   │   │   ├── command_handler.go
│   │   │   └── health_handler.go
│   │   ├── services/
│   │   │   └── message_router.go
│   │   └── validators/
│   │       ├── sensor_validator.go
│   │       ├── command_validator.go
│   │       └── health_validator.go
│   └── interfaces/
│       ├── nats/
│       │   └── message_consumer.go
│       └── http/
│           └── health_check.go
├── pkg/
│   ├── logger/
│   └── metrics/
├── configs/
│   └── config.yaml
├── docker/
│   └── Dockerfile
└── go.mod
```

## Message Flow

### 1. MQTT Message Reception Flow
```
ESP32 Device → MQTT Publish → MQTT Broker → Go-Consumer Subscribe → Message Router → 
Topic Pattern Matching → Handler Selection → Handler Processing
```

### 2. Handler Processing Flow
```
MQTT Message → JSON Parsing → Message Validation → Business Logic Processing → 
Database Persistence → Success/Error Response → MQTT Ack
```

### 3. Device Registration Flow
```
ESP32 Boot → Device Registration Message → /liwaisi/iot/smart-irrigation/device/registration → 
Device Handler → Database Upsert → Registration Confirmation
```

### 4. Retry Mechanism Flow
```
Handler Error → Retry Counter Check → Retry (up to 3 times) → Success OR Dead Letter Queue
```

## Concurrency Model

### Handler Concurrency
- Each topic handler runs in its own goroutine pool
- Configurable number of concurrent workers per handler type
- Message processing is concurrent but maintains order per device when needed

### Resource Management
```go
type HandlerPool struct {
    workers    int
    semaphore  chan struct{}
    handler    MessageHandler
    wg         sync.WaitGroup
}
```

## Configuration

### MQTT Configuration
```yaml
mqtt:
  broker_url: "tcp://localhost:1883"  # MQTT broker URL
  client_id: "go-consumer-service"
  # Topic subscriptions for ESP32 device messages
  topics:
    - "/liwaisi/iot/smart-irrigation/device/registration"
    - "/liwaisi/iot/smart-irrigation/sensor/+"     # Wildcard for all sensor types
    - "/liwaisi/iot/smart-irrigation/command/+"    # Wildcard for all commands
    - "/liwaisi/iot/smart-irrigation/health/+"     # Wildcard for all health metrics
  
  # Quality of Service levels
  qos: 1  # At least once delivery
  
  # Security configuration (if using TLS)
  security:
    tls_enabled: false
    cert_file: "/certs/client.crt"
    key_file: "/certs/client.key"
    ca_file: "/certs/ca.crt"
  
  # Authentication (if broker requires auth)
  auth:
    username: "${MQTT_USERNAME}"
    password: "${MQTT_PASSWORD}"
  
  # Connection options
  connection:
    keep_alive: 60s
    clean_session: true
    connect_timeout: 30s
    reconnect_delay: 5s
    max_reconnect_attempts: 10
  
  # Message handling
  message_handling:
    buffer_size: 1000
    worker_pool_size: 5
    max_message_size: 10240  # 10KB max message size
    enable_persistence: true
    persistent_session: false
```

### Database Configuration
```yaml
database:
  host: "localhost"
  port: 5432
  user: "liwaisi-sis-admin"
  password: "${POSTGRES_PASSWORD:-postgres}"
  dbname: "smart-irrigation-system-db"
  sslmode: "disable"
  max_connections: 25
  max_idle_connections: 5
```

### Handler Configuration
```yaml
handlers:
  device_registration:
    workers: 2
    buffer_size: 50
  sensor_data:
    workers: 5
    buffer_size: 100
  command_response:
    workers: 3
    buffer_size: 50
  health_status:
    workers: 2
    buffer_size: 25
```

## Error Handling Strategy

### Retry Policy
- **Max Retries**: 3 attempts per message
- **Backoff Strategy**: Exponential backoff (1s, 2s, 4s)
- **Discard Policy**: After 3 failed attempts, log error and discard message

### Error Categories
1. **Validation Errors**: Immediate discard, log warning
2. **Database Errors**: Retry with backoff
3. **Business Logic Errors**: Retry once, then discard
4. **Infrastructure Errors**: Retry with full policy

## Enhanced Monitoring and Observability

### Application Metrics
```go
type ApplicationMetrics struct {
    // Message processing metrics
    MessagesProcessed    prometheus.Counter
    MessagesRetried      prometheus.Counter
    MessagesDiscarded    prometheus.Counter
    ProcessingDuration   prometheus.Histogram
    
    // Handler-specific metrics
    HandlerLatency      *prometheus.HistogramVec
    HandlerErrors       *prometheus.CounterVec
    ActiveWorkers       *prometheus.GaugeVec
    
    // Infrastructure metrics
    MQTTConnections     prometheus.Gauge
    DatabaseConnections prometheus.Gauge
    CircuitBreakerState *prometheus.GaugeVec
    
    // Business metrics
    DevicesOnline       prometheus.Gauge
    SensorReadingsRate  prometheus.Counter
    AlertsTriggered     prometheus.Counter
}
```

### Health Checks
```go
type HealthChecker struct {
    mqttClient   mqtt.Client
    database     *gorm.DB
    handlers     map[string]MessageHandler
    lastActivity map[string]time.Time
}

type HealthStatus struct {
    Status      string                 `json:"status"`
    Timestamp   time.Time             `json:"timestamp"`
    Services    map[string]ServiceHealth `json:"services"`
    Metrics     HealthMetrics         `json:"metrics"`
}

type ServiceHealth struct {
    Status      string    `json:"status"`
    LastCheck   time.Time `json:"last_check"`
    Error       string    `json:"error,omitempty"`
    Latency     string    `json:"latency,omitempty"`
}
```

## Security Implementation

### Message-Level Security
```go
type SecureMessageProcessor struct {
    encryptor    MessageEncryptor
    authenticator DeviceAuthenticator
    validator    MessageValidator
}

type MessageSignature struct {
    DeviceID   string    `json:"device_id"`
    Timestamp  time.Time `json:"timestamp"`
    Signature  string    `json:"signature"`
    Algorithm  string    `json:"algorithm"`
}

func (smp *SecureMessageProcessor) ProcessSecureMessage(msg *Message) error {
    // 1. Validate message signature
    if err := smp.authenticator.ValidateSignature(msg); err != nil {
        return fmt.Errorf("invalid signature: %w", err)
    }
    
    // 2. Check timestamp freshness (prevent replay attacks)
    if err := smp.validator.ValidateTimestamp(msg.Timestamp); err != nil {
        return fmt.Errorf("invalid timestamp: %w", err)
    }
    
    // 3. Decrypt message if needed
    if encrypted, ok := msg.Headers["encrypted"]; ok && encrypted == "true" {
        decryptedData, err := smp.encryptor.Decrypt(msg.Data)
        if err != nil {
            return fmt.Errorf("decryption failed: %w", err)
        }
        msg.Data = decryptedData
    }
    
    return nil
}
```

### Device Authentication
```go
type DeviceAuthenticator interface {
    ValidateDevice(deviceID string, token string) error
    ValidateSignature(msg *Message) error
    RotateDeviceKeys(deviceID string) error
}

type JWTDeviceAuth struct {
    publicKeys  map[string]*rsa.PublicKey
    keyRotation time.Duration
}
```

## Performance Optimizations

### Batch Processing Implementation
```go
type BatchProcessor struct {
    batchSize    int
    flushTimeout time.Duration
    buffer       []SensorReading
    repo         SensorReadingRepository
    mu           sync.Mutex
    timer        *time.Timer
    done         chan struct{}
}

func (bp *BatchProcessor) Start() {
    bp.timer = time.NewTimer(bp.flushTimeout)
    go bp.flushLoop()
}

func (bp *BatchProcessor) Add(reading SensorReading) {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    bp.buffer = append(bp.buffer, reading)
    
    if len(bp.buffer) >= bp.batchSize {
        bp.flushBuffer()
    }
}

func (bp *BatchProcessor) flushLoop() {
    for {
        select {
        case <-bp.timer.C:
            bp.mu.Lock()
            if len(bp.buffer) > 0 {
                bp.flushBuffer()
            }
            bp.timer.Reset(bp.flushTimeout)
            bp.mu.Unlock()
        case <-bp.done:
            return
        }
    }
}
```

### Memory Pool Implementation
```go
var (
    messagePool = sync.Pool{
        New: func() interface{} {
            return &Message{
                Headers: make(map[string]string, 8),
                Data:    make([]byte, 0, 1024),
            }
        },
    }
    
    sensorReadingPool = sync.Pool{
        New: func() interface{} {
            return &SensorReading{}
        },
    }
)

func GetMessage() *Message {
    return messagePool.Get().(*Message)
}

func PutMessage(msg *Message) {
    // Reset message fields
    msg.Subject = ""
    msg.Data = msg.Data[:0]
    msg.DeviceID = ""
    msg.Region = ""
    msg.Zone = ""
    msg.MessageType = ""
    msg.SequenceID = 0
    msg.Timestamp = time.Time{}
    
    // Clear headers map efficiently
    for k := range msg.Headers {
        delete(msg.Headers, k)
    }
    
    messagePool.Put(msg)
}
```

## Testing Strategy

### Unit Testing
```go
func TestSensorDataHandler_Handle(t *testing.T) {
    // Arrange
    mockRepo := &MockSensorRepository{}
    mockValidator := &MockValidator{}
    handler := NewSensorDataHandler(mockRepo, mockValidator)
    
    testMessage := &Message{
        Topic:    "/liwaisi/iot/smart-irrigation/sensor/temperature",
        DeviceID: "A0:A3:B3:AB:2F:D8",
        Data:     []byte(`{"value": 25.5, "unit": "celsius", "mac_address": "A0:A3:B3:AB:2F:D8"}`),
    }
    
    // Act
    err := handler.Handle(context.Background(), testMessage)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockRepo.SaveCallCount())
}
```

### Integration Testing
```go
func TestMessageRouter_Integration(t *testing.T) {
    // Setup test NATS server
    server := natstest.RunDefaultServer()
    defer server.Shutdown()
    
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create router with real dependencies
    router := NewMessageRouter(server.ClientURL(), db)
    
    // Test message processing end-to-end
    testIntegration(t, router)
}
```

## Future Considerations

### Scalability Enhancements
- **Horizontal Scaling**: Consumer groups with auto-scaling based on message backlog
- **Geo-distributed Processing**: Regional consumer instances for global IoT deployments  
- **Message Partitioning**: Smart partitioning by device zones for parallel processing
- **Caching Layer**: Redis integration for frequently accessed device states

### Advanced Monitoring
- **Distributed Tracing**: OpenTelemetry integration for request tracing across services
- **Anomaly Detection**: ML-based detection of unusual sensor patterns or device behavior
- **Predictive Alerting**: Early warning system for device failures based on health trends
- **Real-time Dashboards**: Grafana dashboards with real-time IoT metrics

### Enhanced Security
- **Zero-Trust Architecture**: Device identity verification for every message
- **Message Encryption**: End-to-end encryption between ESP32 and consumer
- **Audit Logging**: Comprehensive security event logging and monitoring
- **Device Lifecycle Management**: Automated key rotation and certificate management

### Advanced Features
- **Event Sourcing**: Complete audit trail of all device state changes
- **CQRS Implementation**: Separate read/write models for optimal performance
- **Stream Processing**: Real-time analytics on sensor data streams
- **Machine Learning Integration**: Automated irrigation decisions based on sensor data patterns