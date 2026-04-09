package requestmetadata

// NewForCLI creates normalized metadata for CLI flows.
func NewForCLI(actor string, gen IDGenerator) Metadata {
	return newMetadata(actor, "", "", "", SourceCLI, gen)
}

// NewForWorker creates normalized metadata for worker flows.
func NewForWorker(actor, correlationID, idempotencyKey string, gen IDGenerator) Metadata {
	return newMetadata(actor, correlationID, "", idempotencyKey, SourceWorker, gen)
}

func newMetadata(actor, correlationID, requestID, idempotencyKey string, source Source, gen IDGenerator) Metadata {
	return Metadata{
		Actor:          normalizeText(actor),
		CorrelationID:  resolveID(correlationID, gen),
		IdempotencyKey: normalizeText(idempotencyKey),
		RequestID:      resolveID(requestID, gen),
		Source:         normalizeSource(source),
	}
}

func resolveID(value string, gen IDGenerator) string {
	normalized := normalizeText(value)
	if normalized != "" || gen == nil {
		return normalized
	}
	return normalizeText(gen.NewID())
}
