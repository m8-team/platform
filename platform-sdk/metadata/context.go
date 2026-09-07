package metadata

import "context"

type Principal struct {
	SubjectID, TenantID string
	Roles               []string
}
type principalKey struct{}
type requestKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	p.Roles = append([]string(nil), p.Roles...)
	return context.WithValue(ctx, principalKey{}, p)
}
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	p.Roles = append([]string(nil), p.Roles...)
	return p, ok
}
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey{}, id)
}
func RequestID(ctx context.Context) string { id, _ := ctx.Value(requestKey{}).(string); return id }
