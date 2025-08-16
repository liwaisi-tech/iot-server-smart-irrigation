"""Logger factory and utilities."""

import structlog
from structlog.types import FilteringBoundLogger


def get_logger(name: str) -> FilteringBoundLogger:
    """Get a logger instance for the given name."""
    return structlog.get_logger(name)


class LoggerMixin:
    """Mixin class to add logging capabilities to any class."""
    
    @property
    def logger(self) -> FilteringBoundLogger:
        """Get logger instance for this class."""
        return get_logger(self.__class__.__name__)