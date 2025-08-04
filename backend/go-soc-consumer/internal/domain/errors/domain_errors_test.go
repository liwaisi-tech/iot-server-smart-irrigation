package errors

import (
	"testing"
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

			if err == nil {
				t.Errorf("NewDomainError() returned nil")
				return
			}

			if err.Code != tt.code {
				t.Errorf("NewDomainError() code expected %s, got %s", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("NewDomainError() message expected %s, got %s", tt.message, err.Message)
			}

			if err.Details == nil {
				t.Errorf("NewDomainError() Details should be initialized")
			}

			if len(err.Details) != 0 {
				t.Errorf("NewDomainError() Details should be empty initially, got %d items", len(err.Details))
			}
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

			if errorString != tt.expectedString {
				t.Errorf("Error() expected '%s', got '%s'", tt.expectedString, errorString)
			}
		})
	}
}

func TestDomainError_WithDetails_SingleDetail(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test error message")

	// Add a single detail
	result := err.WithDetails("field", "mac_address")

	// Verify the same instance is returned
	if result != err {
		t.Errorf("WithDetails() should return the same instance")
	}

	// Verify the detail was added
	if len(err.Details) != 1 {
		t.Errorf("WithDetails() expected 1 detail, got %d", len(err.Details))
	}

	value, exists := err.Details["field"]
	if !exists {
		t.Errorf("WithDetails() detail 'field' not found")
	}

	if value != "mac_address" {
		t.Errorf("WithDetails() detail value expected 'mac_address', got %v", value)
	}
}

func TestDomainError_WithDetails_MultipleDetails(t *testing.T) {
	err := NewDomainError("VALIDATION_ERROR", "Multiple validation failures")

	// Add multiple details
	err.WithDetails("field1", "mac_address").
		WithDetails("field2", "device_name").
		WithDetails("field3", "ip_address")

	// Verify all details were added
	if len(err.Details) != 3 {
		t.Errorf("WithDetails() expected 3 details, got %d", len(err.Details))
	}

	expectedDetails := map[string]interface{}{
		"field1": "mac_address",
		"field2": "device_name",
		"field3": "ip_address",
	}

	for key, expectedValue := range expectedDetails {
		actualValue, exists := err.Details[key]
		if !exists {
			t.Errorf("WithDetails() detail '%s' not found", key)
		}

		if actualValue != expectedValue {
			t.Errorf("WithDetails() detail '%s' expected %v, got %v", key, expectedValue, actualValue)
		}
	}
}

func TestDomainError_WithDetails_OverwriteDetail(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test error")

	// Add initial detail
	err.WithDetails("field", "initial_value")

	// Verify initial detail
	if err.Details["field"] != "initial_value" {
		t.Errorf("WithDetails() initial value not set correctly")
	}

	// Overwrite the detail
	err.WithDetails("field", "updated_value")

	// Verify detail was overwritten
	if len(err.Details) != 1 {
		t.Errorf("WithDetails() should still have 1 detail after overwrite, got %d", len(err.Details))
	}

	if err.Details["field"] != "updated_value" {
		t.Errorf("WithDetails() detail not overwritten correctly, got %v", err.Details["field"])
	}
}

func TestDomainError_WithDetails_DifferentTypes(t *testing.T) {
	err := NewDomainError("TEST_ERROR", "Test with different types")

	// Add details with different types
	err.WithDetails("string_field", "string_value").
		WithDetails("int_field", 42).
		WithDetails("bool_field", true).
		WithDetails("float_field", 3.14).
		WithDetails("nil_field", nil)

	// Verify all details were added with correct types
	if len(err.Details) != 5 {
		t.Errorf("WithDetails() expected 5 details, got %d", len(err.Details))
	}

	// Test string field
	if err.Details["string_field"] != "string_value" {
		t.Errorf("WithDetails() string field incorrect")
	}

	// Test int field
	if err.Details["int_field"] != 42 {
		t.Errorf("WithDetails() int field incorrect")
	}

	// Test bool field
	if err.Details["bool_field"] != true {
		t.Errorf("WithDetails() bool field incorrect")
	}

	// Test float field
	if err.Details["float_field"] != 3.14 {
		t.Errorf("WithDetails() float field incorrect")
	}

	// Test nil field
	if err.Details["nil_field"] != nil {
		t.Errorf("WithDetails() nil field incorrect")
	}
}

