package logging

import (
	"context"
	"github.com/m8-team/platform-sdk/metadata"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log/slog"
	"strings"
)

type Info struct{ Name, Version, Namespace, Environment, InstanceID string }
type Config struct {
	Text  bool
	Level slog.Level
}

func New(w io.Writer, info Info, cfg Config) *slog.Logger {
	options := &slog.HandlerOptions{Level: cfg.Level, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if sensitive(a.Key) {
			return slog.String(a.Key, "[REDACTED]")
		}
		a.Value = a.Value.Resolve()
		if a.Value.Kind() == slog.KindAny {
			a.Value = slog.AnyValue(sanitize(a.Value.Any()))
		}
		return a
	}}
	var handler slog.Handler = slog.NewJSONHandler(w, options)
	if cfg.Text {
		handler = slog.NewTextHandler(w, options)
	}
	return slog.New(handler).With("service.name", info.Name, "service.version", info.Version,
		"service.namespace", info.Namespace, "deployment.environment.name", info.Environment, "service.instance.id", info.InstanceID)
}

// WithContext adds metadata, not dependencies. The logger is explicitly supplied.
func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	span := trace.SpanContextFromContext(ctx)
	attrs := []any{"request_id", metadata.RequestID(ctx)}
	if span.IsValid() {
		attrs = append(attrs, "trace_id", span.TraceID().String(), "span_id", span.SpanID().String())
	}
	if p, ok := metadata.PrincipalFrom(ctx); ok {
		attrs = append(attrs, "tenant_id", p.TenantID)
	}
	return logger.With(attrs...)
}

func sensitive(key string) bool {
	key = strings.ToLower(key)
	for _, s := range []string{"authorization", "cookie", "password", "token", "secret", "credit", "card_number", "email", "phone", "personal"} {
		if strings.Contains(key, s) {
			return true
		}
	}
	return false
}
func sanitize(v any) any {
	switch value := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(value))
		for k, v := range value {
			if sensitive(k) {
				result[k] = "[REDACTED]"
			} else {
				result[k] = sanitize(v)
			}
		}
		return result
	case map[string]string:
		result := make(map[string]string, len(value))
		for k, v := range value {
			if sensitive(k) {
				result[k] = "[REDACTED]"
			} else {
				result[k] = v
			}
		}
		return result
	case []any:
		result := make([]any, len(value))
		for i, v := range value {
			result[i] = sanitize(v)
		}
		return result
	case string, bool, int, int64, uint64, float64, nil:
		return value
	default:
		return "[opaque value omitted]"
	}
}
