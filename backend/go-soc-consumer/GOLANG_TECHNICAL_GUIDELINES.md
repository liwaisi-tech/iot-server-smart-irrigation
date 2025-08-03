# Golang Technical Guidelines - Hexagonal Architecture

## Overview

This document establishes technical guidelines for Go projects following hexagonal architecture (ports and adapters pattern) with clean architecture principles. These guidelines ensure maintainable, testable, and production-ready Go microservices that leverage Go's strengths in concurrency, performance, and simplicity.

## Architecture Philosophy

### Core Principles
- **Dependency Inversion**: Domain layer never imports from outer layers
- **Single Responsibility**: Each package/struct has one reason to change
- **Interface Segregation**: Small, focused interfaces over large ones
- **Explicit Dependencies**: Clear dependency injection without magic
- **Context Propagation**: Proper context handling for cancellation and timeouts
- **Error Handling**: Explicit error handling with proper wrapping and context

### Layer Responsibilities

```
project/
├── cmd/                              # Application entry points
│   └── server/                       # Main server application
│       └── main.go
├── internal/                         # Private application code
│   ├── domain/                       # Business core (no external dependencies)
│   │   ├── entities/                 # Business entities with validation
│   │   ├── events/                   # Domain events for event-driven architecture
│   │   ├── ports/                    # Interfaces/contracts (Repository, Services)
│   │   ├── services/                 # Domain services for complex business rules
│   │   └── errors/                   # Domain-specific errors
│   ├── usecases/                     # Application logic
│   │   └── {use_case_name}/          # One package per use case
│   ├── infrastructure/               # External implementations
│   │   ├── persistence/              # Repository implementations
│   │   ├── messaging/                # Event publishers, message brokers
│   │   ├── cache/                    # Cache implementations
│   │   └── externalapis/             # Third-party API clients
│   └── presentation/                 # Delivery layer
│       ├── http/                     # HTTP handlers and middleware
│       └── grpc/                     # gRPC handlers (if needed)
├── pkg/                             # Public packages (reusable across projects)
│   ├── logger/                      # Structured logging utilities
│   └── config/                      # Configuration management
└── config/                          # Configuration files
    └── config.yaml
```

## Domain Layer Guidelines

### 1. Domain Entities

**✅ Recommended**: Pure Go structs with validation methods
```go
package entities

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "net/mail"
    "strings"
    "time"
)

// User represents the user domain entity
type User struct {
    id        UserID
    email     string
    name      string
    createdAt time.Time
}

// UserID is a value object for user identification
type UserID struct {
    value string
}

// NewUserID creates a new UserID with validation
func NewUserID(value string) (UserID, error) {
    if value == "" {
        return UserID{}, fmt.Errorf("user ID cannot be empty")
    }
    if !strings.HasPrefix(value, "user_") {
        return UserID{}, fmt.Errorf("user ID must start with 'user_'")
    }
    return UserID{value: value}, nil
}

// Value returns the string value of the UserID
func (u UserID) Value() string {
    return u.value
}

// String implements the Stringer interface
func (u UserID) String() string {
    return u.value
}

// IDGenerator defines the interface for generating user IDs
type IDGenerator interface {
    GenerateUserID() (UserID, error)
}

// NewUser creates a new user with validation
func NewUser(email, name string, idGen IDGenerator) (*User, error) {
    // Validate email
    if _, err := mail.ParseAddress(email); err != nil {
        return nil, fmt.Errorf("invalid email format: %w", err)
    }
    
    // Validate name
    name = strings.TrimSpace(name)
    if name == "" {
        return nil, fmt.Errorf("name cannot be empty")
    }
    if len(name) > 100 {
        return nil, fmt.Errorf("name cannot exceed 100 characters")
    }
    
    // Generate ID
    id, err := idGen.GenerateUserID()
    if err != nil {
        return nil, fmt.Errorf("failed to generate user ID: %w", err)
    }
    
    return &User{
        id:        id,
        email:     email,
        name:      name,
        createdAt: time.Now(),
    }, nil
}

// Getters (Go doesn't have properties, use explicit getters)
func (u *User) ID() UserID       { return u.id }
func (u *User) Email() string    { return u.email }
func (u *User) Name() string     { return u.name }
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdateEmail updates the user's email with validation
func (u *User) UpdateEmail(newEmail string) error {
    if _, err := mail.ParseAddress(newEmail); err != nil {
        return fmt.Errorf("invalid email format: %w", err)
    }
    u.email = newEmail
    return nil
}

// UpdateName updates the user's name with validation
func (u *User) UpdateName(newName string) error {
    newName = strings.TrimSpace(newName)
    if newName == "" {
        return fmt.Errorf("name cannot be empty")
    }
    if len(newName) > 100 {
        return fmt.Errorf("name cannot exceed 100 characters")
    }
    u.name = newName
    return nil
}

// IsValid validates the current state of the user
func (u *User) IsValid() error {
    if u.id.value == "" {
        return fmt.Errorf("user ID is required")
    }
    if u.email == "" {
        return fmt.Errorf("email is required")
    }
    if _, err := mail.ParseAddress(u.email); err != nil {
        return fmt.Errorf("invalid email format: %w", err)
    }
    if strings.TrimSpace(u.name) == "" {
        return fmt.Errorf("name is required")
    }
    return nil
}

// DefaultIDGenerator implements IDGenerator using crypto/rand
type DefaultIDGenerator struct{}

// GenerateUserID generates a new user ID
func (g *DefaultIDGenerator) GenerateUserID() (UserID, error) {
    bytes := make([]byte, 4)
    if _, err := rand.Read(bytes); err != nil {
        return UserID{}, fmt.Errorf("failed to generate random bytes: %w", err)
    }
    
    id := fmt.Sprintf("user_%s", hex.EncodeToString(bytes))
    return NewUserID(id)
}
```

### 2. Domain Events

```go
package events

import (
    "encoding/json"
    "time"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
)

// DomainEvent represents the interface all domain events must implement
type DomainEvent interface {
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
    EventID() string
    ToJSON() ([]byte, error)
}

// BaseEvent provides common functionality for all domain events
type BaseEvent struct {
    eventID    string
    occurredAt time.Time
}

// NewBaseEvent creates a new base event
func NewBaseEvent() BaseEvent {
    return BaseEvent{
        eventID:    generateEventID(),
        occurredAt: time.Now(),
    }
}

func (e BaseEvent) EventID() string    { return e.eventID }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }

// UserCreated represents the user created domain event
type UserCreated struct {
    BaseEvent
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(user *entities.User) *UserCreated {
    return &UserCreated{
        BaseEvent: NewBaseEvent(),
        UserID:    user.ID().Value(),
        Email:     user.Email(),
        Name:      user.Name(),
        CreatedAt: user.CreatedAt(),
    }
}

// EventType returns the event type
func (e *UserCreated) EventType() string {
    return "user.created"
}

// AggregateID returns the aggregate ID
func (e *UserCreated) AggregateID() string {
    return e.UserID
}

// ToJSON converts the event to JSON
func (e *UserCreated) ToJSON() ([]byte, error) {
    eventData := struct {
        EventID    string    `json:"event_id"`
        EventType  string    `json:"event_type"`
        AggregateID string   `json:"aggregate_id"`
        OccurredAt time.Time `json:"occurred_at"`
        Payload    *UserCreated `json:"payload"`
    }{
        EventID:     e.EventID(),
        EventType:   e.EventType(),
        AggregateID: e.AggregateID(),
        OccurredAt:  e.OccurredAt(),
        Payload:     e,
    }
    
    return json.Marshal(eventData)
}

// generateEventID generates a unique event ID
func generateEventID() string {
    bytes := make([]byte, 4)
    rand.Read(bytes)
    return fmt.Sprintf("evt_%s", hex.EncodeToString(bytes))
}

// UserUpdated represents the user updated domain event
type UserUpdated struct {
    BaseEvent
    UserID     string            `json:"user_id"`
    Changes    map[string]string `json:"changes"`
    UpdatedAt  time.Time         `json:"updated_at"`
}

// NewUserUpdated creates a new UserUpdated event
func NewUserUpdated(userID string, changes map[string]string) *UserUpdated {
    return &UserUpdated{
        BaseEvent: NewBaseEvent(),
        UserID:    userID,
        Changes:   changes,
        UpdatedAt: time.Now(),
    }
}

func (e *UserUpdated) EventType() string { return "user.updated" }
func (e *UserUpdated) AggregateID() string { return e.UserID }

func (e *UserUpdated) ToJSON() ([]byte, error) {
    eventData := struct {
        EventID     string            `json:"event_id"`
        EventType   string            `json:"event_type"`
        AggregateID string            `json:"aggregate_id"`
        OccurredAt  time.Time         `json:"occurred_at"`
        Payload     *UserUpdated      `json:"payload"`
    }{
        EventID:     e.EventID(),
        EventType:   e.EventType(),
        AggregateID: e.AggregateID(),
        OccurredAt:  e.OccurredAt(),
        Payload:     e,
    }
    
    return json.Marshal(eventData)
}
```

