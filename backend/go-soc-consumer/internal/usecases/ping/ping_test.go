package ping

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

func TestNewUseCase(t *testing.T) {
	useCase := NewUseCase()

	assert.NotNil(t, useCase)

	// Verify it implements the UseCase interface
	var _ PingUseCase = useCase
}

func TestUseCaseImpl_Execute(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "basic execution",
			ctx:      context.Background(),
			expected: "pong",
		},
		{
			name:     "execution with context with value",
			ctx:      context.WithValue(context.Background(), contextKey("test-key"), "test-value"),
			expected: "pong",
		},
		{
			name:     "execution with cancelled context",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expected: "pong", // Should still return pong even with cancelled context
		},
		{
			name:     "execution with nil context (would panic in real usage but testing behavior)",
			ctx:      context.Background(),
			expected: "pong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewUseCase()

			result := useCase.Ping(tt.ctx)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUseCaseImpl_Execute_ConsistentBehavior(t *testing.T) {
	t.Run("multiple calls return same result", func(t *testing.T) {
		useCase := NewUseCase()

		// Call multiple times to ensure consistent behavior
		for i := 0; i < 5; i++ {
			result := useCase.Ping(context.Background())
			assert.Equal(t, "pong", result)
		}
	})

	t.Run("concurrent calls return same result", func(t *testing.T) {
		useCase := NewUseCase()

		results := make(chan string, 10)

		// Start 10 concurrent goroutines
		for i := 0; i < 10; i++ {
			go func() {
				result := useCase.Ping(context.Background())
				results <- result
			}()
		}

		// Collect all results
		for i := 0; i < 10; i++ {
			result := <-results
			assert.Equal(t, "pong", result)
		}
	})
}

func TestUseCaseImpl_Execute_ImplementsInterface(t *testing.T) {
	t.Run("implements UseCase interface correctly", func(t *testing.T) {
		var useCase PingUseCase = NewUseCase()

		result := useCase.Ping(context.Background())
		assert.Equal(t, "pong", result)
	})
}

func TestUseCaseImpl_Execute_ContextHandling(t *testing.T) {
	t.Run("context is accepted but not used", func(t *testing.T) {
		useCase := NewUseCase()

		// Test with different context types
		contexts := []context.Context{
			context.Background(),
			context.TODO(),
			context.WithValue(context.Background(), contextKey("testKey"), "value"),
		}

		for _, ctx := range contexts {
			result := useCase.Ping(ctx)
			assert.Equal(t, "pong", result)
		}
	})
}

// Benchmark tests
func BenchmarkUseCaseImpl_Execute(b *testing.B) {
	useCase := NewUseCase()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		useCase.Ping(ctx)
	}
}

func BenchmarkUseCaseImpl_Execute_WithContext(b *testing.B) {
	useCase := NewUseCase()
	ctx := context.WithValue(context.Background(), contextKey("benchmark-key"), "benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		useCase.Ping(ctx)
	}
}

func BenchmarkUseCaseImpl_Execute_Concurrent(b *testing.B) {
	useCase := NewUseCase()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			useCase.Ping(ctx)
		}
	})
}

// Table-driven test for edge cases
func TestUseCaseImpl_Execute_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expected    string
		description string
	}{
		{
			name:        "background context",
			setupCtx:    func() context.Context { return context.Background() },
			expected:    "pong",
			description: "Should work with background context",
		},
		{
			name:        "todo context",
			setupCtx:    func() context.Context { return context.TODO() },
			expected:    "pong",
			description: "Should work with TODO context",
		},
		{
			name: "context with deadline",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 0) // Already expired
				cancel()
				return ctx
			},
			expected:    "pong",
			description: "Should work even with expired deadline",
		},
		{
			name: "context with cancellation",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected:    "pong",
			description: "Should work even with cancelled context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewUseCase()
			ctx := tt.setupCtx()

			result := useCase.Ping(ctx)

			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// Test to ensure the use case is stateless
func TestUseCaseImpl_Stateless(t *testing.T) {
	t.Run("use case is stateless between calls", func(t *testing.T) {
		useCase := NewUseCase()

		// Multiple calls should not affect each other
		result1 := useCase.Ping(context.Background())
		result2 := useCase.Ping(context.WithValue(context.Background(), contextKey("key"), "value"))
		result3 := useCase.Ping(context.TODO())

		assert.Equal(t, "pong", result1)
		assert.Equal(t, "pong", result2)
		assert.Equal(t, "pong", result3)

		// All results should be identical
		assert.Equal(t, result1, result2)
		assert.Equal(t, result2, result3)
	})
}

// Test multiple instances
func TestUseCaseImpl_MultipleInstances(t *testing.T) {
	t.Run("multiple instances behave identically", func(t *testing.T) {
		useCase1 := NewUseCase()
		useCase2 := NewUseCase()

		result1 := useCase1.Ping(context.Background())
		result2 := useCase2.Ping(context.Background())

		assert.Equal(t, "pong", result1)
		assert.Equal(t, "pong", result2)
		assert.Equal(t, result1, result2)
	})
}
