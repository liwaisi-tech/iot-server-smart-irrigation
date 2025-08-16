"""Domain events for the Python AI SIS Assistant."""

from .base import DomainEvent
from .conversation_events import (
    ConversationCompleted,
    ConversationStarted,
    MessageProcessed,
)
from .device_events import (
    DeviceDiscovered,
    DeviceOffline,
    DeviceOnline,
    SensorDataReceived,
)

__all__ = [
    "DomainEvent",
    "ConversationCompleted",
    "ConversationStarted", 
    "MessageProcessed",
    "DeviceDiscovered",
    "DeviceOffline",
    "DeviceOnline",
    "SensorDataReceived",
]