"""Local Assistant Agent implementation."""

import logging
from typing import Dict, Any

from src.domain.entities.agent import SpecializedAgent, AgentResponse
from src.domain.value_objects.agent_types import LOCAL_AGENT_PROFILE
from src.domain.value_objects.conversation import Message, ConversationContext
from src.domain.ports.llm_client import LLMClientPort, LLMRequest, LocalKnowledgePort


logger = logging.getLogger(__name__)


class LocalAssistantAgent(SpecializedAgent):
    """Local Assistant Agent specialized in IoT device management."""
    
    def __init__(
        self, 
        llm_client: LLMClientPort,
        local_knowledge: LocalKnowledgePort
    ):
        super().__init__(LOCAL_AGENT_PROFILE)
        self.llm_client = llm_client
        self.local_knowledge = local_knowledge
    
    async def process_message(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> AgentResponse:
        """Process a message related to local IoT system management."""
        
        # Gather relevant local data
        local_data = await self._gather_local_context(message)
        
        # Generate response with local context
        response_content = await self._generate_response_with_context(
            message, context, local_data
        )
        
        return AgentResponse(
            content=response_content,
            should_switch_agent=False,
            metadata={
                "agent_type": "local_assistant",
                "local_data_used": bool(local_data),
                "data_sources": list(local_data.keys()) if local_data else []
            }
        )
    
    def can_handle_request(self, message: Message, context: ConversationContext) -> bool:
        """Check if this agent can handle IoT-related requests."""
        content_lower = message.content.lower()
        
        local_keywords = [
            "dispositivo", "sensor", "temperatura", "humedad", "riego",
            "estado", "datos", "sistema", "configuración", "alarma",
            "encender", "apagar", "monitoreo"
        ]
        
        return any(keyword in content_lower for keyword in local_keywords)
    
    async def _gather_local_context(self, message: Message) -> Dict[str, Any]:
        """Gather relevant local system data."""
        context_data = {}
        content_lower = message.content.lower()
        
        try:
            # Get system status if requested
            if any(word in content_lower for word in ["estado", "sistema", "status"]):
                system_status = await self.local_knowledge.get_system_status()
                context_data["system_status"] = system_status
            
            # Get sensor data if requested
            if any(word in content_lower for word in ["sensor", "temperatura", "humedad", "datos"]):
                sensor_data = await self.local_knowledge.get_sensor_data()
                context_data["sensor_data"] = sensor_data
            
            # Search for specific device info
            device_keywords = ["dispositivo", "device"]
            if any(word in content_lower for word in device_keywords):
                # Extract potential device identifier from message
                device_info = await self.local_knowledge.search_device_info(message.content)
                context_data["device_info"] = device_info
            
        except Exception as e:
            logger.error(f"Error gathering local context: {e}")
            context_data["error"] = f"No se pudo obtener información del sistema: {str(e)}"
        
        return context_data
    
    async def _generate_response_with_context(
        self, 
        message: Message, 
        context: ConversationContext,
        local_data: Dict[str, Any]
    ) -> str:
        """Generate response using local system context."""
        
        # Build system prompt with local context
        system_prompt = self._build_system_prompt_with_data(local_data)
        
        # Include recent conversation context
        recent_messages = context.get_recent_messages(3)
        all_messages = recent_messages + [message]
        
        llm_request = LLMRequest(
            messages=all_messages,
            system_prompt=system_prompt,
            max_tokens=400,
            temperature=0.6
        )
        
        try:
            response = await self.llm_client.generate_response(llm_request)
            return response.content
        except Exception as e:
            logger.error(f"Error generating local assistant response: {e}")
            return self._generate_fallback_response(local_data)
    
    def _build_system_prompt_with_data(self, local_data: Dict[str, Any]) -> str:
        """Build system prompt including local system data."""
        
        base_prompt = """Eres el Asistente Local del Sistema IoT de Riego Inteligente de Liwaisi.

        Tu especialidad es:
        - Gestión de dispositivos IoT (sensores, actuadores)
        - Monitoreo de datos del sistema (temperatura, humedad)
        - Configuración y estado de los dispositivos
        - Resolución de problemas técnicos locales
        - Análisis de datos de sensores
        
        Características:
        - Conocimiento técnico profundo del sistema local
        - Acceso a datos en tiempo real
        - Hablas español con formalidad mixta/adaptativa
        - Eres preciso y técnico cuando es necesario
        - Siempre incluyes datos específicos cuando están disponibles"""
        
        # Add local data context
        if local_data:
            data_context = "\n\nInformación actual del sistema:"
            
            if "system_status" in local_data:
                data_context += f"\n- Estado del sistema: {local_data['system_status']}"
            
            if "sensor_data" in local_data:
                data_context += f"\n- Datos de sensores: {local_data['sensor_data']}"
            
            if "device_info" in local_data:
                data_context += f"\n- Información de dispositivos: {local_data['device_info']}"
            
            if "error" in local_data:
                data_context += f"\n- Nota: {local_data['error']}"
            
            base_prompt += data_context
        
        return base_prompt
    
    def _generate_fallback_response(self, local_data: Dict[str, Any]) -> str:
        """Generate a fallback response when LLM fails."""
        if "error" in local_data:
            return f"Hay un problema técnico accediendo a los datos del sistema: {local_data['error']}. Te recomiendo intentar de nuevo en unos momentos."
        
        if local_data:
            return "He recopilado información del sistema local. ¿Podrías ser más específico sobre qué aspecto te interesa?"
        
        return "Soy el Asistente Local del sistema IoT. Puedo ayudarte con información sobre dispositivos, sensores y estado del sistema. ¿Qué necesitas saber?"