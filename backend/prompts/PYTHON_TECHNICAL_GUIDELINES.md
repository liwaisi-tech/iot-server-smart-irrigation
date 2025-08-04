# Python Technical Guidelines - Hexagonal Architecture

## Overview

This document establishes technical guidelines for Python projects following hexagonal architecture (ports and adapters pattern) with clean architecture principles. Based on analysis of the reference implementation and expert recommendations, these guidelines ensure maintainable, testable, and production-ready Python microservices.

## Architecture Philosophy

### Core Principles
- **Dependency Inversion**: Domain layer never imports from outer layers
- **Single Responsibility**: Each class/module has one reason to change  
- **Open/Closed**: Open for extension, closed for modification
- **Interface Segregation**: Small, focused interfaces over large ones
- **YAGNI**: You Aren't Gonna Need It - avoid over-engineering
- **KISS**: Keep It Simple, Stupid

### Layer Responsibilities

```
src/
├── domain/                    # Business core (no external dependencies)
│   ├── entities/              # Business entities with pure logic
│   ├── events/                # Domain events for event-driven architecture
│   ├── ports/                 # Interfaces/contracts (Repository, Services)
│   ├── services/              # Domain services for complex business rules
│   └── errors/                # Domain-specific errors and exceptions
├── usecases/                  # Application logic (one use case = one folder)
│   └── {use_case_name}/       # {use_case}.py, dtos.py, exceptions.py
├── infra/                     # External implementations
│   ├── persistence/           # Repository implementations
│   ├── messaging/             # Event publishers, message brokers  
│   ├── cache/                 # Cache implementations
│   └── externalapis/          # Third-party API clients
└── presentation/              # Delivery layer
    ├── api/                   # HTTP handlers and routers
    ├── middleware/            # Cross-cutting concerns
    └── app.py                 # Application factory
```

## Domain Layer Guidelines

### 1. Domain Entities

**❌ Avoid**: Using Pydantic BaseModel for domain entities
```python
# WRONG - Violates domain purity
from pydantic import BaseModel

class User(BaseModel):
    id: str
    email: str
    name: str
```

**✅ Recommended**: Pure Python dataclasses with validation
```python
from dataclasses import dataclass
from datetime import datetime
from typing import Protocol


class IdGenerator(Protocol):
    """Interface for ID generation"""
    def generate_user_id(self) -> str: ...


@dataclass(frozen=True)
class UserId:
    """Value object for user ID with validation"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.startswith("user_"):
            raise ValueError("Invalid user ID format")


@dataclass
class User:
    """Pure domain entity"""
    id: UserId
    email: str
    name: str
    created_at: datetime
    
    @classmethod
    def create(cls, email: str, name: str, id_generator: IdGenerator) -> "User":
        """Factory method for creating new users"""
        if not email or "@" not in email:
            raise ValueError("Invalid email format")
        if not name or len(name.strip()) == 0:
            raise ValueError("Name cannot be empty")
            
        return cls(
            id=UserId(id_generator.generate_user_id()),
            email=email,
            name=name,
            created_at=datetime.now()
        )
    
    def is_valid(self) -> bool:
        """Validate entity state"""
        return (
            self.id.value and 
            self.email and 
            "@" in self.email and 
            self.name and 
            len(self.name.strip()) > 0
        )
```

### 2. Domain Events

```python
from abc import ABC, abstractmethod
from datetime import datetime
from typing import Any, Dict
from dataclasses import dataclass
import uuid


class DomainEvent(ABC):
    """Base domain event interface"""
    
    @abstractmethod
    def event_type(self) -> str:
        """Return event type identifier"""
        pass
    
    @abstractmethod
    def aggregate_id(self) -> str:
        """Return ID of the aggregate that raised this event"""
        pass
    
    @abstractmethod
    def occurred_at(self) -> datetime:
        """Return when this event occurred"""
        pass
    
    @abstractmethod
    def to_dict(self) -> Dict[str, Any]:
        """Serialize event to dictionary"""
        pass


@dataclass
class UserCreated(DomainEvent):
    """User created domain event"""
    user_id: str
    email: str
    name: str
    created_at: datetime
    event_id: str = None
    occurred_at_time: datetime = None
    
    def __post_init__(self):
        if not self.event_id:
            self.event_id = f"evt_{uuid.uuid4().hex[:8]}"
        if not self.occurred_at_time:
            self.occurred_at_time = datetime.now()
    
    def event_type(self) -> str:
        return "user.created"
    
    def aggregate_id(self) -> str:
        return self.user_id
    
    def occurred_at(self) -> datetime:
        return self.occurred_at_time
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "event_id": self.event_id,
            "event_type": self.event_type(),
            "aggregate_id": self.aggregate_id(),
            "occurred_at": self.occurred_at().isoformat(),
            "payload": {
                "user_id": self.user_id,
                "email": self.email,
                "name": self.name,
                "created_at": self.created_at.isoformat()
            }
        }
    
    @classmethod
    def from_user(cls, user: User) -> "UserCreated":
        """Create event from User entity"""
        return cls(
            user_id=user.id.value,
            email=user.email,
            name=user.name,
            created_at=user.created_at
        )
```

