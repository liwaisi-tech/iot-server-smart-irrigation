package entities

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "valid short message",
			content:   "Hello",
			wantError: false,
		},
		{
			name:      "valid long message",
			content:   "This is a longer message with multiple words and sentences. It should be handled correctly by the Message entity.",
			wantError: false,
		},
		{
			name:      "valid message with special characters",
			content:   "Message with special chars: @#$%^&*()_+-=[]{}|;:,.<>?",
			wantError: false,
		},
		{
			name:      "valid message with numbers",
			content:   "Message123 with numbers 456 and symbols",
			wantError: false,
		},
		{
			name:      "valid message with newlines",
			content:   "Multi-line\nmessage\nwith\nnewlines",
			wantError: false,
		},
		{
			name:      "valid message with unicode",
			content:   "Message with unicode: 你好世界 🌍 émojis",
			wantError: false,
		},
		{
			name:      "valid single character",
			content:   "A",
			wantError: false,
		},
		{
			name:      "valid message with only spaces",
			content:   "   ",
			wantError: false,
		},
		{
			name:      "empty content",
			content:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeTime := time.Now()
			
			message, err := NewMessage(tt.content)
			
			afterTime := time.Now()

			if tt.wantError {
				assert.Error(t, err, "NewMessage() expected error but got none")
				assert.Nil(t, message, "NewMessage() expected nil message")
			} else {
				require.NoError(t, err, "NewMessage() unexpected error")
				require.NotNil(t, message, "NewMessage() expected message but got nil")

				// Verify content is set correctly
				assert.Equal(t, tt.content, message.Content, "NewMessage() content mismatch")

				// Verify ID is generated and not empty
				assert.NotEmpty(t, message.ID, "NewMessage() ID should not be empty")

				// Verify ID format (should start with "msg_")
				assert.True(t, strings.HasPrefix(message.ID, "msg_"), "NewMessage() ID should start with 'msg_', got '%s'", message.ID)

				// Verify timestamp is set correctly
				assert.False(t, message.CreatedAt.Before(beforeTime) || message.CreatedAt.After(afterTime), "NewMessage() CreatedAt timestamp not within expected range")
			}
		})
	}
}

func TestNewMessage_UniqueIDs(t *testing.T) {
	// Test that multiple messages created in quick succession have unique IDs
	const numMessages = 1000
	ids := make(map[string]bool)
	content := "Test message for unique ID testing"

	for i := 0; i < numMessages; i++ {
		message, err := NewMessage(content)
		require.NoError(t, err, "NewMessage() unexpected error on iteration %d", i)

		assert.False(t, ids[message.ID], "NewMessage() duplicate ID found: %s", message.ID)

		ids[message.ID] = true
	}

	assert.Equal(t, numMessages, len(ids), "NewMessage() expected %d unique IDs, got %d", numMessages, len(ids))
}

func TestNewMessage_IDFormat(t *testing.T) {
	message, err := NewMessage("Test content")
	require.NoError(t, err, "NewMessage() unexpected error")

	// Verify ID format: "msg_" + unix nano timestamp
	assert.True(t, strings.HasPrefix(message.ID, "msg_"), "NewMessage() ID should start with 'msg_', got '%s'", message.ID)

	// Extract timestamp part and verify it's numeric
	timestampPart := strings.TrimPrefix(message.ID, "msg_")
	assert.NotEmpty(t, timestampPart, "NewMessage() ID should have timestamp part after 'msg_'")

	// The timestamp part should be all digits
	for _, char := range timestampPart {
		assert.True(t, char >= '0' && char <= '9', "NewMessage() ID timestamp part should be numeric, got '%s'", timestampPart)
	}
}

func TestMessage_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		message   *Message
		wantError bool
	}{
		{
			name: "valid message",
			message: &Message{
				ID:        "msg_1234567890",
				Content:   "Valid content",
				CreatedAt: time.Now(),
			},
			wantError: false,
		},
		{
			name: "valid message with long content",
			message: &Message{
				ID:        "msg_9876543210",
				Content:   strings.Repeat("Long content ", 100),
				CreatedAt: time.Now(),
			},
			wantError: false,
		},
		{
			name: "valid message with special characters",
			message: &Message{
				ID:        "msg_5555555555",
				Content:   "Content with special chars: @#$%^&*()",
				CreatedAt: time.Now(),
			},
			wantError: false,
		},
		{
			name: "valid message with spaces only",
			message: &Message{
				ID:        "msg_1111111111",
				Content:   "   ",
				CreatedAt: time.Now(),
			},
			wantError: false,
		},
		{
			name: "missing ID",
			message: &Message{
				ID:        "",
				Content:   "Valid content",
				CreatedAt: time.Now(),
			},
			wantError: true,
		},
		{
			name: "missing content",
			message: &Message{
				ID:        "msg_1234567890",
				Content:   "",
				CreatedAt: time.Now(),
			},
			wantError: true,
		},
		{
			name: "both ID and content missing",
			message: &Message{
				ID:        "",
				Content:   "",
				CreatedAt: time.Now(),
			},
			wantError: true,
		},
		{
			name: "zero timestamp (should still be valid)",
			message: &Message{
				ID:        "msg_1234567890",
				Content:   "Valid content",
				CreatedAt: time.Time{},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.IsValid()

			if tt.wantError {
				assert.Error(t, err, "IsValid() expected error but got none")
			} else {
				assert.NoError(t, err, "IsValid() unexpected error")
			}
		})
	}
}

