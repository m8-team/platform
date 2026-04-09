package requestmetadata

import "context"

type contextKey struct{}

var metadataContextKey contextKey

// IntoContext stores metadata in a context.
func IntoContext(ctx context.Context, meta Metadata) context.Context {
	return context.WithValue(ctx, metadataContextKey, meta.Normalize())
}

// FromContext loads metadata from a context.
func FromContext(ctx context.Context) (Metadata, bool) {
	meta, ok := ctx.Value(metadataContextKey).(Metadata)
	if !ok {
		return Metadata{}, false
	}
	return meta, true
}

// MustFromContext loads metadata from a context and panics when it is absent.
func MustFromContext(ctx context.Context) Metadata {
	meta, ok := FromContext(ctx)
	if !ok {
		panic("requestmeta: metadata missing from context")
	}
	return meta
}
