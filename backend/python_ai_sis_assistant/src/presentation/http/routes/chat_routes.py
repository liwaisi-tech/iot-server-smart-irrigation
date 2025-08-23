"""Chat routes for agent interactions."""

from fastapi import APIRouter, Depends, Query
from dependency_injector.wiring import inject, Provide

from src.app.container import Container
from src.presentation.http.handlers.chat_handler import (
    ChatHandler, ChatRequest, ChatResponse, ConversationHistoryResponse
)


router = APIRouter(prefix="/chat", tags=["chat"])


@router.post("/message", response_model=ChatResponse)
@inject
async def send_message(
    request: ChatRequest,
    chat_handler: ChatHandler = Depends(Provide[Container.chat_handler])
) -> ChatResponse:
    """Send a message to the agent system."""
    return await chat_handler.process_chat_message(request)


@router.get("/conversation/{conversation_id}", response_model=ConversationHistoryResponse)
@inject
async def get_conversation_history(
    conversation_id: str,
    limit: int = Query(default=50, ge=1, le=100, description="Maximum number of messages"),
    chat_handler: ChatHandler = Depends(Provide[Container.chat_handler])
) -> ConversationHistoryResponse:
    """Get conversation history."""
    return await chat_handler.get_conversation_history(conversation_id, limit)


@router.delete("/conversation/{conversation_id}")
@inject
async def end_conversation(
    conversation_id: str,
    chat_handler: ChatHandler = Depends(Provide[Container.chat_handler])
) -> dict:
    """End a conversation."""
    return await chat_handler.end_conversation(conversation_id)