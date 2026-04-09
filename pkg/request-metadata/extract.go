package requestmetadata

// FromCarrier extracts normalized metadata from an arbitrary carrier.
func FromCarrier(actor string, source Source, carrier Carrier, gen IDGenerator) Metadata {
	return newMetadata(
		actor,
		carrierValue(carrier, correlationIDKeys[:]...),
		carrierValue(carrier, requestIDKeys[:]...),
		carrierValue(carrier, idempotencyKeys[:]...),
		source,
		gen,
	)
}

func carrierValue(carrier Carrier, keys ...string) string {
	if carrier == nil {
		return ""
	}
	for _, key := range keys {
		value := normalizeText(carrier.Get(key))
		if value != "" {
			return value
		}
	}
	return ""
}
