"""User domain entities."""

from datetime import datetime
from typing import Any, Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class User(BaseModel):
    """User entity representing a system user."""
    
    id: UUID = Field(default_factory=uuid4)
    username: Optional[str] = None
    email: Optional[str] = None
    full_name: Optional[str] = None
    language_preference: str = Field(default="es-CO", description="Language preference (Colombian Spanish)")
    timezone: str = Field(default="America/Bogota", description="User timezone")
    preferences: dict[str, Any] = Field(default_factory=dict)
    is_active: bool = Field(default=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    last_login: Optional[datetime] = None
    
    def update_preference(self, key: str, value: Any) -> None:
        """Update a user preference."""
        self.preferences[key] = value
        self.updated_at = datetime.utcnow()
    
    def mark_login(self) -> None:
        """Mark user as logged in."""
        self.last_login = datetime.utcnow()
        self.updated_at = datetime.utcnow()
    
    def deactivate(self) -> None:
        """Deactivate the user."""
        self.is_active = False
        self.updated_at = datetime.utcnow()
    
    def activate(self) -> None:
        """Activate the user."""
        self.is_active = True
        self.updated_at = datetime.utcnow()
    
    class Config:
        validate_assignment = True