### 3. Domain Ports (Interfaces)

```go
package ports

import (
    "context"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
    "github.com/yourorg/yourproject/internal/domain/events"
)

// UserRepository defines the contract for user persistence operations
type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, userID entities.UserID) (*entities.User, error)
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
    Delete(ctx context.Context, userID entities.UserID) error
    List(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

// EventPublisher defines the contract for publishing domain events
type EventPublisher interface {
    Publish(ctx context.Context, event events.DomainEvent) error
    PublishBatch(ctx context.Context, events []events.DomainEvent) error
}

// EmailService defines the contract for email operations
type EmailService interface {
    SendWelcomeEmail(ctx context.Context, email, name string) error
    SendPasswordResetEmail(ctx context.Context, email, resetToken string) error
}

// CacheService defines the contract for caching operations
type CacheService interface {
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Get(ctx context.Context, key string, dest interface{}) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

### 4. Domain Services

```go
package services

import (
    "context"
    "fmt"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
    "github.com/yourorg/yourproject/internal/domain/ports"
    "github.com/yourorg/yourproject/internal/domain/errors"
)

// UserDomainService encapsulates complex business rules for users
type UserDomainService struct {
    userRepo ports.UserRepository
}

// NewUserDomainService creates a new UserDomainService
func NewUserDomainService(userRepo ports.UserRepository) *UserDomainService {
    return &UserDomainService{
        userRepo: userRepo,
    }
}

// ValidateUniqueEmail ensures email is unique across all users
func (s *UserDomainService) ValidateUniqueEmail(ctx context.Context, email string, excludeUserID *entities.UserID) error {
    existingUser, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        // If user not found, email is unique
        if errors.IsNotFoundError(err) {
            return nil
        }
        return fmt.Errorf("failed to check email uniqueness: %w", err)
    }
    
    // If we're excluding a specific user (for updates), check if it's the same user
    if excludeUserID != nil && existingUser.ID().Value() == excludeUserID.Value() {
        return nil
    }
    
    return errors.NewUserAlreadyExistsError(email)
}

// CanDeleteUser checks if a user can be deleted based on business rules
func (s *UserDomainService) CanDeleteUser(ctx context.Context, userID entities.UserID) error {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to find user for deletion check: %w", err)
    }
    
    // Add business logic for deletion rules
    // For example: check if user has active orders, subscriptions, etc.
    
    // For now, allow all deletions
    _ = user
    return nil
}

// ValidateUserData performs comprehensive business validation
func (s *UserDomainService) ValidateUserData(user *entities.User) []string {
    var errors []string
    
    // Basic entity validation
    if err := user.IsValid(); err != nil {
        errors = append(errors, err.Error())
    }
    
    // Additional business rules
    if len(user.Name()) < 2 {
        errors = append(errors, "name must be at least 2 characters long")
    }
    
    // Check for inappropriate content in name (simplified example)
    inappropriateWords := []string{"admin", "root", "system"}
    for _, word := range inappropriateWords {
        if strings.Contains(strings.ToLower(user.Name()), word) {
            errors = append(errors, "name contains inappropriate content")
            break
        }
    }
    
    return errors
}

// GenerateUserMetrics calculates business metrics for a user
func (s *UserDomainService) GenerateUserMetrics(ctx context.Context, userID entities.UserID) (*UserMetrics, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to find user for metrics: %w", err)
    }
    
    // Calculate business metrics
    accountAge := time.Since(user.CreatedAt())
    
    return &UserMetrics{
        UserID:     user.ID(),
        AccountAge: accountAge,
        // Add more metrics as needed
    }, nil
}

// UserMetrics represents calculated user metrics
type UserMetrics struct {
    UserID     entities.UserID
    AccountAge time.Duration
    // Add more fields as needed
}
```

### 5. Domain Errors

```go
package errors

import (
    "errors"
    "fmt"
)

// DomainError represents a domain-specific error
type DomainError struct {
    Code    string
    Message string
    Cause   error
}

// Error implements the error interface
func (e *DomainError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements error unwrapping
func (e *DomainError) Unwrap() error {
    return e.Cause
}

// Is implements error comparison
func (e *DomainError) Is(target error) bool {
    t, ok := target.(*DomainError)
    if !ok {
        return false
    }
    return e.Code == t.Code
}

// Predefined domain errors
var (
    ErrUserNotFound = &DomainError{
        Code:    "USER_NOT_FOUND",
        Message: "user not found",
    }
    
    ErrUserAlreadyExists = &DomainError{
        Code:    "USER_ALREADY_EXISTS",
        Message: "user already exists",
    }
    
    ErrInvalidUserData = &DomainError{
        Code:    "INVALID_USER_DATA",
        Message: "invalid user data",
    }
    
    ErrUserValidationFailed = &DomainError{
        Code:    "USER_VALIDATION_FAILED",
        Message: "user validation failed",
    }
)

// NewUserNotFoundError creates a new user not found error
func NewUserNotFoundError(userID string) *DomainError {
    return &DomainError{
        Code:    "USER_NOT_FOUND",
        Message: fmt.Sprintf("user with ID %s not found", userID),
    }
}

// NewUserAlreadyExistsError creates a new user already exists error
func NewUserAlreadyExistsError(email string) *DomainError {
    return &DomainError{
        Code:    "USER_ALREADY_EXISTS", 
        Message: fmt.Sprintf("user with email %s already exists", email),
    }
}

// NewUserValidationError creates a new user validation error
func NewUserValidationError(validationErrors []string) *DomainError {
    return &DomainError{
        Code:    "USER_VALIDATION_FAILED",
        Message: fmt.Sprintf("user validation failed: %s", strings.Join(validationErrors, ", ")),
    }
}

// Error checking helpers
func IsNotFoundError(err error) bool {
    return errors.Is(err, ErrUserNotFound)
}

func IsAlreadyExistsError(err error) bool {
    return errors.Is(err, ErrUserAlreadyExists)
}

func IsValidationError(err error) bool {
    return errors.Is(err, ErrUserValidationFailed)
}

// GetErrorCode extracts error code from domain error
func GetErrorCode(err error) string {
    var domainErr *DomainError
    if errors.As(err, &domainErr) {
        return domainErr.Code
    }
    return "UNKNOWN_ERROR"
}
```

## Use Cases Layer Guidelines

### 1. Use Case Implementation

```go
package createuser

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
    "github.com/yourorg/yourproject/internal/domain/events"
    "github.com/yourorg/yourproject/internal/domain/ports"
    "github.com/yourorg/yourproject/internal/domain/services"
    domainerrors "github.com/yourorg/yourproject/internal/domain/errors"
)

// UseCase handles the create user use case
type UseCase struct {
    userRepo           ports.UserRepository
    eventPublisher     ports.EventPublisher
    userDomainService  *services.UserDomainService
    emailService       ports.EmailService
    idGenerator        entities.IDGenerator
    logger             *slog.Logger
}

// NewUseCase creates a new create user use case
func NewUseCase(
    userRepo ports.UserRepository,
    eventPublisher ports.EventPublisher,
    userDomainService *services.UserDomainService,
    emailService ports.EmailService,
    idGenerator entities.IDGenerator,
    logger *slog.Logger,
) *UseCase {
    return &UseCase{
        userRepo:          userRepo,
        eventPublisher:    eventPublisher,
        userDomainService: userDomainService,
        emailService:      emailService,
        idGenerator:       idGenerator,
        logger:            logger,
    }
}

