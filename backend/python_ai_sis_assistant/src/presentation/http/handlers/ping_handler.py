"""HTTP handler for ping endpoint."""

from fastapi import APIRouter, Depends
from src.domain.ports.ping_service import PingServicePort


class PingHandler:
    """HTTP handler for ping operations."""
    
    def __init__(self, ping_service: PingServicePort):
        """Initialize ping handler.
        
        Args:
            ping_service: The ping service implementation.
        """
        self.ping_service = ping_service
        self.router = APIRouter()
        self._setup_routes()
    
    def _setup_routes(self):
        """Set up HTTP routes for ping operations."""
        
        @self.router.get("/ping", summary="Ping endpoint", tags=["health"])
        def ping():
            """Ping endpoint that responds with 'pong'.
            
            Returns:
                str: The response 'pong'.
            """
            return self.ping_service.ping()