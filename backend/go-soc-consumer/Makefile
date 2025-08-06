.PHONY: help build run test test-postgres clean migrate-up migrate-down

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build the application binary"
	@echo "  run             - Run the application locally"
	@echo "  test            - Run unit tests"
	@echo "  test-postgres   - Run PostgreSQL integration tests (requires root PostgreSQL service)"
	@echo "  clean           - Clean build artifacts"
	@echo "  migrate-up      - Run database migrations (requires root PostgreSQL service)"
	@echo "  migrate-down    - Rollback database migrations (requires root PostgreSQL service)"
	@echo "  dev-info        - Show development environment setup instructions"

# Build the application
build:
	go build -o bin/iot-consumer ./cmd/server

# Run the application locally
run:
	go run ./cmd/server

# Run unit tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run PostgreSQL integration tests (requires root PostgreSQL service running)
test-postgres:
	POSTGRES_TEST_DSN="postgres://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:5432/smart-irrigation-system-db?sslmode=disable" \
	go test -v -race -coverprofile=coverage.out ./internal/infrastructure/persistence

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean


# Migration targets (requires golang-migrate CLI and root PostgreSQL service running)
migrate-up:
	migrate -path migrations -database "postgres://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:5432/smart-irrigation-system-db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:5432/smart-irrigation-system-db?sslmode=disable" down

migrate-version:
	migrate -path migrations -database "postgres://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:5432/smart-irrigation-system-db?sslmode=disable" version

# Development helpers
dev-info:
	@echo "To start the development environment, run from the project root:"
	@echo "  docker-compose up -d postgres pgbouncer nats"
	@echo ""
	@echo "Available services:"
	@echo "  PostgreSQL: postgresql://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:5432/smart-irrigation-system-db"
	@echo "  PgBouncer:  postgresql://liwaisi-sis-admin:$(POSTGRES_PASSWORD)@localhost:6432/smart-irrigation-system-db"
	@echo "  NATS/MQTT:  nats://localhost:4222 (NATS) / mqtt://localhost:1883 (MQTT)"