// Execute executes the create user use case
func (uc *UseCase) Execute(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Log use case execution
    uc.logger.InfoContext(ctx, "Executing create user use case",
        "email", req.Email,
        "name", req.Name,
    )
    
    // Validate request
    if err := req.Validate(); err != nil {
        uc.logger.WarnContext(ctx, "Invalid create user request", "error", err)
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Domain validation - check email uniqueness
    if err := uc.userDomainService.ValidateUniqueEmail(ctx, req.Email, nil); err != nil {
        uc.logger.WarnContext(ctx, "Email already exists", "email", req.Email, "error", err)
        return nil, err
    }
    
    // Create user entity
    user, err := entities.NewUser(req.Email, req.Name, uc.idGenerator)
    if err != nil {
        uc.logger.ErrorContext(ctx, "Failed to create user entity", "error", err)
        return nil, fmt.Errorf("failed to create user entity: %w", err)
    }
    
    // Additional domain validation
    if validationErrors := uc.userDomainService.ValidateUserData(user); len(validationErrors) > 0 {
        uc.logger.WarnContext(ctx, "User data validation failed", "errors", validationErrors)
        return nil, domainerrors.NewUserValidationError(validationErrors)
    }
    
    // Save user
    if err := uc.userRepo.Save(ctx, user); err != nil {
        uc.logger.ErrorContext(ctx, "Failed to save user", "error", err)
        return nil, fmt.Errorf("failed to save user: %w", err)
    }
    
    // Create and publish domain event
    userCreatedEvent := events.NewUserCreated(user)
    if err := uc.eventPublisher.Publish(ctx, userCreatedEvent); err != nil {
        // Log error but don't fail the use case
        // In production, consider using a retry mechanism or dead letter queue
        uc.logger.ErrorContext(ctx, "Failed to publish user created event", "error", err)
    }
    
    // Send welcome email (asynchronously)
    go func() {
        if err := uc.emailService.SendWelcomeEmail(context.Background(), user.Email(), user.Name()); err != nil {
            uc.logger.ErrorContext(context.Background(), "Failed to send welcome email", 
                "user_id", user.ID().Value(), 
                "email", user.Email(), 
                "error", err,
            )
        }
    }()
    
    uc.logger.InfoContext(ctx, "User created successfully", "user_id", user.ID().Value())
    
    return &CreateUserResponse{
        UserID: user.ID().Value(),
    }, nil
}
```

### 2. DTOs with Validation

```go
package createuser

import (
    "fmt"
    "net/mail"
    "strings"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=2,max=100"`
}

// Validate validates the create user request
func (r *CreateUserRequest) Validate() error {
    var errors []string
    
    // Validate email
    if r.Email == "" {
        errors = append(errors, "email is required")
    } else {
        if _, err := mail.ParseAddress(r.Email); err != nil {
            errors = append(errors, "email format is invalid")
        }
    }
    
    // Validate name
    r.Name = strings.TrimSpace(r.Name)
    if r.Name == "" {
        errors = append(errors, "name is required")
    } else {
        if len(r.Name) < 2 {
            errors = append(errors, "name must be at least 2 characters long")
        }
        if len(r.Name) > 100 {
            errors = append(errors, "name cannot exceed 100 characters")
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
    }
    
    return nil
}

// CreateUserResponse represents the response after creating a user
type CreateUserResponse struct {
    UserID string `json:"user_id"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
    UserID string `json:"user_id" validate:"required"`
    Email  string `json:"email,omitempty" validate:"omitempty,email"`
    Name   string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}

// Validate validates the update user request
func (r *UpdateUserRequest) Validate() error {
    var errors []string
    
    // Validate user ID
    if r.UserID == "" {
        errors = append(errors, "user_id is required")
    }
    
    // If email is provided, validate it
    if r.Email != "" {
        if _, err := mail.ParseAddress(r.Email); err != nil {
            errors = append(errors, "email format is invalid")
        }
    }
    
    // If name is provided, validate it
    if r.Name != "" {
        r.Name = strings.TrimSpace(r.Name)
        if len(r.Name) < 2 {
            errors = append(errors, "name must be at least 2 characters long")
        }
        if len(r.Name) > 100 {
            errors = append(errors, "name cannot exceed 100 characters")
        }
    }
    
    // At least one field must be provided for update
    if r.Email == "" && r.Name == "" {
        errors = append(errors, "at least one field (email or name) must be provided for update")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
    }
    
    return nil
}

// UpdateUserResponse represents the response after updating a user
type UpdateUserResponse struct {
    UserID  string `json:"user_id"`
    Updated bool   `json:"updated"`
}

// GetUserRequest represents the request to get a user
type GetUserRequest struct {
    UserID string `json:"user_id" validate:"required"`
}

// Validate validates the get user request
func (r *GetUserRequest) Validate() error {
    if r.UserID == "" {
        return fmt.Errorf("user_id is required")
    }
    return nil
}

// GetUserResponse represents the response when getting a user
type GetUserResponse struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

// NewGetUserResponse creates a GetUserResponse from a User entity
func NewGetUserResponse(user *entities.User) *GetUserResponse {
    return &GetUserResponse{
        UserID:    user.ID().Value(),
        Email:     user.Email(),
        Name:      user.Name(),
        CreatedAt: user.CreatedAt(),
    }
}
```

## Infrastructure Layer Guidelines

### 1. Configuration Management

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Config holds all application configuration
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Logger   LoggerConfig
    Email    EmailConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    Host         string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
    Host         string
    Port         int
    Password     string
    DB           int
    PoolSize     int
    MinIdleConns int
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
    Level  string
    Format string
}

// EmailConfig holds email service configuration
type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    SMTPUser     string
    SMTPPassword string
    FromEmail    string
    FromName     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    config := &Config{
        Server: ServerConfig{
            Port:         getEnvAsInt("SERVER_PORT", 8080),
            ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "10s"),
            WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "10s"),
            IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
            Host:         getEnv("SERVER_HOST", "0.0.0.0"),
        },
        Database: DatabaseConfig{
            Host:            getEnv("DB_HOST", "localhost"),
            Port:            getEnvAsInt("DB_PORT", 5432),
            User:            getEnv("DB_USER", "postgres"),
            Password:        getEnv("DB_PASSWORD", ""),
            DBName:          getEnv("DB_NAME", "myapp"),
            SSLMode:         getEnv("DB_SSL_MODE", "disable"),
            MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
        },
        Redis: RedisConfig{
            Host:         getEnv("REDIS_HOST", "localhost"),
            Port:         getEnvAsInt("REDIS_PORT", 6379),
            Password:     getEnv("REDIS_PASSWORD", ""),
            DB:           getEnvAsInt("REDIS_DB", 0),
            PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
            MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
        },
        Logger: LoggerConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),
        },
        Email: EmailConfig{
            SMTPHost:     getEnv("SMTP_HOST", "localhost"),
            SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
            SMTPUser:     getEnv("SMTP_USER", ""),
            SMTPPassword: getEnv("SMTP_PASSWORD", ""),
            FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),
            FromName:     getEnv("FROM_NAME", "MyApp"),
        },
    }
    
    // Validate required configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
    if c.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    if c.Database.User == "" {
        return fmt.Errorf("database user is required")
    }
    if c.Database.DBName == "" {
        return fmt.Errorf("database name is required")
    }
    return nil
}

// DatabaseURL returns the database connection URL
func (c *DatabaseConfig) DatabaseURL() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    if duration, err := time.ParseDuration(defaultValue); err == nil {
        return duration
    }
    return time.Second * 30
}
```

### 2. Repository Implementation

```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
    domainerrors "github.com/yourorg/yourproject/internal/domain/errors"
)

// UserRepository implements ports.UserRepository using PostgreSQL
type UserRepository struct {
    db *sqlx.DB
}

// UserModel represents the database model for users
type UserModel struct {
    ID        string    `db:"id"`
    Email     string    `db:"email"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Save saves a user to the database
func (r *UserRepository) Save(ctx context.Context, user *entities.User) error {
    model := &UserModel{
        ID:        user.ID().Value(),
        Email:     user.Email(),
        Name:      user.Name(),
        CreatedAt: user.CreatedAt(),
        UpdatedAt: time.Now(),
    }
    
    query := `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES (:id, :email, :name, :created_at, :updated_at)
        ON CONFLICT (id) DO UPDATE SET
            email = EXCLUDED.email,
            name = EXCLUDED.name,
            updated_at = EXCLUDED.updated_at`
    
    _, err := r.db.NamedExecContext(ctx, query, model)
    if err != nil {
        // Handle unique constraint violations
        if pqErr, ok := err.(*pq.Error); ok {
            switch pqErr.Code {
            case "23505": // unique_violation
                if pqErr.Constraint == "users_email_key" {
                    return domainerrors.NewUserAlreadyExistsError(user.Email())
                }
            }
        }
        return fmt.Errorf("failed to save user: %w", err)
    }
    
    return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, userID entities.UserID) (*entities.User, error) {
    var model UserModel
    query := "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1"
    
    err := r.db.GetContext(ctx, &model, query, userID.Value())
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domainerrors.NewUserNotFoundError(userID.Value())
        }
        return nil, fmt.Errorf("failed to find user by ID: %w", err)
    }
    
    return r.modelToEntity(&model)
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
    var model UserModel
    query := "SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1"
    
    err := r.db.GetContext(ctx, &model, query, email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domainerrors.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to find user by email: %w", err)
    }
    
    return r.modelToEntity(&model)
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, userID entities.UserID) error {
    query := "DELETE FROM users WHERE id = $1"
    
    result, err := r.db.ExecContext(ctx, query, userID.Value())
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return domainerrors.NewUserNotFoundError(userID.Value())
    }
    
    return nil
}

// List returns a paginated list of users
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
    var models []UserModel
    query := `
        SELECT id, email, name, created_at, updated_at 
        FROM users 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2`
    
    err := r.db.SelectContext(ctx, &models, query, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to list users: %w", err)
    }
    
    users := make([]*entities.User, len(models))
    for i, model := range models {
        user, err := r.modelToEntity(&model)
        if err != nil {
            return nil, fmt.Errorf("failed to convert model to entity: %w", err)
        }
        users[i] = user
    }
    
    return users, nil
}

// modelToEntity converts a database model to a domain entity
func (r *UserRepository) modelToEntity(model *UserModel) (*entities.User, error) {
    userID, err := entities.NewUserID(model.ID)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID from database: %w", err)
    }
    
    // Create user entity directly (bypassing NewUser validation since data is from DB)
    user := &entities.User{}
    user.SetID(userID)
    user.SetEmail(model.Email)
    user.SetName(model.Name)
    user.SetCreatedAt(model.CreatedAt)
    
    return user, nil
}

// DatabaseManager manages database connections and migrations
type DatabaseManager struct {
    db *sqlx.DB
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(databaseURL string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*DatabaseManager, error) {
    db, err := sqlx.Connect("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(maxOpenConns)
    db.SetMaxIdleConns(maxIdleConns)
    db.SetConnMaxLifetime(connMaxLifetime)
    
    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &DatabaseManager{db: db}, nil
}

// DB returns the database instance
func (dm *DatabaseManager) DB() *sqlx.DB {
    return dm.db
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
    return dm.db.Close()
}

// Migrate runs database migrations
func (dm *DatabaseManager) Migrate() error {
    // This is a simplified example. In production, use a proper migration tool like golang-migrate
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );
    
    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
    CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
    `
    
    _, err := dm.db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    return nil
}
```

### 3. Event Publishing

```go
package messaging

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "sync"
    "time"
    
    "github.com/yourorg/yourproject/internal/domain/events"
)

