"""Tests for health check endpoints."""

def test_health_endpoint(client):
    """Test basic health check endpoint."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert "timestamp" in data
    assert "version" in data
    assert "environment" in data


def test_detailed_health_endpoint(client):
    """Test detailed health check endpoint."""
    response = client.get("/health/detailed")
    assert response.status_code == 200
    data = response.json()
    assert "status" in data
    assert "components" in data
    assert "configuration" in data["components"]
    assert "logging" in data["components"]


def test_ping_endpoint(client):
    """Test ping endpoint."""
    response = client.get("/ping")
    assert response.status_code == 200
    data = response.json()
    assert data["message"] == "pong"


def test_ready_endpoint(client):
    """Test readiness probe endpoint."""
    response = client.get("/ready")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ready"


def test_live_endpoint(client):
    """Test liveness probe endpoint."""
    response = client.get("/live")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "alive"