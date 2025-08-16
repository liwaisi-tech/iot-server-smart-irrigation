"""Dependency injection container for the application."""

from src.domain.ports.ping_service import PingServicePort
from src.usecases.ping.ping_service import PingService


class Container:
    """Simple dependency injection container."""
    
    def __init__(self):
        """Initialize the container with service instances."""
        self._ping_service = None
    
    @property
    def ping_service(self) -> PingServicePort:
        """Get ping service instance.
        
        Returns:
            PingServicePort: The ping service implementation.
        """
        if self._ping_service is None:
            self._ping_service = PingService()
        return self._ping_service