// EventPublisher implements ports.EventPublisher with reliability features
type EventPublisher struct {
    maxRetries   int
    retryDelay   time.Duration
    failedEvents []events.DomainEvent
    mu           sync.RWMutex
    logger       *slog.Logger
    
    // In production, these would be actual message broker connections
    // For example: Kafka producer, RabbitMQ channel, etc.
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(maxRetries int, retryDelay time.Duration, logger *slog.Logger) *EventPublisher {
    return &EventPublisher{
        maxRetries:   maxRetries,
        retryDelay:   retryDelay,
        failedEvents: make([]events.DomainEvent, 0),
        logger:       logger,
    }
}

// Publish publishes a single domain event with retry logic
func (p *EventPublisher) Publish(ctx context.Context, event events.DomainEvent) error {
    for attempt := 0; attempt <= p.maxRetries; attempt++ {
        if err := p.doPublish(ctx, event); err != nil {
            if attempt == p.maxRetries {
                // Final attempt failed, add to failed events
                p.addFailedEvent(event)
                p.logger.ErrorContext(ctx, "Event publishing failed after all retries",
                    "event_type", event.EventType(),
                    "event_id", event.EventID(),
                    "aggregate_id", event.AggregateID(),
                    "attempts", attempt+1,
                    "error", err,
                )
                return fmt.Errorf("failed to publish event after %d attempts: %w", p.maxRetries+1, err)
            }
            
            p.logger.WarnContext(ctx, "Event publishing failed, retrying",
                "event_type", event.EventType(),
                "event_id", event.EventID(),
                "attempt", attempt+1,
                "retry_in", p.retryDelay,
                "error", err,
            )
            
            select {
            case <-time.After(p.retryDelay):
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        
        p.logger.InfoContext(ctx, "Event published successfully",
            "event_type", event.EventType(),
            "event_id", event.EventID(),
            "aggregate_id", event.AggregateID(),
            "attempt", attempt+1,
        )
        return nil
    }
    
    return nil
}

// PublishBatch publishes multiple events atomically
func (p *EventPublisher) PublishBatch(ctx context.Context, events []events.DomainEvent) error {
    p.logger.InfoContext(ctx, "Publishing event batch", "count", len(events))
    
    // In a real implementation, this would use transactional publishing
    for _, event := range events {
        if err := p.doPublish(ctx, event); err != nil {
            // If any event fails, add all to failed list
            p.addFailedEvents(events)
            p.logger.ErrorContext(ctx, "Batch event publishing failed",
                "total_events", len(events),
                "failed_event_type", event.EventType(),
                "error", err,
            )
            return fmt.Errorf("failed to publish event batch: %w", err)
        }
    }
    
    p.logger.InfoContext(ctx, "Event batch published successfully", "count", len(events))
    return nil
}

// doPublish actually publishes the event to the message broker
func (p *EventPublisher) doPublish(ctx context.Context, event events.DomainEvent) error {
    // Convert event to JSON
    eventData, err := event.ToJSON()
    if err != nil {
        return fmt.Errorf("failed to serialize event: %w", err)
    }
    
    // In a real implementation, this is where you'd:
    // 1. Publish to Kafka
    // 2. Send to RabbitMQ
    // 3. Write to event store
    // 4. Send to cloud messaging service (AWS SNS, Google Pub/Sub, etc.)
    
    // For demonstration, we'll just log the event
    p.logger.InfoContext(ctx, "Publishing event to message broker",
        "event_type", event.EventType(),
        "event_id", event.EventID(),
        "payload_size", len(eventData),
    )
    
    // Simulate potential publishing failure for testing
    // if rand.Float32() < 0.1 { // 10% failure rate
    //     return fmt.Errorf("simulated publishing failure") 
    // }
    
    return nil
}

// addFailedEvent adds an event to the failed events list
func (p *EventPublisher) addFailedEvent(event events.DomainEvent) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.failedEvents = append(p.failedEvents, event)
}

// addFailedEvents adds multiple events to the failed events list
func (p *EventPublisher) addFailedEvents(events []events.DomainEvent) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.failedEvents = append(p.failedEvents, events...)
}

// RetryFailedEvents attempts to republish failed events
func (p *EventPublisher) RetryFailedEvents(ctx context.Context) error {
    p.mu.Lock()
    failedCopy := make([]events.DomainEvent, len(p.failedEvents))
    copy(failedCopy, p.failedEvents)
    p.failedEvents = p.failedEvents[:0] // Clear the slice
    p.mu.Unlock()
    
    if len(failedCopy) == 0 {
        return nil
    }
    
    p.logger.InfoContext(ctx, "Retrying failed events", "count", len(failedCopy))
    
    var stillFailed []events.DomainEvent
    for _, event := range failedCopy {
        if err := p.doPublish(ctx, event); err != nil {
            stillFailed = append(stillFailed, event)
            p.logger.WarnContext(ctx, "Failed event retry unsuccessful",
                "event_type", event.EventType(),
                "event_id", event.EventID(),
                "error", err,
            )
        } else {
            p.logger.InfoContext(ctx, "Failed event retry successful",
                "event_type", event.EventType(),
                "event_id", event.EventID(),
            )
        }
    }
    
    // Add still failed events back to the list
    if len(stillFailed) > 0 {
        p.addFailedEvents(stillFailed)
        return fmt.Errorf("failed to retry %d out of %d events", len(stillFailed), len(failedCopy))
    }
    
    return nil
}

// GetFailedEventsCount returns the number of failed events
func (p *EventPublisher) GetFailedEventsCount() int {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return len(p.failedEvents)
}

// GetFailedEvents returns a copy of failed events for inspection
func (p *EventPublisher) GetFailedEvents() []events.DomainEvent {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    failedCopy := make([]events.DomainEvent, len(p.failedEvents))
    copy(failedCopy, p.failedEvents)
    return failedCopy
}

// Close gracefully closes the event publisher
func (p *EventPublisher) Close(ctx context.Context) error {
    p.logger.InfoContext(ctx, "Closing event publisher")
    
    // Attempt to publish any remaining failed events
    if err := p.RetryFailedEvents(ctx); err != nil {
        p.logger.WarnContext(ctx, "Some events could not be published during shutdown", "error", err)
    }
    
    // In a real implementation, close connections to message brokers
    
    return nil
}
```

## Presentation Layer Guidelines

### 1. HTTP Handlers

```go
package handlers

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "log/slog"
    
    "github.com/gorilla/mux"
    
    createuser "github.com/yourorg/yourproject/internal/usecases/create_user"
    updateuser "github.com/yourorg/yourproject/internal/usecases/update_user" 
    getuser "github.com/yourorg/yourproject/internal/usecases/get_user"
    domainerrors "github.com/yourorg/yourproject/internal/domain/errors"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
    createUserUseCase *createuser.UseCase
    updateUserUseCase *updateuser.UseCase
    getUserUseCase    *getuser.UseCase
    logger            *slog.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
    createUserUseCase *createuser.UseCase,
    updateUserUseCase *updateuser.UseCase, 
    getUserUseCase *getuser.UseCase,
    logger *slog.Logger,
) *UserHandler {
    return &UserHandler{
        createUserUseCase: createUserUseCase,
        updateUserUseCase: updateUserUseCase,
        getUserUseCase:    getUserUseCase,
        logger:            logger,
    }
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req createuser.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format", err)
        return
    }
    
    resp, err := h.createUserUseCase.Execute(ctx, &req)
    if err != nil {
        h.handleUseCaseError(w, r, err)
        return
    }
    
    h.respondWithJSON(w, http.StatusCreated, resp)
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    vars := mux.Vars(r)
    userID := vars["id"]
    
    req := &getuser.GetUserRequest{UserID: userID}
    
    resp, err := h.getUserUseCase.Execute(ctx, req)
    if err != nil {
        h.handleUseCaseError(w, r, err)
        return
    }
    
    h.respondWithJSON(w, http.StatusOK, resp)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    vars := mux.Vars(r)
    userID := vars["id"]
    
    var req updateuser.UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format", err)
        return
    }
    
    req.UserID = userID // Set the user ID from the URL
    
    resp, err := h.updateUserUseCase.Execute(ctx, &req)
    if err != nil {
        h.handleUseCaseError(w, r, err)
        return
    }
    
    h.respondWithJSON(w, http.StatusOK, resp)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    
    limit := 20 // default
    if limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }
    
    offset := 0 // default
    if offsetStr != "" {
        if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        }
    }
    
    // Limit the maximum number of results
    if limit > 100 {
        limit = 100
    }
    
    req := &listusers.ListUsersRequest{
        Limit:  limit,
        Offset: offset,
    }
    
    resp, err := h.listUsersUseCase.Execute(ctx, req)
    if err != nil {
        h.handleUseCaseError(w, r, err)
        return
    }
    
    h.respondWithJSON(w, http.StatusOK, resp)
}

