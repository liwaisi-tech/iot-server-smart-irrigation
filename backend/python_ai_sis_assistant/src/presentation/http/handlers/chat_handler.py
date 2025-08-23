"""Chat handler for agent interactions."""

import logging
from typing import Optional
from fastapi import HTTPException
from pydantic import BaseModel, Field

from src.usecases.conversation_manager import (
    ConversationManager, ProcessMessageRequest, ProcessMessageResponse
)
from src.domain.value_objects.conversation import ConversationId


logger = logging.getLogger(__name__)


class ChatRequest(BaseModel):
    """Request model for chat interactions."""
    message: str = Field(..., description="User message", min_length=1, max_length=1000)
    conversation_id: Optional[str] = Field(None, description="Existing conversation ID")
    user_id: Optional[str] = Field(None, description="User identifier")


class ChatResponse(BaseModel):
    """Response model for chat interactions."""
    response: str = Field(..., description="Agent response")
    conversation_id: str = Field(..., description="Conversation identifier")
    agent_type: str = Field(..., description="Type of agent that handled the request")
    metadata: Optional[dict] = Field(None, description="Additional response metadata")


class ConversationHistoryResponse(BaseModel):
    """Response model for conversation history."""
    conversation_id: str
    messages: list[dict]
    current_agent: str
    message_count: int


class ChatHandler:
    """HTTP handler for chat interactions."""
    
    def __init__(self, conversation_manager: ConversationManager):
        self.conversation_manager = conversation_manager
    
    async def process_chat_message(self, request: ChatRequest) -> ChatResponse:
        """Process a chat message and return agent response."""
        
        try:
            # Convert to use case request
            conversation_id = None
            if request.conversation_id:
                conversation_id = ConversationId(request.conversation_id)
            
            use_case_request = ProcessMessageRequest(
                content=request.message,
                conversation_id=conversation_id,
                user_id=request.user_id
            )
            
            # Process with conversation manager
            response = await self.conversation_manager.process_message(use_case_request)
            
            logger.info(f"Processed chat message for conversation {response.conversation_id}")
            
            return ChatResponse(
                response=response.content,
                conversation_id=response.conversation_id.value,
                agent_type=response.agent_type,
                metadata=response.metadata
            )
            
        except ValueError as e:
            logger.error(f"Validation error processing chat: {e}")
            raise HTTPException(status_code=400, detail=str(e))
        except Exception as e:
            logger.error(f"Unexpected error processing chat: {e}")
            raise HTTPException(status_code=500, detail="Error interno procesando el mensaje")
    
    async def get_conversation_history(
        self, 
        conversation_id: str, 
        limit: int = 50
    ) -> ConversationHistoryResponse:
        """Get conversation history."""
        
        try:
            conv_id = ConversationId(conversation_id)
            context = await self.conversation_manager.get_conversation_history(
                conv_id, limit
            )
            
            if not context:
                raise HTTPException(
                    status_code=404, 
                    detail="Conversación no encontrada"
                )
            
            # Convert messages to dict format
            messages = []
            for msg in context.messages:
                messages.append({
                    "id": msg.id,
                    "role": msg.role.value,
                    "content": msg.content,
                    "timestamp": msg.timestamp.isoformat(),
                    "agent_type": msg.agent_type
                })
            
            return ConversationHistoryResponse(
                conversation_id=conversation_id,
                messages=messages,
                current_agent=context.current_agent_type,
                message_count=len(messages)
            )
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Error retrieving conversation history: {e}")
            raise HTTPException(status_code=500, detail="Error obteniendo historial")
    
    async def end_conversation(self, conversation_id: str) -> dict:
        """End a conversation."""
        
        try:
            conv_id = ConversationId(conversation_id)
            success = await self.conversation_manager.end_conversation(conv_id)
            
            if not success:
                raise HTTPException(
                    status_code=404, 
                    detail="Conversación no encontrada"
                )
            
            logger.info(f"Ended conversation {conversation_id}")
            return {"message": "Conversación finalizada", "conversation_id": conversation_id}
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Error ending conversation: {e}")
            raise HTTPException(status_code=500, detail="Error finalizando conversación")