from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI

from src.app.container import Container


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """FastAPI lifespan management for startup and shutdown events."""
    container: Container = app.state.container
    
    # Startup
    try:
        yield
    finally:
        # Shutdown
        pass