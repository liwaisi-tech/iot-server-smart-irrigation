import uvicorn

from src.app.factory import get_app
from src.config.settings import Settings

if __name__ == "__main__":
    settings = Settings()
    app = get_app()
    uvicorn.run(
        "src.app.factory:get_app", 
        factory=True,
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
