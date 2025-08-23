"""Web Search Client - Provides external search capabilities."""

import logging
from typing import Dict, Any, List
import httpx
from urllib.parse import quote_plus

from src.domain.ports.llm_client import WebSearchPort


logger = logging.getLogger(__name__)


class WebSearchClient(WebSearchPort):
    """HTTP client for web search capabilities using various search APIs."""
    
    def __init__(self, api_key: str = "", timeout: int = 30):
        self.api_key = api_key
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
    
    async def search(self, query: str, max_results: int = 5) -> List[Dict[str, Any]]:
        """Perform web search and return results."""
        
        # If no API key is provided, return mock results for development
        if not self.api_key or self.api_key == "your-web-search-api-key":
            logger.warning("No web search API key configured, returning mock results")
            return self._get_mock_search_results(query, max_results)
        
        try:
            # Using SerpAPI as example (you can replace with your preferred search API)
            search_url = "https://serpapi.com/search"
            params = {
                "q": query,
                "api_key": self.api_key,
                "num": min(max_results, 10),
                "hl": "es",  # Spanish language results
                "gl": "co",  # Colombia location
                "google_domain": "google.com.co"
            }
            
            response = await self.client.get(search_url, params=params)
            response.raise_for_status()
            
            data = response.json()
            results = []
            
            # Parse organic search results
            for result in data.get("organic_results", [])[:max_results]:
                results.append({
                    "title": result.get("title", ""),
                    "url": result.get("link", ""),
                    "snippet": result.get("snippet", ""),
                    "source": "web_search"
                })
            
            logger.info(f"Found {len(results)} search results for query: {query}")
            return results
            
        except httpx.HTTPStatusError as e:
            logger.error(f"HTTP error in web search: {e.response.status_code}")
            return self._get_fallback_results(query)
        except httpx.RequestError as e:
            logger.error(f"Network error in web search: {e}")
            return self._get_fallback_results(query)
        except Exception as e:
            logger.error(f"Unexpected error in web search: {e}")
            return self._get_fallback_results(query)
    
    async def fetch_url_content(self, url: str) -> str:
        """Fetch content from a specific URL."""
        try:
            response = await self.client.get(url, follow_redirects=True)
            response.raise_for_status()
            
            # Return first 2000 characters to avoid overwhelming the LLM
            content = response.text[:2000]
            logger.info(f"Successfully fetched content from: {url}")
            return content
            
        except httpx.HTTPStatusError as e:
            logger.error(f"HTTP error fetching URL {url}: {e.response.status_code}")
            return f"Error: Could not fetch content from {url} (HTTP {e.response.status_code})"
        except httpx.RequestError as e:
            logger.error(f"Network error fetching URL {url}: {e}")
            return f"Error: Could not connect to {url}"
        except Exception as e:
            logger.error(f"Unexpected error fetching URL {url}: {e}")
            return f"Error: Unexpected error fetching content from {url}"
    
    def _get_mock_search_results(self, query: str, max_results: int) -> List[Dict[str, Any]]:
        """Return mock search results for development/testing."""
        mock_results = [
            {
                "title": f"Información sobre {query} - Guía Completa",
                "url": f"https://example.com/guia-{quote_plus(query.lower())}",
                "snippet": f"Encuentra toda la información que necesitas sobre {query}. Guía completa y actualizada.",
                "source": "mock_search"
            },
            {
                "title": f"Tutorial: {query} paso a paso",
                "url": f"https://tutoriales.com/{quote_plus(query.lower())}",
                "snippet": f"Aprende todo sobre {query} con nuestro tutorial paso a paso y ejemplos prácticos.",
                "source": "mock_search"
            },
            {
                "title": f"FAQ: Preguntas frecuentes sobre {query}",
                "url": f"https://faq.com/{quote_plus(query.lower())}",
                "snippet": f"Las preguntas más frecuentes sobre {query} con respuestas detalladas de expertos.",
                "source": "mock_search"
            }
        ]
        
        return mock_results[:max_results]
    
    def _get_fallback_results(self, query: str) -> List[Dict[str, Any]]:
        """Return fallback results when search fails."""
        return [
            {
                "title": f"Búsqueda no disponible: {query}",
                "url": "",
                "snippet": "Lo siento, el servicio de búsqueda web no está disponible en este momento. Por favor intenta más tarde o contacta al administrador del sistema.",
                "source": "fallback"
            }
        ]
    
    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()