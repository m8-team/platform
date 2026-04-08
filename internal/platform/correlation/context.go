package correlation

import "context"

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	causationIDKey   contextKey = "causation_id"
)

func WithCorrelationID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, correlationIDKey, value)
}

func WithCausationID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, causationIDKey, value)
}

func CorrelationIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDKey).(string)
	return value
}

func CausationIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(causationIDKey).(string)
	return value
}
