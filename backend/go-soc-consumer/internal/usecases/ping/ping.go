package ping

import "context"

// UseCaseImpl implements the UseCase interface
type UseCaseImpl struct{}

// NewUseCase creates a new ping use case implementation
func NewUseCase() UseCase {
	return &UseCaseImpl{}
}

// Execute returns "pong" response
func (uc *UseCaseImpl) Execute(ctx context.Context) string {
	return "pong"
}