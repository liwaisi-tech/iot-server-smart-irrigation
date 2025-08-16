"""Application factory for creating FastAPI app instance."""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.app.container import Container
from src.app.lifespan import lifespan
from src.config.settings import settings
from src.presentation.http.handlers.health_handler import HealthHandler
from src.presentation.http.handlers.ping_handler import PingHandler
from src.presentation.http.middleware.logging import LoggingMiddleware
from src.presentation.http.middleware.rate_limit import RateLimitMiddleware


def create_app() -> FastAPI:
    """Create and configure FastAPI application.
    
    Returns:
        FastAPI: Configured FastAPI application instance.
    """
    app = FastAPI(
        title="Python AI SIS Assistant",
        description="Conversational AI Agent for IoT Smart Irrigation System in Colombia",
        version="0.1.0",
        lifespan=lifespan,
        docs_url="/docs" if settings.debug else None,
        redoc_url="/redoc" if settings.debug else None,
    )
    
    # Add middleware
    _configure_middleware(app)
    
    # Initialize dependency container
    container = Container()
    
    # Initialize and include routers
    _configure_routes(app, container)
    
    return app


def _configure_middleware(app: FastAPI) -> None:
    """Configure application middleware."""
    
    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Rate limiting middleware
    app.add_middleware(RateLimitMiddleware)
    
    # Logging middleware (should be last to capture all requests)
    app.add_middleware(LoggingMiddleware)


def _configure_routes(app: FastAPI, container: Container) -> None:
    """Configure application routes."""
    
    # Initialize handlers
    health_handler = HealthHandler()
    ping_handler = PingHandler(container.ping_service)
    
    # Include routers
    app.include_router(health_handler.router, tags=["health"])
    app.include_router(ping_handler.router, tags=["legacy"])
    
    # Future: Add conversation, device, and other handlers