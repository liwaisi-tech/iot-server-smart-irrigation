"""Use case for orchestrating agents and routing tasks."""

from typing import List, Optional, Dict, Any
from dataclasses import dataclass

from src.domain.entities.agent import Agent
from src.domain.value_objects.agent_types import AgentType, AgentCapability
from src.domain.value_objects.conversation import Message, ConversationContext
from src.domain.ports.agent_registry import AgentRegistryPort


@dataclass
class AgentHealthStatus:
    """Health status of an agent."""
    agent_type: AgentType
    is_available: bool
    last_check: str
    error_message: Optional[str] = None


@dataclass
class RouteTaskRequest:
    """Request to route a task to an appropriate agent."""
    message: Message
    context: ConversationContext
    preferred_agent_type: Optional[AgentType] = None


@dataclass
class RouteTaskResponse:
    """Response from routing a task."""
    selected_agent_type: AgentType
    confidence_score: float
    reasoning: str


class AgentOrchestrator:
    """Use case for orchestrating multiple agents and routing tasks."""
    
    def __init__(self, agent_registry: AgentRegistryPort):
        self.agent_registry = agent_registry
    
    async def initialize_agents(self) -> Dict[AgentType, bool]:
        """Initialize all agents in the system."""
        results = {}
        
        # Get all registered agents
        agents = await self.agent_registry.get_all_agents()
        
        for agent_type, agent in agents.items():
            try:
                # Perform health check or initialization
                is_available = await self.agent_registry.is_agent_available(agent_type)
                results[agent_type] = is_available
            except Exception as e:
                results[agent_type] = False
                # Log error in a real implementation
        
        return results
    
    async def route_task(self, request: RouteTaskRequest) -> RouteTaskResponse:
        """Route a task to the most appropriate agent."""
        
        # If preferred agent is specified and available, use it
        if request.preferred_agent_type:
            is_available = await self.agent_registry.is_agent_available(
                request.preferred_agent_type
            )
            if is_available:
                return RouteTaskResponse(
                    selected_agent_type=request.preferred_agent_type,
                    confidence_score=1.0,
                    reasoning="Using preferred agent type"
                )
        
        # Analyze message content to determine best agent
        best_agent_type = await self._analyze_message_for_routing(
            request.message, 
            request.context
        )
        
        # Verify agent is available
        is_available = await self.agent_registry.is_agent_available(best_agent_type)
        if not is_available:
            # Fallback to coordinator
            best_agent_type = AgentType.COORDINATOR
        
        confidence_score = await self._calculate_confidence_score(
            request.message, 
            best_agent_type
        )
        
        return RouteTaskResponse(
            selected_agent_type=best_agent_type,
            confidence_score=confidence_score,
            reasoning=f"Selected based on message analysis"
        )
    
    async def get_system_health(self) -> List[AgentHealthStatus]:
        """Get health status of all agents."""
        health_statuses = []
        agents = await self.agent_registry.get_all_agents()
        
        for agent_type in agents.keys():
            try:
                is_available = await self.agent_registry.is_agent_available(agent_type)
                health_statuses.append(
                    AgentHealthStatus(
                        agent_type=agent_type,
                        is_available=is_available,
                        last_check="now"  # In real implementation, use actual timestamp
                    )
                )
            except Exception as e:
                health_statuses.append(
                    AgentHealthStatus(
                        agent_type=agent_type,
                        is_available=False,
                        last_check="now",
                        error_message=str(e)
                    )
                )
        
        return health_statuses
    
    async def get_agents_by_capability(
        self, 
        capability: AgentCapability
    ) -> List[Agent]:
        """Get all agents that have a specific capability."""
        return await self.agent_registry.get_agents_by_capability(capability)
    
    async def _analyze_message_for_routing(
        self, 
        message: Message, 
        context: ConversationContext
    ) -> AgentType:
        """Analyze message content to determine the best agent using Spanish linguistic patterns."""
        import re
        
        content_lower = message.content.lower()
        
        # Local operations - immediate actions on IoT system
        local_action_verbs = [
            "muestra", "ver", "revisar", "verificar", "consultar", "comprobar", 
            "monitorear", "supervisar", "activar", "desactivar", "encender", 
            "apagar", "configurar", "ajustar", "modificar", "cambiar", 
            "programar", "obtener", "listar", "mostrar", "reportar"
        ]
        
        # Web search - research and learning actions
        web_search_verbs = [
            "buscar", "encontrar", "investigar", "explorar", "averiguar", 
            "indagar", "aprender", "estudiar", "entender", "explicar", 
            "enseñar", "informar", "busca", "búsqueda"
        ]
        
        # Domain-specific keywords
        iot_domain_keywords = [
            "dispositivo", "sensor", "temperatura", "humedad", "riego",
            "estado", "datos", "sistema", "configuración", "alarma",
            "medición", "lectura", "valor", "umbral"
        ]
        
        web_research_keywords = [
            "google", "internet", "web", "online", "información", 
            "documentación", "manual", "tutorial", "guía", "wikipedia",
            "artículo", "estudio", "investigación"
        ]
        
        # Colombian Spanish expressions
        colombian_local = ["cómo va", "qué tal", "de una", "ya mismo"]
        colombian_web = ["hacer el favor", "colaborar"]
        
        # Question patterns using regex
        local_question_patterns = [
            r"¿?qué.*(?:temperatura|humedad|estado|dispositivo|sensor)",
            r"¿?cuál.*(?:sensor|dispositivo|configuración|valor)",
            r"¿?cómo está.*(?:sistema|riego|sensor|dispositivo)",
            r"¿?cuándo.*(?:último|riego|medición|dato)"
        ]
        
        web_search_patterns = [
            r"¿?cómo.*(?:funciona|instalar|configurar|usar|hacer|se)",
            r"¿?por qué.*(?:es importante|funciona|se usa)",
            r"¿?qué es.*(?:un|una|el|la)",
            r"¿?para qué sirve"
        ]
        
        # Check Colombian expressions first (more specific)
        for expr in colombian_local:
            if expr in content_lower:
                return AgentType.LOCAL_ASSISTANT
        
        for expr in colombian_web:
            if expr in content_lower:
                return AgentType.WEB_SEARCH
        
        # Check question patterns
        for pattern in local_question_patterns:
            if re.search(pattern, content_lower):
                return AgentType.LOCAL_ASSISTANT
        
        for pattern in web_search_patterns:
            if re.search(pattern, content_lower):
                return AgentType.WEB_SEARCH
        
        # Check action verbs + domain context
        has_local_verb = any(verb in content_lower for verb in local_action_verbs)
        has_web_verb = any(verb in content_lower for verb in web_search_verbs)
        has_iot_context = any(keyword in content_lower for keyword in iot_domain_keywords)
        has_web_context = any(keyword in content_lower for keyword in web_research_keywords)
        
        # Decision logic based on verb + context combinations
        if has_local_verb and has_iot_context:
            return AgentType.LOCAL_ASSISTANT
        elif has_web_verb or has_web_context:
            return AgentType.WEB_SEARCH
        elif has_iot_context and not has_web_context:
            return AgentType.LOCAL_ASSISTANT
        else:
            # Default to coordinator for general questions
            return AgentType.COORDINATOR
    
    async def _calculate_confidence_score(
        self, 
        message: Message, 
        agent_type: AgentType
    ) -> float:
        """Calculate confidence score for agent selection using linguistic pattern matching."""
        import re
        
        content_lower = message.content.lower()
        base_confidence = 0.5
        
        if agent_type == AgentType.WEB_SEARCH:
            # High confidence indicators for web search
            web_verbs = ["buscar", "investigar", "aprender", "estudiar", "explicar"]
            web_keywords = ["información", "documentación", "manual", "tutorial", "guía"]
            web_patterns = [r"¿?cómo.*funciona", r"¿?qué es", r"¿?por qué"]
            colombian_web = ["hacer el favor", "colaborar"]
            
            score = base_confidence
            
            # Verb matches (+0.2 each)
            verb_matches = sum(1 for verb in web_verbs if verb in content_lower)
            score += min(verb_matches * 0.2, 0.4)
            
            # Keyword matches (+0.15 each)
            keyword_matches = sum(1 for keyword in web_keywords if keyword in content_lower)
            score += min(keyword_matches * 0.15, 0.3)
            
            # Pattern matches (+0.25 each)
            pattern_matches = sum(1 for pattern in web_patterns if re.search(pattern, content_lower))
            score += min(pattern_matches * 0.25, 0.5)
            
            # Colombian expressions (+0.3)
            if any(expr in content_lower for expr in colombian_web):
                score += 0.3
                
            return min(score, 1.0)
        
        elif agent_type == AgentType.LOCAL_ASSISTANT:
            # High confidence indicators for local operations
            local_verbs = ["muestra", "ver", "revisar", "activar", "configurar", "obtener"]
            iot_keywords = ["dispositivo", "sensor", "temperatura", "humedad", "riego", "estado"]
            local_patterns = [r"¿?qué.*(?:temperatura|estado)", r"¿?cómo está.*sistema"]
            colombian_local = ["cómo va", "qué tal", "de una", "ya mismo"]
            
            score = base_confidence
            
            # Verb matches (+0.2 each)
            verb_matches = sum(1 for verb in local_verbs if verb in content_lower)
            score += min(verb_matches * 0.2, 0.4)
            
            # IoT keyword matches (+0.15 each)
            keyword_matches = sum(1 for keyword in iot_keywords if keyword in content_lower)
            score += min(keyword_matches * 0.15, 0.3)
            
            # Pattern matches (+0.25 each)
            pattern_matches = sum(1 for pattern in local_patterns if re.search(pattern, content_lower))
            score += min(pattern_matches * 0.25, 0.5)
            
            # Colombian expressions (+0.3)
            if any(expr in content_lower for expr in colombian_local):
                score += 0.3
                
            return min(score, 1.0)
        
        else:  # COORDINATOR
            return 0.7  # Default confidence for coordinator