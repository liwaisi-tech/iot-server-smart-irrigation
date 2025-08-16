"""Conversation repository port interface."""

from abc import ABC, abstractmethod
from typing import Optional

from src.domain.entities.conversation import Conversation, ConversationContext


class ConversationRepository(ABC):
    """Port interface for conversation data operations."""
    
    @abstractmethod
    async def save_conversation(self, conversation: Conversation) -> None:
        """Save or update a conversation."""
        pass
    
    @abstractmethod
    async def get_conversation_by_session_id(self, session_id: str) -> Optional[Conversation]:
        """Get a conversation by session ID."""
        pass
    
    @abstractmethod
    async def save_context(self, context: ConversationContext) -> None:
        """Save or update conversation context."""
        pass
    
    @abstractmethod
    async def get_context_by_session_id(self, session_id: str) -> Optional[ConversationContext]:
        """Get conversation context by session ID."""
        pass
    
    @abstractmethod
    async def delete_conversation(self, session_id: str) -> None:
        """Delete a conversation and its context."""
        pass
    
    @abstractmethod
    async def delete_expired_sessions(self, expiry_hours: int = 24) -> int:
        """Delete expired conversation sessions. Returns count of deleted sessions."""
        pass
    
    @abstractmethod
    async def get_conversation_history(
        self, 
        session_id: str, 
        limit: int = 50
    ) -> list[dict[str, str]]:
        """Get conversation message history for a session."""
        pass