#!/bin/bash
# Code quality checker script

set -e

echo "🎨 Checking Python AI SIS Assistant code quality..."

# Check if virtual environment exists
if [ ! -d ".venv" ]; then
    echo "⚠️ Virtual environment not found. Please run ./scripts/setup_dev.sh first"
    exit 1
fi

echo "🔍 Running code quality checks..."
echo ""

# Black formatting check
echo "📐 Checking code formatting with Black..."
if uv run black --check --diff src/ main.py; then
    echo "✅ Black formatting check passed"
else
    echo "❌ Black formatting check failed"
    echo "💡 Run: uv run black src/ main.py"
    exit 1
fi
echo ""

# isort import sorting check  
echo "📦 Checking import sorting with isort..."
if uv run isort --check-only --diff src/ main.py; then
    echo "✅ isort check passed"
else
    echo "❌ isort check failed"
    echo "💡 Run: uv run isort src/ main.py"
    exit 1
fi
echo ""

# Ruff linting
echo "🔍 Running linting with ruff..."
if uv run ruff check src/ main.py; then
    echo "✅ Ruff linting passed"
else
    echo "❌ Ruff linting failed"
    echo "💡 Run: uv run ruff check --fix src/ main.py"
    exit 1
fi
echo ""

# MyPy type checking
echo "🔬 Running type checking with mypy..."
if uv run mypy src/ main.py; then
    echo "✅ MyPy type checking passed"
else
    echo "❌ MyPy type checking failed"
    exit 1
fi
echo ""

echo "🎉 All code quality checks passed!"