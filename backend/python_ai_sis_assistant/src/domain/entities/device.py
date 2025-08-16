"""Device domain entities."""

from datetime import datetime
from typing import Any, Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field, IPvAnyAddress


class SensorData(BaseModel):
    """Sensor data value object."""
    
    sensor_type: str
    value: float
    unit: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    metadata: dict[str, Any] = Field(default_factory=dict)
    
    class Config:
        frozen = True


class Device(BaseModel):
    """Device entity representing an IoT device."""
    
    id: UUID = Field(default_factory=uuid4)
    mac_address: str = Field(..., description="Device MAC address")
    device_name: str = Field(..., description="Human-readable device name")
    ip_address: IPvAnyAddress = Field(..., description="Device IP address")
    location_description: str = Field(..., description="Device location description")
    device_type: str = Field(default="sensor", description="Type of device")
    status: str = Field(default="unknown", description="Device status (online, offline, unknown)")
    last_seen: Optional[datetime] = None
    sensor_data: list[SensorData] = Field(default_factory=list)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    
    def update_status(self, status: str) -> None:
        """Update device status."""
        self.status = status
        if status == "online":
            self.last_seen = datetime.utcnow()
        self.updated_at = datetime.utcnow()
    
    def add_sensor_data(self, sensor_data: SensorData) -> None:
        """Add sensor data to the device."""
        self.sensor_data.append(sensor_data)
        # Keep only the last 100 readings
        if len(self.sensor_data) > 100:
            self.sensor_data = self.sensor_data[-100:]
        self.updated_at = datetime.utcnow()
    
    def get_latest_sensor_data(self, sensor_type: Optional[str] = None) -> Optional[SensorData]:
        """Get the latest sensor data, optionally filtered by type."""
        if not self.sensor_data:
            return None
        
        if sensor_type:
            filtered_data = [data for data in self.sensor_data if data.sensor_type == sensor_type]
            return filtered_data[-1] if filtered_data else None
        
        return self.sensor_data[-1]
    
    def is_online(self) -> bool:
        """Check if device is online."""
        return self.status == "online"
    
    def is_responsive(self, timeout_minutes: int = 5) -> bool:
        """Check if device has been seen recently."""
        if not self.last_seen:
            return False
        
        time_diff = datetime.utcnow() - self.last_seen
        return time_diff.total_seconds() < (timeout_minutes * 60)
    
    class Config:
        validate_assignment = True