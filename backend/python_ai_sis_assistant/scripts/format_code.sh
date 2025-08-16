#!/bin/bash
# Code formatter script

set -e

echo "🎨 Formatting Python AI SIS Assistant code..."

# Check if virtual environment exists
if [ ! -d ".venv" ]; then
    echo "⚠️ Virtual environment not found. Please run ./scripts/setup_dev.sh first"
    exit 1
fi

echo "📐 Formatting code with Black..."
uv run black src/ main.py

echo "📦 Sorting imports with isort..."
uv run isort src/ main.py

echo "🔧 Auto-fixing linting issues with ruff..."
uv run ruff check --fix src/ main.py

echo "✨ Code formatting completed!"