// handleUseCaseError handles errors from use cases and maps them to HTTP responses
func (h *UserHandler) handleUseCaseError(w http.ResponseWriter, r *http.Request, err error) {
    ctx := r.Context()
    
    var domainErr *domainerrors.DomainError
    if errors.As(err, &domainErr) {
        statusCode := h.getHTTPStatusForDomainError(domainErr)
        
        h.logger.WarnContext(ctx, "Domain error in request",
            "error_code", domainErr.Code,
            "message", domainErr.Message,
            "path", r.URL.Path,
            "method", r.Method,
        )
        
        h.respondWithError(w, statusCode, domainErr.Code, domainErr.Message, domainErr)
        return
    }
    
    // Handle validation errors
    if err.Error() == "validation errors" || 
       err.Error() == "invalid request" {
        h.logger.WarnContext(ctx, "Validation error in request",
            "error", err.Error(),
            "path", r.URL.Path,
            "method", r.Method,
        )
        
        h.respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), err)
        return
    }
    
    // Log unexpected errors
    h.logger.ErrorContext(ctx, "Unexpected error in request",
        "error", err.Error(),
        "path", r.URL.Path,
        "method", r.Method,
    )
    
    h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred", err)
}

// getHTTPStatusForDomainError maps domain error codes to HTTP status codes
func (h *UserHandler) getHTTPStatusForDomainError(err *domainerrors.DomainError) int {
    switch err.Code {
    case "USER_NOT_FOUND":
        return http.StatusNotFound
    case "USER_ALREADY_EXISTS":
        return http.StatusConflict
    case "USER_VALIDATION_FAILED", "INVALID_USER_DATA":
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
    ErrorCode string `json:"error_code"`
    Message   string `json:"message"`
    Details   string `json:"details,omitempty"`
}

// respondWithError sends an error response
func (h *UserHandler) respondWithError(w http.ResponseWriter, statusCode int, errorCode, message string, err error) {
    response := ErrorResponse{
        ErrorCode: errorCode,
        Message:   message,
    }
    
    if err != nil {
        response.Details = err.Error()
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// respondWithJSON sends a JSON response
func (h *UserHandler) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        h.logger.Error("Failed to encode JSON response", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}

// HealthCheck handles GET /health
func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "service":   "golang-hexagonal-architecture",
        "timestamp": time.Now().UTC(),
    }
    
    h.respondWithJSON(w, http.StatusOK, health)
}
```

### 2. HTTP Router Setup

```go
package http

import (
    "context"
    "net/http"
    "time"
    "log/slog"
    
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    
    "github.com/yourorg/yourproject/internal/presentation/http/handlers"
    "github.com/yourorg/yourproject/internal/presentation/http/middleware"
)

// Server represents the HTTP server
type Server struct {
    server      *http.Server
    userHandler *handlers.UserHandler
    logger      *slog.Logger
}

// NewServer creates a new HTTP server
func NewServer(
    addr string,
    userHandler *handlers.UserHandler,
    logger *slog.Logger,
    readTimeout, writeTimeout, idleTimeout time.Duration,
) *Server {
    router := mux.NewRouter()
    
    // Setup middleware
    router.Use(middleware.RequestID)
    router.Use(middleware.Logging(logger))
    router.Use(middleware.Recovery(logger))
    
    // Setup CORS
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"}, // Configure appropriately for production
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        AllowCredentials: true,
    })
    
    handler := c.Handler(router)
    
    // Setup routes
    setupRoutes(router, userHandler)
    
    server := &http.Server{
        Addr:         addr,
        Handler:      handler,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
        IdleTimeout:  idleTimeout,
    }
    
    return &Server{
        server:      server,
        userHandler: userHandler,
        logger:      logger,
    }
}

// setupRoutes configures all the routes
func setupRoutes(router *mux.Router, userHandler *handlers.UserHandler) {
    // Health check
    router.HandleFunc("/health", userHandler.HealthCheck).Methods("GET")
    
    // API routes
    api := router.PathPrefix("/api/v1").Subrouter()
    
    // User routes
    users := api.PathPrefix("/users").Subrouter()
    users.HandleFunc("", userHandler.CreateUser).Methods("POST")
    users.HandleFunc("", userHandler.ListUsers).Methods("GET")
    users.HandleFunc("/{id}", userHandler.GetUser).Methods("GET")
    users.HandleFunc("/{id}", userHandler.UpdateUser).Methods("PUT")
    users.HandleFunc("/{id}", userHandler.DeleteUser).Methods("DELETE")
}

// Start starts the HTTP server
func (s *Server) Start() error {
    s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
    return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("Shutting down HTTP server")
    return s.server.Shutdown(ctx)
}
```

### 3. HTTP Middleware

```go
package middleware

import (
    "context"
    "net/http"
    "runtime/debug"
    "time"
    "log/slog"
    
    "github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Add to response headers
        w.Header().Set("X-Request-ID", requestID)
        
        // Add to request context
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Logging logs HTTP requests and responses
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap the response writer to capture status code
            wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            // Get request ID from context
            requestID, _ := r.Context().Value("request_id").(string)
            
            // Log request
            logger.InfoContext(r.Context(), "HTTP request started",
                "method", r.Method,
                "path", r.URL.Path,
                "query", r.URL.RawQuery,
                "remote_addr", r.RemoteAddr,
                "user_agent", r.UserAgent(),
                "request_id", requestID,
            )
            
            next.ServeHTTP(wrapped, r)
            
            duration := time.Since(start)
            
            // Log response
            logger.InfoContext(r.Context(), "HTTP request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "status_code", wrapped.statusCode,
                "duration_ms", duration.Milliseconds(),
                "request_id", requestID,
            )
        })
    }
}

// Recovery recovers from panics and logs them
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    requestID, _ := r.Context().Value("request_id").(string)
                    
                    logger.ErrorContext(r.Context(), "HTTP request panic recovered",
                        "error", err,
                        "method", r.Method,
                        "path", r.URL.Path,
                        "request_id", requestID,
                        "stack", string(debug.Stack()),
                    )
                    
                    http.Error(w, "Internal server error", http.StatusInternalServerError)
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

## Dependency Injection and Application Setup

### 1. Dependency Injection Container

