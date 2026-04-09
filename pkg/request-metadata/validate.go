package requestmetadata

// ValidateForQuery validates metadata required for query use cases.
func (m Metadata) ValidateForQuery() error {
	return m.validate(false)
}

// ValidateForCommand validates metadata required for command use cases.
func (m Metadata) ValidateForCommand(requireIdempotencyKey bool) error {
	return m.validate(requireIdempotencyKey)
}

func (m Metadata) validate(requireIdempotencyKey bool) error {
	normalized := m.Normalize()
	if !isValidSource(normalized.Source) {
		return ErrInvalidSource
	}
	if normalized.Actor == "" {
		return ErrMissingActor
	}
	if normalized.CorrelationID == "" {
		return ErrMissingCorrelationID
	}
	if normalized.RequestID == "" {
		return ErrMissingRequestID
	}
	if requireIdempotencyKey && normalized.IdempotencyKey == "" {
		return ErrMissingIdempotencyKey
	}
	return nil
}
