package platform

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/m8platform/platform/internal/platform/middleware"
)

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type UUIDGenerator struct{}

func (UUIDGenerator) NewString() string {
	return uuid.NewString()
}

type ActorResolver struct{}

func (ActorResolver) Resolve(ctx context.Context) (string, error) {
	return middleware.FromGRPCContext(ctx).Actor, nil
}
