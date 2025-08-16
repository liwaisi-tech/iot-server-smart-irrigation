#!/bin/bash
# Development environment setup script

set -e

echo "🚀 Setting up Python AI SIS Assistant development environment..."

# Check if uv is installed
if ! command -v uv &> /dev/null; then
    echo "❌ UV package manager is not installed. Please install it first:"
    echo "   curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
fi

echo "✅ UV package manager found"

# Create virtual environment if it doesn't exist
if [ ! -d ".venv" ]; then
    echo "📦 Creating virtual environment..."
    uv venv --python 3.12
fi

echo "📥 Installing dependencies..."
uv sync --dev

echo "🔍 Installing pre-commit hooks..."
uv run pre-commit install

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "⚙️ Creating .env file from example..."
    cp .env.example .env
    echo "📝 Please edit .env file with your configuration"
fi

echo "🧪 Running initial tests..."
uv run pytest --version > /dev/null 2>&1 || echo "⚠️ Pytest not ready, will be available after sync"

echo "🎨 Checking code quality tools..."
echo "  - Black: $(uv run black --version 2>/dev/null || echo 'Not ready')"
echo "  - isort: $(uv run isort --version 2>/dev/null || echo 'Not ready')"
echo "  - ruff: $(uv run ruff --version 2>/dev/null || echo 'Not ready')"
echo "  - mypy: $(uv run mypy --version 2>/dev/null || echo 'Not ready')"

echo ""
echo "✨ Development environment setup complete!"
echo ""
echo "🔧 Next steps:"
echo "  1. Edit .env file with your configuration"
echo "  2. Start the development server: ./scripts/run_dev.sh"
echo "  3. Run tests: ./scripts/run_tests.sh"
echo "  4. Check code quality: ./scripts/check_quality.sh"
echo ""