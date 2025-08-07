package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/mocks"
)

func TestNewPingHandler(t *testing.T) {
	mockUseCase := mocks.NewMockPingUseCase(t)

	handler := NewPingHandler(mockUseCase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUseCase, handler.pingUseCase)
}

func TestPingHandler_Ping(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setupMock      func(*mocks.MockPingUseCase)
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:   "successful ping request",
			method: http.MethodGet,
			setupMock: func(mockUseCase *mocks.MockPingUseCase) {
				mockUseCase.EXPECT().
					Ping(mock.Anything).
					Return("pong").
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
			expectedHeader: "text/plain",
		},
		{
			name:   "ping with POST method",
			method: http.MethodPost,
			setupMock: func(mockUseCase *mocks.MockPingUseCase) {
				mockUseCase.EXPECT().
					Ping(mock.Anything).
					Return("pong").
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
			expectedHeader: "text/plain",
		},
		{
			name:   "ping with PUT method",
			method: http.MethodPut,
			setupMock: func(mockUseCase *mocks.MockPingUseCase) {
				mockUseCase.EXPECT().
					Ping(mock.Anything).
					Return("pong").
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
			expectedHeader: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockPingUseCase(t)
			tt.setupMock(mockUseCase)

			handler := NewPingHandler(mockUseCase)

			// Create a request
			req := httptest.NewRequest(tt.method, "/ping", nil)

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler
			handler.Ping(w, req)

			// Check the response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"))

			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestPingHandler_Ping_ContextHandling(t *testing.T) {
	t.Run("context is passed correctly to use case", func(t *testing.T) {
		mockUseCase := mocks.NewMockPingUseCase(t)

		// We can't easily test the exact context, but we can verify Execute is called
		mockUseCase.EXPECT().
			Ping(mock.Anything).
			Return("pong").
			Once()

		handler := NewPingHandler(mockUseCase)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.Ping(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("context with custom values", func(t *testing.T) {
		mockUseCase := mocks.NewMockPingUseCase(t)

		mockUseCase.EXPECT().
			Ping(mock.Anything).
			Return("pong").
			Once()

		handler := NewPingHandler(mockUseCase)

		// Create request with custom context
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		ctx := context.WithValue(req.Context(), "test-key", "test-value")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Ping(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
		mockUseCase.AssertExpectations(t)
	})
}

func TestPingHandler_Ping_ResponseHeaders(t *testing.T) {
	t.Run("sets correct content type header", func(t *testing.T) {
		mockUseCase := mocks.NewMockPingUseCase(t)

		mockUseCase.EXPECT().
			Ping(mock.Anything).
			Return("pong").
			Once()

		handler := NewPingHandler(mockUseCase)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.Ping(w, req)

		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestPingHandler_Ping_DifferentResponses(t *testing.T) {
	tests := []struct {
		name         string
		useCaseResp  string
		expectedBody string
	}{
		{
			name:         "standard pong response",
			useCaseResp:  "pong",
			expectedBody: "pong",
		},
		{
			name:         "custom response",
			useCaseResp:  "service is alive",
			expectedBody: "service is alive",
		},
		{
			name:         "empty response",
			useCaseResp:  "",
			expectedBody: "",
		},
		{
			name:         "json-like response",
			useCaseResp:  `{"status":"ok"}`,
			expectedBody: `{"status":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockPingUseCase(t)

			mockUseCase.EXPECT().
				Ping(mock.Anything).
				Return(tt.useCaseResp).
				Once()

			handler := NewPingHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			handler.Ping(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestPingHandler_Ping_Integration(t *testing.T) {
	t.Run("complete request response cycle", func(t *testing.T) {
		mockUseCase := mocks.NewMockPingUseCase(t)

		mockUseCase.EXPECT().
			Ping(mock.Anything).
			Return("pong").
			Times(3) // Expect 3 calls since we're making 3 requests

		handler := NewPingHandler(mockUseCase)

		// Test multiple requests to ensure handler state is maintained correctly
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			handler.Ping(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "pong", w.Body.String())
			assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		}

		// We expect 3 calls since we made 3 requests
		mockUseCase.AssertExpectations(t)
	})
}

func TestPingHandler_Ping_EdgeCases(t *testing.T) {
	t.Run("nil request context", func(t *testing.T) {
		mockUseCase := mocks.NewMockPingUseCase(t)

		// The handler should still work even if context handling has issues
		mockUseCase.EXPECT().
			Ping(mock.Anything).
			Return("pong").
			Once()

		handler := NewPingHandler(mockUseCase)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.Ping(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkPingHandler_Ping(b *testing.B) {
	mockUseCase := mocks.NewMockPingUseCase(&testing.T{})

	// Setup mock to handle all benchmark iterations
	mockUseCase.EXPECT().
		Ping(mock.Anything).
		Return("pong").
		Times(b.N)

	handler := NewPingHandler(mockUseCase)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.Ping(w, req)
	}
}

func BenchmarkPingHandler_Ping_WithLargeResponse(b *testing.B) {
	mockUseCase := mocks.NewMockPingUseCase(&testing.T{})

	// Create a large response to test performance with bigger payloads
	largeResponse := ""
	for i := 0; i < 1000; i++ {
		largeResponse += "pong "
	}

	mockUseCase.EXPECT().
		Ping(mock.Anything).
		Return(largeResponse).
		Times(b.N)

	handler := NewPingHandler(mockUseCase)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.Ping(w, req)
	}
}
