"""NLP service port interface."""

from abc import ABC, abstractmethod
from enum import Enum
from typing import Any

from pydantic import BaseModel


class Intent(str, Enum):
    """User intent classification."""
    DEVICE_STATUS = "device_status"
    SENSOR_DATA = "sensor_data"
    SYSTEM_HEALTH = "system_health"
    DEVICE_CONTROL = "device_control"
    TROUBLESHOOTING = "troubleshooting"
    GENERAL_INFO = "general_info"
    GREETING = "greeting"
    GOODBYE = "goodbye"


class EntityType(str, Enum):
    """Entity types for extraction."""
    LOCATION = "location"
    DEVICE_NAME = "device_name"
    SENSOR_TYPE = "sensor_type"
    TIME_RANGE = "time_range"
    DEVICE_ID = "device_id"


class Entity(BaseModel):
    """Extracted entity."""
    type: EntityType
    value: str
    confidence: float
    start_pos: int
    end_pos: int


class NLPService(ABC):
    """Port interface for natural language processing."""
    
    @abstractmethod
    async def classify_intent(self, message: str) -> tuple[Intent, float]:
        """Classify user intent with confidence score."""
        pass
    
    @abstractmethod
    async def extract_entities(self, message: str) -> list[Entity]:
        """Extract entities from user message."""
        pass
    
    @abstractmethod
    async def normalize_text(self, text: str) -> str:
        """Normalize text for processing (lowercase, remove accents, etc.)."""
        pass
    
    @abstractmethod
    async def detect_language(self, text: str) -> str:
        """Detect the language of the input text."""
        pass
    
    @abstractmethod
    async def translate_if_needed(self, text: str, target_language: str = "es") -> str:
        """Translate text to target language if needed."""
        pass