### 3. Domain Ports (Interfaces)

```python
from abc import ABC, abstractmethod
from typing import Optional, List
from ..entities.user import User
from ..events import DomainEvent


class UserRepository(ABC):
    """Abstract repository for user persistence"""
    
    @abstractmethod
    async def save(self, user: User) -> None:
        """Save user to storage"""
        pass
    
    @abstractmethod
    async def find_by_id(self, user_id: str) -> Optional[User]:
        """Find user by ID"""
        pass
    
    @abstractmethod
    async def find_by_email(self, email: str) -> Optional[User]:
        """Find user by email"""
        pass
    
    @abstractmethod
    async def delete(self, user_id: str) -> None:
        """Delete user by ID"""
        pass


class EventPublisher(ABC):
    """Abstract event publisher interface"""
    
    @abstractmethod
    async def publish(self, event: DomainEvent) -> None:
        """Publish domain event"""
        pass
    
    @abstractmethod
    async def publish_batch(self, events: List[DomainEvent]) -> None:
        """Publish multiple events atomically"""
        pass
```

### 4. Domain Services

```python
from typing import List
from ..entities.user import User
from ..ports.user_repository import UserRepository
from ..errors.domain_errors import UserValidationError


class UserDomainService:
    """Domain service for complex user business rules"""
    
    def __init__(self, user_repository: UserRepository):
        self._user_repository = user_repository
    
    async def validate_unique_email(self, email: str, exclude_user_id: str = None) -> None:
        """Validate that email is unique across users"""
        existing_user = await self._user_repository.find_by_email(email)
        if existing_user and existing_user.id.value != exclude_user_id:
            raise UserValidationError(f"Email {email} is already in use")
    
    async def can_delete_user(self, user_id: str) -> bool:
        """Check if user can be deleted based on business rules"""
        user = await self._user_repository.find_by_id(user_id)
        if not user:
            return False
        
        # Add business logic here (e.g., check dependencies, permissions)
        return True
    
    def validate_user_data(self, user: User) -> List[str]:
        """Validate user data according to business rules"""
        errors = []
        
        if not user.is_valid():
            errors.append("User entity is invalid")
        
        if len(user.name) > 100:
            errors.append("Name cannot exceed 100 characters")
        
        # Add more business validation rules
        
        return errors
```

### 5. Domain Errors

```python
from typing import Optional, Dict, Any


class DomainError(Exception):
    """Base class for domain errors"""
    
    def __init__(
        self, 
        message: str, 
        error_code: str = None,
        details: Dict[str, Any] = None,
        cause: Exception = None
    ):
        super().__init__(message)
        self.message = message
        self.error_code = error_code or self.__class__.__name__
        self.details = details or {}
        self.cause = cause
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert error to dictionary representation"""
        return {
            "error_code": self.error_code,
            "message": self.message,
            "details": self.details,
            "type": self.__class__.__name__
        }


class UserNotFoundError(DomainError):
    """User not found error"""
    
    def __init__(self, user_id: str):
        super().__init__(
            message=f"User with ID {user_id} not found",
            error_code="USER_NOT_FOUND",
            details={"user_id": user_id}
        )


class UserAlreadyExistsError(DomainError):
    """User already exists error"""
    
    def __init__(self, email: str):
        super().__init__(
            message=f"User with email {email} already exists",
            error_code="USER_ALREADY_EXISTS",
            details={"email": email}
        )


class UserValidationError(DomainError):
    """User validation error"""
    
    def __init__(self, validation_errors: List[str]):
        super().__init__(
            message="User validation failed",
            error_code="USER_VALIDATION_ERROR",
            details={"validation_errors": validation_errors}
        )
```

## Use Cases Layer Guidelines

### 1. Use Case Implementation

```python
from typing import List, Optional
from dataclasses import dataclass
from ...domain.entities.user import User, UserId, IdGenerator
from ...domain.ports.user_repository import UserRepository
from ...domain.ports.event_publisher import EventPublisher
from ...domain.services.user_domain_service import UserDomainService
from ...domain.events.user_created import UserCreated
from ...domain.errors.domain_errors import UserAlreadyExistsError
from .dtos import CreateUserRequest, CreateUserResponse
import structlog


logger = structlog.get_logger()


class CreateUserUseCase:
    """Use case for creating a new user"""
    
    def __init__(
        self,
        user_repository: UserRepository,
        event_publisher: EventPublisher,
        user_domain_service: UserDomainService,
        id_generator: IdGenerator
    ):
        self._user_repository = user_repository
        self._event_publisher = event_publisher
        self._user_domain_service = user_domain_service
        self._id_generator = id_generator
    
    async def execute(self, request: CreateUserRequest) -> CreateUserResponse:
        """Execute the create user use case"""
        await logger.ainfo("Creating user", email=request.email, name=request.name)
        
        try:
            # Validate request
            request.validate()
            
            # Domain validation
            await self._user_domain_service.validate_unique_email(request.email)
            
            # Create user entity
            user = User.create(
                email=request.email,
                name=request.name,
                id_generator=self._id_generator
            )
            
            # Additional domain validation
            validation_errors = self._user_domain_service.validate_user_data(user)
            if validation_errors:
                raise UserValidationError(validation_errors)
            
            # Save user
            await self._user_repository.save(user)
            
            # Create and publish domain event
            user_created_event = UserCreated.from_user(user)
            await self._event_publisher.publish(user_created_event)
            
            await logger.ainfo("User created successfully", user_id=user.id.value)
            
            return CreateUserResponse(user_id=user.id.value)
            
        except Exception as e:
            await logger.aerror("Failed to create user", error=str(e), email=request.email)
            raise
```

