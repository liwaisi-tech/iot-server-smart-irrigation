package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDomainError(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{
			name:    "simple error",
			code:    "TEST_ERROR",
			message: "This is a test error",
		},
		{
			name:    "empty code",
			code:    "",
			message: "Error with empty code",
		},
		{
			name:    "empty message",
			code:    "EMPTY_MSG",
			message: "",
		},
		{
			name:    "both empty",
			code:    "",
			message: "",
		},
		{
			name:    "complex error",
			code:    "VALIDATION_FAILURE",
			message: "Input validation failed: device name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDomainError(tt.code, tt.message)

			require.NotNil(t, err, "NewDomainError() returned nil")
			assert.Equal(t, tt.code, err.Code, "NewDomainError() code mismatch")
			assert.Equal(t, tt.message, err.Message, "NewDomainError() message mismatch")
			assert.NotNil(t, err.Details, "NewDomainError() Details should be initialized")
			assert.Empty(t, err.Details, "NewDomainError() Details should be empty initially")
		})
	}
}

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		message        string
		expectedString string
	}{
		{
			name:           "normal error",
			code:           "TEST_ERROR",
			message:        "This is a test error",
			expectedString: "domain error [TEST_ERROR]: This is a test error",
		},
		{
			name:           "empty code",
			code:           "",
			message:        "Error with empty code",
			expectedString: "domain error []: Error with empty code",
		},
		{
			name:           "empty message",
			code:           "EMPTY_MSG",
			message:        "",
			expectedString: "domain error [EMPTY_MSG]: ",
		},
		{
			name:           "both empty",
			code:           "",
			message:        "",
			expectedString: "domain error []: ",
		},
		{
			name:           "validation error",
			code:           "VALIDATION_ERROR",
			message:        "Input validation failed",
			expectedString: "domain error [VALIDATION_ERROR]: Input validation failed",
		},
		{
			name:           "not found error",
			code:           "NOT_FOUND",
			message:        "Device with MAC address AA:BB:CC:DD:EE:FF not found",
			expectedString: "domain error [NOT_FOUND]: Device with MAC address AA:BB:CC:DD:EE:FF not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDomainError(tt.code, tt.message)
			errorString := err.Error()

			assert.Equal(t, tt.expectedString, errorString, "Error() string format mismatch")
		})
	}
}

func TestDomainError_WithDetails_SingleDetail(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test error message")

	// Add a single detail
	result := err.WithDetails("field", "mac_address")

	// Verify the same instance is returned
	assert.Same(t, err, result, "WithDetails() should return the same instance")

	// Verify the detail was added
	assert.Len(t, err.Details, 1, "WithDetails() should have 1 detail")

	value, exists := err.Details["field"]
	assert.True(t, exists, "WithDetails() detail 'field' not found")
	assert.Equal(t, "mac_address", value, "WithDetails() detail value mismatch")
}

func TestDomainError_WithDetails_MultipleDetails(t *testing.T) {
	err := NewDomainError("VALIDATION_ERROR", "Multiple validation failures")

	// Add multiple details (using blank identifiers since we don't need the return values in tests)
	_ = err.WithDetails("field1", "mac_address").
		WithDetails("field2", "device_name").
		WithDetails("field3", "ip_address")

	// Verify all details were added
	assert.Len(t, err.Details, 3, "WithDetails() should have 3 details")

	expectedDetails := map[string]interface{}{
		"field1": "mac_address",
		"field2": "device_name",
		"field3": "ip_address",
	}

	for key, expectedValue := range expectedDetails {
		actualValue, exists := err.Details[key]
		assert.True(t, exists, "WithDetails() detail '%s' not found", key)
		assert.Equal(t, expectedValue, actualValue, "WithDetails() detail '%s' value mismatch", key)
	}
}

func TestDomainError_WithDetails_OverwriteDetail(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test error")

	// Add initial detail (using blank identifier since we don't need the return value in tests)
	_ = err.WithDetails("field", "initial_value")

	// Verify initial detail
	assert.Equal(t, "initial_value", err.Details["field"], "WithDetails() initial value not set correctly")

	// Overwrite the detail (using blank identifier since we don't need the return value in tests)
	_ = err.WithDetails("field", "updated_value")

	// Verify detail was overwritten
	assert.Len(t, err.Details, 1, "WithDetails() should still have 1 detail after overwrite")
	assert.Equal(t, "updated_value", err.Details["field"], "WithDetails() detail not overwritten correctly")
}

func TestDomainError_WithDetails_DifferentTypes(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test with different types")

	// Add details with different types (using blank identifier since we don't need the return values in tests)
	_ = err.WithDetails("string_field", "string_value").
		WithDetails("int_field", 42).
		WithDetails("bool_field", true).
		WithDetails("float_field", 3.14).
		WithDetails("nil_field", nil)

	// Verify all details were added with correct types
	assert.Len(t, err.Details, 5, "WithDetails() should have 5 details")

	// Test string field
	assert.Equal(t, "string_value", err.Details["string_field"], "WithDetails() string field incorrect")

	// Test int field
	assert.Equal(t, 42, err.Details["int_field"], "WithDetails() int field incorrect")

	// Test bool field
	assert.Equal(t, true, err.Details["bool_field"], "WithDetails() bool field incorrect")

	// Test float field
	assert.Equal(t, 3.14, err.Details["float_field"], "WithDetails() float field incorrect")

	// Test nil field
	assert.Nil(t, err.Details["nil_field"], "WithDetails() nil field incorrect")
}

