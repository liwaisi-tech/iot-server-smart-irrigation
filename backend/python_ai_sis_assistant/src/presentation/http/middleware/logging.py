"""Logging middleware for HTTP requests and responses."""

import time
import uuid
from typing import Callable

from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from src.infrastructure.logging import get_logger


class LoggingMiddleware(BaseHTTPMiddleware):
    """Middleware for logging HTTP requests and responses."""
    
    def __init__(self, app: Callable) -> None:
        super().__init__(app)
        self.logger = get_logger("http")
    
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """Process HTTP request and log details."""
        
        # Generate correlation ID
        correlation_id = str(uuid.uuid4())
        request.state.correlation_id = correlation_id
        
        # Start timing
        start_time = time.time()
        
        # Log incoming request
        self.logger.info(
            "HTTP request started",
            correlation_id=correlation_id,
            method=request.method,
            url=str(request.url),
            client_host=request.client.host if request.client else None,
            user_agent=request.headers.get("user-agent"),
        )
        
        try:
            # Process request
            response = await call_next(request)
            
            # Calculate duration
            duration = time.time() - start_time
            
            # Log response
            self.logger.info(
                "HTTP request completed",
                correlation_id=correlation_id,
                status_code=response.status_code,
                duration_ms=round(duration * 1000, 2),
            )
            
            # Add correlation ID to response headers
            response.headers["X-Correlation-ID"] = correlation_id
            
            return response
            
        except Exception as e:
            # Calculate duration
            duration = time.time() - start_time
            
            # Log error
            self.logger.error(
                "HTTP request failed",
                correlation_id=correlation_id,
                error=str(e),
                error_type=type(e).__name__,
                duration_ms=round(duration * 1000, 2),
            )
            
            raise