```go
package container

import (
    "fmt"
    "log/slog"
    "os"
    "time"
    
    "github.com/jmoiron/sqlx"
    
    "github.com/yourorg/yourproject/config"
    "github.com/yourorg/yourproject/internal/domain/entities"
    "github.com/yourorg/yourproject/internal/domain/ports"
    "github.com/yourorg/yourproject/internal/domain/services"
    "github.com/yourorg/yourproject/internal/infrastructure/messaging"
    "github.com/yourorg/yourproject/internal/infrastructure/persistence/postgres"
    "github.com/yourorg/yourproject/internal/presentation/http/handlers"
    createuser "github.com/yourorg/yourproject/internal/usecases/create_user"
    getuser "github.com/yourorg/yourproject/internal/usecases/get_user"
    updateuser "github.com/yourorg/yourproject/internal/usecases/update_user"
)

// Container holds all application dependencies
type Container struct {
    Config *config.Config
    Logger *slog.Logger
    
    // Infrastructure
    DatabaseManager *postgres.DatabaseManager
    EventPublisher  ports.EventPublisher
    
    // Repositories
    UserRepository ports.UserRepository
    
    // Domain Services
    UserDomainService *services.UserDomainService
    
    // Use Cases
    CreateUserUseCase *createuser.UseCase
    GetUserUseCase    *getuser.UseCase
    UpdateUserUseCase *updateuser.UseCase
    
    // Handlers
    UserHandler *handlers.UserHandler
    
    // Utilities
    IDGenerator entities.IDGenerator
}

// NewContainer creates and configures a new dependency injection container
func NewContainer() (*Container, error) {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // Setup logger
    logger := setupLogger(cfg.Logger)
    
    // Setup database
    dbManager, err := postgres.NewDatabaseManager(
        cfg.Database.DatabaseURL(),
        cfg.Database.MaxOpenConns,
        cfg.Database.MaxIdleConns,
        cfg.Database.ConnMaxLifetime,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create database manager: %w", err)
    }
    
    // Run migrations
    if err := dbManager.Migrate(); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }
    
    // Setup repositories
    userRepo := postgres.NewUserRepository(dbManager.DB())
    
    // Setup domain services
    userDomainService := services.NewUserDomainService(userRepo)
    
    // Setup infrastructure services
    eventPublisher := messaging.NewEventPublisher(3, time.Second*2, logger)
    idGenerator := &entities.DefaultIDGenerator{}
    
    // Setup use cases
    createUserUseCase := createuser.NewUseCase(
        userRepo,
        eventPublisher,
        userDomainService,
        nil, // email service - implement as needed
        idGenerator,
        logger,
    )
    
    getUserUseCase := getuser.NewUseCase(userRepo, logger)
    updateUserUseCase := updateuser.NewUseCase(userRepo, eventPublisher, userDomainService, logger)
    
    // Setup handlers
    userHandler := handlers.NewUserHandler(
        createUserUseCase,
        updateUserUseCase,
        getUserUseCase,
        logger,
    )
    
    return &Container{
        Config:            cfg,
        Logger:            logger,
        DatabaseManager:   dbManager,
        EventPublisher:    eventPublisher,
        UserRepository:    userRepo,
        UserDomainService: userDomainService,
        CreateUserUseCase: createUserUseCase,
        GetUserUseCase:    getUserUseCase,
        UpdateUserUseCase: updateUserUseCase,
        UserHandler:       userHandler,
        IDGenerator:       idGenerator,
    }, nil
}

// Close closes all resources
func (c *Container) Close() error {
    var errors []error
    
    // Close database
    if c.DatabaseManager != nil {
        if err := c.DatabaseManager.Close(); err != nil {
            errors = append(errors, fmt.Errorf("failed to close database: %w", err))
        }
    }
    
    // Close event publisher
    if publisher, ok := c.EventPublisher.(*messaging.EventPublisher); ok {
        if err := publisher.Close(context.Background()); err != nil {
            errors = append(errors, fmt.Errorf("failed to close event publisher: %w", err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("errors during cleanup: %v", errors)
    }
    
    return nil
}

// setupLogger configures structured logging
func setupLogger(config config.LoggerConfig) *slog.Logger {
    var handler slog.Handler
    
    opts := &slog.HandlerOptions{
        Level: parseLogLevel(config.Level),
    }
    
    switch config.Format {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, opts)
    default:
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return slog.New(handler)
}

// parseLogLevel parses string log level to slog.Level
func parseLogLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn", "warning":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

### 2. Main Application

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yourorg/yourproject/internal/container"
    httpserver "github.com/yourorg/yourproject/internal/presentation/http"
)

func main() {
    // Create dependency injection container
    container, err := container.NewContainer()
    if err != nil {
        log.Fatalf("Failed to create container: %v", err)
    }
    defer func() {
        if err := container.Close(); err != nil {
            container.Logger.Error("Error closing container", "error", err)
        }
    }()
    
    // Create HTTP server
    server := httpserver.NewServer(
        fmt.Sprintf(":%d", container.Config.Server.Port),
        container.UserHandler,
        container.Logger,
        container.Config.Server.ReadTimeout,
        container.Config.Server.WriteTimeout,
        container.Config.Server.IdleTimeout,
    )
    
    // Start server in a goroutine
    go func() {
        container.Logger.Info("Starting HTTP server", "port", container.Config.Server.Port)
        if err := server.Start(); err != nil && err != http.ErrServerClosed {
            container.Logger.Error("Server failed to start", "error", err)
            os.Exit(1)
        }
    }()
    
    // Start background tasks
    go startBackgroundTasks(container)
    
    // Wait for interrupt signal to gracefully shutdown the server
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    container.Logger.Info("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        container.Logger.Error("Server forced to shutdown", "error", err)
        os.Exit(1)
    }
    
    container.Logger.Info("Server shutdown complete")
}

// startBackgroundTasks starts background tasks like retrying failed events
func startBackgroundTasks(container *container.Container) {
    // Retry failed events every 5 minutes
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if publisher, ok := container.EventPublisher.(*messaging.EventPublisher); ok {
                ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
                if err := publisher.RetryFailedEvents(ctx); err != nil {
                    container.Logger.Error("Failed to retry failed events", "error", err)
                }
                cancel()
            }
        }
    }
}
```

## Testing Guidelines

### 1. Unit Testing

