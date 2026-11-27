package event

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"time"

	"buf.build/go/protovalidate"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Attributes contains caller-owned facts that must be stable across retries.
type Attributes struct {
	ID, Type, Source, TenantID string
	AggregateID, AggregateType string
	AggregateRevision          uint64
	CorrelationID, CausationID string
	OccurredAt                 time.Time
	Metadata                   map[string]string
}

// New validates and packs a typed payload. The event owner defines the
// correspondence between the business event type and the payload type.
func New(attrs Attributes, payload proto.Message) (*eventv1.Event, error) {
	if payload == nil || !payload.ProtoReflect().IsValid() || protovalidate.Validate(payload) != nil {
		return nil, fmt.Errorf("%w: payload validation", ErrInvalidEvent)
	}
	packed, err := anypb.New(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: payload serialization", ErrInvalidEvent)
	}
	e := &eventv1.Event{
		Id: attrs.ID, Type: attrs.Type, Source: attrs.Source, TenantId: attrs.TenantID,
		AggregateId: attrs.AggregateID, AggregateType: attrs.AggregateType,
		AggregateRevision: attrs.AggregateRevision, CorrelationId: attrs.CorrelationID,
		CausationId: attrs.CausationID, OccurredAt: timestamppb.New(attrs.OccurredAt),
		Payload: packed, Metadata: maps.Clone(attrs.Metadata),
	}
	if attrs.OccurredAt.IsZero() || !validEnvelope(e) {
		return nil, fmt.Errorf("%w: envelope validation", ErrInvalidEvent)
	}
	return e, nil
}

func validEnvelope(e *eventv1.Event) bool {
	return e.GetOccurredAt() != nil && e.GetOccurredAt().CheckValid() == nil &&
		e.GetPayload() != nil && protovalidate.Validate(e) == nil
}

// routingKey encodes component boundaries before the SDK hashes the bytes.
func routingKey(e *eventv1.Event) string {
	parts := [...]string{e.GetTenantId(), e.GetAggregateType(), e.GetAggregateId()}
	var size int
	for _, part := range parts {
		size += len(part) + binary.MaxVarintLen16
	}
	key := make([]byte, 0, size)
	for _, part := range parts {
		key = binary.AppendUvarint(key, uint64(len(part)))
		key = append(key, part...)
	}
	return string(key)
}

func encodeBatch(ctx context.Context, cfg Config, events []*eventv1.Event) ([]topicwriter.Message, int, error) {
	if len(events) > cfg.MaxBatchEvents {
		return nil, 0, ErrBatchTooLarge
	}
	messages := make([]topicwriter.Message, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	total := 0
	for i, e := range events {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		if e == nil {
			return nil, 0, fmt.Errorf("%w: item %d is nil", ErrInvalidEvent, i)
		}
		size := proto.Size(e)
		if size > cfg.MaxEventBytes {
			return nil, 0, fmt.Errorf("%w: item %d", ErrEventTooLarge, i)
		}
		if size > cfg.MaxBatchBytes-total {
			return nil, 0, ErrBatchTooLarge
		}
		if !validEnvelope(e) {
			return nil, 0, fmt.Errorf("%w: item %d envelope validation", ErrInvalidEvent, i)
		}
		payload, err := e.GetPayload().UnmarshalNew()
		if err != nil {
			return nil, 0, fmt.Errorf("%w: item %d unknown or malformed payload", ErrInvalidEvent, i)
		}
		if err := protovalidate.Validate(payload); err != nil {
			return nil, 0, fmt.Errorf("%w: item %d payload validation", ErrInvalidEvent, i)
		}
		if _, ok := seen[e.GetId()]; ok {
			return nil, 0, fmt.Errorf("%w: item %d duplicate id", ErrInvalidEvent, i)
		}
		seen[e.GetId()] = struct{}{}
		data, err := proto.Marshal(e)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: item %d serialization", ErrInvalidEvent, i)
		}
		if len(data) > cfg.MaxEventBytes {
			return nil, 0, ErrEventTooLarge
		}
		if len(data) > cfg.MaxBatchBytes-total {
			return nil, 0, ErrBatchTooLarge
		}
		total += len(data)
		messages = append(messages, topicwriter.Message{
			Key: routingKey(e), Data: bytes.NewReader(data),
		})
	}
	return messages, total, nil
}
