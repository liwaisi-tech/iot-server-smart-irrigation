package errors

import "fmt"

// DomainError represents a business domain error
type DomainError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return fmt.Sprintf("domain error [%s]: %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the domain error
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	e.Details[key] = value
	return e
}

// Common domain errors
var (
	ErrInternalServer = NewDomainError("INTERNAL_SERVER_ERROR", "An internal server error occurred")
	ErrNotFound       = NewDomainError("NOT_FOUND", "The requested resource was not found")
	ErrInvalidInput   = NewDomainError("INVALID_INPUT", "The provided input is invalid")
)