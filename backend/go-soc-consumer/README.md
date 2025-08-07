# IoT Smart Irrigation - Go SOC Consumer

This repository contains the backend service for the IoT Smart Irrigation System, responsible for consuming and processing messages from IoT devices. This service is built with Go and follows clean architecture principles to ensure a modular, scalable, and maintainable codebase.

## Overview

The Go SOC (System on Chip) Consumer is a critical component that listens for events from IoT devices, such as registration requests and health status updates. It processes these events and interacts with other parts of the system, like the database, to keep the device registry and status up-to-date.

## Features

-   **Device Registration:** Handles the registration of new IoT devices into the system.
-   **Device Health Monitoring:** Processes health check messages from devices to monitor their status.
-   **MQTT Integration:** Consumes messages from an MQTT broker.
-   **Clean Architecture:** Organized into distinct layers (`domain`, `usecases`, `infrastructure`, `presentation`) for separation of concerns.

## Project Structure

The project follows a standard Go project layout with a clean architecture approach:

```
.
├── cmd/                # Application entry point
├── internal/           # Private application and library code
│   ├── app/            # Application-specific logic
│   ├── domain/         # Core domain entities and business rules
│   ├── infrastructure/ # External concerns (database, messaging)
│   ├── presentation/   # API definitions and transport layer
│   └── usecases/       # Application-specific use cases
├── pkg/                # Public library code
├── mocks/              # Generated mocks for testing
├── go.mod              # Go module definition
└── Makefile            # Project commands
```

## Getting Started

### Prerequisites

-   [Go](https://golang.org/doc/install) (version 1.23 or higher)
-   [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/)
-   [make](https://www.gnu.org/software/make/)

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/liwaisi-tech/iot-server-smart-irrigation.git
    cd iot-server-smart-irrigation/backend/go-soc-consumer
    ```

2.  **Set up environment variables:**

    Copy the example environment file and update it with your configuration details.

    ```sh
    cp .env.example .env
    ```

3.  **Install dependencies:**

    ```sh
    go mod tidy
    ```

### Running the Application

To run the application, use the following `make` command:

```sh
make run
```

This will start the service, which will then connect to the configured MQTT broker and start consuming messages.

## Usage

### Makefile Commands

The `Makefile` includes several useful commands for development and testing:

-   `make help`: Displays a help message with all available commands.
-   `make build`: Compiles the application.
-   `make run`: Runs the application.
-   `make test`: Runs the test suite.
-   `make clean`: Cleans up build artifacts.
-   `make check-linter`: Runs the static code analyzer.
-   `make dev-info`: Shows development environment setup instructions.

### Configuration

The application is configured using environment variables. See the `.env.example` file for a complete list of available options. Key configuration areas include:

-   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: PostgreSQL database connection details.
-   `MQTT_BROKER`, `MQTT_CLIENT_ID`, `MQTT_USER`, `MQTT_PASSWORD`: MQTT broker connection details.
-   `LOG_LEVEL`: Logging level (e.g., `debug`, `info`, `warn`, `error`).

## Testing

To run the full suite of tests, use:

```sh
make test
```

### Prerequisites

- Go 1.23+
- PostgreSQL 15+ (available via root docker-compose.yml)
- Docker & Docker Compose (for running infrastructure services)

### Commands

```bash
# Build
make build

# Run the application
make run

# Run tests
make test

# Show development environment info
make dev-info

# Clean build artifacts
make clean
```

### Database Migrations

Migrations are automatically applied when the application starts with PostgreSQL using GORM.


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

## Repository Implementation

### Memory Repository
- In-memory storage using Go maps
- Thread-safe with mutex protection
- Fast access for development and testing
- Data persists only during application runtime

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

This project is part of the IoT Smart Irrigation system and is licensed under the Apache License 2.0. See the LICENSE file for details.