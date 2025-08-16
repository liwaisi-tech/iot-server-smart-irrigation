"""Device-related domain events."""

from typing import Any, Optional

from .base import DomainEvent


class DeviceDiscovered(DomainEvent):
    """Event fired when a new device is discovered."""
    
    def __init__(
        self,
        device_id: str,
        mac_address: str,
        device_name: str,
        ip_address: str,
        location_description: str,
        device_type: str = "sensor",
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=device_id,
            aggregate_type="Device",
            event_data={
                "device_id": device_id,
                "mac_address": mac_address,
                "device_name": device_name,
                "ip_address": ip_address,
                "location_description": location_description,
                "device_type": device_type,
            },
            **kwargs
        )


class DeviceOnline(DomainEvent):
    """Event fired when a device comes online."""
    
    def __init__(
        self,
        device_id: str,
        device_name: str,
        ip_address: str,
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=device_id,
            aggregate_type="Device",
            event_data={
                "device_id": device_id,
                "device_name": device_name,
                "ip_address": ip_address,
            },
            **kwargs
        )


class DeviceOffline(DomainEvent):
    """Event fired when a device goes offline."""
    
    def __init__(
        self,
        device_id: str,
        device_name: str,
        ip_address: str,
        last_seen: Optional[str] = None,
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=device_id,
            aggregate_type="Device",
            event_data={
                "device_id": device_id,
                "device_name": device_name,
                "ip_address": ip_address,
                "last_seen": last_seen,
            },
            **kwargs
        )


class SensorDataReceived(DomainEvent):
    """Event fired when sensor data is received from a device."""
    
    def __init__(
        self,
        device_id: str,
        device_name: str,
        sensor_type: str,
        sensor_data: dict[str, Any],
        **kwargs: Any
    ) -> None:
        super().__init__(
            aggregate_id=device_id,
            aggregate_type="Device",
            event_data={
                "device_id": device_id,
                "device_name": device_name,
                "sensor_type": sensor_type,
                "sensor_data": sensor_data,
            },
            **kwargs
        )