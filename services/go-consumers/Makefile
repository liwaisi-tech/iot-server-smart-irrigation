.PHONY: help install test test-coverage test-watch clean build run migrate mocks mocks-install mock-interface

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

test: ## Run all tests
	go test ./... -coverprofile=coverage.out

test-coverage: ## Run tests with coverage report
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-watch: ## Watch tests and coverage continuously
	@echo "Watching for changes... Press Ctrl+C to stop"
	@while true; do \
		clear; \
		echo "Running tests with coverage..."; \
		go test -cover ./...; \
		echo ""; \
		echo "Waiting for changes... (Press Ctrl+C to stop)"; \
		sleep 2; \
	done

mocks-install: ## Install mockery tool
	@which mockery > /dev/null || (echo "Installing mockery..." && go install github.com/vektra/mockery/v3@v3.5.1)

mocks: ## Generate mocks using mockery
	mockery

mock-interface: ## Generate mock for specific interface (usage: make mock-interface INTERFACE=InterfaceName)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required"; \
		echo "Usage: make mock-interface name=InterfaceName"; \
		echo "Examples:"; \
		echo "  make mock-interface name=DeviceRepository"; \
		echo "  make mock-interface name=SensorService"; \
		exit 1; \
	fi
	mockery --name=$(name)

clean: ## Clean build artifacts and coverage files
	rm -f bin/consumer bin/migrate
	rm -f coverage.out coverage.html
	rm -rf mocks/
	go clean ./...

build: ## Build binaries
	mkdir -p bin
	go build -o bin/consumer cmd/consumer/main.go

run: ## Run the consumer service
	go run cmd/consumer/main.go

dev: ## Start development environment with docker services
	docker-compose up nats postgres pgbouncer -d

dev-logs: ## View development environment logs
	docker-compose logs -f nats postgres pgbouncer

dev-stop: ## Stop development environment
	docker-compose down