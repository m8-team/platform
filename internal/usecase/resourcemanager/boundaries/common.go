package boundaries

type RequestMetadata struct {
	Actor          string
	CorrelationID  string
	CausationID    string
	IdempotencyKey string
}
