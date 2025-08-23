from fastapi import APIRouter, Response

router = APIRouter()


@router.get("/ping", response_class=Response)
async def ping():
    """Ping endpoint that returns 'pong' as plain text."""
    return Response(content="pong", media_type="application/text")