from contextlib import asynccontextmanager
from typing import AsyncGenerator
import logging

from fastapi import FastAPI

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncGenerator[None, None]:
    """FastAPI lifespan management for startup and shutdown events."""
    
    # Startup
    logger.info("Starting IoT Smart Irrigation AI Assistant...")
    
    try:
        logger.info("Application started successfully")
        
        yield
        
    except Exception as e:
        logger.error(f"Error during startup: {e}")
        raise
    finally:
        # Shutdown
        logger.info("Shutting down IoT Smart Irrigation AI Assistant...")