func TestDomainError_WithDetails_ChainedCalls(t *testing.T) {
	// Test that chained calls work correctly and return the same instance
	err := NewDomainError("CHAINED_ERROR", "Test chained calls")

	result := err.WithDetails("key1", "value1").
		WithDetails("key2", "value2").
		WithDetails("key3", "value3")

	// Verify it's the same instance
	assert.Same(t, err, result, "WithDetails() chained calls should return the same instance")

	// Verify all details are present
	assert.Len(t, err.Details, 3, "WithDetails() chained calls should have 3 details")
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name        string
		error       *DomainError
		expectedCode string
		expectedMsg  string
	}{
		{
			name:        "ErrInternalServer",
			error:       ErrInternalServer,
			expectedCode: "INTERNAL_SERVER_ERROR",
			expectedMsg:  "An internal server error occurred",
		},
		{
			name:        "ErrNotFound",
			error:       ErrNotFound,
			expectedCode: "NOT_FOUND",
			expectedMsg:  "The requested resource was not found",
		},
		{
			name:        "ErrInvalidInput",
			error:       ErrInvalidInput,
			expectedCode: "INVALID_INPUT",
			expectedMsg:  "The provided input is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.error, "Predefined error %s is nil", tt.name)
			assert.Equal(t, tt.expectedCode, tt.error.Code, "Predefined error %s code mismatch", tt.name)
			assert.Equal(t, tt.expectedMsg, tt.error.Message, "Predefined error %s message mismatch", tt.name)
			assert.NotNil(t, tt.error.Details, "Predefined error %s Details should be initialized", tt.name)

			// Verify error string format
			expectedErrorString := "domain error [" + tt.expectedCode + "]: " + tt.expectedMsg
			assert.Equal(t, expectedErrorString, tt.error.Error(), "Predefined error %s Error() string mismatch", tt.name)
		})
	}
}

func TestPredefinedErrors_WithDetails(t *testing.T) {
	// Test that predefined errors can have details added
	internalServerErr := ErrInternalServer.WithDetails("operation", "device_save")
	
	assert.Len(t, internalServerErr.Details, 1, "ErrInternalServer.WithDetails() should have 1 detail")
	assert.Equal(t, "device_save", internalServerErr.Details["operation"], "ErrInternalServer.WithDetails() detail not set correctly")

	// Verify the original predefined error is modified (since it's the same instance)
	assert.Same(t, ErrInternalServer, internalServerErr, "ErrInternalServer.WithDetails() should return the same instance")

	// Reset for other tests (this is a side effect we need to handle)
	delete(ErrInternalServer.Details, "operation")
}

func TestPredefinedErrors_Independence(t *testing.T) {
	// Ensure each predefined error is independent
	originalInternalCount := len(ErrInternalServer.Details)
	originalNotFoundCount := len(ErrNotFound.Details)
	originalInvalidInputCount := len(ErrInvalidInput.Details)

	// Add details to different predefined errors (using blank identifiers since we don't need the return values in tests)
	_ = ErrInternalServer.WithDetails("test_internal", "value1")
	_ = ErrNotFound.WithDetails("test_not_found", "value2")
	_ = ErrInvalidInput.WithDetails("test_invalid", "value3")

	// Verify they don't affect each other
	assert.Equal(t, originalInternalCount+1, len(ErrInternalServer.Details), "ErrInternalServer details count incorrect")
	assert.Equal(t, originalNotFoundCount+1, len(ErrNotFound.Details), "ErrNotFound details count incorrect")
	assert.Equal(t, originalInvalidInputCount+1, len(ErrInvalidInput.Details), "ErrInvalidInput details count incorrect")

	// Verify the specific details
	assert.Equal(t, "value1", ErrInternalServer.Details["test_internal"], "ErrInternalServer detail not set correctly")
	assert.Equal(t, "value2", ErrNotFound.Details["test_not_found"], "ErrNotFound detail not set correctly")
	assert.Equal(t, "value3", ErrInvalidInput.Details["test_invalid"], "ErrInvalidInput detail not set correctly")

	// Clean up
	delete(ErrInternalServer.Details, "test_internal")
	delete(ErrNotFound.Details, "test_not_found")
	delete(ErrInvalidInput.Details, "test_invalid")
}

func TestDomainError_AsStandardError(t *testing.T) {
	// Test that DomainError can be used as a standard Go error
	err := NewDomainError("TEST_ERROR", "Test message")

	// Should be able to assign to error interface
	var standardErr error = err

	assert.NotNil(t, standardErr, "DomainError should be assignable to error interface")

	// Error() method should be called
	assert.Equal(t, "domain error [TEST_ERROR]: Test message", standardErr.Error(), "DomainError as standard error should call Error() method")
}