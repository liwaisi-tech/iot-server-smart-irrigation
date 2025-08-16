"""Base domain error classes."""

from typing import Any, Optional


class DomainError(Exception):
    """Base class for all domain errors."""
    
    def __init__(
        self, 
        message: str, 
        error_code: Optional[str] = None,
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.message = message
        self.error_code = error_code
        self.details = details or {}
        super().__init__(message)
    
    def __str__(self) -> str:
        if self.error_code:
            return f"[{self.error_code}] {self.message}"
        return self.message
    
    def to_dict(self) -> dict[str, Any]:
        """Convert error to dictionary representation."""
        return {
            "error_type": self.__class__.__name__,
            "message": self.message,
            "error_code": self.error_code,
            "details": self.details,
        }