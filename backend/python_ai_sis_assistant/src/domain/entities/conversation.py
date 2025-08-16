"""Conversation domain entities."""

from datetime import datetime
from typing import Any, Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class Message(BaseModel):
    """Message entity representing a single conversation message."""
    
    id: UUID = Field(default_factory=uuid4)
    session_id: str
    content: str
    role: str = Field(..., description="Either 'user' or 'assistant'")
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    metadata: dict[str, Any] = Field(default_factory=dict)
    
    class Config:
        frozen = True


class ConversationContext(BaseModel):
    """Conversation context entity."""
    
    session_id: str
    user_id: Optional[str] = None
    current_topic: str = ""
    last_mentioned_devices: list[str] = Field(default_factory=list)
    last_mentioned_location: str = ""
    conversation_history: list[dict[str, Any]] = Field(default_factory=list)
    preferences: dict[str, Any] = Field(default_factory=dict)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    
    def add_device_reference(self, device_name: str) -> None:
        """Add a device reference to the context."""
        if device_name not in self.last_mentioned_devices:
            self.last_mentioned_devices.append(device_name)
            # Keep only the last 5 devices
            if len(self.last_mentioned_devices) > 5:
                self.last_mentioned_devices = self.last_mentioned_devices[-5:]
    
    def set_location(self, location: str) -> None:
        """Set the current location context."""
        self.last_mentioned_location = location
    
    def add_to_history(self, message: str, entities: list[dict[str, Any]]) -> None:
        """Add a message to conversation history."""
        self.conversation_history.append({
            "timestamp": datetime.utcnow().isoformat(),
            "message": message,
            "entities": entities
        })
        # Keep only the last 20 messages
        if len(self.conversation_history) > 20:
            self.conversation_history = self.conversation_history[-20:]
        
        self.updated_at = datetime.utcnow()


class Conversation(BaseModel):
    """Conversation aggregate root."""
    
    id: UUID = Field(default_factory=uuid4)
    session_id: str
    user_id: Optional[str] = None
    messages: list[Message] = Field(default_factory=list)
    context: ConversationContext
    status: str = Field(default="active", description="active, completed, or error")
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    
    def add_message(self, content: str, role: str, metadata: Optional[dict[str, Any]] = None) -> Message:
        """Add a message to the conversation."""
        message = Message(
            session_id=self.session_id,
            content=content,
            role=role,
            metadata=metadata or {}
        )
        self.messages.append(message)
        self.updated_at = datetime.utcnow()
        return message
    
    def get_recent_messages(self, limit: int = 10) -> list[Message]:
        """Get the most recent messages."""
        return self.messages[-limit:] if len(self.messages) > limit else self.messages
    
    def mark_completed(self) -> None:
        """Mark the conversation as completed."""
        self.status = "completed"
        self.updated_at = datetime.utcnow()
    
    def mark_error(self) -> None:
        """Mark the conversation as having an error."""
        self.status = "error"
        self.updated_at = datetime.utcnow()