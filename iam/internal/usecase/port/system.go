package port

import (
	"errors"
	"time"
)

var ErrAuthorizationUnavailable = errors.New("authorization runtime is unavailable")

type Clock interface {
	Now() time.Time
}