```go
package createuser_test

import (
    "context"
    "errors"
    "testing"
    "log/slog"
    "os"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    
    "github.com/yourorg/yourproject/internal/domain/entities"
    "github.com/yourorg/yourproject/internal/domain/events"
    domainerrors "github.com/yourorg/yourproject/internal/domain/errors"
    createuser "github.com/yourorg/yourproject/internal/usecases/create_user"
)

// Mock implementations
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, user *entities.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID entities.UserID) (*entities.User, error) {
    args := m.Called(ctx, userID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, userID entities.UserID) error {
    args := m.Called(ctx, userID)
    return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
    args := m.Called(ctx, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*entities.User), args.Error(1)
}

type MockEventPublisher struct {
    mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event events.DomainEvent) error {
    args := m.Called(ctx, event)
    return args.Error(0)
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []events.DomainEvent) error {
    args := m.Called(ctx, events)
    return args.Error(0)
}

type MockUserDomainService struct {
    mock.Mock
}

func (m *MockUserDomainService) ValidateUniqueEmail(ctx context.Context, email string, excludeUserID *entities.UserID) error {
    args := m.Called(ctx, email, excludeUserID)
    return args.Error(0)
}

func (m *MockUserDomainService) ValidateUserData(user *entities.User) []string {
    args := m.Called(user)
    return args.Get(0).([]string)
}

type MockIDGenerator struct {
    mock.Mock
}

func (m *MockIDGenerator) GenerateUserID() (entities.UserID, error) {
    args := m.Called()
    return args.Get(0).(entities.UserID), args.Error(1)
}

// Test suite
func TestCreateUserUseCase(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
    
    t.Run("Execute_Success", func(t *testing.T) {
        // Arrange
        mockRepo := new(MockUserRepository)
        mockPublisher := new(MockEventPublisher)
        mockDomainService := new(MockUserDomainService)
        mockIDGenerator := new(MockIDGenerator)
        
        useCase := createuser.NewUseCase(
            mockRepo,
            mockPublisher,
            mockDomainService,
            nil, // email service
            mockIDGenerator,
            logger,
        )
        
        ctx := context.Background()
        req := &createuser.CreateUserRequest{
            Email: "test@example.com",
            Name:  "Test User",
        }
        
        userID, _ := entities.NewUserID("user_12345678")
        
        // Mock expectations
        mockDomainService.On("ValidateUniqueEmail", ctx, req.Email, (*entities.UserID)(nil)).Return(nil)
        mockIDGenerator.On("GenerateUserID").Return(userID, nil)
        mockDomainService.On("ValidateUserData", mock.AnythingOfType("*entities.User")).Return([]string{})
        mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.User")).Return(nil)
        mockPublisher.On("Publish", ctx, mock.AnythingOfType("*events.UserCreated")).Return(nil)
        
        // Act
        resp, err := useCase.Execute(ctx, req)
        
        // Assert
        require.NoError(t, err)
        assert.NotNil(t, resp)
        assert.Equal(t, "user_12345678", resp.UserID)
        
        mockRepo.AssertExpectations(t)
        mockPublisher.AssertExpectations(t)
        mockDomainService.AssertExpectations(t)
        mockIDGenerator.AssertExpectations(t)
    })
    
    t.Run("Execute_UserAlreadyExists", func(t *testing.T) {
        // Arrange
        mockRepo := new(MockUserRepository)
        mockPublisher := new(MockEventPublisher)
        mockDomainService := new(MockUserDomainService)
        mockIDGenerator := new(MockIDGenerator)
        
        useCase := createuser.NewUseCase(
            mockRepo,
            mockPublisher,
            mockDomainService,
            nil,
            mockIDGenerator,
            logger,
        )
        
        ctx := context.Background()
        req := &createuser.CreateUserRequest{
            Email: "existing@example.com",
            Name:  "Test User",
        }
        
        // Mock expectations
        mockDomainService.On("ValidateUniqueEmail", ctx, req.Email, (*entities.UserID)(nil)).
            Return(domainerrors.NewUserAlreadyExistsError(req.Email))
        
        // Act
        resp, err := useCase.Execute(ctx, req)
        
        // Assert
        assert.Error(t, err)
        assert.Nil(t, resp)
        
        var domainErr *domainerrors.DomainError
        assert.True(t, errors.As(err, &domainErr))
        assert.Equal(t, "USER_ALREADY_EXISTS", domainErr.Code)
        
        mockDomainService.AssertExpectations(t)
    })
    
    t.Run("Execute_InvalidRequest", func(t *testing.T) {
        // Arrange
        mockRepo := new(MockUserRepository)
        mockPublisher := new(MockEventPublisher)
        mockDomainService := new(MockUserDomainService)
        mockIDGenerator := new(MockIDGenerator)
        
        useCase := createuser.NewUseCase(
            mockRepo,
            mockPublisher,
            mockDomainService,
            nil,
            mockIDGenerator,
            logger,
        )
        
        ctx := context.Background()
        req := &createuser.CreateUserRequest{
            Email: "invalid-email",
            Name:  "",
        }
        
        // Act
        resp, err := useCase.Execute(ctx, req)
        
        // Assert
        assert.Error(t, err)
        assert.Nil(t, resp)
        assert.Contains(t, err.Error(), "validation errors")
    })
    
    t.Run("Execute_ValidationErrors", func(t *testing.T) {
        // Arrange
        mockRepo := new(MockUserRepository)
        mockPublisher := new(MockEventPublisher)
        mockDomainService := new(MockUserDomainService)
        mockIDGenerator := new(MockIDGenerator)
        
        useCase := createuser.NewUseCase(
            mockRepo,
            mockPublisher,
            mockDomainService,
            nil,
            mockIDGenerator,
            logger,
        )
        
        ctx := context.Background()
        req := &createuser.CreateUserRequest{
            Email: "test@example.com",
            Name:  "Test User",
        }
        
        userID, _ := entities.NewUserID("user_12345678")
        validationErrors := []string{"name contains inappropriate content"}
        
        // Mock expectations
        mockDomainService.On("ValidateUniqueEmail", ctx, req.Email, (*entities.UserID)(nil)).Return(nil)
        mockIDGenerator.On("GenerateUserID").Return(userID, nil)
        mockDomainService.On("ValidateUserData", mock.AnythingOfType("*entities.User")).Return(validationErrors)
        
        // Act
        resp, err := useCase.Execute(ctx, req)
        
        // Assert
        assert.Error(t, err)
        assert.Nil(t, resp)
        
        var domainErr *domainerrors.DomainError
        assert.True(t, errors.As(err, &domainErr))
        assert.Equal(t, "USER_VALIDATION_FAILED", domainErr.Code)
        
        mockDomainService.AssertExpectations(t)
        mockIDGenerator.AssertExpectations(t)
    })
}

// Table-driven tests for entity validation
func TestUser_Create(t *testing.T) {
    idGen := &entities.DefaultIDGenerator{}
    
    tests := []struct {
        name        string
        email       string
        username    string
        expectError bool
        errorMsg    string
    }{
        {
            name:        "Valid user",
            email:       "test@example.com",
            username:    "Test User",
            expectError: false,
        },
        {
            name:        "Invalid email",
            email:       "invalid-email",
            username:    "Test User", 
            expectError: true,
            errorMsg:    "invalid email format",
        },
        {
            name:        "Empty name",
            email:       "test@example.com",
            username:    "",
            expectError: true,
            errorMsg:    "name cannot be empty",
        },
        {
            name:        "Name too long",
            email:       "test@example.com",
            username:    string(make([]byte, 101)), // 101 characters
            expectError: true,
            errorMsg:    "name cannot exceed 100 characters",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := entities.NewUser(tt.email, tt.username, idGen)
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.email, user.Email())
                assert.Equal(t, tt.username, user.Name())
                assert.True(t, user.ID().Value() != "")
            }
        })
    }
}
```

### 2. Integration Testing

```go
package integration_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "log/slog"
    "os"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
    
    "github.com/yourorg/yourproject/internal/container"
    httpserver "github.com/yourorg/yourproject/internal/presentation/http"
    createuser "github.com/yourorg/yourproject/internal/usecases/create_user"
)

type IntegrationTestSuite struct {
    suite.Suite
    container *container.Container
    server    *httpserver.Server
    testDB    *sqlx.DB
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Setup test database (you might want to use a different database for tests)
    os.Setenv("DB_NAME", "test_db")
    os.Setenv("LOG_LEVEL", "error")
    
    // Create container with test configuration
    container, err := container.NewContainer()
    require.NoError(suite.T(), err)
    
    suite.container = container
    
    // Create test server
    suite.server = httpserver.NewServer(
        ":0", // Use random port for testing
        container.UserHandler,
        container.Logger,
        10*time.Second,
        10*time.Second,
        60*time.Second,
    )
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    if suite.container != nil {
        suite.container.Close()
    }
}

func (suite *IntegrationTestSuite) SetupTest() {
    // Clean up database before each test
    // In a real implementation, you might truncate tables or use database transactions
}

func (suite *IntegrationTestSuite) TestCreateUser_Success() {
    // Arrange
    requestBody := createuser.CreateUserRequest{
        Email: "integration@example.com",
        Name:  "Integration Test User",
    }
    
    jsonBody, err := json.Marshal(requestBody)
    require.NoError(suite.T(), err)
    
    req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    
    // Act
    suite.server.Handler.ServeHTTP(rr, req)
    
    // Assert
    assert.Equal(suite.T(), http.StatusCreated, rr.Code)
    
    var response createuser.CreateUserResponse
    err = json.Unmarshal(rr.Body.Bytes(), &response)
    require.NoError(suite.T(), err)
    
    assert.NotEmpty(suite.T(), response.UserID)
    assert.True(suite.T(), strings.HasPrefix(response.UserID, "user_"))
}

func (suite *IntegrationTestSuite) TestCreateUser_DuplicateEmail() {
    // Arrange - create first user
    requestBody := createuser.CreateUserRequest{
        Email: "duplicate@example.com",
        Name:  "First User",
    }
    
    jsonBody, err := json.Marshal(requestBody)
    require.NoError(suite.T(), err)
    
    req1 := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
    req1.Header.Set("Content-Type", "application/json")
    rr1 := httptest.NewRecorder()
    
    // Create first user
    suite.server.Handler.ServeHTTP(rr1, req1)
    assert.Equal(suite.T(), http.StatusCreated, rr1.Code)
    
    // Try to create duplicate
    duplicateBody := createuser.CreateUserRequest{
        Email: "duplicate@example.com",
        Name:  "Second User",
    }
    
    jsonBody2, err := json.Marshal(duplicateBody)
    require.NoError(suite.T(), err)
    
    req2 := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody2))
    req2.Header.Set("Content-Type", "application/json")
    rr2 := httptest.NewRecorder()
    
    // Act
    suite.server.Handler.ServeHTTP(rr2, req2)
    
    // Assert
    assert.Equal(suite.T(), http.StatusConflict, rr2.Code)
    
    var errorResponse map[string]interface{}
    err = json.Unmarshal(rr2.Body.Bytes(), &errorResponse)
    require.NoError(suite.T(), err)
    
    assert.Equal(suite.T(), "USER_ALREADY_EXISTS", errorResponse["error_code"])
}

func (suite *IntegrationTestSuite) TestCreateUser_InvalidData() {
    // Arrange
    requestBody := createuser.CreateUserRequest{
        Email: "invalid-email",
        Name:  "",
    }
    
    jsonBody, err := json.Marshal(requestBody)
    require.NoError(suite.T(), err)
    
    req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    
    // Act
    suite.server.Handler.ServeHTTP(rr, req)
    
    // Assert
    assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
    
    var errorResponse map[string]interface{}
    err = json.Unmarshal(rr.Body.Bytes(), &errorResponse)
    require.NoError(suite.T(), err)
    
    assert.Contains(suite.T(), errorResponse["error_code"], "VALIDATION_ERROR")
}

func (suite *IntegrationTestSuite) TestHealthCheck() {
    // Arrange
    req := httptest.NewRequest("GET", "/health", nil)
    rr := httptest.NewRecorder()
    
    // Act
    suite.server.Handler.ServeHTTP(rr, req)
    
    // Assert
    assert.Equal(suite.T(), http.StatusOK, rr.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    require.NoError(suite.T(), err)
    
    assert.Equal(suite.T(), "healthy", response["status"])
}

func TestIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}
```

