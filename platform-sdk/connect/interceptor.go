// Package connect provides a unary server policy chain without owning handlers.
package connect

import (
	"buf.build/go/protovalidate"
	rpc "connectrpc.com/connect"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/m8-team/platform-sdk/fault"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/metadata"
	"github.com/m8-team/platform-sdk/telemetry"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type Principal = metadata.Principal
type Authenticator interface {
	Authenticate(context.Context, http.Header) (Principal, error)
}
type CheckRequest struct {
	Principal Principal
	Procedure string
}
type Authorizer interface {
	Check(context.Context, CheckRequest) error
}
type Limiter interface {
	Allow(context.Context, Principal, string) error
}
type Config struct {
	Logger         *slog.Logger
	Telemetry      *telemetry.Providers
	Authentication Authenticator
	Authorization  Authorizer
	Limit          Limiter
	RequestTimeout time.Duration
}
type Interceptor struct {
	cfg       Config
	validator protovalidate.Validator
	calls     metric.Int64Counter
	duration  metric.Float64Histogram
}

func New(cfg Config) (*Interceptor, error) {
	if cfg.Logger == nil || cfg.Telemetry == nil || cfg.Authentication == nil || cfg.Authorization == nil {
		return nil, errors.New("logger, telemetry, authentication and authorization required")
	}
	if cfg.RequestTimeout < 0 {
		return nil, errors.New("request timeout must be positive")
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 20 * time.Second
	}
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	calls, err := cfg.Telemetry.Meter.Int64Counter("rpc_server_requests_total")
	if err != nil {
		return nil, err
	}
	duration, err := cfg.Telemetry.Meter.Float64Histogram("rpc_server_request_duration_seconds", metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	return &Interceptor{cfg: cfg, validator: v, calls: calls, duration: duration}, nil
}
func (i *Interceptor) WrapUnary(next rpc.UnaryFunc) rpc.UnaryFunc {
	return func(ctx context.Context, request rpc.AnyRequest) (response rpc.AnyResponse, err error) {
		ctx, cancel := context.WithTimeout(ctx, i.cfg.RequestTimeout)
		defer cancel()
		started := time.Now()
		procedure := request.Spec().Procedure
		ctx = i.cfg.Telemetry.Propagator.Extract(ctx, propagation.HeaderCarrier(request.Header()))
		id := request.Header().Get("X-Request-ID")
		if len(id) != 32 {
			var b [16]byte
			_, _ = rand.Read(b[:])
			id = hex.EncodeToString(b[:])
		} else if _, e := hex.DecodeString(id); e != nil {
			var b [16]byte
			_, _ = rand.Read(b[:])
			id = hex.EncodeToString(b[:])
		}
		ctx = metadata.WithRequestID(ctx, id)
		ctx, span := i.cfg.Telemetry.Tracer.Start(ctx, procedure, trace.WithSpanKind(trace.SpanKindServer))
		defer func() {
			elapsed := time.Since(started).Seconds()
			code := rpc.CodeOf(err).String()
			if err == nil {
				code = "ok"
			} else {
				span.SetStatus(otelcodes.Error, code)
			}
			attrs := metric.WithAttributes(attribute.String("rpc.method", procedure), attribute.String("rpc.status", code))
			i.calls.Add(ctx, 1, attrs)
			i.duration.Record(ctx, elapsed, attrs)
			logging.WithContext(ctx, i.cfg.Logger).Info("rpc completed", "operation", procedure, "status", code, "duration_seconds", elapsed)
			span.End()
		}()
		defer func() {
			if recover() != nil {
				logging.WithContext(ctx, i.cfg.Logger).Error("handler panic", "stack", string(debug.Stack()))
				err = fault.New(fault.Internal, "internal error", nil)
				response = nil
			}
			if err != nil {
				mapped := ToError(err)
				mapped.Meta().Set("X-Request-ID", id)
				err = mapped
			} else if response != nil {
				response.Header().Set("X-Request-ID", id)
			}
		}()
		principal, err := i.cfg.Authentication.Authenticate(ctx, request.Header())
		if err != nil {
			return nil, err
		}
		if principal.SubjectID == "" || principal.TenantID == "" {
			return nil, fault.New(fault.Unauthenticated, "authentication required", nil)
		}
		ctx = metadata.WithPrincipal(ctx, principal)
		if err := i.cfg.Authorization.Check(ctx, CheckRequest{Principal: principal, Procedure: procedure}); err != nil {
			return nil, err
		}
		if i.cfg.Limit != nil {
			if err := i.cfg.Limit.Allow(ctx, principal, procedure); err != nil {
				return nil, err
			}
		}
		message, ok := request.Any().(proto.Message)
		if !ok {
			return nil, fault.New(fault.Internal, "protobuf request required", nil)
		}
		if err := i.validator.Validate(message); err != nil {
			// Protovalidate's native violations retain field paths without request values.
			invalid := rpc.NewError(rpc.CodeInvalidArgument, errors.New("invalid request"))
			var validation *protovalidate.ValidationError
			if errors.As(err, &validation) {
				if detail, e := rpc.NewErrorDetail(validation.ToProto()); e == nil {
					invalid.AddDetail(detail)
				}
			}
			return nil, invalid
		}
		return next(ctx, request)
	}
}
func (i *Interceptor) WrapStreamingClient(next rpc.StreamingClientFunc) rpc.StreamingClientFunc {
	return next
}
func (i *Interceptor) WrapStreamingHandler(_ rpc.StreamingHandlerFunc) rpc.StreamingHandlerFunc {
	return func(context.Context, rpc.StreamingHandlerConn) error {
		return rpc.NewError(rpc.CodeUnimplemented, errors.New("streaming requires a dedicated policy chain"))
	}
}

func ToError(err error) *rpc.Error {
	if errors.Is(err, context.Canceled) {
		return rpc.NewError(rpc.CodeCanceled, errors.New("request canceled"))
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return rpc.NewError(rpc.CodeDeadlineExceeded, errors.New("deadline exceeded"))
	}
	var existing *rpc.Error
	if errors.As(err, &existing) {
		if existing.Code() == rpc.CodeInternal || existing.Code() == rpc.CodeUnknown {
			return rpc.NewError(rpc.CodeInternal, errors.New("internal error"))
		}
		return existing
	}
	var app *fault.Error
	if !errors.As(err, &app) {
		return rpc.NewError(rpc.CodeInternal, errors.New("internal error"))
	}
	code := map[fault.Code]rpc.Code{fault.InvalidArgument: rpc.CodeInvalidArgument, fault.NotFound: rpc.CodeNotFound, fault.AlreadyExists: rpc.CodeAlreadyExists, fault.PermissionDenied: rpc.CodePermissionDenied, fault.Unauthenticated: rpc.CodeUnauthenticated, fault.Conflict: rpc.CodeAborted, fault.FailedPrecondition: rpc.CodeFailedPrecondition, fault.ResourceExhausted: rpc.CodeResourceExhausted, fault.Unavailable: rpc.CodeUnavailable}[app.Code]
	if code == 0 {
		return rpc.NewError(rpc.CodeInternal, errors.New("internal error"))
	}
	result := rpc.NewError(code, errors.New(app.Message))
	if len(app.Violations) > 0 {
		detail := &errdetails.BadRequest{}
		for _, v := range app.Violations {
			detail.FieldViolations = append(detail.FieldViolations, &errdetails.BadRequest_FieldViolation{Field: v.Field, Description: v.Description})
		}
		if d, e := rpc.NewErrorDetail(detail); e == nil {
			result.AddDetail(d)
		}
	}
	return result
}

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch ToError(err).Code() {
	case rpc.CodeInvalidArgument:
		return 400
	case rpc.CodeUnauthenticated:
		return 401
	case rpc.CodePermissionDenied:
		return 403
	case rpc.CodeNotFound:
		return 404
	case rpc.CodeAlreadyExists, rpc.CodeAborted:
		return 409
	case rpc.CodeFailedPrecondition:
		return 412
	case rpc.CodeResourceExhausted:
		return 429
	case rpc.CodeUnavailable:
		return 503
	case rpc.CodeDeadlineExceeded:
		return 504
	default:
		return 500
	}
}
