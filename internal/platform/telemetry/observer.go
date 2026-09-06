// Package telemetry instruments application boundaries without global providers.
package telemetry

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

type counters struct{ calls, failures, nanoseconds atomic.Int64 }

// Observer uses bounded operation names, never IDs, paths or payloads as metric
// dimensions. A nil Observer is disabled (useful for pure application tests).
type Observer struct {
	logger  *slog.Logger
	tracer  trace.Tracer
	mu      sync.Mutex
	metrics map[string]*counters
}

func New(logger *slog.Logger, provider trace.TracerProvider) *Observer {
	return &Observer{logger: logger, tracer: provider.Tracer("m8.application"), metrics: make(map[string]*counters)}
}

func (o *Observer) Start(ctx context.Context, operation string) (context.Context, func(error)) {
	if o == nil {
		return ctx, func(error) {}
	}
	started := time.Now()
	ctx, span := o.tracer.Start(ctx, operation)
	o.mu.Lock()
	count := o.metrics[operation]
	if count == nil {
		count = &counters{}
		o.metrics[operation] = count
	}
	o.mu.Unlock()
	return ctx, func(err error) {
		elapsed := time.Since(started)
		count.calls.Add(1)
		count.nanoseconds.Add(elapsed.Nanoseconds())
		outcome := "ok"
		level := slog.LevelInfo
		if err != nil {
			outcome = "error"
			level = slog.LevelWarn
			count.failures.Add(1)
			// Error strings may contain caller-controlled values; keep traces/logs safe.
			span.SetStatus(codes.Error, "operation failed")
		}
		o.logger.Log(ctx, level, "application operation",
			"operation", operation, "outcome", outcome, "duration", elapsed,
			"request_id", ctx.Value(requestIDKey{}), "trace_id", span.SpanContext().TraceID().String())
		span.End()
	}
}

// ServeHTTP exposes process-local counters as JSON on the management listener.
func (o *Observer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	result := make(map[string]map[string]float64)
	o.mu.Lock()
	for name, count := range o.metrics {
		result[name] = map[string]float64{"calls": float64(count.calls.Load()), "failures": float64(count.failures.Load()), "duration_seconds": float64(count.nanoseconds.Load()) / 1e9}
	}
	o.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func requestContext(ctx context.Context, requestID string) context.Context {
	if _, err := uuid.Parse(requestID); err != nil {
		requestID = uuid.NewString()
	}
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx = requestContext(ctx, r.Header.Get("X-Request-ID"))
		w.Header().Set("X-Request-ID", ctx.Value(requestIDKey{}).(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Unary(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	carrier := propagation.MapCarrier{}
	for _, key := range []string{"traceparent", "tracestate", "x-request-id"} {
		if values := md.Get(key); len(values) > 0 {
			carrier[key] = values[0]
		}
	}
	ctx = propagation.TraceContext{}.Extract(ctx, carrier)
	return handler(requestContext(ctx, carrier["x-request-id"]), req)
}
