package telemetry

import (
	"context"
	"errors"
	"github.com/m8-team/platform-sdk/logging"
	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Config struct {
	Endpoint    string
	Insecure    bool
	SampleRatio float64
}
type Providers struct {
	Traces     *sdktrace.TracerProvider
	Metrics    *sdkmetric.MeterProvider
	Tracer     trace.Tracer
	Meter      metric.Meter
	Handler    http.Handler
	Propagator propagation.TextMapPropagator
}

func New(ctx context.Context, info logging.Info, cfg Config) (*Providers, error) {
	if cfg.SampleRatio < 0 || cfg.SampleRatio > 1 {
		return nil, errors.New("sampling must be between zero and one")
	}
	res := resource.NewSchemaless(attribute.String("service.name", info.Name), attribute.String("service.version", info.Version), attribute.String("service.namespace", info.Namespace), attribute.String("deployment.environment.name", info.Environment), attribute.String("service.instance.id", info.InstanceID))
	registry := promclient.NewRegistry()
	metrics, err := prometheus.New(prometheus.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(metrics))
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res), sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio)))}
	if cfg.Endpoint != "" {
		exporterOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
		if err != nil {
			_ = meterProvider.Shutdown(ctx)
			return nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(exporter, sdktrace.WithMaxQueueSize(2048)))
	}
	traces := sdktrace.NewTracerProvider(opts...)
	p := &Providers{Traces: traces, Metrics: meterProvider, Tracer: traces.Tracer("platform-sdk"), Meter: meterProvider.Meter("platform-sdk"), Handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}), Propagator: propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})}
	infoMetric, err := p.Meter.Int64Gauge("application_info")
	if err != nil {
		_ = p.Shutdown(ctx)
		return nil, err
	}
	infoMetric.Record(ctx, 1)
	return p, nil
}
func (p *Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.Traces.Shutdown(ctx), p.Metrics.Shutdown(ctx))
}
