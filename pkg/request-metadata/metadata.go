package requestmetadata

// Metadata is canonical request-scoped operational metadata.
type Metadata struct {
	Actor          string
	CorrelationID  string
	IdempotencyKey string
	RequestID      string
	Source         Source
}

// HasIdempotencyKey reports whether metadata contains a non-empty idempotency key.
func (m Metadata) HasIdempotencyKey() bool {
	return normalizeText(m.IdempotencyKey) != ""
}