func TestDomainError_WithDetails_ChainedCalls(t *testing.T) {
	// Test that chained calls work correctly and return the same instance
	err := NewDomainError("CHAINED_ERROR", "Test chained calls")

	result := err.WithDetails("key1", "value1").
		WithDetails("key2", "value2").
		WithDetails("key3", "value3")

	// Verify it's the same instance
	if result != err {
		t.Errorf("WithDetails() chained calls should return the same instance")
	}

	// Verify all details are present
	if len(err.Details) != 3 {
		t.Errorf("WithDetails() chained calls expected 3 details, got %d", len(err.Details))
	}
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
			if tt.error == nil {
				t.Errorf("Predefined error %s is nil", tt.name)
				return
			}

			if tt.error.Code != tt.expectedCode {
				t.Errorf("Predefined error %s code expected %s, got %s", tt.name, tt.expectedCode, tt.error.Code)
			}

			if tt.error.Message != tt.expectedMsg {
				t.Errorf("Predefined error %s message expected %s, got %s", tt.name, tt.expectedMsg, tt.error.Message)
			}

			if tt.error.Details == nil {
				t.Errorf("Predefined error %s Details should be initialized", tt.name)
			}

			// Verify error string format
			expectedErrorString := "domain error [" + tt.expectedCode + "]: " + tt.expectedMsg
			if tt.error.Error() != expectedErrorString {
				t.Errorf("Predefined error %s Error() expected '%s', got '%s'", tt.name, expectedErrorString, tt.error.Error())
			}
		})
	}
}

func TestPredefinedErrors_WithDetails(t *testing.T) {
	// Test that predefined errors can have details added
	internalServerErr := ErrInternalServer.WithDetails("operation", "device_save")
	
	if len(internalServerErr.Details) != 1 {
		t.Errorf("ErrInternalServer.WithDetails() expected 1 detail, got %d", len(internalServerErr.Details))
	}

	if internalServerErr.Details["operation"] != "device_save" {
		t.Errorf("ErrInternalServer.WithDetails() detail not set correctly")
	}

	// Verify the original predefined error is modified (since it's the same instance)
	if ErrInternalServer != internalServerErr {
		t.Errorf("ErrInternalServer.WithDetails() should return the same instance")
	}

	// Reset for other tests (this is a side effect we need to handle)
	delete(ErrInternalServer.Details, "operation")
}

func TestPredefinedErrors_Independence(t *testing.T) {
	// Ensure each predefined error is independent
	originalInternalCount := len(ErrInternalServer.Details)
	originalNotFoundCount := len(ErrNotFound.Details)
	originalInvalidInputCount := len(ErrInvalidInput.Details)

	// Add details to different predefined errors
	ErrInternalServer.WithDetails("test_internal", "value1")
	ErrNotFound.WithDetails("test_not_found", "value2")
	ErrInvalidInput.WithDetails("test_invalid", "value3")

	// Verify they don't affect each other
	if len(ErrInternalServer.Details) != originalInternalCount + 1 {
		t.Errorf("ErrInternalServer details count incorrect")
	}

	if len(ErrNotFound.Details) != originalNotFoundCount + 1 {
		t.Errorf("ErrNotFound details count incorrect")
	}

	if len(ErrInvalidInput.Details) != originalInvalidInputCount + 1 {
		t.Errorf("ErrInvalidInput details count incorrect")
	}

	// Verify the specific details
	if ErrInternalServer.Details["test_internal"] != "value1" {
		t.Errorf("ErrInternalServer detail not set correctly")
	}

	if ErrNotFound.Details["test_not_found"] != "value2" {
		t.Errorf("ErrNotFound detail not set correctly")
	}

	if ErrInvalidInput.Details["test_invalid"] != "value3" {
		t.Errorf("ErrInvalidInput detail not set correctly")
	}

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

	if standardErr == nil {
		t.Errorf("DomainError should be assignable to error interface")
	}

	// Error() method should be called
	if standardErr.Error() != "domain error [TEST_ERROR]: Test message" {
		t.Errorf("DomainError as standard error should call Error() method")
	}
}