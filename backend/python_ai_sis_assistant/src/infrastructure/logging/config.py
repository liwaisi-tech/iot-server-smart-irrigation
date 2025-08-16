"""Logging configuration using structlog."""

import logging
import sys
from typing import Any

import structlog
from structlog.types import Processor

from src.config.settings import settings


def configure_logging() -> None:
    """Configure structured logging with structlog."""
    
    # Determine log level
    log_level = getattr(logging, settings.log_level.upper(), logging.INFO)
    
    # Configure processors based on format
    processors: list[Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        structlog.dev.set_exc_info,
    ]
    
    if settings.log_format.lower() == "json":
        # Production JSON logging
        processors.extend([
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer()
        ])
    else:
        # Development console logging
        processors.extend([
            structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S"),
            structlog.dev.ConsoleRenderer(colors=True)
        ])
    
    # Configure structlog
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(log_level),
        logger_factory=structlog.WriteLoggerFactory(file=sys.stdout),
        cache_logger_on_first_use=True,
    )
    
    # Configure standard library logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=log_level,
    )
    
    # Reduce noise from third-party libraries
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("httpcore").setLevel(logging.WARNING)
    logging.getLogger("asyncio").setLevel(logging.WARNING)
    logging.getLogger("urllib3").setLevel(logging.WARNING)


def add_correlation_id_processor(_logger: Any, _method_name: str, event_dict: dict[str, Any]) -> dict[str, Any]:
    """Add correlation ID to log events."""
    # This will be enhanced when we add request tracking
    return event_dict


def add_service_context_processor(_logger: Any, _method_name: str, event_dict: dict[str, Any]) -> dict[str, Any]:
    """Add service context information to log events."""
    event_dict.update({
        "service": "python-ai-sis-assistant",
        "environment": settings.environment,
    })
    return event_dict