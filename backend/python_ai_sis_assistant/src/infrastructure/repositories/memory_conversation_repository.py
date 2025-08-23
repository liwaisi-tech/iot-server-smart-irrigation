"""In-memory conversation repository implementation."""

import logging
from typing import Optional, List, Dict

from src.domain.value_objects.conversation import ConversationId, ConversationContext, Message
from src.domain.ports.conversation_repository import ConversationRepositoryPort


logger = logging.getLogger(__name__)


class MemoryConversationRepository(ConversationRepositoryPort):
    """In-memory implementation of conversation repository."""
    
    def __init__(self):
        self._conversations: Dict[str, ConversationContext] = {}
    
    async def save_conversation(self, context: ConversationContext) -> None:
        """Save a conversation context."""
        self._conversations[context.conversation_id.value] = context
        logger.debug(f"Saved conversation {context.conversation_id}")
    
    async def get_conversation(self, conversation_id: ConversationId) -> Optional[ConversationContext]:
        """Retrieve a conversation by ID."""
        context = self._conversations.get(conversation_id.value)
        if context:
            logger.debug(f"Retrieved conversation {conversation_id}")
        else:
            logger.debug(f"Conversation {conversation_id} not found")
        return context
    
    async def update_conversation(self, context: ConversationContext) -> None:
        """Update an existing conversation."""
        self._conversations[context.conversation_id.value] = context
        logger.debug(f"Updated conversation {context.conversation_id}")
    
    async def add_message(self, conversation_id: ConversationId, message: Message) -> None:
        """Add a message to a conversation."""
        context = await self.get_conversation(conversation_id)
        if context:
            context.add_message(message)
            await self.update_conversation(context)
            logger.debug(f"Added message to conversation {conversation_id}")
        else:
            logger.warning(f"Cannot add message: conversation {conversation_id} not found")
    
    async def get_active_conversations(self, limit: int = 50) -> List[ConversationContext]:
        """Get active conversations."""
        from src.domain.value_objects.conversation import ConversationState
        
        active_conversations = [
            context for context in self._conversations.values()
            if context.state == ConversationState.ACTIVE
        ]
        
        # Sort by last_updated descending
        active_conversations.sort(key=lambda x: x.last_updated, reverse=True)
        
        return active_conversations[:limit]
    
    async def delete_conversation(self, conversation_id: ConversationId) -> bool:
        """Delete a conversation."""
        if conversation_id.value in self._conversations:
            del self._conversations[conversation_id.value]
            logger.debug(f"Deleted conversation {conversation_id}")
            return True
        else:
            logger.debug(f"Cannot delete: conversation {conversation_id} not found")
            return False
    
    def get_conversation_count(self) -> int:
        """Get total number of stored conversations."""
        return len(self._conversations)
    
    def clear_all(self) -> None:
        """Clear all conversations (useful for testing)."""
        self._conversations.clear()
        logger.debug("Cleared all conversations")