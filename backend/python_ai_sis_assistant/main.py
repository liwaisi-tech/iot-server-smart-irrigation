"""Main entry point for the Python AI SIS Assistant."""

import uvicorn

from src.app.factory import create_app


# Create the app instance for uvicorn
app = create_app()


def main() -> None:
    """Start the FastAPI server."""
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8081,  # Using port 8081 to avoid conflict with Go service on 8080
        log_level="info",
        reload=True,
    )


if __name__ == "__main__":
    main()
