"""Domain errors for the Python AI SIS Assistant."""

from .base import DomainError
from .system_errors import (
    ConversationServiceError,
    DatabaseError,
    DeviceConnectionError,
    ExternalServiceError,
)
from .user_errors import (
    AmbiguousRequestError,
    InvalidSessionError,
    NoDevicesFoundError,
    UserExperienceError,
)

__all__ = [
    "DomainError",
    "ConversationServiceError",
    "DatabaseError", 
    "DeviceConnectionError",
    "ExternalServiceError",
    "AmbiguousRequestError",
    "InvalidSessionError",
    "NoDevicesFoundError",
    "UserExperienceError",
]