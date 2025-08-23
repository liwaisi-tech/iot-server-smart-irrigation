"""Port for agent registry and management."""

from abc import ABC, abstractmethod
from typing import Dict, List, Optional

from src.domain.entities.agent import Agent
from src.domain.value_objects.agent_types import AgentType, AgentCapability


class AgentRegistryPort(ABC):
    """Port for managing and discovering agents."""
    
    @abstractmethod
    async def register_agent(self, agent: Agent) -> None:
        """Register an agent in the system."""
        pass
    
    @abstractmethod
    async def get_agent(self, agent_type: AgentType) -> Optional[Agent]:
        """Get an agent by type."""
        pass
    
    @abstractmethod
    async def get_agents_by_capability(self, capability: AgentCapability) -> List[Agent]:
        """Get all agents that have a specific capability."""
        pass
    
    @abstractmethod
    async def get_all_agents(self) -> Dict[AgentType, Agent]:
        """Get all registered agents."""
        pass
    
    @abstractmethod
    async def get_coordinator_agent(self) -> Optional[Agent]:
        """Get the coordinator agent."""
        pass
    
    @abstractmethod
    async def is_agent_available(self, agent_type: AgentType) -> bool:
        """Check if an agent is available and healthy."""
        pass