func TestMessage_IsValid_ErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		message         *Message
		expectedError   string
	}{
		{
			name: "missing ID error message",
			message: &Message{
				ID:        "",
				Content:   "Valid content",
				CreatedAt: time.Now(),
			},
			expectedError: "message ID is required",
		},
		{
			name: "missing content error message",
			message: &Message{
				ID:        "msg_1234567890",
				Content:   "",
				CreatedAt: time.Now(),
			},
			expectedError: "message content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.IsValid()

			require.Error(t, err, "IsValid() expected error but got none")
			assert.Equal(t, tt.expectedError, err.Error(), "IsValid() error message mismatch")
		})
	}
}

func TestMessage_IsValid_MultipleValidationErrors(t *testing.T) {
	// Test message with both ID and content missing
	// The validation should return the first error it encounters (ID check comes first)
	message := &Message{
		ID:        "",
		Content:   "",
		CreatedAt: time.Now(),
	}

	err := message.IsValid()
	require.Error(t, err, "IsValid() expected error for message with missing ID and content")

	// Should return the ID error first since it's checked first
	expectedError := "message ID is required"
	assert.Equal(t, expectedError, err.Error(), "IsValid() expected first validation error mismatch")
}

func TestMessage_CreatedFromNewMessage_IsAlwaysValid(t *testing.T) {
	// Test that any message created via NewMessage() will always pass IsValid()
	testContents := []string{
		"Simple message",
		"Message with numbers 123",
		"Message with special chars: @#$%^&*()",
		strings.Repeat("Long message ", 100),
		"Multi-line\nmessage\nwith\nbreaks",
		"Unicode message: 你好世界 🌍",
		"   ", // spaces only
		"A",   // single character
	}

	for _, content := range testContents {
		t.Run("content: "+content[:min(len(content), 20)], func(t *testing.T) {
			message, err := NewMessage(content)
			require.NoError(t, err, "NewMessage() unexpected error")

			err = message.IsValid()
			assert.NoError(t, err, "Message created by NewMessage() should always be valid")
		})
	}
}

func TestMessage_Fields(t *testing.T) {
	content := "Test message content"
	beforeTime := time.Now()
	
	message, err := NewMessage(content)
	require.NoError(t, err, "NewMessage() unexpected error")
	
	afterTime := time.Now()

	// Test that all fields are properly set
	assert.Equal(t, content, message.Content, "Message.Content mismatch")
	assert.NotEmpty(t, message.ID, "Message.ID should not be empty")
	assert.False(t, message.CreatedAt.Before(beforeTime) || message.CreatedAt.After(afterTime), "Message.CreatedAt should be set to current time")

	// Test that CreatedAt has reasonable precision (nanoseconds)
	if message.CreatedAt.Nanosecond() == 0 {
		// This could theoretically fail if the nanosecond portion is exactly 0,
		// but it's extremely unlikely and indicates the timestamp has good precision
		t.Logf("Message.CreatedAt nanosecond portion is 0, which is unlikely but possible")
	}
}

func TestMessage_Immutability_Concept(t *testing.T) {
	// While Go doesn't enforce immutability, test that the message structure
	// is designed to be used in an immutable way
	content := "Original content"
	
	message, err := NewMessage(content)
	require.NoError(t, err, "NewMessage() unexpected error")

	originalID := message.ID
	originalContent := message.Content
	originalCreatedAt := message.CreatedAt

	// Simulate external modification (this would be bad practice in real code)
	message.Content = "Modified content"
	message.ID = "modified_id"
	message.CreatedAt = time.Now().Add(time.Hour)

	// Verify that the fields can be modified (showing why immutability patterns would be good)
	assert.NotEqual(t, originalContent, message.Content, "Message.Content was not modified (this test verifies mutability exists)")
	assert.NotEqual(t, originalID, message.ID, "Message.ID was not modified (this test verifies mutability exists)")
	assert.False(t, message.CreatedAt.Equal(originalCreatedAt), "Message.CreatedAt was not modified (this test verifies mutability exists)")

	// This test demonstrates that while the Message struct can be modified,
	// in a well-designed system, messages should be treated as immutable
	// after creation to maintain data integrity
}

func TestMessage_ZeroValue(t *testing.T) {
	// Test behavior of zero-value Message
	var message Message

	err := message.IsValid()
	require.Error(t, err, "Zero-value Message should not be valid")

	// Zero-value should fail validation for missing ID
	expectedError := "message ID is required"
	assert.Equal(t, expectedError, err.Error(), "Zero-value Message validation error mismatch")
}

// Helper function for min (since Go doesn't have a built-in min for strings)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}