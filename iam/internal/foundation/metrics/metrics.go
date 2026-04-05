package metrics

import (
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics = observability.Metrics

func New(registry prometheus.Registerer) *Metrics {
	return observability.NewMetrics(registry)
}
