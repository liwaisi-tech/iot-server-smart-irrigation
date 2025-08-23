#!/usr/bin/env python3
"""
ADK CLI tool for testing agents directly
"""
import os
import sys
from google import adk

# Load environment variables
from dotenv import load_dotenv
load_dotenv()

def get_weather(city: str) -> str:
    """Get weather information for a city."""
    if city.lower() in ["bogotá", "bogota", "medellín", "medellin", "cali"]:
        return f"El clima en {city}: 24°C, parcialmente nublado con probabilidad de lluvia del 30%"
    return f"No tengo información del clima para {city} en este momento."

def get_current_time(city: str = "bogotá") -> str:
    """Get current time for a city."""
    from datetime import datetime
    current_time = datetime.now().strftime("%H:%M")
    return f"La hora actual en {city} es {current_time}"

def create_irrigation_agent():
    """Create the main irrigation agent with Spanish tools."""
    
    # Configure ADK with API key
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        print("❌ Error: GEMINI_API_KEY no encontrada en .env")
        sys.exit(1)
    
    try:
        # Create agent with Spanish system instruction
        agent = adk.Agent(
            model_id="gemini-1.5-flash",
            system_instruction="""
            Eres Peja, un asistente inteligente para sistemas de riego agrícola en Colombia.
            
            Puedes ayudar con:
            - Consultas sobre el clima local
            - Información de tiempo y horarios
            - Preguntas sobre riego y agricultura
            
            Responde siempre en español colombiano, siendo amigable y útil.
            Adapta tu nivel de formalidad al del usuario.
            """,
            tools=[get_weather, get_current_time]
        )
        
        return agent
        
    except Exception as e:
        print(f"❌ Error creando el agente: {e}")
        sys.exit(1)

def main():
    """Run the interactive ADK CLI."""
    print("🌱 ADK CLI - Asistente de Riego Inteligente")
    print("=" * 50)
    print("Escribe 'salir' para terminar")
    print()
    
    # Create the agent
    agent = create_irrigation_agent()
    
    # Interactive loop
    while True:
        try:
            user_input = input("👤 Usuario: ").strip()
            
            if user_input.lower() in ['salir', 'exit', 'quit']:
                print("👋 ¡Hasta luego!")
                break
            
            if not user_input:
                continue
            
            # Send message to agent
            print("🤖 Peja: ", end="")
            response = agent.send_message(user_input)
            
            # Print response
            for chunk in response:
                print(chunk, end="", flush=True)
            print()  # New line after response
            print()
            
        except KeyboardInterrupt:
            print("\n👋 ¡Hasta luego!")
            break
        except Exception as e:
            print(f"❌ Error: {e}")

if __name__ == "__main__":
    main()