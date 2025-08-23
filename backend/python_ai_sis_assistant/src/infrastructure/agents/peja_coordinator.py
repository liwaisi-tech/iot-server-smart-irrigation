"""Peja - Coordinator Agent implementation."""

import logging

from src.domain.entities.agent import CoordinatorAgent, AgentResponse
from src.domain.value_objects.agent_types import (
    PEJA_COORDINATOR_PROFILE, AgentType
)
from src.domain.value_objects.conversation import Message, ConversationContext
from src.domain.ports.llm_client import LLMClientPort, LLMRequest


logger = logging.getLogger(__name__)


class PejaCoordinatorAgent(CoordinatorAgent):
    """Peja - The coordinator agent for the IoT Smart Irrigation System."""
    
    def __init__(self, llm_client: LLMClientPort):
        super().__init__(PEJA_COORDINATOR_PROFILE)
        self.llm_client = llm_client
    
    async def process_message(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> AgentResponse:
        """Process a message as the coordinator."""
        
        # Check if we should route to a specialized agent
        _, target_agent = await self.should_switch_agent(message, context)
        should_switch = target_agent is not None
        
        if should_switch and target_agent:
            # Generate a routing response
            response_content = await self._generate_routing_response(message, target_agent)
            return AgentResponse(
                content=response_content,
                should_switch_agent=True,
                target_agent_type=target_agent,
                metadata={"routing_decision": True}
            )
        
        # Handle the message directly
        response_content = await self._generate_direct_response(message, context)
        
        return AgentResponse(
            content=response_content,
            should_switch_agent=False,
            metadata={"handled_by": "coordinator"}
        )
    
    def can_handle_request(self, message: Message, context: ConversationContext) -> bool:
        """Coordinator can handle any request as fallback."""
        return True
    
    async def route_request(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> str:
        """Route a request to the appropriate agent."""
        should_switch, target_agent = await self.should_switch_agent(message, context)
        return target_agent if target_agent else AgentType.COORDINATOR.value
    
    async def should_switch_agent(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> tuple[bool, str | None]:
        """Determine if we should switch to another agent."""
        content_lower = message.content.lower()
        
        # Web search keywords (Spanish)
        web_search_keywords = [
            "busca", "búsqueda", "buscar", "información sobre", "qué es",
            "documentación", "manual", "tutorial", "guía", "aprende",
            "internet", "web", "online", "google"
        ]
        
        # Local IoT keywords (Spanish)
        local_keywords = [
            "dispositivo", "sensor", "temperatura", "humedad", "riego",
            "estado del sistema", "datos", "configuración", "alarma",
            "encender", "apagar", "monitoreo", "estado"
        ]
        
        # Check for web search intent
        if any(keyword in content_lower for keyword in web_search_keywords):
            return True, AgentType.WEB_SEARCH.value
        
        # Check for local system management intent
        if any(keyword in content_lower for keyword in local_keywords):
            return True, AgentType.LOCAL_ASSISTANT.value
        
        # Check for explicit agent requests
        if "asistente local" in content_lower or "sistema local" in content_lower:
            return True, AgentType.LOCAL_ASSISTANT.value
        
        if "buscar en internet" in content_lower or "búsqueda web" in content_lower:
            return True, AgentType.WEB_SEARCH.value
        
        return False, None
    
    async def _generate_routing_response(self, message: Message, target_agent: str) -> str:
        """Generate a response when routing to another agent."""
        agent_names = {
            AgentType.LOCAL_ASSISTANT.value: "Asistente Local",
            AgentType.WEB_SEARCH.value: "Especialista en Búsqueda Web"
        }
        
        agent_name = agent_names.get(target_agent, "especialista apropiado")
        
        system_prompt = f"""Eres Peja, el coordinador del Sistema IoT de Riego Inteligente.
        
        El usuario ha hecho una pregunta que requiere derivar al {agent_name}.
        Genera una respuesta breve y amigable explicando que estás derivando la consulta
        al especialista apropiado. Mantén un tono profesional pero cercano."""
        
        llm_request = LLMRequest(
            messages=[message],
            system_prompt=system_prompt,
            max_tokens=150,
            temperature=0.7
        )
        
        try:
            response = await self.llm_client.generate_response(llm_request)
            return response.content
        except Exception as e:
            logger.error(f"Error generating routing response: {e}")
            return f"Te voy a derivar con el {agent_name} quien puede ayudarte mejor con tu consulta."
    
    async def _generate_direct_response(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> str:
        """Generate a direct response as coordinator."""
        
        system_prompt = """Eres Peja, el coordinador del Sistema IoT de Riego Inteligente de Liwaisi.

        Tu rol principal es:
        - Coordinar las consultas entre diferentes especialistas
        - Proporcionar información general sobre el sistema
        - Ayudar con preguntas administrativas y de navegación
        - Mantener conversaciones amigables sobre el sistema de riego
        
        Características de tu personalidad:
        - Profesional pero cercano
        - Conoces el sistema a nivel general
        - Siempre derivarás a especialistas cuando sea necesario
        - Hablas en español con formalidad mixta/adaptativa
        
        Si la pregunta requiere conocimiento técnico específico o búsqueda de información externa,
        indica que puedes derivar al especialista apropiado."""
        
        # Include recent conversation context
        recent_messages = context.get_recent_messages(5)
        all_messages = recent_messages + [message]
        
        llm_request = LLMRequest(
            messages=all_messages,
            system_prompt=system_prompt,
            max_tokens=300,
            temperature=0.7
        )
        
        try:
            response = await self.llm_client.generate_response(llm_request)
            return response.content
        except Exception as e:
            logger.error(f"Error generating direct response: {e}")
            return "Hola, soy Peja, coordinador del sistema. ¿En qué puedo ayudarte hoy?"