// Package kafka keeps the native franz-go client accessible and retry policy explicit.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/m8-team/platform-sdk/outbox"
	"github.com/m8-team/platform-sdk/retry"
	"github.com/m8-team/platform-sdk/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"runtime/debug"
	"time"
)

type Message struct {
	Topic      string
	Key, Value []byte
	Headers    map[string]string
}
type Producer struct {
	Client    *kgo.Client
	telemetry *telemetry.Providers
}

// NewProducer accepts native options: callers explicitly choose brokers, TLS,
// SASL, partitioning, compression, delivery timeout and retry budget. Idempotent
// writes and all-ISR acknowledgements remain franz-go defaults. Cancellation of
// in-flight idempotent records is enabled so a lost broker response cannot exceed
// the caller's budget indefinitely. A cancelled record may already be delivered:
// replay must retain its event ID. Use a native client for Kafka transactions.
func NewProducer(t *telemetry.Providers, options ...kgo.Opt) (*Producer, error) {
	if t == nil {
		return nil, errors.New("telemetry required")
	}
	options = append([]kgo.Opt{kgo.AllowIdempotentProduceCancellation()}, options...)
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, err
	}
	return &Producer{client, t}, nil
}
func (p *Producer) Publish(ctx context.Context, event outbox.Event) error {
	ctx = p.telemetry.Propagator.Extract(ctx, propagation.MapCarrier(event.Headers))
	headers := map[string]string{}
	for k, v := range event.Headers {
		headers[k] = v
	}
	headers["event_id"] = event.ID
	headers["message_id"] = event.ID
	headers["event_type"] = event.Type
	headers["aggregate_type"] = event.AggregateType
	headers["aggregate_id"] = event.AggregateID
	return p.Send(ctx, Message{Topic: event.Topic, Key: []byte(event.AggregateID), Value: event.Payload, Headers: headers})
}
func (p *Producer) Send(ctx context.Context, msg Message) (result error) {
	calls, err := p.telemetry.Meter.Int64Counter("kafka_producer_messages_total")
	if err != nil {
		return err
	}
	defer func() {
		outcome := "success"
		if result != nil {
			outcome = "error"
		}
		calls.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
	}()
	ctx, span := p.telemetry.Tracer.Start(ctx, "kafka publish", trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(attribute.String("messaging.system", "kafka")))
	defer span.End()
	carrier := propagation.MapCarrier{}
	for k, v := range msg.Headers {
		carrier[k] = v
	}
	p.telemetry.Propagator.Inject(ctx, carrier)
	record := &kgo.Record{Topic: msg.Topic, Key: msg.Key, Value: msg.Value}
	for k, v := range carrier {
		record.Headers = append(record.Headers, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}
	if err := p.Client.ProduceSync(ctx, record).FirstErr(); err != nil {
		span.SetStatus(codes.Error, "publish failed")
		return fmt.Errorf("kafka delivery: %w", err)
	}
	return nil
}

type HandlerFunc func(context.Context, Message) error
type Consumer struct {
	Client    *kgo.Client
	Handler   HandlerFunc
	Telemetry *telemetry.Providers
	Logger    *slog.Logger
	Timeout   time.Duration
	Retry     retry.Policy
	// DeadLetter must confirm durable publication before returning nil. Nil means
	// stop on exhausted/terminal failure without committing the source offset.
	DeadLetter func(context.Context, Message, error) error
}

// NewConsumer disables automatic commits and blocks rebalances until processing
// and commit finish. Processing is deliberately serial (bounded concurrency=1).
// Scale with partitions/instances; do not commit past unfinished partition work.
func NewConsumer(options ...kgo.Opt) (*kgo.Client, error) {
	options = append(options, kgo.DisableAutoCommit(), kgo.BlockRebalanceOnPoll())
	return kgo.NewClient(options...)
}
func (c *Consumer) Run(ctx context.Context) error {
	if c.Client == nil || c.Handler == nil || c.Telemetry == nil || c.Logger == nil || c.Timeout <= 0 || c.Retry.MaxAttempts < 1 {
		return errors.New("invalid consumer configuration")
	}
	for {
		fetches := c.Client.PollRecords(ctx, 1)
		if ctx.Err() != nil {
			c.Client.AllowRebalance()
			return ctx.Err()
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			c.Client.AllowRebalance()
			return fmt.Errorf("poll kafka: %w", errs[0].Err)
		}
		records := fetches.Records()
		for _, record := range records {
			err := c.process(ctx, record)
			if err != nil {
				c.Client.AllowRebalance()
				return err
			}
		}
		c.Client.AllowRebalance()
	}
}
func (c *Consumer) process(ctx context.Context, record *kgo.Record) (result error) {
	calls, err := c.Telemetry.Meter.Int64Counter("kafka_consumer_messages_total")
	if err != nil {
		return err
	}
	duration, err := c.Telemetry.Meter.Float64Histogram("kafka_consumer_processing_duration_seconds", metric.WithUnit("s"))
	if err != nil {
		return err
	}
	started := time.Now()
	defer func() {
		outcome := "success"
		if result != nil {
			outcome = "error"
		}
		attrs := metric.WithAttributes(attribute.String("outcome", outcome))
		calls.Add(ctx, 1, attrs)
		duration.Record(ctx, time.Since(started).Seconds(), attrs)
	}()
	msg := Message{Topic: record.Topic, Key: record.Key, Value: record.Value, Headers: map[string]string{}}
	for _, h := range record.Headers {
		msg.Headers[h.Key] = string(h.Value)
	}
	ctx = c.Telemetry.Propagator.Extract(ctx, propagation.MapCarrier(msg.Headers))
	ctx, span := c.Telemetry.Tracer.Start(ctx, "kafka process", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	attemptCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	err = retry.Do(attemptCtx, c.Retry, func(ctx context.Context) (err error) {
		defer func() {
			if recover() != nil {
				c.Logger.ErrorContext(ctx, "consumer panic", "stack", string(debug.Stack()))
				err = errors.New("consumer panic")
			}
		}()
		return c.Handler(ctx, msg)
	})
	if err == nil && attemptCtx.Err() != nil {
		err = attemptCtx.Err()
	}
	if err != nil {
		span.SetStatus(codes.Error, "handler failed")
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if c.DeadLetter == nil {
			return fmt.Errorf("handle kafka: %w", err)
		}
		dlqCtx, cancelDLQ := context.WithTimeout(ctx, c.Timeout)
		defer cancelDLQ()
		if err = c.DeadLetter(dlqCtx, msg, err); err != nil {
			return fmt.Errorf("publish dead letter: %w", err)
		}
	}
	if err := c.Client.CommitRecords(ctx, record); err != nil {
		return fmt.Errorf("commit kafka offset: %w", err)
	}
	return nil
}
