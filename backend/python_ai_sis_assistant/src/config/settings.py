from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""
    
    # Server configuration
    host: str = Field(default="0.0.0.0", description="Server host")
    port: int = Field(default=8081, description="Server port")
    debug: bool = Field(default=False, description="Debug mode")
    environment: str = Field(default="dev", description="Environment name")
    
    # AI/LLM configuration
    gemini_api_key: str = Field(..., description="Google Gemini API key")
    gemini_model: str = Field(default="gemini-pro", description="Gemini model name")
    
    # Agent configuration
    max_conversation_history: int = Field(default=50, description="Max conversation history to keep")
    agent_response_timeout: int = Field(default=30, description="Agent response timeout in seconds")
    
    # External services (placeholders for future implementation)
    web_search_api_key: str = Field(default="", description="Web search API key")
    iot_backend_url: str = Field(default="http://localhost:8080", description="IoT backend service URL")
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"