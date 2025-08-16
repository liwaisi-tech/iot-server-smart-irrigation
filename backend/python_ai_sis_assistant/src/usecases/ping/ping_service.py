"""Ping service use case implementation."""

from src.domain.ports.ping_service import PingServicePort


class PingService(PingServicePort):
    """Implementation of ping service use case."""
    
    def ping(self) -> str:
        """Execute ping operation and return response.
        
        Returns:
            str: The ping response message 'pong'.
        """
        return "pong"