// Package outbox publishes durable events with at-least-once semantics.
package outbox

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"log/slog"
	mathrand "math/rand/v2"
	"time"
)

type Event struct {
	ID            string            `json:"id"`
	Topic         string            `json:"topic"`
	AggregateType string            `json:"aggregate_type"`
	AggregateID   string            `json:"aggregate_id"`
	Type          string            `json:"type"`
	Payload       []byte            `json:"payload"`
	Headers       map[string]string `json:"headers"`
}
type Publisher interface {
	Publish(context.Context, Event) error
}

func Insert(ctx context.Context, tx *sql.Tx, event Event) error {
	if event.ID == "" || event.Topic == "" || event.AggregateID == "" || event.AggregateType == "" || event.Type == "" {
		return errors.New("incomplete outbox event")
	}
	headers, err := json.Marshal(event.Headers)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO platform.outbox (id,topic,aggregate_type,aggregate_id,event_type,payload,headers) VALUES ($1,$2,$3,$4,$5,$6,$7)`, event.ID, event.Topic, event.AggregateType, event.AggregateID, event.Type, event.Payload, headers)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

type Config struct {
	PollInterval, PublishTimeout, Lease time.Duration
	MaxAttempts                         int
	Meter                               metric.Meter
}
type Worker struct {
	db        *sql.DB
	publisher Publisher
	log       *slog.Logger
	cfg       Config
	published metric.Int64Counter
}

func New(db *sql.DB, p Publisher, log *slog.Logger, cfg Config) (*Worker, error) {
	if db == nil || p == nil || log == nil || cfg.PollInterval <= 0 || cfg.PublishTimeout <= 0 || cfg.Lease <= cfg.PublishTimeout || cfg.MaxAttempts < 1 {
		return nil, errors.New("invalid outbox dependencies or limits")
	}
	if cfg.Meter == nil {
		cfg.Meter = noop.NewMeterProvider().Meter("outbox")
	}
	published, err := cfg.Meter.Int64Counter("outbox_publish_total")
	if err != nil {
		return nil, err
	}
	return &Worker{db: db, publisher: p, log: log, cfg: cfg, published: published}, nil
}

// Run owns no detached goroutines. A cancelled publish remains reclaimable after
// its lease. Concurrent workers fence updates by token, not by event ID alone.
func (w *Worker) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		found, err := w.Step(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			w.log.ErrorContext(ctx, "outbox iteration failed", "error", err)
		}
		if !found || err != nil {
			timer := time.NewTimer(w.cfg.PollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}

func (w *Worker) Step(ctx context.Context) (found bool, result error) {
	ctx, cancelStep := context.WithTimeout(ctx, w.cfg.Lease)
	defer cancelStep()
	defer func() {
		if found {
			outcome := "success"
			if result != nil {
				outcome = "error"
			}
			w.published.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
		}
	}()
	token := rand.Text()
	var event Event
	var headers []byte
	var attempts int
	err := w.db.QueryRowContext(ctx, `WITH candidate AS (
 SELECT id FROM platform.outbox WHERE published_at IS NULL AND dead_at IS NULL AND next_attempt_at<=clock_timestamp() AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1
 ) UPDATE platform.outbox o SET lease_token=$1,lease_until=clock_timestamp()+$2::bigint*interval '1 millisecond',attempts=o.attempts+1 FROM candidate c WHERE o.id=c.id RETURNING o.id,o.topic,o.aggregate_type,o.aggregate_id,o.event_type,o.payload,o.headers,o.attempts`, token, w.cfg.Lease.Milliseconds()).Scan(&event.ID, &event.Topic, &event.AggregateType, &event.AggregateID, &event.Type, &event.Payload, &headers, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim outbox: %w", err)
	}
	err = json.Unmarshal(headers, &event.Headers)
	if err == nil {
		publishCtx, cancel := context.WithTimeout(ctx, w.cfg.PublishTimeout)
		err = w.publisher.Publish(publishCtx, event)
		cancel()
	}
	if err != nil {
		delay := min(time.Second*time.Duration(1<<min(attempts, 8)), 5*time.Minute)
		delay = delay/2 + time.Duration(mathrand.Int64N(int64(delay/2)))
		_, saveErr := w.db.ExecContext(ctx, `UPDATE platform.outbox SET lease_until=NULL,lease_token=NULL,next_attempt_at=clock_timestamp()+$3::bigint*interval '1 millisecond',dead_at=CASE WHEN attempts>=$4 THEN clock_timestamp() ELSE NULL END WHERE id=$1 AND lease_token=$2`, event.ID, token, delay.Milliseconds(), w.cfg.MaxAttempts)
		return true, errors.Join(fmt.Errorf("publish outbox: %w", err), saveErr)
	}
	updated, err := w.db.ExecContext(ctx, `UPDATE platform.outbox SET published_at=clock_timestamp(),lease_until=NULL,lease_token=NULL WHERE id=$1 AND lease_token=$2`, event.ID, token)
	if err != nil {
		return true, fmt.Errorf("ack outbox: %w", err)
	}
	n, err := updated.RowsAffected()
	if err != nil {
		return true, err
	}
	if n != 1 {
		return true, errors.New("outbox lease lost after publish; duplicate delivery possible")
	}
	return true, nil
}