### 2. DTOs with Validation

```python
from dataclasses import dataclass
from typing import List
import re


@dataclass
class CreateUserRequest:
    """Request DTO for creating a user"""
    email: str
    name: str
    
    def validate(self) -> None:
        """Validate request data"""
        errors = []
        
        if not self.email:
            errors.append("Email is required")
        elif not self._is_valid_email(self.email):
            errors.append("Invalid email format")
        
        if not self.name:
            errors.append("Name is required")
        elif not self.name.strip():
            errors.append("Name cannot be empty")
        elif len(self.name) > 100:
            errors.append("Name cannot exceed 100 characters")
        
        if errors:
            raise ValueError(f"Validation errors: {', '.join(errors)}")
    
    def _is_valid_email(self, email: str) -> bool:
        """Basic email validation"""
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        return re.match(pattern, email) is not None


@dataclass
class CreateUserResponse:
    """Response DTO for creating a user"""
    user_id: str
    
    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization"""
        return {"user_id": self.user_id}
```

## Infrastructure Layer Guidelines

### 1. Configuration Management

```python
from pydantic_settings import BaseSettings
from typing import Optional, List
from enum import Enum


class LogLevel(str, Enum):
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"


class Environment(str, Enum):
    DEVELOPMENT = "development"
    STAGING = "staging"
    PRODUCTION = "production"


class DatabaseConfig(BaseSettings):
    """Database configuration"""
    url: str = "sqlite:///./app.db"
    pool_size: int = 10
    max_overflow: int = 20
    pool_timeout: int = 30
    pool_recycle: int = 3600
    echo: bool = False
    
    class Config:
        env_prefix = "DB_"


class RedisConfig(BaseSettings):
    """Redis configuration"""
    url: Optional[str] = None
    host: str = "localhost"
    port: int = 6379
    db: int = 0
    password: Optional[str] = None
    
    class Config:
        env_prefix = "REDIS_"


class AppConfig(BaseSettings):
    """Main application configuration"""
    environment: Environment = Environment.DEVELOPMENT
    debug: bool = False
    log_level: LogLevel = LogLevel.INFO
    cors_origins: List[str] = ["*"]
    
    # Database
    database: DatabaseConfig = DatabaseConfig()
    
    # Redis
    redis: RedisConfig = RedisConfig()
    
    # API
    api_prefix: str = "/api/v1"
    docs_url: Optional[str] = "/docs"
    redoc_url: Optional[str] = "/redoc"
    
    class Config:
        env_file = ".env"
        case_sensitive = False
        
    @property
    def is_production(self) -> bool:
        return self.environment == Environment.PRODUCTION
    
    @property
    def is_development(self) -> bool:
        return self.environment == Environment.DEVELOPMENT
```

### 2. Repository Implementation

```python
import asyncio
from typing import Optional, Dict
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import declarative_base, Mapped, mapped_column
from sqlalchemy import String, DateTime, select
from contextlib import asynccontextmanager
from ...domain.entities.user import User, UserId
from ...domain.ports.user_repository import UserRepository
from ...domain.errors.domain_errors import UserNotFoundError
import structlog


logger = structlog.get_logger()
Base = declarative_base()


class UserModel(Base):
    """SQLAlchemy model for User"""
    __tablename__ = "users"
    
    id: Mapped[str] = mapped_column(String, primary_key=True)
    email: Mapped[str] = mapped_column(String, unique=True, nullable=False)
    name: Mapped[str] = mapped_column(String, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    
    def to_entity(self) -> User:
        """Convert model to domain entity"""
        return User(
            id=UserId(self.id),
            email=self.email,
            name=self.name,
            created_at=self.created_at
        )
    
    @classmethod
    def from_entity(cls, user: User) -> "UserModel":
        """Create model from domain entity"""
        return cls(
            id=user.id.value,
            email=user.email,
            name=user.name,
            created_at=user.created_at
        )


class SqlAlchemyUserRepository(UserRepository):
    """SQLAlchemy implementation of UserRepository"""
    
    def __init__(self, session_factory: async_sessionmaker[AsyncSession]):
        self._session_factory = session_factory
    
    @asynccontextmanager
    async def _get_session(self):
        """Get database session with automatic cleanup"""
        async with self._session_factory() as session:
            try:
                yield session
                await session.commit()
            except Exception:
                await session.rollback()
                raise
    
    async def save(self, user: User) -> None:
        """Save user to database"""
        async with self._get_session() as session:
            # Check if user exists
            stmt = select(UserModel).where(UserModel.id == user.id.value)
            result = await session.execute(stmt)
            existing = result.scalar_one_or_none()
            
            if existing:
                # Update existing user
                existing.email = user.email
                existing.name = user.name
            else:
                # Create new user
                user_model = UserModel.from_entity(user)
                session.add(user_model)
            
            await logger.ainfo("User saved", user_id=user.id.value)
    
    async def find_by_id(self, user_id: str) -> Optional[User]:
        """Find user by ID"""
        async with self._get_session() as session:
            stmt = select(UserModel).where(UserModel.id == user_id)
            result = await session.execute(stmt)
            user_model = result.scalar_one_or_none()
            
            if not user_model:
                return None
            
            return user_model.to_entity()
    
    async def find_by_email(self, email: str) -> Optional[User]:
        """Find user by email"""
        async with self._get_session() as session:
            stmt = select(UserModel).where(UserModel.email == email)
            result = await session.execute(stmt)
            user_model = result.scalar_one_or_none()
            
            if not user_model:
                return None
            
            return user_model.to_entity()
    
    async def delete(self, user_id: str) -> None:
        """Delete user by ID"""
        async with self._get_session() as session:
            stmt = select(UserModel).where(UserModel.id == user_id)
            result = await session.execute(stmt)
            user_model = result.scalar_one_or_none()
            
            if not user_model:
                raise UserNotFoundError(user_id)
            
            await session.delete(user_model)
            await logger.ainfo("User deleted", user_id=user_id)


class DatabaseManager:
    """Database connection and session management"""
    
    def __init__(self, database_url: str):
        self.engine = create_async_engine(
            database_url,
            echo=False,
            pool_size=10,
            max_overflow=20
        )
        self.session_factory = async_sessionmaker(
            self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )
    
    async def create_tables(self):
        """Create database tables"""
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
    
    async def close(self):
        """Close database connections"""
        await self.engine.dispose()
```

