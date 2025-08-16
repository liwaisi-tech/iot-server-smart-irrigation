#!/bin/bash
# Development server startup script

set -e

echo "🚀 Starting Python AI SIS Assistant development server..."

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "⚠️ .env file not found. Please run ./scripts/setup_dev.sh first"
    exit 1
fi

# Check if virtual environment exists
if [ ! -d ".venv" ]; then
    echo "⚠️ Virtual environment not found. Please run ./scripts/setup_dev.sh first"
    exit 1
fi

# Set development environment
export ENVIRONMENT=development
export DEBUG=true
export LOG_FORMAT=console

echo "🔧 Environment: development"
echo "🐛 Debug mode: enabled"
echo "📋 Log format: console"
echo ""

# Start the server with reload
echo "🌐 Server will be available at:"
echo "  - Main app: http://localhost:8081"
echo "  - Health check: http://localhost:8081/health"
echo "  - API docs: http://localhost:8081/docs"
echo ""

uv run uvicorn main:app \
    --host 0.0.0.0 \
    --port 8081 \
    --reload \
    --reload-dir src \
    --log-level info