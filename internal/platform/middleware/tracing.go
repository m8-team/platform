package middleware

import "context"

func StartSpan(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}
