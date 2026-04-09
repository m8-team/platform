package requestmetadata

// IDGenerator generates request-scoped identifiers.
type IDGenerator interface {
	NewID() string
}

// Carrier exposes metadata values from an arbitrary transport-specific carrier.
type Carrier interface {
	Get(key string) string
}
