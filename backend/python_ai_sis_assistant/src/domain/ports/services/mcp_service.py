"""MCP service port interface."""

from abc import ABC, abstractmethod
from typing import Any

from src.domain.entities.device import Device


class MCPService(ABC):
    """Port interface for MCP tool execution."""
    
    @abstractmethod
    async def discover_devices(self) -> list[Device]:
        """Execute device discovery MCP tool."""
        pass
    
    @abstractmethod
    async def collect_sensor_data(
        self, 
        device_ip: str, 
        sensor_type: str = "temperature-and-humidity"
    ) -> dict[str, Any]:
        """Execute sensor data collection MCP tool."""
        pass
    
    @abstractmethod
    async def execute_tool(
        self, 
        tool_name: str, 
        parameters: dict[str, Any]
    ) -> dict[str, Any]:
        """Execute a registered MCP tool."""
        pass
    
    @abstractmethod
    async def register_tool(self, tool_name: str, tool_instance: Any) -> None:
        """Register a new MCP tool."""
        pass
    
    @abstractmethod
    async def list_available_tools(self) -> list[str]:
        """List all available MCP tools."""
        pass