### 3. Event Publishing

```python
import asyncio
import json
from typing import List, Optional
from dataclasses import asdict
from ...domain.ports.event_publisher import EventPublisher
from ...domain.events import DomainEvent
import structlog


logger = structlog.get_logger()


class ReliableEventPublisher(EventPublisher):
    """Event publisher with retry mechanism and error handling"""
    
    def __init__(self, max_retries: int = 3, retry_delay: float = 1.0):
        self._max_retries = max_retries
        self._retry_delay = retry_delay
        self._failed_events: List[DomainEvent] = []
        self._event_handlers = {}
    
    async def publish(self, event: DomainEvent) -> None:
        """Publish single domain event with retry logic"""
        for attempt in range(self._max_retries + 1):
            try:
                await self._do_publish(event)
                await logger.ainfo(
                    "Event published successfully",
                    event_type=event.event_type(),
                    event_id=getattr(event, 'event_id', None),
                    attempt=attempt + 1
                )
                return
                
            except Exception as e:
                if attempt == self._max_retries:
                    # Final attempt failed, add to failed events
                    self._failed_events.append(event)
                    await logger.aerror(
                        "Event publishing failed after all retries",
                        event_type=event.event_type(),
                        error=str(e),
                        attempts=attempt + 1
                    )
                    raise
                
                await logger.awarn(
                    "Event publishing failed, retrying",
                    event_type=event.event_type(),
                    error=str(e),
                    attempt=attempt + 1,
                    retry_in=self._retry_delay
                )
                await asyncio.sleep(self._retry_delay)
    
    async def publish_batch(self, events: List[DomainEvent]) -> None:
        """Publish multiple events atomically"""
        try:
            for event in events:
                await self._do_publish(event)
            
            await logger.ainfo("Batch events published successfully", count=len(events))
            
        except Exception as e:
            # Add all events to failed list
            self._failed_events.extend(events)
            await logger.aerror("Batch event publishing failed", error=str(e), count=len(events))
            raise
    
    async def _do_publish(self, event: DomainEvent) -> None:
        """Actually publish the event (implement based on your message broker)"""
        # Example implementation for in-memory processing
        event_data = event.to_dict()
        
        # Here you would typically:
        # 1. Send to message broker (RabbitMQ, Kafka, etc.)
        # 2. Store in event store
        # 3. Trigger event handlers
        
        # For demo purposes, just log the event
        await logger.ainfo("Publishing event", event_data=event_data)
        
        # Simulate potential failure for testing
        # if random.random() < 0.1:  # 10% failure rate
        #     raise Exception("Simulated publishing failure")
    
    async def retry_failed_events(self) -> None:
        """Retry publishing failed events"""
        if not self._failed_events:
            return
        
        failed_copy = list(self._failed_events)
        self._failed_events.clear()
        
        for event in failed_copy:
            try:
                await self.publish(event)
            except Exception:
                # If retry fails, keep in failed list
                pass
    
    def get_failed_events_count(self) -> int:
        """Get count of failed events"""
        return len(self._failed_events)
```

## Presentation Layer Guidelines

### 1. FastAPI Application Factory

