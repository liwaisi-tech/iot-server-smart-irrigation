"""Use case for managing conversations with agents."""

from typing import Optional
from dataclasses import dataclass

from src.domain.entities.agent import Agent, AgentResponse
from src.domain.value_objects.conversation import (
    ConversationId, ConversationContext, Message, ConversationState
)
from src.domain.value_objects.agent_types import AgentType
from src.domain.ports.conversation_repository import ConversationRepositoryPort
from src.domain.ports.agent_registry import AgentRegistryPort


@dataclass
class ProcessMessageRequest:
    """Request to process a message."""
    content: str
    conversation_id: Optional[ConversationId] = None
    user_id: Optional[str] = None


@dataclass
class ProcessMessageResponse:
    """Response from processing a message."""
    content: str
    conversation_id: ConversationId
    agent_type: str
    metadata: dict = None


class ConversationManager:
    """Use case for managing conversations between users and agents."""
    
    def __init__(
        self,
        conversation_repo: ConversationRepositoryPort,
        agent_registry: AgentRegistryPort
    ):
        self.conversation_repo = conversation_repo
        self.agent_registry = agent_registry
    
    async def process_message(self, request: ProcessMessageRequest) -> ProcessMessageResponse:
        """Process a user message and return agent response."""
        
        # Get or create conversation context
        if request.conversation_id:
            context = await self.conversation_repo.get_conversation(request.conversation_id)
            if not context:
                raise ValueError(f"Conversation {request.conversation_id} not found")
        else:
            context = await self._create_new_conversation()
        
        # Create user message
        user_message = Message.create_user_message(request.content)
        context.add_message(user_message)
        
        # Get appropriate agent
        agent = await self._get_agent_for_context(context)
        if not agent:
            raise ValueError("No suitable agent found")
        
        # Process message with agent
        agent_response = await agent.process_message(user_message, context)
        
        # Create assistant message
        assistant_message = Message.create_assistant_message(
            content=agent_response.content,
            agent_type=agent.agent_type
        )
        context.add_message(assistant_message)
        
        # Handle agent switching if needed
        if agent_response.should_switch_agent and agent_response.target_agent_type:
            context.switch_agent(agent_response.target_agent_type)
        
        # Save conversation
        await self.conversation_repo.update_conversation(context)
        
        return ProcessMessageResponse(
            content=agent_response.content,
            conversation_id=context.conversation_id,
            agent_type=context.current_agent_type,
            metadata=agent_response.metadata
        )
    
    async def get_conversation_history(
        self, 
        conversation_id: ConversationId,
        limit: int = 50
    ) -> Optional[ConversationContext]:
        """Get conversation history."""
        context = await self.conversation_repo.get_conversation(conversation_id)
        if context:
            # Return only recent messages
            context.messages = context.get_recent_messages(limit)
        return context
    
    async def end_conversation(self, conversation_id: ConversationId) -> bool:
        """End a conversation."""
        context = await self.conversation_repo.get_conversation(conversation_id)
        if not context:
            return False
        
        context.state = ConversationState.COMPLETED
        await self.conversation_repo.update_conversation(context)
        return True
    
    async def _create_new_conversation(self) -> ConversationContext:
        """Create a new conversation context."""
        context = ConversationContext(
            conversation_id=ConversationId(),
            state=ConversationState.ACTIVE,
            current_agent_type=AgentType.COORDINATOR.value
        )
        await self.conversation_repo.save_conversation(context)
        return context
    
    async def _get_agent_for_context(self, context: ConversationContext) -> Optional[Agent]:
        """Get the appropriate agent for the current context."""
        agent_type = AgentType(context.current_agent_type)
        return await self.agent_registry.get_agent(agent_type)