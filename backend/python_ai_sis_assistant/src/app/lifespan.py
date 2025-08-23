from contextlib import asynccontextmanager
from typing import AsyncGenerator
import logging

from fastapi import FastAPI

from src.app.container import Container
from src.domain.value_objects.agent_types import AgentType

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """FastAPI lifespan management for startup and shutdown events."""
    container: Container = app.state.container
    
    # Startup
    logger.info("Starting IoT Smart Irrigation AI Assistant...")
    
    try:
        # Initialize and register agents
        await _initialize_agents(container)
        logger.info("All agents initialized successfully")
        
        yield
        
    except Exception as e:
        logger.error(f"Error during startup: {e}")
        raise
    finally:
        # Shutdown
        logger.info("Shutting down IoT Smart Irrigation AI Assistant...")


async def _initialize_agents(container: Container) -> None:
    """Initialize and register all agents."""
    agent_registry = container.agent_registry()
    
    # Get agent instances
    peja_coordinator = container.peja_coordinator()
    local_assistant = container.local_assistant()
    web_search_agent = container.web_search_agent()
    
    # Register agents
    await agent_registry.register_agent(peja_coordinator)
    await agent_registry.register_agent(local_assistant)
    await agent_registry.register_agent(web_search_agent)
    
    logger.info("Registered agents: Peja Coordinator, Local Assistant, Web Search Agent")