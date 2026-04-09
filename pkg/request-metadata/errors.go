package requestmetadata

import "errors"

var (
	ErrMissingActor          = errors.New("requestmeta: missing actor")
	ErrMissingCorrelationID  = errors.New("requestmeta: missing correlation id")
	ErrMissingRequestID      = errors.New("requestmeta: missing request id")
	ErrMissingIdempotencyKey = errors.New("requestmeta: missing idempotency key")
	ErrInvalidSource         = errors.New("requestmeta: invalid source")
)
