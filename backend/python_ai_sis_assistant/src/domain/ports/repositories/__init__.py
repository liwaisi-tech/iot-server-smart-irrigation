"""Repository port interfaces."""

from .conversation_repository import ConversationRepository
from .device_repository import DeviceRepository

__all__ = [
    "ConversationRepository",
    "DeviceRepository",
]