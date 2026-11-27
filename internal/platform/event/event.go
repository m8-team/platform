package event

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"buf.build/go/protovalidate"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
	"google.golang.org/protobuf/proto"
)

// routingKey uses length prefixes so component boundaries cannot collide.
// The hash is a routing hint, not an event identity or deduplication key.
func routingKey(e *eventv1.Event) string {
	var key []byte
	for _, part := range []string{e.GetTenantId(), e.GetAggregateType(), e.GetAggregateId()} {
		key = binary.AppendUvarint(key, uint64(len(part)))
		key = append(key, part...)
	}
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:])
}

func encodeBatch(ctx context.Context, cfg Config, events []*eventv1.Event) ([]topicwriter.Message, error) {
	if len(events) > cfg.MaxBatchEvents {
		return nil, ErrBatchTooLarge
	}
	messages := make([]topicwriter.Message, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	total := 0
	for i, e := range events {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if e == nil {
			return nil, fmt.Errorf("%w: item %d is nil", ErrInvalidEvent, i)
		}
		size := proto.Size(e)
		if size > cfg.MaxEventBytes {
			return nil, fmt.Errorf("%w: item %d", ErrEventTooLarge, i)
		}
		if size > cfg.MaxBatchBytes-total {
			return nil, ErrBatchTooLarge
		}
		if err := protovalidate.Validate(e); err != nil {
			// Validation diagnostics can contain payload values. Do not expose them.
			return nil, fmt.Errorf("%w: item %d envelope validation", ErrInvalidEvent, i)
		}
		if e.GetOccurredAt() == nil || e.GetOccurredAt().CheckValid() != nil || e.GetPayload() == nil {
			return nil, fmt.Errorf("%w: item %d timestamp or payload", ErrInvalidEvent, i)
		}
		payload, err := e.GetPayload().UnmarshalNew()
		if err != nil {
			return nil, fmt.Errorf("%w: item %d unknown or malformed payload", ErrInvalidEvent, i)
		}
		if err := protovalidate.Validate(payload); err != nil {
			return nil, fmt.Errorf("%w: item %d payload validation", ErrInvalidEvent, i)
		}
		if _, ok := seen[e.GetId()]; ok {
			return nil, fmt.Errorf("%w: item %d duplicate id", ErrInvalidEvent, i)
		}
		seen[e.GetId()] = struct{}{}
		data, err := proto.Marshal(e)
		if err != nil {
			return nil, fmt.Errorf("%w: item %d serialization", ErrInvalidEvent, i)
		}
		if len(data) > cfg.MaxEventBytes {
			return nil, ErrEventTooLarge
		}
		if len(data) > cfg.MaxBatchBytes-total {
			return nil, ErrBatchTooLarge
		}
		total += len(data)
		messages = append(messages, topicwriter.Message{
			Key: routingKey(e), Data: bytes.NewReader(data),
		})
	}
	return messages, nil
}