```python
from fastapi import FastAPI, Request, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager
from ...domain.errors.domain_errors import DomainError
from ..config.app_config import AppConfig
from .routers import user_router
from .middleware.request_id import RequestIDMiddleware
from .middleware.logging import LoggingMiddleware
import structlog


logger = structlog.get_logger()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan management"""
    # Startup
    await logger.ainfo("Application starting up")
    yield
    # Shutdown
    await logger.ainfo("Application shutting down")


def create_app(config: AppConfig) -> FastAPI:
    """Create and configure FastAPI application"""
    
    app = FastAPI(
        title="Python Hexagonal Architecture",
        description="Clean Architecture with FastAPI",
        version="1.0.0",
        docs_url=config.docs_url if not config.is_production else None,
        redoc_url=config.redoc_url if not config.is_production else None,
        lifespan=lifespan
    )
    
    # Add middleware
    app.add_middleware(RequestIDMiddleware)
    app.add_middleware(LoggingMiddleware)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=config.cors_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Add exception handlers
    add_exception_handlers(app)
    
    # Include routers
    app.include_router(user_router, prefix=f"{config.api_prefix}/users", tags=["users"])
    
    # Health check endpoint
    @app.get("/health")
    async def health_check():
        return {"status": "healthy", "service": "python-hexagonal-architecture"}
    
    @app.get("/")
    async def root():
        return {"message": "Python Hexagonal Architecture API"}
    
    return app


def add_exception_handlers(app: FastAPI):
    """Add global exception handlers"""
    
    @app.exception_handler(DomainError)
    async def domain_error_handler(request: Request, exc: DomainError):
        """Handle domain errors"""
        status_code = get_status_code_for_domain_error(exc)
        
        await logger.aerror(
            "Domain error occurred",
            error_code=exc.error_code,
            message=exc.message,
            path=request.url.path,
            method=request.method
        )
        
        return JSONResponse(
            status_code=status_code,
            content=exc.to_dict()
        )
    
    @app.exception_handler(ValueError)
    async def value_error_handler(request: Request, exc: ValueError):
        """Handle validation errors"""
        await logger.aerror(
            "Validation error occurred",
            error=str(exc),
            path=request.url.path,
            method=request.method
        )
        
        return JSONResponse(
            status_code=400,
            content={
                "error_code": "VALIDATION_ERROR",
                "message": str(exc),
                "type": "ValueError"
            }
        )
    
    @app.exception_handler(Exception)
    async def general_error_handler(request: Request, exc: Exception):
        """Handle unexpected errors"""
        await logger.aerror(
            "Unexpected error occurred",
            error=str(exc),
            error_type=type(exc).__name__,
            path=request.url.path,
            method=request.method
        )
        
        return JSONResponse(
            status_code=500,
            content={
                "error_code": "INTERNAL_SERVER_ERROR",
                "message": "An unexpected error occurred",
                "type": "InternalServerError"
            }
        )


def get_status_code_for_domain_error(error: DomainError) -> int:
    """Map domain error codes to HTTP status codes"""
    status_map = {
        "USER_NOT_FOUND": 404,
        "USER_ALREADY_EXISTS": 409,
        "USER_VALIDATION_ERROR": 400,
        "INVALID_USER_DATA": 400,
    }
    return status_map.get(error.error_code, 500)
```

### 2. HTTP Handlers

```python
from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from typing import Annotated
from ...usecases.create_user.create_user import CreateUserUseCase
from ...usecases.create_user.dtos import CreateUserRequest, CreateUserResponse
from ...domain.errors.domain_errors import DomainError
from ..dependencies import get_create_user_use_case
import structlog


logger = structlog.get_logger()
router = APIRouter()


@router.post(
    "",
    response_model=CreateUserResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Create a new user",
    description="Create a new user with email and name"
)
async def create_user(
    request: CreateUserRequest,
    use_case: Annotated[CreateUserUseCase, Depends(get_create_user_use_case)]
) -> CreateUserResponse:
    """Create a new user endpoint"""
    try:
        response = await use_case.execute(request)
        return response
        
    except DomainError:
        # Let the global exception handler deal with domain errors
        raise
    
    except Exception as e:
        await logger.aerror("Unexpected error in create_user endpoint", error=str(e))
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="An unexpected error occurred"
        )


@router.get(
    "/{user_id}",
    summary="Get user by ID",
    description="Retrieve a user by their unique identifier"
)
async def get_user(
    user_id: str,
    # Add get user use case dependency here
):
    """Get user by ID endpoint"""
    # Implementation here
    pass
```

### 3. Dependency Injection

