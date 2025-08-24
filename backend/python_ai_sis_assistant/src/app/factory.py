from fastapi import FastAPI
from dependency_injector.wiring import inject, Provide

from src.app.container import Container
from src.app.lifespan import lifespan
from src.config.settings import Settings
from src.presentation.http.handlers import ping_handler


@inject
def create_app(
    settings: Settings = Provide[Container.settings],
) -> FastAPI:
    """Create and configure the FastAPI application."""
    container = Container()
    container.config.from_dict({})
    
    app = FastAPI(
        title="IoT Smart Irrigation AI Assistant",
        description="Python-based AI Assistant for IoT Smart Irrigation System with hexagonal architecture",
        version="1.0.0",
        lifespan=lifespan,
    )
    
    # Store container in app state
    app.state.container = container
    
    # Configure dependency injection
    container.wire(modules=["src.app.factory"])
    
    # Include routers
    app.include_router(ping_handler.router, tags=["ping"])
    
    return app


def get_app() -> FastAPI:
    """Get configured FastAPI application instance."""
    return create_app()