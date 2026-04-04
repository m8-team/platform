package observability

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	CheckAccessLatency        *prometheus.HistogramVec
	CheckAccessCacheHitRatio  prometheus.Gauge
	BindingUpdateLatency      *prometheus.HistogramVec
	SpiceDBWriteLatency       *prometheus.HistogramVec
	SpiceDBCheckLatency       *prometheus.HistogramVec
	TopicConsumerLag          prometheus.Gauge
	ProjectionRebuildDuration prometheus.Histogram
	SupportGrantAutoRevoke    prometheus.Histogram
	WorkflowActivityFailures  *prometheus.CounterVec
}

func NewMetrics(registry prometheus.Registerer) *Metrics {
	metrics := &Metrics{
		CheckAccessLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "check_access_latency_seconds",
				Help:    "Latency of CheckAccess requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"decision"},
		),
		CheckAccessCacheHitRatio: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "check_access_cache_hit_ratio",
				Help: "Current ratio of CheckAccess cache hits.",
			},
		),
		BindingUpdateLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "binding_update_latency_seconds",
				Help:    "Latency of access binding updates.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		SpiceDBWriteLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "spicedb_write_latency_seconds",
				Help:    "Latency of SpiceDB relationship writes.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"result"},
		),
		SpiceDBCheckLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "spicedb_check_latency_seconds",
				Help:    "Latency of SpiceDB permission checks.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"result"},
		),
		TopicConsumerLag: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "topic_consumer_lag",
				Help: "Lag of YDB Topics consumers.",
			},
		),
		ProjectionRebuildDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "projection_rebuild_duration_seconds",
				Help:    "Duration of projection rebuild workflows.",
				Buckets: prometheus.DefBuckets,
			},
		),
		SupportGrantAutoRevoke: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "support_grant_auto_revoke_latency_seconds",
				Help:    "Latency of support grant auto revoke activities.",
				Buckets: prometheus.DefBuckets,
			},
		),
		WorkflowActivityFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "workflow_activity_failures_total",
				Help: "Count of workflow activity failures.",
			},
			[]string{"workflow", "activity"},
		),
	}

	registry.MustRegister(
		metrics.CheckAccessLatency,
		metrics.CheckAccessCacheHitRatio,
		metrics.BindingUpdateLatency,
		metrics.SpiceDBWriteLatency,
		metrics.SpiceDBCheckLatency,
		metrics.TopicConsumerLag,
		metrics.ProjectionRebuildDuration,
		metrics.SupportGrantAutoRevoke,
		metrics.WorkflowActivityFailures,
	)

	return metrics
}
