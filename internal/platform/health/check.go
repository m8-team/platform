package health

import "context"

type Check func(ctx context.Context) Result
