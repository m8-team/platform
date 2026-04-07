package ports

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

type UUIDGenerator interface {
	NewString() string
}

type ActorResolver interface {
	Resolve(ctx context.Context) (string, error)
}

type FilterParser interface {
	Validate(raw string) error
}

type OrderParser interface {
	Validate(raw string) error
}
