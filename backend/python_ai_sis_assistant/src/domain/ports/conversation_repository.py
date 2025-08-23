"""Port for conversation persistence."""

from abc import ABC, abstractmethod
from typing import Optional, List

from src.domain.value_objects.conversation import ConversationId, ConversationContext, Message


class ConversationRepositoryPort(ABC):
    """Port for persisting and retrieving conversation data."""
    
    @abstractmethod
    async def save_conversation(self, context: ConversationContext) -> None:
        """Save a conversation context."""
        pass
    
    @abstractmethod
    async def get_conversation(self, conversation_id: ConversationId) -> Optional[ConversationContext]:
        """Retrieve a conversation by ID."""
        pass
    
    @abstractmethod
    async def update_conversation(self, context: ConversationContext) -> None:
        """Update an existing conversation."""
        pass
    
    @abstractmethod
    async def add_message(self, conversation_id: ConversationId, message: Message) -> None:
        """Add a message to a conversation."""
        pass
    
    @abstractmethod
    async def get_active_conversations(self, limit: int = 50) -> List[ConversationContext]:
        """Get active conversations."""
        pass
    
    @abstractmethod
    async def delete_conversation(self, conversation_id: ConversationId) -> bool:
        """Delete a conversation."""
        pass