// Package eventfeed decodes trusted Kafka/Outbox projection envelopes. Actual
// broker/DB clients supply transport-specific Receive and acknowledgement.
package eventfeed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

var ErrInvalidEnvelope = errors.New("invalid projection event envelope")

const maxEventBytes = 64 << 10

// Message is a single fetched Kafka record or claimed outbox row. Ack commits
// its offset/marks its row published. It MUST be called only after checkpoint.
type Message struct {
	Payload     []byte
	Acknowledge func(context.Context) error
}
type Receiver interface {
	Receive(context.Context) (Message, error)
}
type Feed struct{ receiver Receiver }

func NewKafka(receiver Receiver) (*Feed, error)  { return newFeed(receiver) }
func NewOutbox(receiver Receiver) (*Feed, error) { return newFeed(receiver) }
func newFeed(receiver Receiver) (*Feed, error) {
	if receiver == nil {
		return nil, ErrInvalidEnvelope
	}
	return &Feed{receiver: receiver}, nil
}

var _ projection.EventFeed = (*Feed)(nil)

type delivery struct {
	event projection.Event
	ack   func(context.Context) error
}

func (d delivery) Event() projection.Event       { return d.event }
func (d delivery) Ack(ctx context.Context) error { return d.ack(ctx) }

type envelope struct {
	EventID     string          `json:"event_id"`
	TenantID    string          `json:"tenant_id"`
	Kind        projection.Kind `json:"projection_kind"`
	AggregateID string          `json:"aggregate_id"`
	Revision    int64           `json:"projection_revision"`
}

func (f *Feed) Next(ctx context.Context) (projection.Delivery, error) {
	if f == nil || f.receiver == nil {
		return nil, ErrInvalidEnvelope
	}
	raw, err := f.receiver.Receive(ctx)
	if err != nil {
		return nil, err
	}
	if raw.Acknowledge == nil || len(raw.Payload) == 0 || len(raw.Payload) > maxEventBytes {
		return nil, ErrInvalidEnvelope
	}
	decoder := json.NewDecoder(bytes.NewReader(raw.Payload))
	decoder.DisallowUnknownFields()
	var item envelope
	if err := decoder.Decode(&item); err != nil {
		return nil, fmt.Errorf("%w: malformed JSON", ErrInvalidEnvelope)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, ErrInvalidEnvelope
	}
	tn, err := domain.NewTenantID(item.TenantID)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	event := projection.Event{EventID: item.EventID, Key: projection.Key{TenantID: tn, Kind: item.Kind, AggregateID: item.AggregateID}, Revision: item.Revision}
	if err := event.Validate(); err != nil {
		return nil, ErrInvalidEnvelope
	}
	return delivery{event: event, ack: raw.Acknowledge}, nil
}
