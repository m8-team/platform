package middleware

import (
	"context"
	"net/http"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/ports"
	"google.golang.org/grpc/metadata"
)

const (
	HeaderIdempotencyKey   = "Idempotency-Key"
	MetadataIdempotencyKey = "idempotency-key"
	MetadataCorrelationID  = "correlation-id"
	MetadataCausationID    = "causation-id"
)

type MetadataActorResolver struct {
	Keys []string
}

func NewMetadataActorResolver(keys ...string) MetadataActorResolver {
	if len(keys) == 0 {
		keys = []string{"actor", "x-actor", "principal"}
	}
	return MetadataActorResolver{Keys: keys}
}

func (r MetadataActorResolver) Resolve(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	for _, key := range r.Keys {
		values := md.Get(key)
		if len(values) > 0 && values[0] != "" {
			return values[0], nil
		}
	}
	return "", nil
}

func CommandMetadata(ctx context.Context, resolver ports.ActorResolver) appcommon.Metadata {
	actor := ""
	if resolver != nil {
		actor, _ = resolver.Resolve(ctx)
	}
	return appcommon.Metadata{
		Actor:          actor,
		IdempotencyKey: firstMetadata(ctx, MetadataIdempotencyKey),
		CorrelationID:  firstMetadata(ctx, MetadataCorrelationID),
		CausationID:    firstMetadata(ctx, MetadataCausationID),
	}
}

func IdempotencyKeyFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get(HeaderIdempotencyKey)
}

func firstMetadata(ctx context.Context, key string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
