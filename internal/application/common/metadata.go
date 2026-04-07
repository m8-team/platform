package common

// Metadata carries transport-level execution attributes that participate in
// command handling, auditing, and event publication.
type Metadata struct {
	Actor          string
	CorrelationID  string
	CausationID    string
	IdempotencyKey string
}
