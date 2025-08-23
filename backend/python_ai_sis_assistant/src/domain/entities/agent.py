"""Agent domain entity."""

from abc import ABC, abstractmethod
from typing import Dict, Any
from dataclasses import dataclass

from src.domain.value_objects.agent_types import AgentProfile, AgentCapability
from src.domain.value_objects.conversation import Message, ConversationContext


@dataclass
class AgentResponse:
    """Response from an agent."""
    content: str
    should_switch_agent: bool = False
    target_agent_type: str | None = None
    metadata: Dict[str, Any] | None = None


class Agent(ABC):
    """Base class for all agents in the system."""
    
    def __init__(self, profile: AgentProfile):
        self.profile = profile
    
    @property
    def name(self) -> str:
        """Get agent name."""
        return self.profile.name
    
    @property
    def agent_type(self) -> str:
        """Get agent type as string."""
        return self.profile.agent_type.value
    
    def has_capability(self, capability: AgentCapability) -> bool:
        """Check if agent has a specific capability."""
        return self.profile.can_handle_capability(capability)
    
    @abstractmethod
    async def process_message(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> AgentResponse:
        """Process a message and return a response."""
        pass
    
    @abstractmethod
    def can_handle_request(self, message: Message, context: ConversationContext) -> bool:
        """Determine if this agent can handle the given request."""
        pass


class CoordinatorAgent(Agent):
    """Coordinator agent that manages other agents and routes requests."""
    
    def __init__(self, profile: AgentProfile):
        if not profile.is_coordinator():
            raise ValueError("CoordinatorAgent requires a coordinator profile")
        super().__init__(profile)
    
    @abstractmethod
    async def route_request(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> str:
        """Route a request to the appropriate agent."""
        pass
    
    @abstractmethod
    async def should_switch_agent(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> tuple[bool, str | None]:
        """Determine if we should switch to another agent."""
        pass


class SpecializedAgent(Agent):
    """Base class for specialized agents."""
    
    def __init__(self, profile: AgentProfile):
        if profile.is_coordinator():
            raise ValueError("SpecializedAgent cannot use a coordinator profile")
        super().__init__(profile)