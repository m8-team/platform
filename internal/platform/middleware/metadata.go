package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"

	"github.com/m8platform/platform/internal/platform/correlation"
)

type RequestMetadata struct {
	Actor          string
	CorrelationID  string
	CausationID    string
	IdempotencyKey string
}

func FromGRPCContext(ctx context.Context) RequestMetadata {
	out := RequestMetadata{
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
		CausationID:   correlation.CausationIDFromContext(ctx),
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return out
	}
	if values := md.Get("x-actor"); len(values) > 0 {
		out.Actor = values[0]
	}
	if values := md.Get("x-correlation-id"); len(values) > 0 {
		out.CorrelationID = values[0]
	}
	if values := md.Get("x-causation-id"); len(values) > 0 {
		out.CausationID = values[0]
	}
	if values := md.Get("idempotency-key"); len(values) > 0 {
		out.IdempotencyKey = values[0]
	}
	return out
}

func FromHTTPRequest(r *http.Request) RequestMetadata {
	return RequestMetadata{
		Actor:          r.Header.Get("X-Actor"),
		CorrelationID:  firstNonEmpty(r.Header.Get("X-Correlation-Id"), correlation.CorrelationIDFromContext(r.Context())),
		CausationID:    firstNonEmpty(r.Header.Get("X-Causation-Id"), correlation.CausationIDFromContext(r.Context())),
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
