"""Device repository port interface."""

from abc import ABC, abstractmethod
from typing import Optional

from src.domain.entities.device import Device


class DeviceRepository(ABC):
    """Port interface for device data operations."""
    
    @abstractmethod
    async def save_device(self, device: Device) -> None:
        """Save or update a device."""
        pass
    
    @abstractmethod
    async def get_device_by_id(self, device_id: str) -> Optional[Device]:
        """Get a device by ID."""
        pass
    
    @abstractmethod
    async def get_device_by_mac_address(self, mac_address: str) -> Optional[Device]:
        """Get a device by MAC address."""
        pass
    
    @abstractmethod
    async def get_device_by_ip(self, ip_address: str) -> Optional[Device]:
        """Get a device by IP address."""
        pass
    
    @abstractmethod
    async def get_devices_by_location(self, location: str) -> list[Device]:
        """Get devices filtered by location."""
        pass
    
    @abstractmethod
    async def get_all_devices(self) -> list[Device]:
        """Get all devices."""
        pass
    
    @abstractmethod
    async def get_online_devices(self) -> list[Device]:
        """Get all online devices."""
        pass
    
    @abstractmethod
    async def get_offline_devices(self) -> list[Device]:
        """Get all offline devices."""
        pass
    
    @abstractmethod
    async def delete_device(self, device_id: str) -> None:
        """Delete a device."""
        pass
    
    @abstractmethod
    async def update_device_status(self, device_id: str, status: str) -> None:
        """Update device status."""
        pass
    
    @abstractmethod
    async def save_discovered_devices(self, devices: list[Device]) -> None:
        """Save or update multiple discovered devices."""
        pass