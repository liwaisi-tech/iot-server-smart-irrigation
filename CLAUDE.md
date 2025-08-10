# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an IoT Smart Irrigation System with a Go-based backend service that follows hexagonal architecture principles. The system consumes MQTT messages from IoT devices and manages device registration, health monitoring, and sensor data processing.

## Common Development Commands

### Go SOC Consumer (Primary Backend Service)
Navigate to `backend/go-soc-consumer/` for all commands:

```bash
# Build the application
make build

# Run the application locally
make run

# Run tests with coverage
make test

# Run static code analysis
make check-linter

# Clean build artifacts
make clean

# Show development environment setup info
make dev-info
```

### Infrastructure Services
From project root:

```bash
# Start all infrastructure services (PostgreSQL, NATS, PgBouncer)
docker-compose up -d postgres nats

# Start specific service
docker-compose up -d postgres
docker-compose up -d nats

# Start the Go consumer service via Docker
docker-compose up -d go-soc-consumer

# View service logs
docker-compose logs -f go-soc-consumer
docker-compose logs -f postgres
docker-compose logs -f nats
```

## Architecture

The project follows Hexagonal Architecture (Ports and Adapters):

### Core Structure
- **Domain Layer** (`internal/domain/`): Business entities, domain errors, and ports (interfaces)
- **Use Cases Layer** (`internal/usecases/`): Application-specific business logic
- **Infrastructure Layer** (`internal/infrastructure/`): External concerns (database, messaging, HTTP)
- **Presentation Layer** (`internal/presentation/`): API handlers and transport layer
- **Application Layer** (`internal/app/`): Dependency injection and application startup

### Key Components
- **MQTT Consumer**: Listens to device messages on MQTT topics
- **NATS Publisher/Subscriber**: Event-driven communication between services  
- **PostgreSQL**: Primary data store with GORM for ORM
- **Device Registration**: Handles new IoT device onboarding
- **Device Health Monitoring**: Tracks device status and availability
- **Sensor Data Processing**: Manages temperature/humidity sensor readings

### Dependencies
- **GORM**: Database ORM with PostgreSQL driver
- **NATS**: Message broker for event-driven architecture
- **Paho MQTT**: MQTT client for IoT device communication
- **Zap**: Structured logging
- **UUID**: Unique identifier generation
- **Testify**: Testing framework with mocks

## Key Configuration

### Environment Variables
Set up via `.env` file in `backend/go-soc-consumer/`:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL connection
- `MQTT_BROKER_URL`: MQTT broker URL (default: `tcp://localhost:1883`)
- `NATS_URL`: NATS server URL (default: `nats://localhost:4222`)
- `LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`)

### Service Endpoints
- **HTTP Server**: `localhost:8080` (health checks, ping endpoint)
- **PostgreSQL**: `localhost:5432` (direct) / `localhost:6432` (via PgBouncer)
- **NATS**: `localhost:4222` (NATS) / `localhost:1883` (MQTT)
- **NATS Monitoring**: `localhost:8222` (HTTP interface)

## Testing

- Database migrations are automatically applied via GORM on startup
- Use `make test` to run the full test suite with race detection and coverage
- Mocks are generated for all interfaces and stored in `/mocks` directory
- Tests follow Go naming conventions with `_test.go` suffix

## MQTT Message Format

Device registration messages published to `/liwaisi/iot/smart-irrigation/device/registration`:

```json
{
  "mac_address": "AA:BB:CC:DD:EE:FF",
  "device_name": "Sensor Node 1", 
  "ip_address": "192.168.1.100",
  "location_description": "Garden Zone A"
}
```

## Development Workflow

1. Start infrastructure: `docker-compose up -d postgres nats`
2. Navigate to backend: `cd backend/go-soc-consumer`
3. Install dependencies: `go mod tidy`
4. Run tests: `make test`
5. Run linter: `make check-linter`
6. Start service: `make run`

## Code Quality

- Run `make check-linter` before committing changes
- Maintain test coverage with `make test`
- Follow existing error handling patterns using domain-specific errors
- Use the logger factory for structured logging with appropriate log levels