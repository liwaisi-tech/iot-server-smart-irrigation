# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Build and Run
- `make build` - Build the application binary to `bin/iot-consumer`
- `make run` - Run the application locally with `go run ./cmd/server`
- `go build -o bin/iot-consumer ./cmd/server` - Direct build command

### Testing
- `make test` - Run all unit tests with race detection and coverage
- `go test -v ./...` - Run all tests verbosely
- `go test -v ./internal/domain/entities` - Run specific package tests
- Generated mocks are in `mocks/` directory using testify/mock

### Mock Generation
- Mocks are generated using mockery and located in `mocks/` directory
- Key interfaces: `DeviceRepository`, `MessageConsumer`, `DeviceRegistrationUseCase`, `PingUseCase`

### Dependencies
- `go mod tidy` - Clean up dependencies
- `go mod download` - Download dependencies

## Architecture Overview

This project implements **Hexagonal Architecture** (Ports and Adapters) with Domain-Driven Design:

### Core Structure
```
internal/
├── domain/           # Business logic layer (entities, ports, errors)
├── usecases/         # Application use cases
├── infrastructure/   # External concerns (database, messaging, mappers)
└── presentation/     # HTTP handlers and external interfaces
```

### Key Domain Concepts
- **Device Entity**: Thread-safe IoT device with MAC address as identifier
- **Device Repository**: GORM PostgreSQL implementation with thread safety
- **MQTT Consumer**: Eclipse Paho MQTT client for device registration messages
- **Device Registration Use Case**: Handles create/update logic for device registration

### Database Integration
- Uses GORM with PostgreSQL driver
- Auto-migrations run on startup
- Connection pooling and configuration via environment variables
- Repository pattern implementation with context support

### Message Processing
- MQTT topic: `/liwaisi/iot/smart-irrigation/device/registration`
- JSON payload with device registration data (MAC, name, IP, location)
- Message handlers process registration with validation and persistence

## Configuration

### Environment Variables
- `MQTT_BROKER_URL` - MQTT broker connection (default: tcp://localhost:1883)
- `MQTT_CLIENT_ID` - MQTT client identifier (default: iot-go-soc-consumer)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `HTTP_PORT` - HTTP server port (hardcoded to 8080)

### Infrastructure Dependencies
The application expects these services (configured in root docker-compose.yml):
- PostgreSQL on port 5432
- NATS/MQTT broker on ports 1883 (MQTT) and 4222 (NATS)
- PgBouncer connection pooler on port 6432

## Testing Patterns

### Unit Testing
- Uses testify/assert and testify/require
- Comprehensive validation testing for domain entities
- Mock-based testing for repositories and use cases
- Table-driven tests for comprehensive coverage

### Repository Testing
- Separate test suite for PostgreSQL integration tests
- Uses DATA-DOG/go-sqlmock for database mocking
- Thread-safe operations testing

### Message Handler Testing  
- Mock MQTT clients for integration testing
- Context-aware message processing validation
- Error handling and timeout scenarios

## Development Notes

### Key Patterns
- Thread-safe domain entities with sync.RWMutex
- Context propagation throughout all operations
- Structured logging with slog package
- Graceful shutdown with signal handling
- Interface segregation with ports pattern

### Code Generation
- Mocks are generated, not hand-written
- Auto-migration handled by GORM on startup
- Environment-based configuration with sensible defaults

### Error Handling
- Domain-specific errors with structured messages
- Validation errors bubble up from entities
- Repository errors wrapped with context
- Graceful degradation for missing dependencies