## Project Structure

```
project/
├── cmd/                              # Application entry points
│   └── server/                       # Main server application
│       └── main.go                   # Application bootstrap
├── internal/                         # Private application code
│   ├── domain/                       # Domain layer (business logic)
│   │   ├── entities/                 # Domain entities
│   │   │   ├── user.go               # User entity with business rules
│   │   │   └── base.go               # Base entity interfaces/types
│   │   ├── events/                   # Domain events
│   │   │   ├── user_events.go        # User-related events
│   │   │   └── base.go               # Base event interfaces
│   │   ├── ports/                    # Interfaces/contracts
│   │   │   ├── user_repository.go    # User repository interface
│   │   │   ├── event_publisher.go    # Event publisher interface
│   │   │   └── email_service.go      # Email service interface
│   │   ├── services/                 # Domain services
│   │   │   └── user_domain_service.go # Complex user business rules
│   │   └── errors/                   # Domain errors
│   │       └── domain_errors.go      # Domain-specific exceptions
│   ├── usecases/                     # Application layer
│   │   ├── create_user/              # Create user use case
│   │   │   ├── create_user.go        # Use case implementation
│   │   │   └── dtos.go               # Request/response DTOs
│   │   ├── get_user/                 # Get user use case
│   │   │   ├── get_user.go
│   │   │   └── dtos.go
│   │   └── update_user/              # Update user use case
│   │       ├── update_user.go
│   │       └── dtos.go
│   ├── infrastructure/               # Infrastructure layer
│   │   ├── persistence/              # Data persistence
│   │   │   ├── memory/               # In-memory implementations
│   │   │   │   └── user_repository.go
│   │   │   └── postgres/             # PostgreSQL implementations
│   │   │       ├── user_repository.go
│   │   │       └── database_manager.go
│   │   ├── messaging/                # Event publishing
│   │   │   ├── event_publisher.go    # Reliable event publisher
│   │   │   └── kafka/                # Kafka-specific implementation
│   │   │       └── kafka_publisher.go
│   │   ├── cache/                    # Caching implementations
│   │   │   └── redis/
│   │   │       └── redis_cache.go
│   │   └── externalapis/             # External API clients
│   │       └── email/
│   │           └── smtp_service.go
│   ├── presentation/                 # Presentation layer
│   │   ├── http/                     # HTTP API
│   │   │   ├── handlers/             # HTTP handlers
│   │   │   │   ├── user_handler.go   # User endpoints
│   │   │   │   └── health_handler.go
│   │   │   ├── middleware/           # HTTP middleware
│   │   │   │   ├── request_id.go     # Request ID middleware
│   │   │   │   ├── logging.go        # Logging middleware
│   │   │   │   └── recovery.go       # Panic recovery middleware
│   │   │   └── server.go             # HTTP server setup
│   │   └── grpc/                     # gRPC API (if needed)
│   │       ├── handlers/
│   │       └── server.go
│   └── container/                    # Dependency injection
│       └── container.go              # DI container
├── pkg/                             # Public packages
│   ├── logger/                      # Logging utilities
│   │   └── logger.go
│   └── validator/                   # Validation utilities
│       └── validator.go
├── config/                          # Configuration
│   ├── config.go                    # Configuration structs
│   └── config.yaml                  # Default configuration
├── migrations/                      # Database migrations
│   ├── 001_create_users_table.up.sql
│   └── 001_create_users_table.down.sql
├── tests/                           # Test files
│   ├── unit/                        # Unit tests
│   │   ├── domain/
│   │   │   ├── entities/
│   │   │   │   └── user_test.go
│   │   │   └── services/
│   │   │       └── user_domain_service_test.go
│   │   ├── usecases/
│   │   │   └── create_user/
│   │   │       └── create_user_test.go
│   │   └── infrastructure/
│   │       └── persistence/
│   │           └── postgres/
│   │               └── user_repository_test.go
│   ├── integration/                 # Integration tests
│   │   └── user_api_test.go
│   └── testutils/                   # Test utilities
│       ├── fixtures.go              # Test fixtures
│       └── mocks.go                 # Test mocks
├── scripts/                         # Build and deployment scripts
│   ├── build.sh                     # Build script
│   ├── test.sh                      # Test script
│   └── migrate.sh                   # Migration script
├── docs/                            # Documentation  
│   ├── architecture.md              # Architecture documentation
│   └── api.md                       # API documentation
├── .env.example                     # Environment variables example
├── .gitignore                       # Git ignore file
├── .golangci.yml                    # Linter configuration
├── docker-compose.yml               # Docker Compose for development
├── Dockerfile                       # Docker configuration
├── go.mod                           # Go modules
├── go.sum                           # Go modules checksum
├── Makefile                         # Build automation
└── README.md                        # Project documentation
```

## Dependencies (go.mod)

```go
module github.com/yourorg/yourproject

go 1.21

require (
    github.com/gorilla/mux v1.8.0
    github.com/jmoiron/sqlx v1.3.5
    github.com/lib/pq v1.10.9
    github.com/google/uuid v1.4.0
    github.com/rs/cors v1.10.1
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    github.com/stretchr/objx v0.5.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

## Development Workflow

### 1. Makefile

```makefile
.PHONY: build test test-unit test-integration lint fmt vet clean run migrate docker-build docker-run

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run all tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run unit tests only
test-unit:
	go test -v -race -short ./...

# Run integration tests only
test-integration:
	go test -v -race -run Integration ./...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Run the application
run:
	go run cmd/server/main.go

# Run database migrations
migrate:
	./scripts/migrate.sh

# Build Docker image
docker-build:
	docker build -t yourorg/yourproject .

# Run with Docker Compose
docker-run:
	docker-compose up --build

# Generate mocks
generate-mocks:
	mockgen -source=internal/domain/ports/user_repository.go -destination=tests/testutils/mocks/user_repository_mock.go
	mockgen -source=internal/domain/ports/event_publisher.go -destination=tests/testutils/mocks/event_publisher_mock.go

# Install tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest

# Run code coverage
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Check for security vulnerabilities
security:
	gosec ./...

# All quality checks
quality: fmt vet lint test security
```

### 2. Setup Instructions

```bash
# Clone repository
git clone <repository-url>
cd golang-hexagonal-architecture

# Install dependencies
go mod download

# Install development tools
make install-tools

# Setup database (PostgreSQL)
createdb myapp_dev
createdb myapp_test

# Run migrations
make migrate

# Run tests
make test

# Build application
make build

# Run application
make run

# Or run with Docker
make docker-run
```

### 3. Code Quality Tools

```yaml
# .golangci.yml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - varcheck
    - deadcode
    - structcheck
    - gocyclo
    - gofmt
    - goimports
    - misspell
    - unparam
    - unconvert
    - goconst
    - gosec

linters-settings:
  gocyclo:
    min-complexity: 10
  govet:
    check-shadowing: true
  misspell:
    locale: US
```

## Best Practices Summary

1. **Clean Architecture**: Maintain strict dependency inversion with domain at the center
2. **Explicit Error Handling**: Use Go's explicit error handling with proper wrapping and context
3. **Context Propagation**: Pass context.Context for cancellation and timeouts
4. **Interface Segregation**: Keep interfaces small and focused
5. **Dependency Injection**: Use explicit constructor injection for testability
6. **Structured Logging**: Use structured logging (slog) with proper context
7. **Graceful Shutdown**: Handle SIGTERM/SIGINT for graceful shutdown
8. **Configuration**: Environment-based configuration with validation
9. **Testing**: Comprehensive unit, integration, and table-driven tests
10. **Concurrency**: Leverage goroutines and channels where appropriate
11. **Resource Management**: Proper cleanup of database connections and other resources
12. **Monitoring**: Health checks and metrics for observability
13. **Security**: Input validation, SQL injection prevention, proper error messages
14. **Performance**: Connection pooling, efficient database queries, caching where needed
15. **Documentation**: Clear package documentation and API documentation

This comprehensive guide provides a robust foundation for building maintainable, testable, and production-ready Go microservices using hexagonal architecture principles while leveraging Go's strengths in simplicity, performance, and concurrency.