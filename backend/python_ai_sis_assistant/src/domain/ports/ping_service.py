"""Domain port for ping service operations."""

from abc import ABC, abstractmethod


class PingServicePort(ABC):
    """Port defining the interface for ping service operations."""
    
    @abstractmethod
    def ping(self) -> str:
        """Execute ping operation and return response.
        
        Returns:
            str: The ping response message.
        """
        pass