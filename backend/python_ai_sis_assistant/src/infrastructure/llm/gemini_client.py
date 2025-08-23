"""Google Gemini LLM client implementation."""

import logging
from typing import Dict, Any
import google.adk as adk

from src.domain.ports.llm_client import LLMClientPort, LLMRequest, LLMResponse
from src.domain.value_objects.conversation import Message, MessageRole


logger = logging.getLogger(__name__)


class GeminiClient(LLMClientPort):
    """Google Gemini LLM client using ADK."""
    
    def __init__(self, api_key: str, model_name: str = "gemini-pro"):
        self.api_key = api_key
        self.model_name = model_name
        self._client = None
    
    def _ensure_client(self):
        """Ensure the ADK client is initialized."""
        if self._client is None:
            try:
                adk.configure(api_key=self.api_key)
                self._client = adk.client()
                logger.info(f"Initialized Gemini client with model: {self.model_name}")
            except Exception as e:
                logger.error(f"Failed to initialize Gemini client: {e}")
                raise
    
    async def generate_response(self, request: LLMRequest) -> LLMResponse:
        """Generate a response using Gemini."""
        self._ensure_client()
        
        try:
            # Convert messages to ADK format
            adk_messages = self._convert_messages_to_adk(request.messages)
            
            # Prepare the prompt
            full_prompt = self._build_full_prompt(request.system_prompt, adk_messages)
            
            # Make the request to Gemini
            response = await self._client.generate_content(
                prompt=full_prompt,
                max_tokens=request.max_tokens,
                temperature=request.temperature
            )
            
            logger.info(f"Generated response with {response.get('tokens_used', 0)} tokens")
            
            return LLMResponse(
                content=response.get("content", ""),
                tokens_used=response.get("tokens_used"),
                metadata={"model": self.model_name}
            )
            
        except Exception as e:
            logger.error(f"Error generating Gemini response: {e}")
            # Return a fallback response
            return LLMResponse(
                content="Lo siento, no pude procesar tu solicitud en este momento. Intenta de nuevo más tarde.",
                tokens_used=0,
                metadata={"error": str(e)}
            )
    
    async def health_check(self) -> bool:
        """Check if Gemini service is available."""
        try:
            self._ensure_client()
            # Simple test request
            test_request = LLMRequest(
                messages=[Message.create_user_message("test")],
                system_prompt="Respond with 'OK'",
                max_tokens=10
            )
            response = await self.generate_response(test_request)
            return response.content is not None and "error" not in response.metadata
        except Exception as e:
            logger.error(f"Gemini health check failed: {e}")
            return False
    
    def _convert_messages_to_adk(self, messages: list[Message]) -> list[Dict[str, Any]]:
        """Convert domain messages to ADK format."""
        adk_messages = []
        for msg in messages:
            adk_messages.append({
                "role": self._map_role(msg.role),
                "content": msg.content,
                "timestamp": msg.timestamp.isoformat()
            })
        return adk_messages
    
    def _map_role(self, role: MessageRole) -> str:
        """Map domain roles to ADK roles."""
        mapping = {
            MessageRole.USER: "user",
            MessageRole.ASSISTANT: "assistant", 
            MessageRole.SYSTEM: "system"
        }
        return mapping.get(role, "user")
    
    def _build_full_prompt(self, system_prompt: str, messages: list[Dict[str, Any]]) -> str:
        """Build the full prompt for Gemini."""
        prompt_parts = [system_prompt]
        
        for msg in messages:
            role = msg["role"]
            content = msg["content"]
            prompt_parts.append(f"{role.title()}: {content}")
        
        return "\n\n".join(prompt_parts)