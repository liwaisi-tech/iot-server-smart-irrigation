import uvicorn

from src.app.factory import get_app
from src.config.settings import Settings

# Create app instance for uvicorn
app = get_app()

if __name__ == "__main__":
    settings = Settings()
    uvicorn.run(
        "main:app", 
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
