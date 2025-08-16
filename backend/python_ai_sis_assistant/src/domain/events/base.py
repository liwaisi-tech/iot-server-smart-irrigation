"""Base domain event classes."""

from datetime import datetime
from typing import Any
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class DomainEvent(BaseModel):
    """Base class for all domain events."""
    
    event_id: UUID = Field(default_factory=uuid4)
    event_type: str = Field(..., description="Type of the event")
    aggregate_id: str = Field(..., description="ID of the aggregate that generated this event")
    aggregate_type: str = Field(..., description="Type of the aggregate")
    event_data: dict[str, Any] = Field(default_factory=dict)
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: int = Field(default=1, description="Event schema version")
    
    class Config:
        frozen = True
    
    def __init__(self, **data: Any) -> None:
        if "event_type" not in data:
            data["event_type"] = self.__class__.__name__
        super().__init__(**data)