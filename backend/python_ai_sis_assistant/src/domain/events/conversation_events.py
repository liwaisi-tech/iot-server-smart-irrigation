"""Conversation-related domain events."""

from typing import Any, Optional

from .base import DomainEvent


class ConversationStarted(DomainEvent):
    """Event fired when a new conversation is started."""
    
    def __init__(
        self,
        session_id: str,
        user_id: Optional[str] = None,
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=session_id,
            aggregate_type="Conversation",
            event_data={
                "session_id": session_id,
                "user_id": user_id,
            },
            **kwargs
        )


class MessageProcessed(DomainEvent):
    """Event fired when a message is processed in a conversation."""
    
    def __init__(
        self,
        session_id: str,
        message_id: str,
        message_content: str,
        role: str,
        intent: Optional[str] = None,
        entities: Optional[list[dict[str, Any]]] = None,
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=session_id,
            aggregate_type="Conversation",
            event_data={
                "session_id": session_id,
                "message_id": message_id,
                "message_content": message_content,
                "role": role,
                "intent": intent,
                "entities": entities or [],
            },
            **kwargs
        )


class ConversationCompleted(DomainEvent):
    """Event fired when a conversation is completed."""
    
    def __init__(
        self,
        session_id: str,
        message_count: int,
        duration_seconds: float,
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=session_id,
            aggregate_type="Conversation",
            event_data={
                "session_id": session_id,
                "message_count": message_count,
                "duration_seconds": duration_seconds,
            },
            **kwargs
        )