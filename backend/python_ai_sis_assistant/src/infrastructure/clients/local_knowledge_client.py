"""Local Knowledge Client - Connects to IoT backend service."""

import logging
from typing import Dict, Any
import httpx

from src.domain.ports.llm_client import LocalKnowledgePort


logger = logging.getLogger(__name__)


class LocalKnowledgeClient(LocalKnowledgePort):
    """HTTP client for accessing local IoT system knowledge."""
    
    def __init__(self, backend_url: str, timeout: int = 30):
        self.backend_url = backend_url.rstrip('/')
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
    
    async def search_device_info(self, query: str) -> Dict[str, Any]:
        """Search for device information in the IoT backend."""
        try:
            # Call the Go backend API to search devices
            response = await self.client.get(
                f"{self.backend_url}/api/v1/devices/search",
                params={"query": query}
            )
            response.raise_for_status()
            
            data = response.json()
            logger.info(f"Found {len(data.get('devices', []))} devices matching query: {query}")
            
            return {
                "devices": data.get("devices", []),
                "total_count": data.get("total_count", 0),
                "query": query,
                "status": "success"
            }
            
        except httpx.HTTPStatusError as e:
            logger.error(f"HTTP error searching devices: {e.response.status_code}")
            return {
                "devices": [],
                "total_count": 0,
                "query": query,
                "status": "error",
                "error": f"Backend service returned {e.response.status_code}"
            }
        except httpx.RequestError as e:
            logger.error(f"Network error searching devices: {e}")
            return {
                "devices": [],
                "total_count": 0,
                "query": query,
                "status": "error",
                "error": "Cannot connect to IoT backend service"
            }
    
    async def get_system_status(self) -> Dict[str, Any]:
        """Get current IoT system status from the backend."""
        try:
            # Get overall system health and status
            response = await self.client.get(f"{self.backend_url}/health")
            response.raise_for_status()
            
            health_data = response.json()
            
            # Get device statistics
            devices_response = await self.client.get(f"{self.backend_url}/api/v1/devices/stats")
            devices_data = devices_response.json() if devices_response.status_code == 200 else {}
            
            logger.info("Retrieved system status successfully")
            
            return {
                "system_health": health_data.get("status", "unknown"),
                "uptime": health_data.get("uptime", "unknown"),
                "database_status": health_data.get("database", "unknown"),
                "mqtt_status": health_data.get("mqtt", "unknown"),
                "nats_status": health_data.get("nats", "unknown"),
                "total_devices": devices_data.get("total_devices", 0),
                "active_devices": devices_data.get("active_devices", 0),
                "recent_alerts": devices_data.get("recent_alerts", []),
                "status": "success",
                "timestamp": health_data.get("timestamp")
            }
            
        except httpx.HTTPStatusError as e:
            logger.error(f"HTTP error getting system status: {e.response.status_code}")
            return {
                "system_health": "error",
                "status": "error",
                "error": f"Backend service returned {e.response.status_code}"
            }
        except httpx.RequestError as e:
            logger.error(f"Network error getting system status: {e}")
            return {
                "system_health": "error", 
                "status": "error",
                "error": "Cannot connect to IoT backend service"
            }
    
    async def get_sensor_data(self, device_id: str | None = None) -> Dict[str, Any]:
        """Get sensor data from IoT devices."""
        try:
            # Build the API endpoint
            if device_id:
                endpoint = f"{self.backend_url}/api/v1/devices/{device_id}/sensors/data"
            else:
                endpoint = f"{self.backend_url}/api/v1/sensors/data/recent"
            
            response = await self.client.get(endpoint)
            response.raise_for_status()
            
            data = response.json()
            logger.info(f"Retrieved sensor data for device: {device_id or 'all devices'}")
            
            return {
                "sensor_data": data.get("sensors", []),
                "device_id": device_id,
                "total_readings": data.get("total_readings", 0),
                "last_updated": data.get("last_updated"),
                "status": "success"
            }
            
        except httpx.HTTPStatusError as e:
            logger.error(f"HTTP error getting sensor data: {e.response.status_code}")
            return {
                "sensor_data": [],
                "device_id": device_id,
                "total_readings": 0,
                "status": "error", 
                "error": f"Backend service returned {e.response.status_code}"
            }
        except httpx.RequestError as e:
            logger.error(f"Network error getting sensor data: {e}")
            return {
                "sensor_data": [],
                "device_id": device_id,
                "total_readings": 0,
                "status": "error",
                "error": "Cannot connect to IoT backend service"
            }
    
    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()