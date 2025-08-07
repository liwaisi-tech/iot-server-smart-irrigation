package ping

import "context"

// PingUseCase defines the contract for ping use case operations
type PingUseCase interface {
	Ping(ctx context.Context) string
}

// UseCaseImpl implements the UseCase interface
type useCaseImpl struct{}

// NewUseCase creates a new ping use case implementation
func NewUseCase() PingUseCase {
	return &useCaseImpl{}
}

// Ping returns "pong" response
func (uc *useCaseImpl) Ping(ctx context.Context) string {
	return "pong"
}
