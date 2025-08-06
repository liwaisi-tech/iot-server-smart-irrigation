# IoT Smart Irrigation - Go SOC Consumer

This is a Go-based message consumer for the IoT Smart Irrigation system. It handles device registration messages via MQTT and provides a REST API for device management.

## Features

- **Dual Repository Support**: Choose between in-memory or PostgreSQL storage
- **MQTT Message Consumption**: Handles device registration messages
- **REST API**: Ping endpoint for health checks
- **Hexagonal Architecture**: Clean separation of concerns with domain-driven design
- **Docker Support**: Complete Docker Compose setup for development
- **Comprehensive Testing**: Unit and integration tests included
- **Database Migrations**: Automated PostgreSQL schema management

## Quick Start

### Using In-Memory Repository (Default)

```bash
# Clone and build
go build -o bin/iot-consumer ./cmd/server

# Run with default in-memory storage
./bin/iot-consumer
```

### Using PostgreSQL Repository

1. **Start PostgreSQL from project root:**
   ```bash
   # From the project root directory
   cd ../../
   docker-compose up -d postgres pgbouncer nats
   ```

2. **Run with PostgreSQL:**
   ```bash
   # From go-soc-consumer directory
   DB_TYPE=postgres go run ./cmd/server
   ```

## Configuration

The application uses environment variables for configuration. Copy `.env.example` to `.env` and modify as needed:

```bash
cp .env.example .env
```

**Important**: This application uses the PostgreSQL service configured in the root `docker-compose.yml`. Make sure to set the `POSTGRES_PASSWORD` environment variable that matches your root configuration.

### Key Environment Variables

- `DB_TYPE`: Repository type (`memory` or `postgres`)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL connection details
- `MQTT_BROKER_URL`: MQTT broker connection string
- `HTTP_PORT`: HTTP server port (default: 8080)
- `POSTGRES_PASSWORD`: Password for the PostgreSQL service (configured in root docker-compose.yml)

## API Endpoints

- `GET /ping` - Health check endpoint

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL 15+ (available via root docker-compose.yml)
- Docker & Docker Compose (for running infrastructure services)

### Commands

```bash
# Build
make build

# Run with memory repository
make run

# Run with PostgreSQL
make run-postgres

# Run tests
make test

# Run PostgreSQL integration tests (requires root PostgreSQL service)
make test-postgres

# Show development environment info
make dev-info

# Clean build artifacts
make clean
```

### Database Migrations

Migrations are automatically applied when the application starts with PostgreSQL. Manual migration commands (requires root PostgreSQL service running):

```bash
# Apply migrations
POSTGRES_PASSWORD=yourpassword make migrate-up

# Rollback migrations
POSTGRES_PASSWORD=yourpassword make migrate-down

# Check migration version
POSTGRES_PASSWORD=yourpassword make migrate-version
```

## Architecture

This project follows Hexagonal Architecture (Ports and Adapters) principles:

```
cmd/server/          # Application entry point
internal/
  domain/            # Business domain layer
    entities/        # Domain entities
    errors/          # Domain-specific errors
    ports/           # Repository interfaces
  infrastructure/    # Infrastructure layer
    database/        # Database connection management
    messaging/       # MQTT consumer implementation
    persistence/     # Repository implementations
  presentation/      # Presentation layer
    http/           # HTTP handlers
  usecases/         # Application use cases
pkg/                # Shared packages
  config/           # Configuration management
migrations/         # Database migration files
```

## Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/domain/entities
```

### Integration Tests

PostgreSQL integration tests require the root PostgreSQL service to be running:

```bash
# From project root, start PostgreSQL service
cd ../../
docker-compose up -d postgres

# From go-soc-consumer directory, run integration tests
POSTGRES_PASSWORD=yourpassword make test-postgres
```

## MQTT Message Format

Device registration messages should be published to topic `/liwaisi/iot/smart-irrigation/device/registration` with JSON payload:

```json
{
  "mac_address": "AA:BB:CC:DD:EE:FF",
  "device_name": "Sensor Node 1",
  "ip_address": "192.168.1.100",
  "location_description": "Garden Zone A"
}
```

## Repository Implementations

### Memory Repository
- In-memory storage using Go maps
- Thread-safe with mutex protection
- Suitable for development and testing

### PostgreSQL Repository
- Persistent storage with PostgreSQL
- ACID compliance and transactions
- Connection pooling and proper error handling
- Prepared statements for security

## Infrastructure Services

The root Docker Compose configuration provides:

- **postgres**: PostgreSQL 15 with pgvector (port 5432)
- **pgbouncer**: Connection pooler for PostgreSQL (port 6432)
- **nats**: NATS server with MQTT support (ports 4222, 1883, 8222)

To start these services from the project root:
```bash
docker-compose up -d postgres pgbouncer nats
```

## Contributing

1. Follow Go conventions and best practices
2. Write tests for new functionality
3. Use the existing error handling patterns
4. Update documentation as needed

## License

This project is part of the IoT Smart Irrigation system.