```python
from functools import lru_cache
from fastapi import Depends
from sqlalchemy.ext.asyncio import async_sessionmaker, AsyncSession
from typing import Annotated

from ..config.app_config import AppConfig
from ...infra.persistence.sqlalchemy.user_repository import SqlAlchemyUserRepository, DatabaseManager
from ...infra.messaging.reliable_event_publisher import ReliableEventPublisher
from ...infra.id_generation.uuid_id_generator import UuidIdGenerator
from ...domain.services.user_domain_service import UserDomainService
from ...usecases.create_user.create_user import CreateUserUseCase


@lru_cache()
def get_config() -> AppConfig:
    """Get application configuration"""
    return AppConfig()


@lru_cache()
def get_database_manager(config: Annotated[AppConfig, Depends(get_config)]) -> DatabaseManager:
    """Get database manager"""
    return DatabaseManager(config.database.url)


def get_user_repository(
    db_manager: Annotated[DatabaseManager, Depends(get_database_manager)]
) -> SqlAlchemyUserRepository:
    """Get user repository"""
    return SqlAlchemyUserRepository(db_manager.session_factory)


@lru_cache()
def get_event_publisher() -> ReliableEventPublisher:
    """Get event publisher"""
    return ReliableEventPublisher()


@lru_cache()
def get_id_generator() -> UuidIdGenerator:
    """Get ID generator"""
    return UuidIdGenerator()


def get_user_domain_service(
    user_repo: Annotated[SqlAlchemyUserRepository, Depends(get_user_repository)]
) -> UserDomainService:
    """Get user domain service"""
    return UserDomainService(user_repo)


def get_create_user_use_case(
    user_repo: Annotated[SqlAlchemyUserRepository, Depends(get_user_repository)],
    event_publisher: Annotated[ReliableEventPublisher, Depends(get_event_publisher)],
    user_domain_service: Annotated[UserDomainService, Depends(get_user_domain_service)],
    id_generator: Annotated[UuidIdGenerator, Depends(get_id_generator)]
) -> CreateUserUseCase:
    """Get create user use case"""
    return CreateUserUseCase(
        user_repository=user_repo,
        event_publisher=event_publisher,
        user_domain_service=user_domain_service,
        id_generator=id_generator
    )
```

## Testing Guidelines

### 1. Unit Testing

```python
import pytest
from unittest.mock import AsyncMock, Mock
from datetime import datetime

from src.domain.entities.user import User, UserId
from src.domain.errors.domain_errors import UserAlreadyExistsError
from src.usecases.create_user.create_user import CreateUserUseCase
from src.usecases.create_user.dtos import CreateUserRequest


class TestCreateUserUseCase:
    """Unit tests for CreateUserUseCase"""
    
    @pytest.fixture
    def mock_user_repository(self):
        return AsyncMock()
    
    @pytest.fixture
    def mock_event_publisher(self):
        return AsyncMock()
    
    @pytest.fixture
    def mock_user_domain_service(self):
        return AsyncMock()
    
    @pytest.fixture
    def mock_id_generator(self):
        mock = Mock()
        mock.generate_user_id.return_value = "user_12345678"
        return mock
    
    @pytest.fixture
    def use_case(self, mock_user_repository, mock_event_publisher, 
                mock_user_domain_service, mock_id_generator):
        return CreateUserUseCase(
            user_repository=mock_user_repository,
            event_publisher=mock_event_publisher,
            user_domain_service=mock_user_domain_service,
            id_generator=mock_id_generator
        )
    
    @pytest.mark.asyncio
    async def test_execute_success(self, use_case, mock_user_repository, 
                                  mock_event_publisher, mock_user_domain_service):
        # Arrange
        request = CreateUserRequest(email="test@example.com", name="Test User")
        
        mock_user_domain_service.validate_unique_email.return_value = None
        mock_user_domain_service.validate_user_data.return_value = []
        mock_user_repository.save.return_value = None
        mock_event_publisher.publish.return_value = None
        
        # Act
        response = await use_case.execute(request)
        
        # Assert
        assert response.user_id == "user_12345678"
        mock_user_domain_service.validate_unique_email.assert_called_once_with("test@example.com")
        mock_user_repository.save.assert_called_once()
        mock_event_publisher.publish.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_execute_user_already_exists(self, use_case, mock_user_domain_service):
        # Arrange
        request = CreateUserRequest(email="existing@example.com", name="Test User")
        mock_user_domain_service.validate_unique_email.side_effect = UserAlreadyExistsError("existing@example.com")
        
        # Act & Assert
        with pytest.raises(UserAlreadyExistsError):
            await use_case.execute(request)
    
    @pytest.mark.asyncio
    async def test_execute_invalid_request(self, use_case):
        # Arrange
        request = CreateUserRequest(email="", name="")
        
        # Act & Assert
        with pytest.raises(ValueError):
            await use_case.execute(request)


class TestUser:
    """Unit tests for User entity"""
    
    def test_create_valid_user(self):
        # Arrange
        mock_id_generator = Mock()
        mock_id_generator.generate_user_id.return_value = "user_12345678"
        
        # Act
        user = User.create("test@example.com", "Test User", mock_id_generator)
        
        # Assert
        assert user.id.value == "user_12345678"
        assert user.email == "test@example.com"
        assert user.name == "Test User"
        assert isinstance(user.created_at, datetime)
    
    def test_create_invalid_email(self):
        # Arrange
        mock_id_generator = Mock()
        
        # Act & Assert
        with pytest.raises(ValueError, match="Invalid email format"):
            User.create("invalid-email", "Test User", mock_id_generator)
    
    def test_create_empty_name(self):
        # Arrange
        mock_id_generator = Mock()
        
        # Act & Assert
        with pytest.raises(ValueError, match="Name cannot be empty"):
            User.create("test@example.com", "", mock_id_generator)
    
    def test_is_valid(self):
        # Arrange
        user = User(
            id=UserId("user_12345678"),
            email="test@example.com",
            name="Test User",
            created_at=datetime.now()
        )
        
        # Act & Assert
        assert user.is_valid() is True
```

