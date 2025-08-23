"""Agent management handler."""

import logging
from typing import List, Optional
from fastapi import HTTPException
from pydantic import BaseModel, Field

from src.usecases.agent_orchestrator import AgentOrchestrator, AgentHealthStatus
from src.domain.value_objects.agent_types import AgentType, AgentCapability


logger = logging.getLogger(__name__)


class AgentInfo(BaseModel):
    """Agent information model."""
    agent_type: str
    name: str
    description: str
    capabilities: List[str]
    is_available: bool


class AgentHealthResponse(BaseModel):
    """Agent health status response."""
    agent_type: str
    is_available: bool
    last_check: str
    error_message: Optional[str] = None


class SystemStatusResponse(BaseModel):
    """System status response."""
    total_agents: int
    available_agents: int
    agent_health: List[AgentHealthResponse]
    system_ready: bool


class AgentHandler:
    """HTTP handler for agent management."""
    
    def __init__(self, agent_orchestrator: AgentOrchestrator):
        self.agent_orchestrator = agent_orchestrator
    
    async def get_system_status(self) -> SystemStatusResponse:
        """Get overall system status."""
        
        try:
            health_statuses = await self.agent_orchestrator.get_system_health()
            
            # Convert to response models
            agent_health = []
            available_count = 0
            
            for status in health_statuses:
                if status.is_available:
                    available_count += 1
                
                agent_health.append(AgentHealthResponse(
                    agent_type=status.agent_type.value,
                    is_available=status.is_available,
                    last_check=status.last_check,
                    error_message=status.error_message
                ))
            
            system_ready = available_count > 0  # At least one agent available
            
            logger.info(f"System status: {available_count}/{len(health_statuses)} agents available")
            
            return SystemStatusResponse(
                total_agents=len(health_statuses),
                available_agents=available_count,
                agent_health=agent_health,
                system_ready=system_ready
            )
            
        except Exception as e:
            logger.error(f"Error getting system status: {e}")
            raise HTTPException(status_code=500, detail="Error obteniendo estado del sistema")
    
    async def initialize_system(self) -> dict:
        """Initialize all agents in the system."""
        
        try:
            initialization_results = await self.agent_orchestrator.initialize_agents()
            
            success_count = sum(1 for success in initialization_results.values() if success)
            total_count = len(initialization_results)
            
            logger.info(f"Agent initialization: {success_count}/{total_count} successful")
            
            return {
                "message": "Sistema inicializado",
                "total_agents": total_count,
                "successful_initializations": success_count,
                "results": {
                    agent_type.value: success 
                    for agent_type, success in initialization_results.items()
                }
            }
            
        except Exception as e:
            logger.error(f"Error initializing system: {e}")
            raise HTTPException(status_code=500, detail="Error inicializando sistema")
    
    async def get_agents_by_capability(self, capability: str) -> List[AgentInfo]:
        """Get agents that have a specific capability."""
        
        try:
            # Validate capability
            try:
                agent_capability = AgentCapability(capability)
            except ValueError:
                raise HTTPException(
                    status_code=400, 
                    detail=f"Capacidad no válida: {capability}"
                )
            
            # Get agents with capability
            agents = await self.agent_orchestrator.get_agents_by_capability(agent_capability)
            
            # Convert to response models
            agent_infos = []
            for agent in agents:
                # Check if agent is available
                try:
                    agent_type = AgentType(agent.agent_type)
                    is_available = await self.agent_orchestrator.agent_registry.is_agent_available(agent_type)
                except:
                    is_available = False
                
                agent_infos.append(AgentInfo(
                    agent_type=agent.agent_type,
                    name=agent.name,
                    description=agent.profile.description,
                    capabilities=[cap.value for cap in agent.profile.capabilities],
                    is_available=is_available
                ))
            
            logger.info(f"Found {len(agent_infos)} agents with capability: {capability}")
            return agent_infos
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Error getting agents by capability: {e}")
            raise HTTPException(status_code=500, detail="Error obteniendo agentes")
    
    async def get_available_capabilities(self) -> List[str]:
        """Get list of all available capabilities in the system."""
        
        try:
            # Return all defined capabilities
            capabilities = [capability.value for capability in AgentCapability]
            
            logger.debug(f"Retrieved {len(capabilities)} available capabilities")
            return capabilities
            
        except Exception as e:
            logger.error(f"Error getting capabilities: {e}")
            raise HTTPException(status_code=500, detail="Error obteniendo capacidades")