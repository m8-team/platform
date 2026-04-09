package requestmetadata

const (
	HeaderCorrelationID  = "X-Correlation-Id"
	HeaderRequestID      = "X-Request-Id"
	HeaderIdempotencyKey = "Idempotency-Key"

	MetadataCorrelationID  = "x-correlation-id"
	MetadataRequestID      = "x-request-id"
	MetadataIdempotencyKey = "idempotency-key"
)

var (
	correlationIDKeys = [...]string{HeaderCorrelationID, MetadataCorrelationID}
	requestIDKeys     = [...]string{HeaderRequestID, MetadataRequestID}
	idempotencyKeys   = [...]string{HeaderIdempotencyKey, MetadataIdempotencyKey}
)
