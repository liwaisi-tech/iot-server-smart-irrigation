"""Agent management routes."""

from typing import List
from fastapi import APIRouter, Depends, Query
from dependency_injector.wiring import inject, Provide

from src.app.container import Container
from src.presentation.http.handlers.agent_handler import (
    AgentHandler, AgentInfo, SystemStatusResponse
)


router = APIRouter(prefix="/agents", tags=["agents"])


@router.get("/status", response_model=SystemStatusResponse)
@inject
async def get_system_status(
    agent_handler: AgentHandler = Depends(Provide[Container.agent_handler])
) -> SystemStatusResponse:
    """Get overall system status."""
    return await agent_handler.get_system_status()


@router.post("/initialize")
@inject
async def initialize_system(
    agent_handler: AgentHandler = Depends(Provide[Container.agent_handler])
) -> dict:
    """Initialize all agents in the system."""
    return await agent_handler.initialize_system()


@router.get("/by-capability/{capability}", response_model=List[AgentInfo])
@inject
async def get_agents_by_capability(
    capability: str,
    agent_handler: AgentHandler = Depends(Provide[Container.agent_handler])
) -> List[AgentInfo]:
    """Get agents that have a specific capability."""
    return await agent_handler.get_agents_by_capability(capability)


@router.get("/capabilities", response_model=List[str])
@inject
async def get_available_capabilities(
    agent_handler: AgentHandler = Depends(Provide[Container.agent_handler])
) -> List[str]:
    """Get list of all available capabilities in the system."""
    return await agent_handler.get_available_capabilities()