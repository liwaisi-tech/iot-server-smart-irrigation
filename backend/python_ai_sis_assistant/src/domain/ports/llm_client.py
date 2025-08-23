"""Port for LLM (Large Language Model) client interactions."""

from abc import ABC, abstractmethod
from typing import Dict, Any, List
from dataclasses import dataclass

from src.domain.value_objects.conversation import Message


@dataclass
class LLMRequest:
    """Request to the LLM."""
    messages: List[Message]
    system_prompt: str
    max_tokens: int = 1000
    temperature: float = 0.7
    metadata: Dict[str, Any] | None = None


@dataclass
class LLMResponse:
    """Response from the LLM."""
    content: str
    tokens_used: int | None = None
    metadata: Dict[str, Any] | None = None


class LLMClientPort(ABC):
    """Port for communicating with Large Language Model services."""
    
    @abstractmethod
    async def generate_response(self, request: LLMRequest) -> LLMResponse:
        """Generate a response using the LLM."""
        pass
    
    @abstractmethod
    async def health_check(self) -> bool:
        """Check if the LLM service is available."""
        pass


class LocalKnowledgePort(ABC):
    """Port for accessing local IoT system knowledge."""
    
    @abstractmethod
    async def search_device_info(self, query: str) -> Dict[str, Any]:
        """Search for device information."""
        pass
    
    @abstractmethod
    async def get_system_status(self) -> Dict[str, Any]:
        """Get current system status."""
        pass
    
    @abstractmethod
    async def get_sensor_data(self, device_id: str | None = None) -> Dict[str, Any]:
        """Get sensor data."""
        pass


class WebSearchPort(ABC):
    """Port for web search capabilities."""
    
    @abstractmethod
    async def search(self, query: str, max_results: int = 5) -> List[Dict[str, Any]]:
        """Perform web search and return results."""
        pass
    
    @abstractmethod
    async def fetch_url_content(self, url: str) -> str:
        """Fetch content from a specific URL."""
        pass