### 2. Integration Testing

```python
import pytest
import asyncio
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

from src.presentation.app import create_app
from src.config.app_config import AppConfig
from src.infra.persistence.sqlalchemy.user_repository import Base


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
async def test_db_engine():
    """Create test database engine"""
    engine = create_async_engine(
        "sqlite+aiosqlite:///./test.db",
        echo=False
    )
    
    # Create tables
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    yield engine
    
    # Cleanup
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
    await engine.dispose()


@pytest.fixture
async def test_app(test_db_engine):
    """Create test FastAPI application"""
    config = AppConfig(
        database_url="sqlite+aiosqlite:///./test.db",
        environment="testing"
    )
    
    app = create_app(config)
    return app


@pytest.fixture
async def client(test_app):
    """Create test HTTP client"""
    async with AsyncClient(app=test_app, base_url="http://test") as ac:
        yield ac


class TestUserAPI:
    """Integration tests for User API"""
    
    @pytest.mark.asyncio
    async def test_create_user_success(self, client):
        # Arrange
        user_data = {
            "email": "test@example.com",
            "name": "Test User"
        }
        
        # Act
        response = await client.post("/api/v1/users", json=user_data)
        
        # Assert
        assert response.status_code == 201
        data = response.json()
        assert "user_id" in data
        assert data["user_id"].startswith("user_")
    
    @pytest.mark.asyncio
    async def test_create_user_duplicate_email(self, client):
        # Arrange
        user_data = {
            "email": "duplicate@example.com",
            "name": "First User"
        }
        
        # Create first user
        await client.post("/api/v1/users", json=user_data)
        
        # Try to create duplicate
        duplicate_data = {
            "email": "duplicate@example.com",
            "name": "Second User"
        }
        
        # Act
        response = await client.post("/api/v1/users", json=duplicate_data)
        
        # Assert
        assert response.status_code == 409
        data = response.json()
        assert data["error_code"] == "USER_ALREADY_EXISTS"
    
    @pytest.mark.asyncio
    async def test_create_user_invalid_data(self, client):
        # Arrange
        invalid_data = {
            "email": "invalid-email",
            "name": ""
        }
        
        # Act
        response = await client.post("/api/v1/users", json=invalid_data)
        
        # Assert
        assert response.status_code == 400
    
    @pytest.mark.asyncio
    async def test_health_check(self, client):
        # Act
        response = await client.get("/health")
        
        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
```

## Project Structure

```
src/
├── domain/                           # Domain layer (business logic)
│   ├── __init__.py
│   ├── entities/                     # Domain entities
│   │   ├── __init__.py
│   │   ├── user.py                   # User entity with business rules
│   │   └── base.py                   # Base entity classes
│   ├── events/                       # Domain events
│   │   ├── __init__.py
│   │   ├── user_created.py           # User created event
│   │   └── base.py                   # Base event classes
│   ├── ports/                        # Interfaces/contracts
│   │   ├── __init__.py
│   │   ├── user_repository.py        # User repository interface
│   │   └── event_publisher.py        # Event publisher interface
│   ├── services/                     # Domain services
│   │   ├── __init__.py
│   │   └── user_domain_service.py    # Complex user business rules
│   └── errors/                       # Domain errors
│       ├── __init__.py
│       └── domain_errors.py          # Domain-specific exceptions
├── usecases/                         # Application layer
│   ├── __init__.py
│   └── create_user/                  # Create user use case
│       ├── __init__.py
│       ├── create_user.py            # Use case implementation
│       └── dtos.py                   # Request/response DTOs
├── infra/                           # Infrastructure layer
│   ├── __init__.py
│   ├── persistence/                  # Data persistence
│   │   ├── __init__.py
│   │   ├── memory/                   # In-memory implementations
│   │   │   ├── __init__.py
│   │   │   └── user_repository.py    # In-memory user repository
│   │   └── sqlalchemy/               # SQLAlchemy implementations
│   │       ├── __init__.py
│   │       └── user_repository.py    # SQLAlchemy user repository
│   ├── messaging/                    # Event publishing
│   │   ├── __init__.py
│   │   ├── in_memory_event_publisher.py
│   │   └── reliable_event_publisher.py
│   ├── cache/                        # Caching implementations
│   │   ├── __init__.py
│   │   └── redis_cache.py
│   ├── externalapis/                 # External API clients
│   │   ├── __init__.py
│   │   └── email_service.py
│   └── id_generation/                # ID generation
│       ├── __init__.py
│       └── uuid_id_generator.py
├── presentation/                     # Presentation layer
│   ├── __init__.py
│   ├── api/                         # HTTP API
│   │   ├── __init__.py
│   │   ├── routers/                 # FastAPI routers
│   │   │   ├── __init__.py
│   │   │   └── user_router.py       # User endpoints
│   │   └── dependencies.py          # FastAPI dependencies
│   ├── middleware/                   # HTTP middleware
│   │   ├── __init__.py
│   │   ├── request_id.py            # Request ID middleware
│   │   └── logging.py               # Logging middleware
│   └── app.py                       # FastAPI application factory
├── config/                          # Configuration
│   ├── __init__.py
│   └── app_config.py                # Application configuration
└── main.py                          # Application entry point

tests/                               # Test files
├── __init__.py
├── unit/                            # Unit tests
│   ├── __init__.py
│   ├── domain/                      # Domain layer tests
│   │   ├── __init__.py
│   │   ├── test_user.py
│   │   └── test_user_domain_service.py
│   ├── usecases/                    # Use case tests
│   │   ├── __init__.py
│   │   └── test_create_user.py
│   └── infra/                       # Infrastructure tests
│       ├── __init__.py
│       └── test_user_repository.py
├── integration/                     # Integration tests
│   ├── __init__.py
│   └── test_user_api.py
└── conftest.py                      # Test configuration

requirements/                        # Dependencies
├── base.txt                         # Base dependencies
├── dev.txt                          # Development dependencies
└── test.txt                         # Test dependencies

pyproject.toml                       # Project configuration
pytest.ini                          # Pytest configuration
.env.example                         # Environment variables example
Dockerfile                          # Docker configuration
docker-compose.yml                   # Docker Compose configuration
README.md                           # Project documentation
```

