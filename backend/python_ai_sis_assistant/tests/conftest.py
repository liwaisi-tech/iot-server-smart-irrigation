"""Test configuration and fixtures."""

import pytest
from fastapi.testclient import TestClient

from src.app.factory import create_app


@pytest.fixture
def app():
    """Create FastAPI app for testing."""
    return create_app()


@pytest.fixture
def client(app):
    """Create test client."""
    return TestClient(app)