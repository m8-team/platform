package health

type Status string

const (
	StatusUnknown   Status = "UNKNOWN"
	StatusHealthy   Status = "HEALTHY"
	StatusDegraded  Status = "DEGRADED"
	StatusUnhealthy Status = "UNHEALTHY"
)
