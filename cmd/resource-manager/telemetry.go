package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/m8-team/platform/internal/platform/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func telemetryModule(endpoint string) fx.Option {
	return fx.Module("telemetry", fx.Provide(func(lifecycle fx.Lifecycle) (*telemetry.Observer, error) {
		options := []sdktrace.TracerProviderOption{
			sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", "m8-resource-manager-api"))),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
		}
		if endpoint != "" {
			parsed, err := url.Parse(endpoint)
			if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				return nil, fmt.Errorf("invalid OTLP HTTP endpoint")
			}
			exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(endpoint))
			if err != nil {
				return nil, err
			}
			options = append(options, sdktrace.WithBatcher(exporter))
		}
		provider := sdktrace.NewTracerProvider(options...)
		lifecycle.Append(fx.Hook{OnStop: provider.Shutdown})
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return telemetry.New(logger, provider), nil
	}))
}
