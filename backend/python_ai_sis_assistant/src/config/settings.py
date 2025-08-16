"""Application settings and configuration management."""

from typing import Optional

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings."""

    # Server Configuration
    host: str = Field(default="0.0.0.0", description="Server host")
    port: int = Field(default=8081, description="Server port")
    debug: bool = Field(default=False, description="Debug mode")
    environment: str = Field(default="development", description="Environment")

    # Google ADK Configuration
    google_adk_project_id: Optional[str] = Field(
        default=None, description="Google ADK project ID"
    )
    google_adk_location: str = Field(
        default="us-central1", description="Google ADK location"
    )
    google_adk_api_key: Optional[str] = Field(
        default=None, description="Google ADK API key"
    )
    google_adk_model_name: str = Field(
        default="gemini-1.5-pro", description="Google ADK model name"
    )

    # Database Configuration
    database_url: str = Field(
        default="postgresql+asyncpg://user:password@localhost:5432/sis_assistant",
        description="Database URL",
    )
    database_pool_size: int = Field(default=10, description="Database pool size")
    database_max_overflow: int = Field(
        default=20, description="Database max overflow"
    )

    # Redis Configuration
    redis_url: str = Field(default="redis://localhost:6379/0", description="Redis URL")
    redis_pool_size: int = Field(default=10, description="Redis pool size")

    # Session Configuration
    session_timeout_minutes: int = Field(
        default=30, description="Session timeout in minutes"
    )
    max_conversation_history: int = Field(
        default=100, description="Maximum conversation history"
    )

    # MCP Configuration
    device_discovery_timeout: float = Field(
        default=5.0, description="Device discovery timeout in seconds"
    )
    max_concurrent_device_queries: int = Field(
        default=10, description="Maximum concurrent device queries"
    )
    device_registry_cache_ttl: int = Field(
        default=300, description="Device registry cache TTL in seconds"
    )

    # Go Backend Integration
    go_backend_url: str = Field(
        default="http://localhost:8080", description="Go backend URL"
    )
    go_backend_timeout: float = Field(
        default=10.0, description="Go backend timeout in seconds"
    )

    # Logging Configuration
    log_level: str = Field(default="INFO", description="Log level")
    log_format: str = Field(default="json", description="Log format")

    # API Configuration
    api_key: Optional[str] = Field(default=None, description="API key for authentication")
    cors_origins: list[str] = Field(
        default=["*"], description="CORS allowed origins"
    )
    rate_limit_requests: int = Field(default=60, description="Rate limit requests per minute")

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )


# Global settings instance
settings = Settings()