"""Rate limiting middleware."""

import time
from typing import Callable

from fastapi import HTTPException, Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from src.config.settings import settings
from src.infrastructure.logging import get_logger


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Simple in-memory rate limiting middleware."""
    
    def __init__(self, app: Callable) -> None:
        super().__init__(app)
        self.logger = get_logger("rate_limit")
        self.requests: dict[str, list[float]] = {}
        self.max_requests = settings.rate_limit_requests
        self.window_seconds = 60  # 1 minute window
    
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """Apply rate limiting to requests."""
        
        # Skip rate limiting for health checks
        if request.url.path in ["/health", "/ping", "/docs", "/redoc", "/openapi.json"]:
            return await call_next(request)
        
        # Get client identifier
        client_ip = self._get_client_ip(request)
        current_time = time.time()
        
        # Clean old requests
        self._cleanup_old_requests(client_ip, current_time)
        
        # Check rate limit
        if self._is_rate_limited(client_ip):
            self.logger.warning(
                "Rate limit exceeded",
                client_ip=client_ip,
                requests_count=len(self.requests.get(client_ip, [])),
                max_requests=self.max_requests,
            )
            
            raise HTTPException(
                status_code=429,
                detail={
                    "error": "Rate limit exceeded",
                    "message": f"Maximum {self.max_requests} requests per minute allowed",
                    "retry_after": self.window_seconds
                }
            )
        
        # Add current request
        self._add_request(client_ip, current_time)
        
        return await call_next(request)
    
    def _get_client_ip(self, request: Request) -> str:
        """Get client IP address."""
        # Check for forwarded headers first
        forwarded_for = request.headers.get("X-Forwarded-For")
        if forwarded_for:
            return forwarded_for.split(",")[0].strip()
        
        real_ip = request.headers.get("X-Real-IP")
        if real_ip:
            return real_ip
        
        # Fallback to direct client IP
        return request.client.host if request.client else "unknown"
    
    def _cleanup_old_requests(self, client_ip: str, current_time: float) -> None:
        """Remove requests outside the time window."""
        if client_ip in self.requests:
            cutoff_time = current_time - self.window_seconds
            self.requests[client_ip] = [
                req_time for req_time in self.requests[client_ip]
                if req_time > cutoff_time
            ]
            
            # Clean up empty entries
            if not self.requests[client_ip]:
                del self.requests[client_ip]
    
    def _is_rate_limited(self, client_ip: str) -> bool:
        """Check if client has exceeded rate limit."""
        if client_ip not in self.requests:
            return False
        
        return len(self.requests[client_ip]) >= self.max_requests
    
    def _add_request(self, client_ip: str, current_time: float) -> None:
        """Add current request to tracking."""
        if client_ip not in self.requests:
            self.requests[client_ip] = []
        
        self.requests[client_ip].append(current_time)