# Technical Documentation: Python AI SIS Assistant
## Conversational AI Agent for IoT Smart Irrigation System

### Document Version: 1.0
### Date: August 16, 2025
### Target Audience: Python Software Engineers, System Architects

---

## Table of Contents

1. [Project Overview and Objectives](#1-project-overview-and-objectives)
2. [System Architecture and Components](#2-system-architecture-and-components)
3. [ADK Integration Strategy](#3-adk-integration-strategy)
4. [MCP Tool Design for Device Discovery](#4-mcp-tool-design-for-device-discovery)
5. [Natural Language Processing Requirements](#5-natural-language-processing-requirements)
6. [Conversation Flow Design](#6-conversation-flow-design)
7. [Error Handling and User Experience](#7-error-handling-and-user-experience)
8. [Implementation Phases](#8-implementation-phases)
9. [Technology Stack and Dependencies](#9-technology-stack-and-dependencies)
10. [API Design and Interfaces](#10-api-design-and-interfaces)

---

## 1. Project Overview and Objectives

### 1.1 Project Context

The Python AI SIS Assistant is a conversational AI agent component of the IoT Smart Irrigation System, designed to provide farmers and agricultural users in Colombia with an intuitive natural language interface for interacting with their IoT irrigation infrastructure.

### 1.2 Primary Objectives

- **Conversational Interface**: Enable farmers to query and control their irrigation system using natural Colombian Spanish
- **Device Discovery**: Automatically discover and manage connected IoT devices through MCP tools
- **Real-time Data Access**: Provide immediate access to sensor data (temperature, humidity, soil moisture)
- **User-Friendly Experience**: Deliver patient, helpful responses with natural conversation flow
- **System Integration**: Seamlessly integrate with existing Go backend and PostgreSQL infrastructure

### 1.3 Target Users

- **Primary**: Farmers and agricultural workers in Colombia
- **Secondary**: Agricultural technicians and system administrators
- **Language**: Colombian Spanish (neutral dialect)
- **Technical Level**: Non-technical to semi-technical users

### 1.4 Core Use Cases

1. **Device Status Queries**: "¿Cómo están mis sensores en la zona norte?"
2. **Environmental Data**: "¿Cuál es la temperatura actual del invernadero?"
3. **System Health**: "¿Todos mis dispositivos están funcionando bien?"
4. **Location-based Queries**: "Muéstrame los datos del sensor en el jardín"
5. **Troubleshooting**: "Mi sensor no responde, ¿qué puedo hacer?"

---

## 2. System Architecture and Components

### 2.1 Overall Architecture

The Python AI SIS Assistant follows **Hexagonal Architecture** principles, ensuring clean separation of concerns and high testability. The system operates as a standalone backend service that integrates with the existing IoT infrastructure.

```
┌─────────────────────────────────────────────────────────────┐
│                     External Systems                        │
├─────────────────────────────────────────────────────────────┤
│  Google ADK API │ IoT Devices │ Go Backend │ PostgreSQL     │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
├─────────────────────────────────────────────────────────────┤
│ • HTTP Client (ADK)    • MCP Tools      • Database Client  │
│ • Device Discovery     • Event Bus      • Cache Layer      │
│ • External APIs        • Messaging      • Persistence      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
├─────────────────────────────────────────────────────────────┤
│ • Conversation Service    • Device Service                  │
│ • NLP Processing         • Data Aggregation                │
│ • Context Management     • Response Generation             │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
├─────────────────────────────────────────────────────────────┤
│ • Conversation Entity    • Device Entity                    │
│ • User Intent           • Sensor Data                      │
│ • Business Rules        • Domain Events                    │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Presentation Layer                        │
├─────────────────────────────────────────────────────────────┤
│ • FastAPI Endpoints     • WebSocket Handler                │
│ • Request Validation    • Response Formatting              │
│ • Authentication        • Error Handling                   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Core Components

#### 2.2.1 Conversation Engine
- **Purpose**: Process natural language inputs and generate contextual responses
- **Components**: Intent recognition, entity extraction, response generation
- **Integration**: Google ADK for advanced NLP capabilities

#### 2.2.2 Device Management System
- **Purpose**: Discover, monitor, and interact with IoT devices
- **Components**: Device registry, health monitoring, data collection
- **Integration**: MCP tools for device discovery via `/whoami` endpoints

#### 2.2.3 Context Management
- **Purpose**: Maintain conversation state and user context
- **Components**: Session management, conversation history, user preferences
- **Storage**: Redis cache for session data, PostgreSQL for persistent context

#### 2.2.4 Data Aggregation Service
- **Purpose**: Collect and process sensor data from multiple devices
- **Components**: Data collectors, processors, formatters
- **Sources**: Direct device APIs, Go backend, cached data

### 2.3 Current Infrastructure Integration

The assistant integrates with existing components:

- **Go SOC Consumer**: Primary backend service on port 8080
- **PostgreSQL**: Device registry and historical data storage
- **NATS**: Event-driven communication
- **MQTT**: IoT device messaging
- **PgBouncer**: Database connection pooling

---

## 3. ADK Integration Strategy

### 3.1 Google ADK Overview

Google's AI Development Kit (ADK) provides advanced natural language processing capabilities specifically designed for conversational AI applications.

### 3.2 Integration Architecture

```python
# ADK Integration Layer
class ADKService:
    """Service for Google ADK integration."""
    
    def __init__(self, api_key: str, model_config: ADKConfig):
        self.client = ADKClient(api_key)
        self.config = model_config
    
    async def process_conversation(
        self, 
        message: str, 
        context: ConversationContext
    ) -> ConversationResponse:
        """Process user message through ADK."""
        pass
    
    async def extract_intent(self, message: str) -> Intent:
        """Extract user intent from message."""
        pass
    
    async def extract_entities(self, message: str) -> List[Entity]:
        """Extract entities (devices, locations, etc.)."""
        pass
```

### 3.3 ADK Configuration

#### 3.3.1 Model Selection
- **Primary Model**: Gemini 1.5 Pro for complex reasoning
- **Fallback Model**: Gemini 1.5 Flash for quick responses
- **Specialized Models**: Fine-tuned models for agricultural domain

#### 3.3.2 Prompt Engineering
```python
SYSTEM_PROMPT = """
Eres un asistente inteligente para sistemas de riego IoT en Colombia. 
Tu objetivo es ayudar a agricultores a monitorear y controlar sus 
sistemas de riego de manera natural y amigable.

Características principales:
- Responde en español colombiano neutro
- Sé paciente y explicativo
- Proporciona información técnica de manera simple
- Sugiere acciones cuando sea apropiado
- Maneja errores con empatía

Contexto del sistema:
- Tienes acceso a dispositivos IoT con sensores
- Puedes consultar datos en tiempo real
- Conoces la ubicación de los dispositivos
- Puedes detectar problemas de conectividad
"""

USER_PROMPT_TEMPLATE = """
Usuario: {user_message}
Contexto de la conversación: {conversation_context}
Dispositivos disponibles: {available_devices}
Datos recientes: {recent_data}

Responde de manera útil y natural.
"""
```

### 3.4 ADK Features Utilization

1. **Intent Classification**: Categorize user requests (status, data, control)
2. **Entity Recognition**: Extract device names, locations, sensor types
3. **Context Understanding**: Maintain conversation flow and references
4. **Response Generation**: Create natural, contextual Spanish responses
5. **Error Recovery**: Handle misunderstandings and guide users

---

## 4. MCP Tool Design for Device Discovery

### 4.1 Model Context Protocol (MCP) Overview

MCP tools enable the AI agent to interact with external systems and gather real-time information about IoT devices in the irrigation system.

### 4.2 Device Discovery MCP Tool

#### 4.2.1 Tool Definition

```python
from typing import List, Dict, Any
from pydantic import BaseModel

class DeviceInfo(BaseModel):
    """Device information model."""
    mac_address: str
    device_name: str
    ip_address: str
    location_description: str
    status: str
    device_type: str
    last_seen: str

class MCPDeviceDiscoveryTool:
    """MCP tool for discovering IoT devices."""
    
    name = "device_discovery"
    description = "Discover and get information about IoT irrigation devices"
    
    async def execute(self, parameters: Dict[str, Any]) -> List[DeviceInfo]:
        """Execute device discovery."""
        devices = await self._discover_devices()
        return devices
    
    async def _discover_devices(self) -> List[DeviceInfo]:
        """Discover devices via /whoami endpoints."""
        discovered_devices = []
        
        # Get device list from Go backend or network scan
        device_ips = await self._get_device_ips()
        
        for ip in device_ips:
            try:
                device_info = await self._query_device_whoami(ip)
                if device_info:
                    discovered_devices.append(device_info)
            except Exception as e:
                logger.warning(f"Failed to query device at {ip}: {e}")
        
        return discovered_devices
    
    async def _query_device_whoami(self, ip: str) -> DeviceInfo:
        """Query device /whoami endpoint."""
        url = f"http://{ip}/whoami"
        async with httpx.AsyncClient() as client:
            response = await client.get(url, timeout=5.0)
            response.raise_for_status()
            data = response.json()
            
            return DeviceInfo(
                mac_address=data["mac_address"],
                device_name=data["device_name"],
                ip_address=ip,
                location_description=data["location_description"],
                status="online",
                device_type=data.get("device_type", "sensor"),
                last_seen=datetime.utcnow().isoformat()
            )
```

#### 4.2.2 Sensor Data MCP Tool

```python
class MCPSensorDataTool:
    """MCP tool for collecting sensor data."""
    
    name = "sensor_data"
    description = "Get real-time sensor data from IoT devices"
    
    async def execute(self, parameters: Dict[str, Any]) -> Dict[str, Any]:
        """Get sensor data from specified device."""
        device_ip = parameters.get("device_ip")
        sensor_type = parameters.get("sensor_type", "temperature-and-humidity")
        
        return await self._get_sensor_data(device_ip, sensor_type)
    
    async def _get_sensor_data(self, ip: str, sensor_type: str) -> Dict[str, Any]:
        """Get sensor data from device endpoint."""
        url = f"http://{ip}/{sensor_type}"
        async with httpx.AsyncClient() as client:
            response = await client.get(url, timeout=5.0)
            response.raise_for_status()
            return response.json()
```

### 4.3 Device Registry Integration

#### 4.3.1 Database Schema
```sql
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mac_address VARCHAR(17) UNIQUE NOT NULL,
    device_name VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    location_description TEXT,
    device_type VARCHAR(50) DEFAULT 'sensor',
    status VARCHAR(20) DEFAULT 'unknown',
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_devices_ip_address ON devices(ip_address);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_location ON devices(location_description);
```

#### 4.3.2 Device Repository
```python
class DeviceRepository:
    """Repository for device data operations."""
    
    async def save_discovered_devices(self, devices: List[DeviceInfo]) -> None:
        """Save or update discovered devices."""
        for device in devices:
            await self._upsert_device(device)
    
    async def get_devices_by_location(self, location: str) -> List[DeviceInfo]:
        """Get devices filtered by location."""
        query = """
        SELECT * FROM devices 
        WHERE location_description ILIKE %s 
        AND status = 'online'
        ORDER BY device_name
        """
        # Execute query and return results
        pass
    
    async def get_all_active_devices(self) -> List[DeviceInfo]:
        """Get all active devices."""
        # Implementation
        pass
```

---

## 5. Natural Language Processing Requirements

### 5.1 Language Specifications

#### 5.1.1 Colombian Spanish Characteristics
- **Dialect**: Neutral Colombian Spanish
- **Formality**: Informal but respectful (use "tú" not "usted")
- **Tone**: Friendly, helpful, patient
- **Vocabulary**: Agricultural and technical terms in Colombian context

#### 5.1.2 Domain-Specific Vocabulary
```python
AGRICULTURAL_TERMS = {
    "invernadero": ["greenhouse", "hothouse"],
    "cultivo": ["crop", "cultivation"],
    "riego": ["irrigation", "watering"],
    "sensor": ["sensor", "detector"],
    "humedad": ["humidity", "moisture"],
    "temperatura": ["temperature"],
    "zona": ["zone", "area", "sector"],
    "parcela": ["plot", "parcel"],
    "semillero": ["seedbed", "nursery"]
}

DEVICE_SYNONYMS = {
    "sensor": ["detector", "medidor", "dispositivo"],
    "zona": ["área", "sector", "región", "parte"],
    "temperatura": ["temp", "calor"],
    "humedad": ["humedad relativa", "hr"]
}
```

### 5.2 Intent Classification

#### 5.2.1 Primary Intents
```python
class Intent(Enum):
    """User intent classification."""
    DEVICE_STATUS = "device_status"
    SENSOR_DATA = "sensor_data"
    SYSTEM_HEALTH = "system_health"
    DEVICE_CONTROL = "device_control"
    TROUBLESHOOTING = "troubleshooting"
    GENERAL_INFO = "general_info"
    GREETING = "greeting"
    GOODBYE = "goodbye"

INTENT_PATTERNS = {
    Intent.DEVICE_STATUS: [
        "¿cómo están mis sensores?",
        "estado de los dispositivos",
        "¿funcionan bien los sensores?",
        "dispositivos conectados"
    ],
    Intent.SENSOR_DATA: [
        "¿cuál es la temperatura?",
        "datos de humedad",
        "¿qué temperatura hay en el invernadero?",
        "muéstrame los datos del sensor"
    ],
    Intent.SYSTEM_HEALTH: [
        "¿todo está funcionando?",
        "¿hay problemas en el sistema?",
        "estado general del sistema",
        "¿algún sensor desconectado?"
    ]
}
```

#### 5.2.2 Entity Extraction
```python
class EntityType(Enum):
    """Entity types for extraction."""
    LOCATION = "location"
    DEVICE_NAME = "device_name"
    SENSOR_TYPE = "sensor_type"
    TIME_RANGE = "time_range"
    DEVICE_ID = "device_id"

class Entity(BaseModel):
    """Extracted entity."""
    type: EntityType
    value: str
    confidence: float
    start_pos: int
    end_pos: int

# Location patterns
LOCATION_PATTERNS = [
    r"en el? (?P<location>invernadero|jardín|zona [a-zA-Z0-9]+|sector [a-zA-Z0-9]+)",
    r"del? (?P<location>cultivo|semillero|parcela [a-zA-Z0-9]+)",
    r"(?P<location>zona norte|zona sur|zona este|zona oeste)"
]
```

### 5.3 Context Management

#### 5.3.1 Conversation Context
```python
class ConversationContext(BaseModel):
    """Conversation context state."""
    session_id: str
    user_id: str
    current_topic: str
    last_mentioned_devices: List[str]
    last_mentioned_location: str
    conversation_history: List[Dict[str, Any]]
    preferences: Dict[str, Any]
    created_at: datetime
    updated_at: datetime

class ContextManager:
    """Manages conversation context."""
    
    async def update_context(
        self, 
        session_id: str, 
        message: str, 
        extracted_entities: List[Entity]
    ) -> ConversationContext:
        """Update conversation context with new information."""
        context = await self.get_context(session_id)
        
        # Update referenced devices and locations
        for entity in extracted_entities:
            if entity.type == EntityType.LOCATION:
                context.last_mentioned_location = entity.value
            elif entity.type == EntityType.DEVICE_NAME:
                if entity.value not in context.last_mentioned_devices:
                    context.last_mentioned_devices.append(entity.value)
        
        # Add to conversation history
        context.conversation_history.append({
            "timestamp": datetime.utcnow(),
            "message": message,
            "entities": [e.dict() for e in extracted_entities]
        })
        
        context.updated_at = datetime.utcnow()
        await self.save_context(context)
        return context
```

---

## 6. Conversation Flow Design

### 6.1 Conversation Patterns

#### 6.1.1 Device Discovery Flow
```
User: "¿Qué sensores tengo conectados?"
Assistant: [Executes MCP device discovery]
Assistant: "Tienes 3 sensores conectados:
           • Sensor Zona Norte (192.168.1.239) - Online
           • Sensor Invernadero (192.168.1.240) - Online  
           • Sensor Jardín Principal (192.168.1.241) - Offline
           
           ¿Te gustaría ver los datos de alguno en particular?"

User: "Muéstrame los datos del invernadero"
Assistant: [Executes MCP sensor data collection]
Assistant: "Datos actuales del Sensor Invernadero:
           🌡️ Temperatura: 24.5°C
           💧 Humedad: 68%
           📅 Última actualización: hace 2 minutos
           
           Los niveles están dentro del rango óptimo para tus cultivos."
```

#### 6.1.2 Troubleshooting Flow
```
User: "Mi sensor no responde"
Assistant: "Entiendo tu preocupación. Vamos a revisar paso a paso:
           
           1. ¿Podrías decirme qué sensor específicamente?
           2. ¿Desde cuándo notaste el problema?
           
           Mientras tanto, voy a verificar el estado de todos tus dispositivos."```

[Executes MCP device discovery and health check]
Assistant: "He revisado todos tus dispositivos. Encontré que el Sensor Jardín Principal 
           (192.168.1.241) no responde desde hace 15 minutos.
           
           Posibles causas:
           • Problema de conectividad WiFi
           • Dispositivo desconectado de la corriente
           • Reinicio necesario del dispositivo
           
           Te recomiendo:
           1. Verificar que el dispositivo esté conectado a la corriente
           2. Revisar la señal WiFi en esa zona
           3. Si persiste, puedo ayudarte a reiniciar el dispositivo remotamente"
```

### 6.2 Response Generation Strategy

#### 6.2.1 Response Templates
```python
RESPONSE_TEMPLATES = {
    "device_status": {
        "all_online": "Todos tus {device_count} dispositivos están funcionando correctamente:\n{device_list}",
        "some_offline": "De tus {device_count} dispositivos, {online_count} están online y {offline_count} presentan problemas:\n{status_summary}",
        "no_devices": "No he podido encontrar dispositivos conectados. ¿Te gustaría que busque nuevamente?"
    },
    "sensor_data": {
        "current_data": "Datos actuales de {device_name}:\n🌡️ Temperatura: {temperature}°C\n💧 Humedad: {humidity}%\n📅 Última actualización: {last_update}",
        "no_data": "No puedo obtener datos del {device_name} en este momento. El dispositivo podría estar desconectado.",
        "historical_data": "Resumen de las últimas 24 horas en {device_name}:\n📊 Temperatura promedio: {avg_temp}°C\n📊 Humedad promedio: {avg_humidity}%"
    },
    "error": {
        "device_offline": "El dispositivo {device_name} no está respondiendo. Te ayudo a diagnosticar el problema.",
        "network_error": "Hay un problema de conectividad. Estoy intentando reconectar...",
        "unknown_device": "No conozco ese dispositivo. ¿Podrías verificar el nombre? Estos son los dispositivos disponibles: {available_devices}"
    }
}
```

#### 6.2.2 Context-Aware Responses
```python
class ResponseGenerator:
    """Generates contextual responses based on user intent and conversation history."""
    
    def __init__(self, adk_service: ADKService, template_engine: TemplateEngine):
        self.adk_service = adk_service
        self.template_engine = template_engine
    
    async def generate_response(
        self,
        intent: Intent,
        entities: List[Entity],
        context: ConversationContext,
        data: Dict[str, Any]
    ) -> str:
        """Generate contextual response."""
        
        # Get base template
        template = self._get_template(intent, data)
        
        # Add contextual information
        enhanced_data = await self._enhance_with_context(data, context)
        
        # Generate response using ADK
        response = await self.adk_service.generate_response(
            template=template,
            data=enhanced_data,
            conversation_history=context.conversation_history[-5:]  # Last 5 messages
        )
        
        return self._post_process_response(response)
    
    def _enhance_with_context(self, data: Dict[str, Any], context: ConversationContext) -> Dict[str, Any]:
        """Enhance response data with conversation context."""
        enhanced_data = data.copy()
        
        # Add previously mentioned devices if relevant
        if context.last_mentioned_devices and not enhanced_data.get('device_name'):
            enhanced_data['device_name'] = context.last_mentioned_devices[-1]
        
        # Add location context
        if context.last_mentioned_location:
            enhanced_data['location_context'] = context.last_mentioned_location
        
        return enhanced_data
```

---

## 7. Error Handling and User Experience

### 7.1 Error Categories and Handling

#### 7.1.1 System Errors
```python
class SystemError(Exception):
    """Base class for system-level errors."""
    pass

class DeviceConnectionError(SystemError):
    """Device is unreachable or offline."""
    
    def __init__(self, device_ip: str, device_name: str = None):
        self.device_ip = device_ip
        self.device_name = device_name
        super().__init__(f"Cannot connect to device {device_name or device_ip}")

class ADKServiceError(SystemError):
    """Google ADK service error."""
    pass

class DatabaseError(SystemError):
    """Database operation error."""
    pass
```

#### 7.1.2 User Experience Errors
```python
class UserExperienceError(Exception):
    """Base class for UX-related errors."""
    pass

class AmbiguousRequestError(UserExperienceError):
    """User request is ambiguous and needs clarification."""
    
    def __init__(self, ambiguous_entities: List[str]):
        self.ambiguous_entities = ambiguous_entities
        super().__init__(f"Ambiguous request involving: {', '.join(ambiguous_entities)}")

class NoDevicesFoundError(UserExperienceError):
    """No devices found matching user criteria."""
    pass
```

#### 7.1.3 Error Recovery Strategies
```python
class ErrorRecoveryService:
    """Service for handling errors and guiding user recovery."""
    
    async def handle_device_connection_error(
        self, 
        error: DeviceConnectionError, 
        context: ConversationContext
    ) -> str:
        """Handle device connection errors with helpful suggestions."""
        
        suggestions = [
            "Verificar que el dispositivo esté conectado a la corriente",
            "Revisar la conexión WiFi en esa zona",
            "Intentar reiniciar el dispositivo"
        ]
        
        response = f"""
        No puedo conectar con el {error.device_name or 'dispositivo'} en este momento.
        
        Esto puede deberse a varios factores. Te sugiero:
        {chr(10).join(f'{i+1}. {suggestion}' for i, suggestion in enumerate(suggestions))}
        
        ¿Te gustaría que intente conectar nuevamente en unos minutos?
        """
        
        return response.strip()
    
    async def handle_ambiguous_request(
        self, 
        error: AmbiguousRequestError, 
        context: ConversationContext
    ) -> str:
        """Handle ambiguous user requests."""
        
        if "sensor" in error.ambiguous_entities:
            available_devices = await self._get_available_devices()
            device_list = '\n'.join(f'• {device.device_name} ({device.location_description})' 
                                   for device in available_devices)
            
            return f"""
            Tienes varios sensores disponibles. ¿Te refieres a alguno de estos?
            
            {device_list}
            
            O puedes ser más específico mencionando la ubicación, como "el sensor del invernadero".
            """
        
        return "No estoy seguro a qué te refieres. ¿Podrías ser más específico?"
```

### 7.2 Progressive Disclosure

#### 7.2.1 Information Layering
```python
class ProgressiveDisclosureService:
    """Service for managing information complexity."""
    
    def format_device_summary(self, devices: List[DeviceInfo], detail_level: str = "basic") -> str:
        """Format device information with appropriate detail level."""
        
        if detail_level == "basic":
            return self._format_basic_summary(devices)
        elif detail_level == "detailed":
            return self._format_detailed_summary(devices)
        elif detail_level == "technical":
            return self._format_technical_summary(devices)
    
    def _format_basic_summary(self, devices: List[DeviceInfo]) -> str:
        """Basic summary for non-technical users."""
        online_count = sum(1 for d in devices if d.status == "online")
        total_count = len(devices)
        
        summary = f"Tienes {total_count} dispositivos, {online_count} funcionando correctamente"
        
        if online_count < total_count:
            offline_devices = [d.device_name for d in devices if d.status != "online"]
            summary += f" y {len(offline_devices)} con problemas: {', '.join(offline_devices)}"
        
        return summary + "."
    
    def _format_detailed_summary(self, devices: List[DeviceInfo]) -> str:
        """Detailed summary with location and status."""
        lines = ["Estado detallado de tus dispositivos:\n"]
        
        for device in devices:
            status_emoji = "✅" if device.status == "online" else "❌"
            lines.append(f"{status_emoji} {device.device_name}")
            lines.append(f"   📍 {device.location_description}")
            lines.append(f"   🔌 {device.status.title()}")
            lines.append("")
        
        return "\n".join(lines)
```

### 7.3 Conversation Recovery

#### 7.3.1 Dead End Prevention
```python
class ConversationRecoveryService:
    """Service for preventing and recovering from conversation dead ends."""
    
    async def suggest_next_actions(self, context: ConversationContext) -> List[str]:
        """Suggest relevant next actions based on conversation context."""
        
        suggestions = []
        
        # Based on last mentioned devices
        if context.last_mentioned_devices:
            device_name = context.last_mentioned_devices[-1]
            suggestions.extend([
                f"Ver datos actuales del {device_name}",
                f"Verificar el historial del {device_name}",
                "Revisar el estado de todos los dispositivos"
            ])
        
        # Based on conversation topic
        if "problema" in context.current_topic.lower():
            suggestions.extend([
                "Ejecutar diagnóstico completo del sistema",
                "Ver dispositivos que necesitan atención",
                "Programar mantenimiento preventivo"
            ])
        
        # General suggestions
        suggestions.extend([
            "Mostrar resumen del día",
            "Configurar alertas",
            "Ver ayuda disponible"
        ])
        
        return suggestions[:3]  # Return top 3 suggestions
    
    async def handle_confusion(self, user_message: str, context: ConversationContext) -> str:
        """Handle user confusion with helpful guidance."""
        
        # Detect confusion patterns
        confusion_indicators = ["no entiendo", "no sé", "cómo", "ayuda", "qué puedo"]
        
        if any(indicator in user_message.lower() for indicator in confusion_indicators):
            suggestions = await self.suggest_next_actions(context)
            
            return f"""
            Te entiendo, déjame ayudarte. Aquí tienes algunas cosas que puedes hacer:
            
            {chr(10).join(f'• {suggestion}' for suggestion in suggestions)}
            
            También puedes preguntarme cosas como:
            • "¿Cómo están mis sensores?"
            • "¿Cuál es la temperatura del invernadero?"
            • "¿Hay algún problema en el sistema?"
            
            ¿Qué te gustaría hacer?
            """
        
        return "No estoy seguro de cómo ayudarte con eso. ¿Podrías reformular tu pregunta?"
```

---

## 8. Implementation Phases

### 8.1 Phase 1: Foundation (Weeks 1-2)

#### 8.1.1 Core Infrastructure Setup
- **Hexagonal Architecture Implementation**
  - Domain layer: Entities, value objects, ports
  - Use cases layer: Application services
  - Infrastructure layer: External integrations
  - Presentation layer: FastAPI endpoints

- **Basic ADK Integration**
  - ADK client setup and authentication
  - Basic conversation processing pipeline
  - Simple intent classification (greeting, goodbye, help)

- **MCP Tool Framework**
  - Tool registration and execution framework
  - Basic device discovery tool
  - Simple sensor data collection tool

#### 8.1.2 Deliverables
```python
# Domain entities
class Conversation(Entity):
    """Conversation domain entity."""
    pass

class Device(Entity):
    """Device domain entity."""
    pass

# Use cases
class ProcessConversationUseCase:
    """Process user conversation use case."""
    pass

class DiscoverDevicesUseCase:
    """Discover IoT devices use case."""
    pass

# Infrastructure
class ADKClient:
    """Google ADK client implementation."""
    pass

class MCPToolRegistry:
    """MCP tool registry and executor."""
    pass
```

### 8.2 Phase 2: Core Functionality (Weeks 3-4)

#### 8.2.1 Advanced NLP Processing
- **Intent Classification Enhancement**
  - Colombian Spanish specific training data
  - Agricultural domain vocabulary
  - Context-aware intent recognition

- **Entity Extraction**
  - Location entity extraction
  - Device name normalization
  - Time range parsing

- **Context Management**
  - Conversation state persistence
  - Session management
  - User preference learning

#### 8.2.2 Device Integration
- **Enhanced MCP Tools**
  - Real-time sensor data collection
  - Device health monitoring
  - Historical data aggregation

- **Error Handling**
  - Device connectivity issues
  - Data validation and correction
  - Graceful degradation

### 8.3 Phase 3: User Experience (Weeks 5-6)

#### 8.3.1 Response Generation
- **Natural Response Generation**
  - Colombian Spanish response templates
  - Context-aware responses
  - Personalized communication style

- **Progressive Disclosure**
  - Information layering based on user expertise
  - Adaptive detail levels
  - Smart summarization

#### 8.3.2 Conversation Flow
- **Multi-turn Conversations**
  - Context preservation across turns
  - Reference resolution
  - Conversation thread management

- **Error Recovery**
  - Confusion detection and handling
  - Dead end prevention
  - Helpful suggestions

### 8.4 Phase 4: Advanced Features (Weeks 7-8)

#### 8.4.1 Intelligent Features
- **Proactive Monitoring**
  - Anomaly detection in sensor data
  - Predictive maintenance alerts
  - Environmental condition warnings

- **Learning and Adaptation**
  - User preference learning
  - Conversation pattern optimization
  - Performance improvement

#### 8.4.2 Integration and Testing
- **System Integration**
  - Go backend integration
  - Database synchronization
  - Event-driven communication

- **Testing and Optimization**
  - Performance testing
  - User acceptance testing
  - Load testing and scaling

---

## 9. Technology Stack and Dependencies

### 9.1 Core Technologies

#### 9.1.1 Python Ecosystem
```toml
[project]
name = "python-ai-sis-assistant"
version = "0.1.0"
requires-python = ">=3.12"

dependencies = [
    # Web Framework
    "fastapi>=0.116.1",
    "uvicorn[standard]>=0.24.0",
    "pydantic>=2.5.0",
    
    # Google ADK Integration
    "google-cloud-aiplatform>=1.38.0",
    "google-auth>=2.23.0",
    "google-api-core>=2.11.0",
    
    # HTTP Client
    "httpx>=0.25.0",
    "aiohttp>=3.9.0",
    
    # Database
    "asyncpg>=0.29.0",
    "databases[postgresql]>=0.8.0",
    "sqlalchemy[asyncio]>=2.0.0",
    "alembic>=1.12.0",
    
    # Caching
    "redis[hiredis]>=5.0.0",
    "aioredis>=2.0.0",
    
    # NLP and Text Processing
    "spacy>=3.7.0",
    "transformers>=4.35.0",
    "torch>=2.1.0",
    
    # Data Validation and Serialization
    "pydantic>=2.5.0",
    "pydantic-settings>=2.1.0",
    
    # Logging and Monitoring
    "structlog>=23.2.0",
    "prometheus-client>=0.19.0",
    
    # Testing
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "pytest-cov>=4.1.0",
    "httpx>=0.25.0",  # For testing
    "factory-boy>=3.3.0",
    
    # Development
    "black>=23.0.0",
    "isort>=5.12.0",
    "mypy>=1.7.0",
    "ruff>=0.1.6",
]

[project.optional-dependencies]
dev = [
    "pre-commit>=3.5.0",
    "bandit>=1.7.5",
    "safety>=2.3.0",
]
```

#### 9.1.2 External Services
- **Google ADK**: Advanced NLP and conversation processing
- **PostgreSQL**: Primary database for conversation history and device registry
- **Redis**: Session management and conversation context caching
- **NATS**: Event-driven communication with Go backend
- **Prometheus + Grafana**: Monitoring and observability

### 9.2 Development Environment

#### 9.2.1 UV Package Manager Setup
```bash
# Initialize project with uv
uv init python-ai-sis-assistant
cd python-ai-sis-assistant

# Create virtual environment
uv venv --python 3.12

# Activate virtual environment
source .venv/bin/activate

# Install dependencies
uv add fastapi uvicorn[standard] pydantic
uv add google-cloud-aiplatform google-auth
uv add httpx aiohttp
uv add asyncpg databases[postgresql] sqlalchemy[asyncio] alembic
uv add redis[hiredis] aioredis
uv add spacy transformers torch
uv add structlog prometheus-client

# Install development dependencies
uv add --dev pytest pytest-asyncio pytest-cov httpx factory-boy
uv add --dev black isort mypy ruff pre-commit bandit safety

# Generate lock file
uv lock
```

#### 9.2.2 Environment Configuration
```python
# src/config/settings.py
from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import List, Optional

class Settings(BaseSettings):
    """Application settings."""
    
    # Server Configuration
    host: str = "0.0.0.0"
    port: int = 8081
    debug: bool = False
    environment: str = "development"
    
    # Google ADK Configuration
    google_adk_project_id: str
    google_adk_location: str = "us-central1"
    google_adk_api_key: Optional[str] = None
    google_adk_model_name: str = "gemini-1.5-pro"
    
    # Database Configuration
    database_url: str = "postgresql+asyncpg://user:password@localhost:5432/sis_assistant"
    database_pool_size: int = 10
    database_max_overflow: int = 20
    
    # Redis Configuration
    redis_url: str = "redis://localhost:6379/0"
    redis_pool_size: int = 10
    
    # Session Configuration
    session_timeout_minutes: int = 30
    max_conversation_history: int = 100
    
    # MCP Configuration
    device_discovery_timeout: float = 5.0
    max_concurrent_device_queries: int = 10
    device_registry_cache_ttl: int = 300  # 5 minutes
    
    # Go Backend Integration
    go_backend_url: str = "http://localhost:8080"
    go_backend_timeout: float = 10.0
    
    # Logging Configuration
    log_level: str = "INFO"
    log_format: str = "json"
    
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )

settings = Settings()
```

### 9.3 Project Structure

```
python_ai_sis_assistant/
├── README.md
├── TECHNICAL_DOCUMENTATION.md
├── pyproject.toml
├── uv.lock
├── .env.example
├── .gitignore
├── main.py
├── src/
│   ├── __init__.py
│   ├── config/
│   │   ├── __init__.py
│   │   └── settings.py
│   ├── domain/
│   │   ├── __init__.py
│   │   ├── entities/
│   │   │   ├── __init__.py
│   │   │   ├── conversation.py
│   │   │   ├── device.py
│   │   │   └── user.py
│   │   ├── errors/
│   │   │   ├── __init__.py
│   │   │   ├── base.py
│   │   │   ├── system_errors.py
│   │   │   └── user_errors.py
│   │   ├── events/
│   │   │   ├── __init__.py
│   │   │   ├── conversation_events.py
│   │   │   └── device_events.py
│   │   └── ports/
│   │       ├── __init__.py
│   │       ├── repositories/
│   │       │   ├── __init__.py
│   │       │   ├── conversation_repository.py
│   │       │   └── device_repository.py
│   │       └── services/
│   │           ├── __init__.py
│   │           ├── adk_service.py
│   │           ├── nlp_service.py
│   │           └── mcp_service.py
│   ├── usecases/
│   │   ├── __init__.py
│   │   ├── conversation/
│   │   │   ├── __init__.py
│   │   │   ├── process_message.py
│   │   │   ├── manage_context.py
│   │   │   └── generate_response.py
│   │   ├── device/
│   │   │   ├── __init__.py
│   │   │   ├── discover_devices.py
│   │   │   ├── collect_sensor_data.py
│   │   │   └── monitor_health.py
│   │   └── nlp/
│   │       ├── __init__.py
│   │       ├── extract_intent.py
│   │       ├── extract_entities.py
│   │       └── understand_context.py
│   ├── infrastructure/
│   │   ├── __init__.py
│   │   ├── adk/
│   │   │   ├── __init__.py
│   │   │   ├── client.py
│   │   │   ├── models.py
│   │   │   └── prompts.py
│   │   ├── mcp/
│   │   │   ├── __init__.py
│   │   │   ├── registry.py
│   │   │   ├── tools/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── device_discovery.py
│   │   │   │   └── sensor_data.py
│   │   │   └── executor.py
│   │   ├── database/
│   │   │   ├── __init__.py
│   │   │   ├── connection.py
│   │   │   ├── models/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── conversation.py
│   │   │   │   └── device.py
│   │   │   └── repositories/
│   │   │       ├── __init__.py
│   │   │       ├── conversation_repository.py
│   │   │       └── device_repository.py
│   │   ├── cache/
│   │   │   ├── __init__.py
│   │   │   ├── redis_client.py
│   │   │   └── session_store.py
│   │   ├── http/
│   │   │   ├── __init__.py
│   │   │   ├── client.py
│   │   │   └── device_client.py
│   │   └── logging/
│   │       ├── __init__.py
│   │       ├── config.py
│   │       └── logger.py
│   ├── presentation/
│   │   ├── __init__.py
│   │   ├── http/
│   │   │   ├── __init__.py
│   │   │   ├── handlers/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── conversation_handler.py
│   │   │   │   ├── device_handler.py
│   │   │   │   └── health_handler.py
│   │   │   ├── middleware/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── cors.py
│   │   │   │   ├── auth.py
│   │   │   │   └── logging.py
│   │   │   └── models/
│   │   │       ├── __init__.py
│   │   │       ├── requests.py
│   │   │       └── responses.py
│   │   └── websocket/
│   │       ├── __init__.py
│   │       ├── handler.py
│   │       └── connection_manager.py
│   └── app/
│       ├── __init__.py
│       ├── container.py
│       ├── factory.py
│       └── lifespan.py
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   ├── unit/
│   │   ├── __init__.py
│   │   ├── domain/
│   │   ├── usecases/
│   │   └── infrastructure/
│   ├── integration/
│   │   ├── __init__.py
│   │   ├── database/
│   │   ├── http/
│   │   └── external/
│   └── e2e/
│       ├── __init__.py
│       └── test_conversation_flow.py
├── scripts/
│   ├── setup_dev.sh
│   ├── run_tests.sh
│   └── deploy.sh
├── docs/
│   ├── api/
│   ├── architecture/
│   └── user_guide/
└── migrations/
    └── versions/
```

---

## 10. API Design and Interfaces

### 10.1 REST API Endpoints

#### 10.1.1 Conversation Endpoints
```python
# src/presentation/http/models/requests.py
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any

class ConversationRequest(BaseModel):
    """Request model for conversation processing."""
    
    message: str = Field(..., min_length=1, max_length=1000)
    session_id: Optional[str] = Field(None, description="Session identifier for context")
    user_id: Optional[str] = Field(None, description="User identifier")
    metadata: Optional[Dict[str, Any]] = Field(default_factory=dict)

class ConversationResponse(BaseModel):
    """Response model for conversation processing."""
    
    response: str = Field(..., description="Assistant response message")
    session_id: str = Field(..., description="Session identifier")
    intent: Optional[str] = Field(None, description="Detected user intent")
    entities: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    suggestions: Optional[List[str]] = Field(default_factory=list)
    metadata: Optional[Dict[str, Any]] = Field(default_factory=dict)

# src/presentation/http/handlers/conversation_handler.py
from fastapi import APIRouter, HTTPException, Depends
from src.usecases.conversation.process_message import ProcessMessageUseCase

class ConversationHandler:
    """HTTP handler for conversation endpoints."""
    
    def __init__(self, process_message_use_case: ProcessMessageUseCase):
        self.process_message_use_case = process_message_use_case
        self.router = APIRouter(prefix="/api/v1/conversation", tags=["conversation"])
        self._setup_routes()
    
    def _setup_routes(self):
        """Setup API routes."""
        
        @self.router.post("/message", response_model=ConversationResponse)
        async def process_message(request: ConversationRequest):
            """Process user message and return assistant response."""
            try:
                result = await self.process_message_use_case.execute(
                    message=request.message,
                    session_id=request.session_id,
                    user_id=request.user_id,
                    metadata=request.metadata
                )
                return ConversationResponse(**result)
            except Exception as e:
                raise HTTPException(status_code=500, detail=str(e))
        
        @self.router.get("/session/{session_id}/history")
        async def get_conversation_history(session_id: str):
            """Get conversation history for a session."""
            # Implementation
            pass
        
        @self.router.delete("/session/{session_id}")
        async def clear_session(session_id: str):
            """Clear conversation session."""
            # Implementation
            pass
```

#### 10.1.2 Device Management Endpoints
```python
# src/presentation/http/models/device_models.py
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime

class DeviceResponse(BaseModel):
    """Device information response model."""
    
    mac_address: str
    device_name: str
    ip_address: str
    location_description: str
    device_type: str
    status: str
    last_seen: datetime
    
class SensorDataResponse(BaseModel):
    """Sensor data response model."""
    
    device_name: str
    device_ip: str
    timestamp: datetime
    data: Dict[str, Any]

# src/presentation/http/handlers/device_handler.py
class DeviceHandler:
    """HTTP handler for device-related endpoints."""
    
    def __init__(self, discover_devices_use_case, collect_sensor_data_use_case):
        self.discover_devices_use_case = discover_devices_use_case
        self.collect_sensor_data_use_case = collect_sensor_data_use_case
        self.router = APIRouter(prefix="/api/v1/devices", tags=["devices"])
        self._setup_routes()
    
    def _setup_routes(self):
        
        @self.router.get("/", response_model=List[DeviceResponse])
        async def list_devices():
            """List all discovered devices."""
            devices = await self.discover_devices_use_case.execute()
            return [DeviceResponse(**device.dict()) for device in devices]
        
        @self.router.post("/discover")
        async def trigger_discovery():
            """Trigger device discovery process."""
            result = await self.discover_devices_use_case.execute(force_refresh=True)
            return {"discovered_count": len(result), "devices": result}
        
        @self.router.get("/{device_ip}/data", response_model=SensorDataResponse)
        async def get_sensor_data(device_ip: str, sensor_type: str = "temperature-and-humidity"):
            """Get sensor data from specific device."""
            data = await self.collect_sensor_data_use_case.execute(device_ip, sensor_type)
            return SensorDataResponse(**data)
```

### 10.2 WebSocket Interface

#### 10.2.1 Real-time Conversation
```python
# src/presentation/websocket/handler.py
from fastapi import WebSocket, WebSocketDisconnect
from typing import Dict
import json

class WebSocketHandler:
    """WebSocket handler for real-time conversation."""
    
    def __init__(self, process_message_use_case):
        self.process_message_use_case = process_message_use_case
        self.active_connections: Dict[str, WebSocket] = {}
    
    async def connect(self, websocket: WebSocket, session_id: str):
        """Accept WebSocket connection."""
        await websocket.accept()
        self.active_connections[session_id] = websocket
    
    def disconnect(self, session_id: str):
        """Remove connection."""
        if session_id in self.active_connections:
            del self.active_connections[session_id]
    
    async def handle_message(self, websocket: WebSocket, session_id: str):
        """Handle incoming WebSocket messages."""
        try:
            while True:
                # Receive message
                data = await websocket.receive_text()
                message_data = json.loads(data)
                
                # Process message
                result = await self.process_message_use_case.execute(
                    message=message_data["message"],
                    session_id=session_id,
                    user_id=message_data.get("user_id"),
                    metadata=message_data.get("metadata", {})
                )
                
                # Send response
                await websocket.send_text(json.dumps(result))
                
        except WebSocketDisconnect:
            self.disconnect(session_id)
        except Exception as e:
            await websocket.send_text(json.dumps({
                "error": str(e),
                "type": "processing_error"
            }))
```

### 10.3 Integration APIs

#### 10.3.1 Go Backend Integration
```python
# src/infrastructure/http/go_backend_client.py
import httpx
from typing import List, Dict, Any
from src.config.settings import settings

class GoBackendClient:
    """Client for Go backend integration."""
    
    def __init__(self):
        self.base_url = settings.go_backend_url
        self.timeout = settings.go_backend_timeout
    
    async def get_registered_devices(self) -> List[Dict[str, Any]]:
        """Get registered devices from Go backend."""
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{self.base_url}/api/devices",
                timeout=self.timeout
            )
            response.raise_for_status()
            return response.json()
    
    async def get_device_health_status(self) -> Dict[str, Any]:
        """Get device health status from Go backend."""
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{self.base_url}/api/health/devices",
                timeout=self.timeout
            )
            response.raise_for_status()
            return response.json()
    
    async def get_sensor_history(
        self, 
        device_id: str, 
        hours: int = 24
    ) -> List[Dict[str, Any]]:
        """Get sensor data history from Go backend."""
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{self.base_url}/api/devices/{device_id}/history",
                params={"hours": hours},
                timeout=self.timeout
            )
            response.raise_for_status()
            return response.json()
```

### 10.4 Authentication and Security

#### 10.4.1 API Key Authentication
```python
# src/presentation/http/middleware/auth.py
from fastapi import HTTPException, Security, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from src.config.settings import settings

security = HTTPBearer()

async def verify_api_key(credentials: HTTPAuthorizationCredentials = Security(security)):
    """Verify API key for authentication."""
    
    if credentials.credentials != settings.api_key:
        raise HTTPException(
            status_code=401,
            detail="Invalid API key",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    return credentials.credentials

# Usage in handlers
@router.post("/message", dependencies=[Depends(verify_api_key)])
async def process_message(request: ConversationRequest):
    # Handler implementation
    pass
```

#### 10.4.2 Rate Limiting
```python
# src/presentation/http/middleware/rate_limit.py
from fastapi import HTTPException, Request
from typing import Dict
import time

class RateLimiter:
    """Simple in-memory rate limiter."""
    
    def __init__(self, max_requests: int = 60, window_seconds: int = 60):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self.requests: Dict[str, List[float]] = {}
    
    async def check_rate_limit(self, request: Request):
        """Check if request is within rate limit."""
        client_ip = request.client.host
        current_time = time.time()
        
        # Clean old requests
        if client_ip in self.requests:
            self.requests[client_ip] = [
                req_time for req_time in self.requests[client_ip]
                if current_time - req_time < self.window_seconds
            ]
        else:
            self.requests[client_ip] = []
        
        # Check rate limit
        if len(self.requests[client_ip]) >= self.max_requests:
            raise HTTPException(
                status_code=429,
                detail="Rate limit exceeded"
            )
        
        # Add current request
        self.requests[client_ip].append(current_time)
```

---

## Conclusion

This Technical Documentation provides a comprehensive guide for implementing the Python AI SIS Assistant, a conversational AI agent for the IoT Smart Irrigation System. The document covers all aspects from architecture design to implementation details, ensuring that software engineers have the necessary information to build a robust, scalable, and user-friendly solution.

### Key Success Factors

1. **Clean Architecture**: Following hexagonal architecture principles ensures maintainability and testability
2. **Google ADK Integration**: Leveraging advanced NLP capabilities for natural conversation flow
3. **MCP Tools**: Enabling real-time device discovery and data collection
4. **Colombian Spanish Focus**: Providing culturally appropriate and natural language interactions
5. **Progressive Implementation**: Phased approach allows for iterative development and testing

### Next Steps

1. Review and approve this technical documentation
2. Set up development environment with UV package manager
3. Begin Phase 1 implementation: Foundation setup
4. Establish CI/CD pipeline and testing framework
5. Coordinate with DevOps team for deployment strategy

This documentation should serve as the definitive guide for the Python AI SIS Assistant implementation, providing both high-level architectural guidance and detailed implementation specifications.