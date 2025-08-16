"""User experience related domain errors."""

from typing import Any, Optional

from .base import DomainError


class UserExperienceError(DomainError):
    """Base class for UX-related errors."""
    pass


class AmbiguousRequestError(UserExperienceError):
    """User request is ambiguous and needs clarification."""
    
    def __init__(
        self, 
        ambiguous_entities: list[str],
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.ambiguous_entities = ambiguous_entities
        message = f"Ambiguous request involving: {', '.join(ambiguous_entities)}"
        super().__init__(
            message=message,
            error_code="AMBIGUOUS_REQUEST",
            details={
                "ambiguous_entities": ambiguous_entities,
                **(details or {})
            }
        )


class NoDevicesFoundError(UserExperienceError):
    """No devices found matching user criteria."""
    
    def __init__(
        self, 
        criteria: Optional[str] = None,
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.criteria = criteria
        message = f"No devices found{f' matching: {criteria}' if criteria else ''}"
        super().__init__(
            message=message,
            error_code="NO_DEVICES_FOUND",
            details={
                "criteria": criteria,
                **(details or {})
            }
        )


class InvalidSessionError(UserExperienceError):
    """Session is invalid or expired."""
    
    def __init__(
        self, 
        session_id: str,
        reason: str = "Session expired or not found",
        details: Optional[dict[str, Any]] = None
    ) -> None:
        self.session_id = session_id
        self.reason = reason
        message = f"Invalid session {session_id}: {reason}"
        super().__init__(
            message=message,
            error_code="INVALID_SESSION",
            details={
                "session_id": session_id,
                "reason": reason,
                **(details or {})
            }
        )