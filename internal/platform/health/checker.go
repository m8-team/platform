package health

import "context"

type Checker interface {
	Check(ctx context.Context) Result
}

type CheckerFunc func(ctx context.Context) Result

func (f CheckerFunc) Check(ctx context.Context) Result {
	return f(ctx)
}
