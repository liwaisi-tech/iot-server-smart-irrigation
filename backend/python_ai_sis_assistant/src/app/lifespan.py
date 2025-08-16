"""Application lifespan management."""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI

from src.infrastructure.logging import configure_logging, get_logger


@asynccontextmanager
async def lifespan(_app: FastAPI) -> AsyncGenerator[None, None]:
    """Manage application lifespan events."""
    logger = get_logger("lifespan")
    
    # Startup
    logger.info("Starting Python AI SIS Assistant")
    
    try:
        # Configure logging
        configure_logging()
        logger.info("Logging configured")
        
        # Future: Initialize database connections, cache, etc.
        logger.info("Application startup completed")
        
        yield
        
    finally:
        # Shutdown
        logger.info("Shutting down Python AI SIS Assistant")
        
        # Future: Cleanup resources, close connections, etc.
        logger.info("Application shutdown completed")