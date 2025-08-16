#!/bin/bash
# Test runner script

set -e

echo "🧪 Running Python AI SIS Assistant tests..."

# Check if virtual environment exists
if [ ! -d ".venv" ]; then
    echo "⚠️ Virtual environment not found. Please run ./scripts/setup_dev.sh first"
    exit 1
fi

# Run tests with coverage
echo "📊 Running tests with coverage..."
uv run pytest \
    --cov=src \
    --cov-report=term-missing \
    --cov-report=html:htmlcov \
    --cov-fail-under=80 \
    -v \
    ${@}

echo ""
echo "✅ Tests completed!"
echo "📊 Coverage report generated in htmlcov/"