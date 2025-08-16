"""System-level domain errors."""

from typing import Any, Optional

from .base import DomainError


class ExternalServiceError(DomainError):
    """Base class for external service errors."""
    pass


class DeviceConnectionError(ExternalServiceError):
    """Device is unreachable or offline."""
    
    def __init__(
        self, 
        device_ip: str, 
        device_name: Optional[str] = None,
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.device_ip = device_ip
        self.device_name = device_name
        message = f"Cannot connect to device {device_name or device_ip}"
        super().__init__(
            message=message,
            error_code="DEVICE_CONNECTION_ERROR",
            details={
                "device_ip": device_ip,
                "device_name": device_name,
                **(details or {})
            }
        )


class ConversationServiceError(ExternalServiceError):
    """Conversation service error."""
    
    def __init__(
        self, 
        message: str, 
        status_code: Optional[int] = None,
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.status_code = status_code
        super().__init__(
            message=f"Conversation service error: {message}",
            error_code="CONVERSATION_SERVICE_ERROR",
            details={
                "status_code": status_code,
                **(details or {})
            }
        )


class DatabaseError(DomainError):
    """Database operation error."""
    
    def __init__(
        self, 
        operation: str, 
        message: str,
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.operation = operation
        super().__init__(
            message=f"Database {operation} failed: {message}",
            error_code="DATABASE_ERROR",
            details={
                "operation": operation,
                **(details or {})
            }
        )