package requestmeta

// RequestMetadata is stable request-scoped operational metadata shared by use cases.
type RequestMetadata struct {
	Actor          string
	CorrelationID  string
	CausationID    string
	IdempotencyKey string
}
