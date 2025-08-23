"""Agent registry implementation."""

import logging
from typing import Dict, List, Optional

from src.domain.entities.agent import Agent
from src.domain.value_objects.agent_types import AgentType, AgentCapability
from src.domain.ports.agent_registry import AgentRegistryPort


logger = logging.getLogger(__name__)


class InMemoryAgentRegistry(AgentRegistryPort):
    """In-memory implementation of agent registry."""
    
    def __init__(self):
        self._agents: Dict[AgentType, Agent] = {}
    
    async def register_agent(self, agent: Agent) -> None:
        """Register an agent in the system."""
        agent_type = AgentType(agent.agent_type)
        self._agents[agent_type] = agent
        logger.info(f"Registered agent: {agent.name} ({agent_type.value})")
    
    async def get_agent(self, agent_type: AgentType) -> Optional[Agent]:
        """Get an agent by type."""
        agent = self._agents.get(agent_type)
        if agent:
            logger.debug(f"Retrieved agent: {agent_type.value}")
        else:
            logger.debug(f"Agent not found: {agent_type.value}")
        return agent
    
    async def get_agents_by_capability(self, capability: AgentCapability) -> List[Agent]:
        """Get all agents that have a specific capability."""
        matching_agents = []
        
        for agent in self._agents.values():
            if agent.has_capability(capability):
                matching_agents.append(agent)
        
        logger.debug(f"Found {len(matching_agents)} agents with capability: {capability.value}")
        return matching_agents
    
    async def get_all_agents(self) -> Dict[AgentType, Agent]:
        """Get all registered agents."""
        logger.debug(f"Retrieved all agents: {len(self._agents)} total")
        return self._agents.copy()
    
    async def get_coordinator_agent(self) -> Optional[Agent]:
        """Get the coordinator agent."""
        coordinator = self._agents.get(AgentType.COORDINATOR)
        if coordinator:
            logger.debug("Retrieved coordinator agent")
        else:
            logger.warning("Coordinator agent not found")
        return coordinator
    
    async def is_agent_available(self, agent_type: AgentType) -> bool:
        """Check if an agent is available and healthy."""
        agent = self._agents.get(agent_type)
        if not agent:
            return False
        
        try:
            # For now, just check if agent exists
            # In a real implementation, you might check health status
            logger.debug(f"Agent {agent_type.value} is available")
            return True
        except Exception as e:
            logger.error(f"Agent {agent_type.value} health check failed: {e}")
            return False
    
    def get_registered_agent_types(self) -> List[AgentType]:
        """Get list of all registered agent types."""
        return list(self._agents.keys())
    
    def get_agent_count(self) -> int:
        """Get total number of registered agents."""
        return len(self._agents)
    
    def clear_all(self) -> None:
        """Clear all registered agents (useful for testing)."""
        self._agents.clear()
        logger.debug("Cleared all registered agents")