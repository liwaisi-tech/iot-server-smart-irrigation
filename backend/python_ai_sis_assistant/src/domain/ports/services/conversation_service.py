"""Conversation service port interface."""

from abc import ABC, abstractmethod
from typing import Any, Optional

from src.domain.entities.conversation import ConversationContext


class ConversationService(ABC):
    """Port interface for conversation processing."""
    
    @abstractmethod
    async def process_conversation(
        self, 
        message: str, 
        context: ConversationContext
    ) -> dict[str, Any]:
        """Process user message and return conversation response."""
        pass
    
    @abstractmethod
    async def extract_intent(self, message: str) -> Optional[str]:
        """Extract user intent from message."""
        pass
    
    @abstractmethod
    async def extract_entities(self, message: str) -> list[dict[str, Any]]:
        """Extract entities (devices, locations, etc.) from message."""
        pass
    
    @abstractmethod
    async def generate_response(
        self,
        intent: str,
        entities: list[dict[str, Any]],
        context: ConversationContext,
        data: dict[str, Any]
    ) -> str:
        """Generate natural language response."""
        pass
    
    @abstractmethod
    async def health_check(self) -> bool:
        """Check if conversation service is available."""
        pass