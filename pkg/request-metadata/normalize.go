package requestmetadata

import "strings"

// Normalize returns a trimmed, normalized copy of metadata.
func (m Metadata) Normalize() Metadata {
	return Metadata{
		Actor:          normalizeText(m.Actor),
		CorrelationID:  normalizeText(m.CorrelationID),
		IdempotencyKey: normalizeText(m.IdempotencyKey),
		RequestID:      normalizeText(m.RequestID),
		Source:         normalizeSource(m.Source),
	}
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}
