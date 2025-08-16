"""HTTP handler for health check endpoints."""

from datetime import datetime
from typing import Any

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from src.config.settings import settings
from src.infrastructure.logging import get_logger


class HealthResponse(BaseModel):
    """Health check response model."""
    status: str
    timestamp: datetime
    version: str
    environment: str


class DetailedHealthResponse(BaseModel):
    """Detailed health check response model."""
    status: str
    timestamp: datetime
    version: str
    environment: str
    components: dict[str, dict[str, Any]]
    

class HealthHandler:
    """HTTP handler for health check operations."""
    
    def __init__(self) -> None:
        """Initialize health handler."""
        self.router = APIRouter()
        self.logger = get_logger("health")
        self._setup_routes()
    
    def _setup_routes(self) -> None:
        """Set up HTTP routes for health operations."""
        
        @self.router.get("/health", response_model=HealthResponse, summary="Basic health check")
        async def health() -> HealthResponse:
            """Basic health check endpoint.
            
            Returns:
                HealthResponse: Basic health status
            """
            return HealthResponse(
                status="healthy",
                timestamp=datetime.utcnow(),
                version="0.1.0",
                environment=settings.environment
            )
        
        @self.router.get("/health/detailed", response_model=DetailedHealthResponse, summary="Detailed health check")
        async def detailed_health() -> DetailedHealthResponse:
            """Detailed health check endpoint with component status.
            
            Returns:
                DetailedHealthResponse: Detailed health status with components
            """
            components = await self._check_components()
            overall_status = "healthy" if all(
                comp["status"] == "healthy" for comp in components.values()
            ) else "unhealthy"
            
            return DetailedHealthResponse(
                status=overall_status,
                timestamp=datetime.utcnow(),
                version="0.1.0",
                environment=settings.environment,
                components=components
            )
        
        @self.router.get("/ping", summary="Simple ping endpoint")
        async def ping() -> dict[str, str]:
            """Simple ping endpoint.
            
            Returns:
                dict: Pong response
            """
            return {"message": "pong"}
        
        @self.router.get("/ready", summary="Readiness probe")
        async def ready() -> dict[str, str]:
            """Kubernetes readiness probe endpoint.
            
            Returns:
                dict: Ready status
            
            Raises:
                HTTPException: If service is not ready
            """
            components = await self._check_components()
            
            # Check critical components for readiness
            critical_components = ["configuration", "logging"]
            for component in critical_components:
                if components.get(component, {}).get("status") != "healthy":
                    self.logger.warning(f"Readiness check failed: {component} not healthy")
                    raise HTTPException(status_code=503, detail=f"Service not ready: {component}")
            
            return {"status": "ready"}
        
        @self.router.get("/live", summary="Liveness probe") 
        async def live() -> dict[str, str]:
            """Kubernetes liveness probe endpoint.
            
            Returns:
                dict: Live status
            """
            return {"status": "alive"}
    
    async def _check_components(self) -> dict[str, dict[str, Any]]:
        """Check the health of various application components.
        
        Returns:
            dict: Component health status
        """
        components = {}
        
        # Configuration check
        components["configuration"] = await self._check_configuration()
        
        # Logging check
        components["logging"] = await self._check_logging()
        
        # Future: Add database, cache, external services checks
        # components["database"] = await self._check_database()
        # components["cache"] = await self._check_cache()
        # components["adk_service"] = await self._check_adk_service()
        
        return components
    
    async def _check_configuration(self) -> dict[str, Any]:
        """Check configuration health."""
        try:
            # Validate that required settings are present
            required_settings = ["host", "port", "environment"]
            for setting in required_settings:
                if not hasattr(settings, setting):
                    return {
                        "status": "unhealthy",
                        "error": f"Missing required setting: {setting}"
                    }
            
            return {
                "status": "healthy",
                "details": {
                    "environment": settings.environment,
                    "debug": settings.debug,
                }
            }
        except Exception as e:
            return {
                "status": "unhealthy", 
                "error": str(e)
            }
    
    async def _check_logging(self) -> dict[str, Any]:
        """Check logging system health."""
        try:
            # Test that we can log
            test_logger = get_logger("health_test")
            test_logger.info("Health check logging test")
            
            return {
                "status": "healthy",
                "details": {
                    "log_level": settings.log_level,
                    "log_format": settings.log_format,
                }
            }
        except Exception as e:
            return {
                "status": "unhealthy",
                "error": str(e)
            }