"""Web Search Agent implementation."""

import logging
from typing import Dict, Any, List

from src.domain.entities.agent import SpecializedAgent, AgentResponse
from src.domain.value_objects.agent_types import WEB_SEARCH_AGENT_PROFILE
from src.domain.value_objects.conversation import Message, ConversationContext
from src.domain.ports.llm_client import LLMClientPort, LLMRequest, WebSearchPort


logger = logging.getLogger(__name__)


class WebSearchAgent(SpecializedAgent):
    """Web Search Agent specialized in finding external information."""
    
    def __init__(
        self, 
        llm_client: LLMClientPort,
        web_search: WebSearchPort
    ):
        super().__init__(WEB_SEARCH_AGENT_PROFILE)
        self.llm_client = llm_client
        self.web_search = web_search
    
    async def process_message(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> AgentResponse:
        """Process a message requiring web search."""
        
        # Extract search query from message
        search_query = await self._extract_search_query(message)
        
        # Perform web search
        search_results = await self._perform_search(search_query)
        
        # Generate response with search results
        response_content = await self._generate_response_with_results(
            message, context, search_results, search_query
        )
        
        return AgentResponse(
            content=response_content,
            should_switch_agent=False,
            metadata={
                "agent_type": "web_search",
                "search_query": search_query,
                "results_count": len(search_results),
                "sources": [result.get("url", "") for result in search_results]
            }
        )
    
    def can_handle_request(self, message: Message, context: ConversationContext) -> bool:
        """Check if this agent can handle web search requests."""
        content_lower = message.content.lower()
        
        web_search_keywords = [
            "busca", "búsqueda", "buscar", "información sobre", "qué es",
            "documentación", "manual", "tutorial", "guía", "aprende",
            "internet", "web", "online", "google"
        ]
        
        return any(keyword in content_lower for keyword in web_search_keywords)
    
    async def _extract_search_query(self, message: Message) -> str:
        """Extract search query from user message."""
        system_prompt = """Extrae la consulta de búsqueda principal del siguiente mensaje del usuario.
        
        Reglas:
        - Devuelve solo los términos de búsqueda más relevantes
        - Elimina palabras como "busca", "búsqueda", "información sobre"
        - Mantén los términos técnicos importantes
        - Responde solo con la consulta, sin explicaciones
        - Si es sobre riego o IoT, incluye esos términos
        
        Ejemplo:
        Usuario: "Busca información sobre sensores de humedad para riego"
        Respuesta: "sensores humedad riego IoT"
        """
        
        llm_request = LLMRequest(
            messages=[message],
            system_prompt=system_prompt,
            max_tokens=50,
            temperature=0.3
        )
        
        try:
            response = await self.llm_client.generate_response(llm_request)
            query = response.content.strip().strip('"\'')
            return query if query else message.content
        except Exception as e:
            logger.error(f"Error extracting search query: {e}")
            # Fallback: use original message content
            return message.content
    
    async def _perform_search(self, query: str) -> List[Dict[str, Any]]:
        """Perform web search and return results."""
        try:
            results = await self.web_search.search(query, max_results=5)
            logger.info(f"Found {len(results)} search results for query: {query}")
            return results
        except Exception as e:
            logger.error(f"Error performing web search: {e}")
            return []
    
    async def _generate_response_with_results(
        self, 
        message: Message, 
        context: ConversationContext,
        search_results: List[Dict[str, Any]],
        search_query: str
    ) -> str:
        """Generate response using web search results."""
        
        # Build system prompt with search results
        system_prompt = self._build_system_prompt_with_results(
            search_results, search_query
        )
        
        # Include recent conversation context
        recent_messages = context.get_recent_messages(3)
        all_messages = recent_messages + [message]
        
        llm_request = LLMRequest(
            messages=all_messages,
            system_prompt=system_prompt,
            max_tokens=500,
            temperature=0.7
        )
        
        try:
            response = await self.llm_client.generate_response(llm_request)
            return response.content
        except Exception as e:
            logger.error(f"Error generating web search response: {e}")
            return self._generate_fallback_response(search_results, search_query)
    
    def _build_system_prompt_with_results(
        self, 
        search_results: List[Dict[str, Any]], 
        search_query: str
    ) -> str:
        """Build system prompt including search results."""
        
        base_prompt = """Eres el Especialista en Búsqueda Web del Sistema IoT de Riego Inteligente de Liwaisi.

        Tu especialidad es:
        - Búsqueda y análisis de información externa
        - Documentación técnica y manuales
        - Investigación sobre tecnologías IoT y riego
        - Comparación de productos y soluciones
        - Tendencias y mejores prácticas del sector
        
        Características:
        - Siempre citas las fuentes de información
        - Proporcionas información actualizada y verificada
        - Hablas español con formalidad mixta/adaptativa
        - Relacionas la información encontrada con el contexto IoT/riego cuando es relevante
        - Eres crítico y analítico con las fuentes"""
        
        # Add search results context
        if search_results:
            results_context = f"\n\nResultados de búsqueda para '{search_query}':\n"
            
            for i, result in enumerate(search_results[:5], 1):
                title = result.get("title", "Sin título")
                url = result.get("url", "")
                snippet = result.get("snippet", "Sin descripción")
                
                results_context += f"\n{i}. {title}\n"
                results_context += f"   URL: {url}\n"
                results_context += f"   Resumen: {snippet}\n"
            
            base_prompt += results_context
            base_prompt += "\n\nUsa esta información para responder la consulta del usuario, citando las fuentes apropiadas."
        else:
            base_prompt += f"\n\nNo se encontraron resultados para la búsqueda '{search_query}'. Informa al usuario y sugiere términos alternativos o fuentes conocidas."
        
        return base_prompt
    
    def _generate_fallback_response(
        self, 
        search_results: List[Dict[str, Any]], 
        search_query: str
    ) -> str:
        """Generate a fallback response when LLM fails."""
        if not search_results:
            return f"No pude encontrar resultados relevantes para '{search_query}'. Te sugiero probar con términos más específicos o contactar directamente con proveedores especializados en IoT y riego."
        
        response = f"Encontré {len(search_results)} resultados para '{search_query}':\n\n"
        
        for i, result in enumerate(search_results[:3], 1):
            title = result.get("title", "Resultado sin título")
            url = result.get("url", "")
            snippet = result.get("snippet", "")
            
            response += f"{i}. **{title}**\n"
            if snippet:
                response += f"   {snippet}\n"
            if url:
                response += f"   Fuente: {url}\n"
            response += "\n"
        
        response += "¿Te gustaría que profundice en alguno de estos resultados?"
        return response