package ports

import "context"

type ActorResolver interface {
	Resolve(ctx context.Context) (string, error)
}
