package ping

import "context"

// UseCase defines the contract for ping use case operations
type UseCase interface {
	Execute(ctx context.Context) string
}