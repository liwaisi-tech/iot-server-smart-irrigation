# Python AI SIS Assistant Makefile

.PHONY: help install install-dev test test-cov lint format check clean run dev setup

help: ## Show this help message
	@echo "Python AI SIS Assistant - Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Set up development environment
	@./scripts/setup_dev.sh

install: ## Install production dependencies
	@uv sync

install-dev: ## Install development dependencies
	@uv sync --dev

test: ## Run tests
	@uv run pytest

test-cov: ## Run tests with coverage
	@./scripts/run_tests.sh

lint: ## Run linting
	@uv run ruff check src/ main.py

format: ## Format code
	@./scripts/format_code.sh

check: ## Run all code quality checks
	@./scripts/check_quality.sh

clean: ## Clean up build artifacts
	@rm -rf .pytest_cache/
	@rm -rf htmlcov/
	@rm -rf .coverage
	@rm -rf .mypy_cache/
	@rm -rf .ruff_cache/
	@find . -type d -name __pycache__ -exec rm -rf {} +
	@find . -type f -name "*.pyc" -delete

run: ## Run the application
	@uv run python main.py

dev: ## Run development server with auto-reload
	@./scripts/run_dev.sh

build: ## Build the application (placeholder)
	@echo "Build functionality to be implemented"

deploy: ## Deploy the application (placeholder)
	@echo "Deploy functionality to be implemented"