package common

type RequestMetadata struct {
	Actor          string
	CorrelationID  string
	CausationID    string
	IdempotencyKey string
}