## Dependencies (pyproject.toml)

```toml
[project]
name = "python-hexagonal-architecture"
version = "0.1.0"
description = "Python implementation of hexagonal architecture"
readme = "README.md"
requires-python = ">=3.10"
dependencies = [
    "fastapi>=0.104.0",
    "uvicorn[standard]>=0.24.0",
    "pydantic>=2.5.0",
    "pydantic-settings>=2.1.0",
    "sqlalchemy[asyncio]>=2.0.0",
    "alembic>=1.13.0",
    "asyncpg>=0.29.0",  # PostgreSQL async driver
    "aiosqlite>=0.19.0",  # SQLite async driver
    "structlog>=23.2.0",
    "httpx>=0.25.0",
    "redis>=5.0.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "pytest-cov>=4.1.0",
    "black>=23.0.0",
    "isort>=5.12.0",
    "flake8>=6.0.0",
    "mypy>=1.7.0",
    "pre-commit>=3.5.0",
]

test = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "pytest-cov>=4.1.0",
    "pytest-mock>=3.12.0",
    "httpx>=0.25.0",
    "factory-boy>=3.3.0",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = [
    "-v",
    "--strict-markers",
    "--strict-config",
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-report=xml",
]

[tool.black]
line-length = 100
target-version = ['py310']
include = '\.pyi?$'
extend-exclude = '''
/(
  # directories
  \.eggs
  | \.git
  | \.hg
  | \.mypy_cache
  | \.tox
  | \.venv
  | build
  | dist
)/
'''

[tool.isort]
profile = "black"
line_length = 100
multi_line_output = 3
include_trailing_comma = true
force_grid_wrap = 0
use_parentheses = true
ensure_newline_before_comments = true

[tool.mypy]
python_version = "3.10"
check_untyped_defs = true
ignore_missing_imports = true
warn_unused_ignores = true
warn_redundant_casts = true
warn_unused_configs = true
strict_optional = true
```

## Development Workflow

### 1. Setup Instructions

```bash
# Clone repository
git clone <repository-url>
cd python-hexagonal-architecture

# Install dependencies with uv (recommended) or pip
uv sync --all-extras

# Or with pip
pip install -e ".[dev,test]"

# Setup pre-commit hooks
pre-commit install

# Run database migrations
alembic upgrade head

# Start development server
uv run uvicorn src.main:app --reload --host 0.0.0.0 --port 8000
```

### 2. Code Quality Tools

```bash
# Format code
black src tests
isort src tests

# Lint code
flake8 src tests
mypy src

# Run tests
pytest

# Run tests with coverage
pytest --cov=src --cov-report=html

# Run specific test
pytest tests/unit/domain/test_user.py::TestUser::test_create_valid_user
```

### 3. Adding New Features

1. **Start with Domain**: Define entities, events, and ports
2. **Add Use Case**: Implement application logic with DTOs
3. **Implement Infrastructure**: Add repository and external service implementations
4. **Add Presentation**: Create HTTP endpoints and handlers
5. **Write Tests**: Unit tests for each layer, integration tests for complete flows
6. **Update Documentation**: README, API docs, architectural decisions

## Best Practices Summary

1. **Keep Domain Pure**: No external dependencies in domain layer
2. **Use Dependency Injection**: Constructor injection for testability
3. **Handle Errors Properly**: Domain-specific exceptions with proper HTTP mapping
4. **Log Structured Data**: Use structured logging for observability
5. **Test Each Layer**: Unit tests, integration tests, and end-to-end tests
6. **Configure via Environment**: Use pydantic-settings for configuration
7. **Use Type Hints**: Full type annotations for better IDE support
8. **Follow SOLID Principles**: Single responsibility, dependency inversion, etc.
9. **Event-Driven Architecture**: Use domain events for decoupling
10. **Async Throughout**: Use async/await for scalable I/O operations

This comprehensive guide provides a robust foundation for building maintainable, testable, and production-ready Python microservices using hexagonal architecture principles.