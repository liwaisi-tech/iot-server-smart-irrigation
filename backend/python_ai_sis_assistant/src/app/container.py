from dependency_injector import containers, providers
from dependency_injector.wiring import Provide, inject

from src.config.settings import Settings

# Domain imports
from src.domain.value_objects.agent_types import (
    PEJA_COORDINATOR_PROFILE, LOCAL_AGENT_PROFILE, WEB_SEARCH_AGENT_PROFILE
)

# Use case imports
from src.usecases.conversation_manager import ConversationManager
from src.usecases.agent_orchestrator import AgentOrchestrator

# Infrastructure imports
from src.infrastructure.llm.gemini_client import GeminiClient
from src.infrastructure.agents.peja_coordinator import PejaCoordinatorAgent
from src.infrastructure.agents.local_assistant import LocalAssistantAgent
from src.infrastructure.agents.web_search_agent import WebSearchAgent
from src.infrastructure.repositories.memory_conversation_repository import MemoryConversationRepository
from src.infrastructure.repositories.agent_registry import InMemoryAgentRegistry
from src.infrastructure.clients.local_knowledge_client import LocalKnowledgeClient
from src.infrastructure.clients.web_search_client import WebSearchClient

# Presentation imports
from src.presentation.http.handlers.chat_handler import ChatHandler
from src.presentation.http.handlers.agent_handler import AgentHandler


class Container(containers.DeclarativeContainer):
    """Dependency injection container for the IoT Smart Irrigation System."""
    
    wiring_config = containers.WiringConfiguration(
        modules=[
            "src.app.factory",
            "src.presentation.http.routes.chat_routes",
            "src.presentation.http.routes.agent_routes",
        ]
    )
    
    # Configuration
    config = providers.Configuration()
    
    # Settings
    settings = providers.Singleton(Settings)
    
    # Infrastructure - LLM Client
    gemini_client = providers.Singleton(
        GeminiClient,
        api_key=settings.provided.gemini_api_key,
        model_name=settings.provided.gemini_model
    )
    
    # Infrastructure - Repositories
    conversation_repository = providers.Singleton(MemoryConversationRepository)
    agent_registry = providers.Singleton(InMemoryAgentRegistry)
    
    # Infrastructure - External Service Clients
    local_knowledge_client = providers.Singleton(
        LocalKnowledgeClient,
        backend_url=settings.provided.iot_backend_url,
        timeout=settings.provided.agent_response_timeout
    )
    
    web_search_client = providers.Singleton(
        WebSearchClient,
        api_key=settings.provided.web_search_api_key,
        timeout=settings.provided.agent_response_timeout
    )
    
    # Infrastructure - Agents
    peja_coordinator = providers.Singleton(
        PejaCoordinatorAgent,
        llm_client=gemini_client
    )
    
    local_assistant = providers.Singleton(
        LocalAssistantAgent,
        llm_client=gemini_client,
        local_knowledge=local_knowledge_client
    )
    
    web_search_agent = providers.Singleton(
        WebSearchAgent,
        llm_client=gemini_client,
        web_search=web_search_client
    )
    
    # Use Cases
    conversation_manager = providers.Singleton(
        ConversationManager,
        conversation_repo=conversation_repository,
        agent_registry=agent_registry
    )
    
    agent_orchestrator = providers.Singleton(
        AgentOrchestrator,
        agent_registry=agent_registry
    )
    
    # Presentation - Handlers
    chat_handler = providers.Singleton(
        ChatHandler,
        conversation_manager=conversation_manager
    )
    
    agent_handler = providers.Singleton(
        AgentHandler,
        agent_orchestrator=agent_orchestrator
    )