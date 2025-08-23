"""Value objects for conversation management."""

from enum import Enum
from typing import Dict, Any, Optional
from dataclasses import dataclass, field
from datetime import datetime
import uuid


class MessageRole(Enum):
    """Roles in a conversation."""
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class ConversationState(Enum):
    """States of a conversation."""
    ACTIVE = "active"
    PENDING_AGENT_SWITCH = "pending_agent_switch"
    COMPLETED = "completed"
    ERROR = "error"


@dataclass(frozen=True)
class ConversationId:
    """Unique identifier for a conversation."""
    value: str = field(default_factory=lambda: str(uuid.uuid4()))
    
    def __str__(self) -> str:
        return self.value


@dataclass(frozen=True)
class Message:
    """A message in a conversation."""
    id: str
    role: MessageRole
    content: str
    timestamp: datetime
    metadata: Dict[str, Any] = field(default_factory=dict)
    agent_type: Optional[str] = None
    
    @classmethod
    def create_user_message(cls, content: str) -> "Message":
        """Create a user message."""
        return cls(
            id=str(uuid.uuid4()),
            role=MessageRole.USER,
            content=content,
            timestamp=datetime.utcnow()
        )
    
    @classmethod
    def create_assistant_message(cls, content: str, agent_type: str) -> "Message":
        """Create an assistant message."""
        return cls(
            id=str(uuid.uuid4()),
            role=MessageRole.ASSISTANT,
            content=content,
            timestamp=datetime.utcnow(),
            agent_type=agent_type
        )


@dataclass
class ConversationContext:
    """Context information for a conversation."""
    conversation_id: ConversationId
    state: ConversationState
    current_agent_type: str
    messages: list[Message] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.utcnow)
    last_updated: datetime = field(default_factory=datetime.utcnow)
    
    def add_message(self, message: Message) -> None:
        """Add a message to the conversation."""
        self.messages.append(message)
        self.last_updated = datetime.utcnow()
    
    def get_recent_messages(self, limit: int = 10) -> list[Message]:
        """Get the most recent messages."""
        return self.messages[-limit:]
    
    def switch_agent(self, new_agent_type: str) -> None:
        """Switch to a different agent."""
        self.current_agent_type = new_agent_type
        self.last_updated = datetime.utcnow()