"""Value objects for agent types and capabilities."""

from enum import Enum
from typing import Set
from dataclasses import dataclass


class AgentType(Enum):
    """Types of agents in the system."""
    COORDINATOR = "coordinator"
    LOCAL_ASSISTANT = "local_assistant" 
    WEB_SEARCH = "web_search"


class AgentCapability(Enum):
    """Capabilities that agents can possess."""
    TASK_ROUTING = "task_routing"
    LOCAL_KNOWLEDGE = "local_knowledge"
    IOT_MANAGEMENT = "iot_management"
    WEB_SEARCH = "web_search"
    CONVERSATION_MANAGEMENT = "conversation_management"
    CONTEXT_SWITCHING = "context_switching"


@dataclass(frozen=True)
class AgentProfile:
    """Profile defining an agent's characteristics and capabilities."""
    agent_type: AgentType
    name: str
    description: str
    capabilities: Set[AgentCapability]
    
    def can_handle_capability(self, capability: AgentCapability) -> bool:
        """Check if agent has a specific capability."""
        return capability in self.capabilities
    
    def is_coordinator(self) -> bool:
        """Check if this is a coordinator agent."""
        return self.agent_type == AgentType.COORDINATOR


# Predefined agent profiles
PEJA_COORDINATOR_PROFILE = AgentProfile(
    agent_type=AgentType.COORDINATOR,
    name="Peja",
    description="Coordinador principal del sistema IoT de riego inteligente",
    capabilities={
        AgentCapability.TASK_ROUTING,
        AgentCapability.CONVERSATION_MANAGEMENT,
        AgentCapability.CONTEXT_SWITCHING,
    }
)

LOCAL_AGENT_PROFILE = AgentProfile(
    agent_type=AgentType.LOCAL_ASSISTANT,
    name="Asistente Local",
    description="Especialista en gestión local de dispositivos IoT y datos del sistema",
    capabilities={
        AgentCapability.LOCAL_KNOWLEDGE,
        AgentCapability.IOT_MANAGEMENT,
    }
)

WEB_SEARCH_AGENT_PROFILE = AgentProfile(
    agent_type=AgentType.WEB_SEARCH,
    name="Buscador Web",
    description="Especialista en búsqueda de información externa y documentación técnica",
    capabilities={
        AgentCapability.WEB_SEARCH,
    }
)